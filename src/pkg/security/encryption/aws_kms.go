package encryption

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// AWSKMS implements the Provider interface for AWS KMS encryption
type AWSKMS struct {
	client *kms.Client
	region string
	keyID  string
}

// AWSOpts defines AWS KMS-specific options
type AWSOpts struct {
	// Region is the AWS region where the KMS key is located
	Region string

	// KeyID is the ARN or ID of the default KMS key
	KeyID string

	// RoleARN is an optional IAM role to assume for KMS operations
	RoleARN string
	
	// Profile is an optional AWS profile to use for credentials
	Profile string
	
	// Endpoint is an optional custom endpoint for the KMS service
	Endpoint string
}

// NewAWSKMS creates a new AWS KMS encryption provider
func NewAWSKMS(ctx context.Context, opts AWSOpts) (*AWSKMS, error) {
	var cfgOpts []func(*config.LoadOptions) error
	
	// Configure region
	if opts.Region != "" {
		cfgOpts = append(cfgOpts, config.WithRegion(opts.Region))
	}
	
	// Configure profile if specified
	if opts.Profile != "" {
		cfgOpts = append(cfgOpts, config.WithSharedConfigProfile(opts.Profile))
	}
	
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Configure custom endpoint if provided
	var kmsOpts []func(*kms.Options)
	if opts.Endpoint != "" {
		kmsOpts = append(kmsOpts, func(o *kms.Options) {
			o.EndpointResolver = kms.EndpointResolverFromURL(opts.Endpoint)
		})
	}
	
	// Create the KMS client
	var kmsClient *kms.Client
	
	// If role ARN is provided, assume that role for KMS operations
	if opts.RoleARN != "" {
		// Create STS client for assuming role
		stsClient := sts.NewFromConfig(cfg)
		
		// Create the credentials provider for assuming the role
		provider := stscreds.NewAssumeRoleProvider(stsClient, opts.RoleARN)
		
		// Create new config with the assumed role credentials
		roleCfg := aws.Config{
			Credentials: aws.NewCredentialsCache(provider),
			Region:      cfg.Region,
		}
		
		// Create KMS client with the assumed role
		kmsClient = kms.NewFromConfig(roleCfg, kmsOpts...)
	} else {
		// Use the default credentials
		kmsClient = kms.NewFromConfig(cfg, kmsOpts...)
	}

	return &AWSKMS{
		client: kmsClient,
		region: opts.Region,
		keyID:  opts.KeyID,
	}, nil
}

// Name returns the provider name
func (a *AWSKMS) Name() string {
	return "aws-kms"
}

// Encrypt encrypts the plaintext using AWS KMS
func (a *AWSKMS) Encrypt(ctx context.Context, plaintext []byte, keyID string) ([]byte, error) {
	// Use default key if none specified
	if keyID == "" {
		keyID = a.keyID
	}
	
	if keyID == "" {
		return nil, fmt.Errorf("no KMS key ID specified for encryption")
	}

	input := &kms.EncryptInput{
		KeyId:     aws.String(keyID),
		Plaintext: plaintext,
	}

	result, err := a.client.Encrypt(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with AWS KMS: %w", err)
	}

	return result.CiphertextBlob, nil
}

// Decrypt decrypts the ciphertext using AWS KMS
func (a *AWSKMS) Decrypt(ctx context.Context, ciphertext []byte, keyID string) ([]byte, error) {
	input := &kms.DecryptInput{
		CiphertextBlob: ciphertext,
	}

	// Specify the key ID if provided to ensure we're using the expected key
	if keyID != "" {
		input.KeyId = aws.String(keyID)
	}

	result, err := a.client.Decrypt(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with AWS KMS: %w", err)
	}

	return result.Plaintext, nil
}

// GenerateDataKey generates a data key for envelope encryption
func (a *AWSKMS) GenerateDataKey(ctx context.Context, keyID string, keyLength int) (*DataKey, error) {
	// Use default key if none specified
	if keyID == "" {
		keyID = a.keyID
	}
	
	if keyID == "" {
		return nil, fmt.Errorf("no KMS key ID specified for data key generation")
	}

	// Default key length to AES-256 if not specified
	if keyLength <= 0 {
		keyLength = 32 // 256 bits
	}

	// Determine the key spec based on the requested length
	var keySpec types.DataKeySpec
	if keyLength == 16 {
		keySpec = types.DataKeySpecAes128
	} else {
		keySpec = types.DataKeySpecAes256
	}

	input := &kms.GenerateDataKeyInput{
		KeyId:   aws.String(keyID),
		KeySpec: keySpec,
	}

	result, err := a.client.GenerateDataKey(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to generate data key with AWS KMS: %w", err)
	}

	return &DataKey{
		Plaintext:  result.Plaintext,
		Ciphertext: result.CiphertextBlob,
	}, nil
}

// ReEncrypt re-encrypts already encrypted data with a new key
func (a *AWSKMS) ReEncrypt(ctx context.Context, ciphertext []byte, sourceKeyID, destinationKeyID string) ([]byte, error) {
	// Use default key if none specified for destination
	if destinationKeyID == "" {
		destinationKeyID = a.keyID
	}
	
	if destinationKeyID == "" {
		return nil, fmt.Errorf("no destination KMS key ID specified for re-encryption")
	}

	input := &kms.ReEncryptInput{
		CiphertextBlob:   ciphertext,
		DestinationKeyId: aws.String(destinationKeyID),
	}

	// Specify the source key ID if provided
	if sourceKeyID != "" {
		input.SourceKeyId = aws.String(sourceKeyID)
	}

	result, err := a.client.ReEncrypt(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to re-encrypt with AWS KMS: %w", err)
	}

	return result.CiphertextBlob, nil
}

// GetKeyInfo retrieves information about a KMS key
func (a *AWSKMS) GetKeyInfo(ctx context.Context, keyID string) (*KeyInfo, error) {
	// Use default key if none specified
	if keyID == "" {
		keyID = a.keyID
	}
	
	if keyID == "" {
		return nil, fmt.Errorf("no KMS key ID specified for key info")
	}

	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyID),
	}

	result, err := a.client.DescribeKey(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get key info from AWS KMS: %w", err)
	}

	// Create key info from response
	keyInfo := &KeyInfo{
		ID:         *result.KeyMetadata.KeyId,
		ARN:        *result.KeyMetadata.Arn,
		Algorithm:  string(result.KeyMetadata.KeySpec),
		State:      string(result.KeyMetadata.KeyState),
		Enabled:    result.KeyMetadata.Enabled,
		Provider:   "aws-kms",
		Region:     a.region,
		CreateTime: *result.KeyMetadata.CreationDate,
	}

	// Check if key is customer managed
	keyInfo.CustomerManaged = true // All keys accessed through KMS API are either AWS managed or customer managed
	if result.KeyMetadata.KeyManager == types.KeyManagerTypeAws {
		keyInfo.CustomerManaged = false // AWS managed key
	}

	return keyInfo, nil
}
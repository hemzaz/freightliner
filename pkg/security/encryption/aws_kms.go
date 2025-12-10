package encryption

import (
	"context"
	"strings"

	"freightliner/pkg/helper/errors"

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

	// Profile is an optional AWS profile to use for authentication
	Profile string
}

// NewAWSKMS creates a new AWS KMS provider
func NewAWSKMS(ctx context.Context, opts AWSOpts) (*AWSKMS, error) {
	if opts.Region == "" {
		return nil, errors.InvalidInputf("AWS region is required")
	}

	// Load AWS config
	var configOpts []func(*config.LoadOptions) error
	configOpts = append(configOpts, config.WithRegion(opts.Region))

	// Use profile if specified
	if opts.Profile != "" {
		configOpts = append(configOpts, config.WithSharedConfigProfile(opts.Profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load AWS config")
	}

	var kmsClient *kms.Client

	// If a role ARN is provided, assume the role and create a KMS client with the assumed role credentials
	if opts.RoleARN != "" {
		// Create an STS client to assume the role
		stsClient := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsClient, opts.RoleARN)

		// Create a new config with the role credentials
		roleCfg := aws.Config{
			Credentials: aws.NewCredentialsCache(provider),
			Region:      cfg.Region,
		}

		// Create a KMS client with the assumed role credentials
		kmsClient = kms.NewFromConfig(roleCfg)
	} else {
		// Use default credentials
		kmsClient = kms.NewFromConfig(cfg)
	}

	return &AWSKMS{
		client: kmsClient,
		region: opts.Region,
		keyID:  opts.KeyID,
	}, nil
}

// Encrypt encrypts plaintext using AWS KMS
func (a *AWSKMS) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	// Validate key ID
	if a.keyID == "" {
		return nil, errors.InvalidInputf("no KMS key ID specified for encryption")
	}

	// Validate plaintext
	if len(plaintext) == 0 {
		return nil, errors.InvalidInputf("plaintext cannot be empty")
	}

	// Create the encryption request
	input := &kms.EncryptInput{
		KeyId:     aws.String(a.keyID),
		Plaintext: plaintext,
	}

	// Call KMS to encrypt the data
	result, err := a.client.Encrypt(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt with AWS KMS")
	}

	return result.CiphertextBlob, nil
}

// Decrypt decrypts ciphertext using AWS KMS
func (a *AWSKMS) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	// Validate ciphertext
	if len(ciphertext) == 0 {
		return nil, errors.InvalidInputf("ciphertext cannot be empty")
	}

	// Create the decryption request
	input := &kms.DecryptInput{
		CiphertextBlob: ciphertext,
	}

	// Specify the key ID if available
	if a.keyID != "" {
		input.KeyId = aws.String(a.keyID)
	}

	// Call KMS to decrypt the data
	result, err := a.client.Decrypt(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt with AWS KMS")
	}

	return result.Plaintext, nil
}

// GenerateDataKey generates a data key that can be used to encrypt data locally
func (a *AWSKMS) GenerateDataKey(ctx context.Context, keySize int) ([]byte, []byte, error) {
	// Validate key ID
	if a.keyID == "" {
		return nil, nil, errors.InvalidInputf("no KMS key ID specified for data key generation")
	}

	// Validate key size
	if keySize <= 0 {
		return nil, nil, errors.InvalidInputf("key size must be positive, got %d", keySize)
	}

	// AWS KMS supports specific key sizes (128, 256, 512 bits)
	var keySpecValue types.DataKeySpec
	switch keySize {
	case 16: // 128 bits
		keySpecValue = types.DataKeySpecAes128
	case 32: // 256 bits
		keySpecValue = types.DataKeySpecAes256
	default:
		return nil, nil, errors.InvalidInputf("unsupported key size %d, must be 16 (128 bits) or 32 (256 bits)", keySize)
	}

	// Create the generate data key request
	input := &kms.GenerateDataKeyInput{
		KeyId:   aws.String(a.keyID),
		KeySpec: keySpecValue,
	}

	// Call KMS to generate the data key
	result, err := a.client.GenerateDataKey(ctx, input)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate data key with AWS KMS")
	}

	return result.Plaintext, result.CiphertextBlob, nil
}

// Name returns the provider name
func (a *AWSKMS) Name() string {
	return "aws-kms"
}

// GetKeyName returns the key ID or ARN
func (a *AWSKMS) GetKeyName() string {
	return a.keyID
}

// GetKeyInfo returns information about the KMS key
func (a *AWSKMS) GetKeyInfo() map[string]string {
	return map[string]string{
		"provider": "aws-kms",
		"region":   a.region,
		"keyID":    a.keyID,
	}
}

// ValidateKeyARN validates an AWS KMS key ARN
func ValidateKeyARN(arn string) error {
	// Basic validation for ARN format
	// arn:aws:kms:{region}:{account}:key/{key-id}
	if arn == "" {
		return errors.InvalidInputf("key ARN cannot be empty")
	}

	// Manual ARN parsing since aws.SplitARN is not available
	parts := strings.Split(arn, ":")
	if len(parts) != 6 {
		return errors.InvalidInputf("invalid ARN format: %s", arn)
	}

	// Check prefix
	if parts[0] != "arn" {
		return errors.InvalidInputf("invalid ARN prefix, expected 'arn', got '%s'", parts[0])
	}

	// Check partition
	if parts[1] != "aws" && parts[1] != "aws-cn" && parts[1] != "aws-us-gov" {
		return errors.InvalidInputf("invalid ARN partition, expected 'aws', 'aws-cn', or 'aws-us-gov', got '%s'", parts[1])
	}

	// Check service
	if parts[2] != "kms" {
		return errors.InvalidInputf("invalid ARN service, expected 'kms', got '%s'", parts[2])
	}

	// Check region
	region := parts[3]
	if region == "" {
		return errors.InvalidInputf("key ARN missing region")
	}

	// Check account
	accountID := parts[4]
	if accountID == "" {
		return errors.InvalidInputf("key ARN missing account ID")
	}

	// Check resource
	resource := parts[5]
	if !strings.HasPrefix(resource, "key/") {
		return errors.InvalidInputf("key ARN resource must start with 'key/', got '%s'", resource)
	}

	return nil
}

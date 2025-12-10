package cmd

// Config represents the main CLI configuration
type Config struct {
	// ECR specific configuration
	EcrRegion    string
	EcrAccountID string

	// GCR specific configuration
	GcrProject  string
	GcrLocation string
}

// RegistryCredentials represents the structure of registry credentials stored in the secrets manager
type RegistryCredentials struct {
	ECR struct {
		AccessKey    string `json:"access_key"`
		SecretKey    string `json:"secret_key"`
		AccountID    string `json:"account_id"`
		Region       string `json:"region"`
		SessionToken string `json:"session_token,omitempty"`
	} `json:"ecr"`

	GCR struct {
		Project      string `json:"project"`
		Location     string `json:"location"`
		Credentials  string `json:"credentials,omitempty"` // Base64 encoded JSON credentials
		TokenSource  string `json:"token_source,omitempty"`
		ClientEmail  string `json:"client_email,omitempty"`
		PrivateKeyID string `json:"private_key_id,omitempty"`
		PrivateKey   string `json:"private_key,omitempty"`
	} `json:"gcr"`
}

// EncryptionKeys represents the structure of encryption keys stored in the secrets manager
type EncryptionKeys struct {
	AWS struct {
		KMSKeyID string `json:"kms_key_id"`
		Region   string `json:"region"`
	} `json:"aws"`

	GCP struct {
		KMSKeyID    string `json:"kms_key_id"`
		Project     string `json:"project"`
		Location    string `json:"location"`
		KeyRing     string `json:"key_ring"`
		Key         string `json:"key"`
		Credentials string `json:"credentials,omitempty"` // Base64 encoded JSON credentials
	} `json:"gcp"`
}

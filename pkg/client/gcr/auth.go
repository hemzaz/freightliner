package gcr

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"freightliner/pkg/helper/errors"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Options configures a GCR authenticator
type Options struct {
	// CredentialsFile is the path to a JSON credentials file
	CredentialsFile string

	// TokenSource is an optional explicit OAuth2 token source
	TokenSource oauth2.TokenSource

	// Registry is the GCR registry to connect to (e.g. gcr.io, us.gcr.io)
	Registry string
}

// GCRAuthenticator implements the go-containerregistry authn.Authenticator interface for GCR
type GCRAuthenticator struct {
	ts oauth2.TokenSource
}

// NewAuthenticator creates a new authenticator for GCR with the given options
func NewAuthenticator(opts Options) (*GCRAuthenticator, error) {
	var ts oauth2.TokenSource
	var err error

	if opts.TokenSource != nil {
		ts = opts.TokenSource
	} else {
		// Set up the token source from the credentials file or default credentials
		ctx := context.Background()
		ts, err = DefaultTokenSource(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create token source")
		}
	}

	return &GCRAuthenticator{
		ts: ts,
	}, nil
}

// NewGCRAuthenticator creates a new authenticator for GCR
func NewGCRAuthenticator(ts oauth2.TokenSource) *GCRAuthenticator {
	return &GCRAuthenticator{
		ts: ts,
	}
}

// Authorization returns the authorization for GCR
func (a *GCRAuthenticator) Authorization() (*authn.AuthConfig, error) {
	// Get a token from the token source
	token, err := a.ts.Token()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get GCP authentication token")
	}

	return &authn.AuthConfig{
		Username: "oauth2accesstoken", // Special username for GCR
		Password: token.AccessToken,
	}, nil
}

// DefaultTokenSource returns the default token source for GCR
func DefaultTokenSource(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
	if len(scopes) == 0 {
		scopes = []string{
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/devstorage.read_write",
		}
	}

	ts, err := google.DefaultTokenSource(ctx, scopes...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create default token source")
	}

	return ts, nil
}

// RegistryAuthenticator creates an authenticator for a specific GCR registry
func (a *GCRAuthenticator) RegistryAuthenticator(registry string) (authn.Authenticator, error) {
	// Check if the registry is a GCR registry
	if !strings.HasSuffix(registry, "gcr.io") && !strings.Contains(registry, ".gcr.io") {
		return nil, errors.InvalidInputf("not a GCR registry: %s", registry)
	}

	// For GCR we can use the same authenticator for all registries
	// (unlike ECR which requires region-specific authentication)
	return a, nil
}

// GCRKeychain implements the go-containerregistry authn.Keychain interface for GCR
type GCRKeychain struct {
	ts oauth2.TokenSource
}

// Resolve returns an authenticator for the given resource
func (k *GCRKeychain) Resolve(resource authn.Resource) (authn.Authenticator, error) {
	registry := resource.RegistryStr()

	// Check if the registry is a GCR registry
	if !strings.HasSuffix(registry, "gcr.io") && !strings.Contains(registry, ".gcr.io") {
		return nil, errors.InvalidInputf("not a GCR registry: %s", registry)
	}

	return &GCRAuthenticator{ts: k.ts}, nil
}

// gcrTransport is an http.RoundTripper that adds GCP credentials to requests
type gcrTransport struct {
	base http.RoundTripper
	src  oauth2.TokenSource
}

// RoundTrip implements http.RoundTripper
func (t *gcrTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.src.Token()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token")
	}

	req2 := req.Clone(req.Context())
	token.SetAuthHeader(req2)
	return t.base.RoundTrip(req2)
}

// GetRepository returns a Repository object for a GCR repository
func (a *GCRAuthenticator) GetRepository(project, repoName string) (name.Repository, error) {
	if project == "" {
		return name.Repository{}, errors.InvalidInputf("GCP project cannot be empty")
	}
	if repoName == "" {
		return name.Repository{}, errors.InvalidInputf("repository name cannot be empty")
	}

	registryPath := fmt.Sprintf("gcr.io/%s/%s", project, repoName)
	repo, err := name.NewRepository(registryPath)
	if err != nil {
		return name.Repository{}, errors.Wrap(err, "failed to create GCR repository reference")
	}

	return repo, nil
}

// isGCRPath checks if a path is a valid GCR repository path
func isGCRPath(repository string) bool {
	// GCR paths should at least have a project/repo format
	if repository == "" {
		return false
	}

	parts := strings.Split(repository, "/")
	return len(parts) >= 2
}

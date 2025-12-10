package common

import (
	"context"
	"fmt"
	"net/http"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// CreateTransport creates a transport for registry operations with authentication
// This centralizes the common transport creation logic used in both ECR and GCR clients
func CreateTransport(registry name.Registry, auth authn.Authenticator, logger log.Logger) (http.RoundTripper, error) {
	scopes := []string{
		fmt.Sprintf("repository:%s:pull,push", registry.String()),
	}

	// Create transport with authentication and scopes
	rt, err := transport.NewWithContext(
		context.Background(),
		registry,
		auth,
		http.DefaultTransport,
		scopes,
	)
	if err != nil {
		logger.WithField("registry", registry.String()).Error("Failed to create transport", err)
		return nil, errors.Wrap(err, "failed to create transport")
	}

	return rt, nil
}

// CommonAuthOptions provides shared authentication options
type CommonAuthOptions struct {
	Registry string
	Region   string
	Project  string
	Account  string
}

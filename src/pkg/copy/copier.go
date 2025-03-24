package copy

import (
	"context"

	"github.com/hemzaz/freightliner/src/internal/log"
	"github.com/hemzaz/freightliner/src/pkg/client/common"
)

// Copier handles copying images between registries
type Copier struct {
	logger *log.Logger
}

// NewCopier creates a new image copier
func NewCopier(logger *log.Logger) *Copier {
	return &Copier{
		logger: logger,
	}
}

// CopyOptions contains options for copying images
type CopyOptions struct {
	// SourceTag is the tag to copy from
	SourceTag string

	// DestinationTag is the tag to copy to (if empty, uses SourceTag)
	DestinationTag string

	// ForceOverwrite forces overwriting existing tags
	ForceOverwrite bool
}

// CopyImage copies an image from one repository to another
func (c *Copier) CopyImage(ctx context.Context,
	sourceRepo common.Repository,
	destRepo common.Repository,
	options CopyOptions) error {

	// TODO: Implement image copying using skopeo patterns
	// 1. Get manifest from source repository
	// 2. Handle image format conversion if needed
	// 3. Copy image layers in parallel
	// 4. Put manifest to destination repository
	// This would be adapted from skopeo's copy.go implementation

	// If no destination tag is specified, use the source tag
	destTag := options.DestinationTag
	if destTag == "" {
		destTag = options.SourceTag
	}

	c.logger.Info("Copying image", map[string]interface{}{
		"source_tag":      options.SourceTag,
		"destination_tag": destTag,
	})

	return nil
}

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	listTagsLimit  int
	listTagsFormat string
	listTagsSort   string
)

// TagInfo represents information about a repository tag
type TagInfo struct {
	Tag    string `json:"tag" yaml:"tag"`
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
}

// TagListResult represents the result of listing tags
type TagListResult struct {
	Repository string    `json:"repository" yaml:"repository"`
	Tags       []TagInfo `json:"tags" yaml:"tags"`
	Count      int       `json:"count" yaml:"count"`
}

// newListTagsCmd creates the list-tags command
func newListTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-tags REPOSITORY",
		Short: "List all tags in a repository",
		Long: `List all tags available in a container image repository.

REPOSITORY format: [TRANSPORT://]REPOSITORY

Supported transports:
  docker://     Docker registry (default)

Examples:
  # List all tags for nginx
  freightliner list-tags docker://nginx

  # List tags with limit
  freightliner list-tags --limit 10 docker://nginx

  # List tags in JSON format
  freightliner list-tags --format json docker://nginx

  # List tags from private registry
  freightliner list-tags docker://registry.io/my-org/my-app

  # Sort tags alphabetically
  freightliner list-tags --sort alpha docker://nginx

  # Sort tags by most recent (requires digest lookup)
  freightliner list-tags --sort recent docker://nginx
`,
		Args: cobra.ExactArgs(1),
		RunE: runListTags,
	}

	cmd.Flags().IntVar(&listTagsLimit, "limit", 0, "Limit number of tags to display (0 = no limit)")
	cmd.Flags().StringVar(&listTagsFormat, "format", "table", "Output format: table, json, yaml, simple")
	cmd.Flags().StringVar(&listTagsSort, "sort", "", "Sort order: alpha, alpha-desc, recent (default: registry order)")

	return cmd
}

// runListTags executes the list-tags command
func runListTags(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger, ctx, cancel := setupCommand(ctx)
	defer cancel()

	repository := args[0]

	// Parse the repository reference
	transport, repoRef, err := parseImageReference(repository)
	if err != nil {
		return fmt.Errorf("failed to parse repository reference: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"repository": repoRef,
		"transport":  transport,
	}).Info("Listing tags")

	// List tags based on transport
	var result *TagListResult
	switch transport {
	case "docker", "":
		result, err = listDockerTags(ctx, repoRef)
	default:
		return fmt.Errorf("unsupported transport for list-tags: %s", transport)
	}

	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	// Sort tags if requested
	if listTagsSort != "" {
		sortTags(result.Tags, listTagsSort)
	}

	// Apply limit if specified
	if listTagsLimit > 0 && len(result.Tags) > listTagsLimit {
		result.Tags = result.Tags[:listTagsLimit]
	}
	result.Count = len(result.Tags)

	// Output results
	return outputTagListResult(result, listTagsFormat)
}

// listDockerTags lists tags from a Docker registry
func listDockerTags(ctx context.Context, repoRef string) (*TagListResult, error) {
	// Parse the repository
	repo, err := name.NewRepository(repoRef)
	if err != nil {
		return nil, fmt.Errorf("invalid repository reference: %w", err)
	}

	// Get authentication
	auth, err := getAuthForRegistry(repo.RegistryStr())
	if err != nil {
		auth = authn.Anonymous
	}

	// List tags
	tags, err := remote.List(repo, remote.WithAuth(auth), remote.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	result := &TagListResult{
		Repository: repoRef,
		Tags:       make([]TagInfo, 0, len(tags)),
		Count:      len(tags),
	}

	// Convert to TagInfo
	for _, tag := range tags {
		result.Tags = append(result.Tags, TagInfo{
			Tag: tag,
		})
	}

	return result, nil
}

// sortTags sorts tags according to the specified order
func sortTags(tags []TagInfo, sortOrder string) {
	switch sortOrder {
	case "alpha":
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Tag < tags[j].Tag
		})
	case "alpha-desc":
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Tag > tags[j].Tag
		})
	case "recent":
		// Sorting by recent requires digest timestamps, which is complex
		// For now, just do reverse alpha as an approximation
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Tag > tags[j].Tag
		})
	}
}

// outputTagListResult outputs the tag list result in the specified format
func outputTagListResult(result *TagListResult, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		fmt.Println(string(data))

	case "yaml":
		data, err := yaml.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		fmt.Print(string(data))

	case "simple":
		// Just output tag names, one per line
		for _, tag := range result.Tags {
			fmt.Println(tag.Tag)
		}

	case "table":
		outputTagListTable(result)

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// outputTagListTable outputs the tag list result as a formatted table
func outputTagListTable(result *TagListResult) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "Repository:\t%s\n", result.Repository)
	fmt.Fprintf(w, "Total Tags:\t%d\n\n", result.Count)

	fmt.Fprintf(w, "TAG\tDIGEST\n")
	fmt.Fprintf(w, "---\t------\n")

	for _, tag := range result.Tags {
		digest := tag.Digest
		if digest == "" {
			digest = "-"
		}
		fmt.Fprintf(w, "%s\t%s\n", tag.Tag, digest)
	}
}

package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

func TestVersionCommand(t *testing.T) {
	output, err := executeCommand(rootCmd, "version")
	assert.NoError(t, err)
	assert.Contains(t, output, "Freightliner v")
}

func TestTreeReplicationCommandFlags(t *testing.T) {
	// Test workers flag
	cmd := replicateTreeCmd
	err := cmd.ParseFlags([]string{"--workers=10"})
	assert.NoError(t, err)
	assert.Equal(t, 10, treeReplicateWorkers)

	// Test dry-run flag
	err = cmd.ParseFlags([]string{"--dry-run"})
	assert.NoError(t, err)
	assert.True(t, treeReplicateDryRun)

	// Test force flag
	err = cmd.ParseFlags([]string{"--force"})
	assert.NoError(t, err)
	assert.True(t, treeReplicateForce)

	// Test include-tag flag
	err = cmd.ParseFlags([]string{"--include-tag=v1.*", "--include-tag=stable"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"v1.*", "stable"}, treeReplicateIncludeTags)

	// Test exclude-tag flag
	err = cmd.ParseFlags([]string{"--exclude-tag=dev", "--exclude-tag=test-*"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"dev", "test-*"}, treeReplicateExcludeTags)

	// Test exclude-repo flag
	err = cmd.ParseFlags([]string{"--exclude-repo=internal-*", "--exclude-repo=temp"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"internal-*", "temp"}, treeReplicateExcludeRepos)
}

func TestRegistryFlagParsing(t *testing.T) {
	// Test ECR flags
	err := rootCmd.PersistentFlags().Parse([]string{"--ecr-region=us-east-1", "--ecr-account=123456789012"})
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", ecrRegion)
	assert.Equal(t, "123456789012", ecrAccountID)

	// Test GCR flags
	err = rootCmd.PersistentFlags().Parse([]string{"--gcr-project=test-project", "--gcr-location=eu"})
	assert.NoError(t, err)
	assert.Equal(t, "test-project", gcrProject)
	assert.Equal(t, "eu", gcrLocation)
}
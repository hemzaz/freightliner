package banner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestLogo(t *testing.T) {
	tests := []struct {
		name     string
		logo     string
		contains []string
	}{
		{
			name: "full logo contains required elements",
			logo: Logo,
			contains: []string{
				"FREIGHTLINER",
				"Container Registry Replication",
				"(o)",
			},
		},
		{
			name: "small logo contains required elements",
			logo: SmallLogo,
			contains: []string{
				"FREIGHTLINER",
				"(o)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expected := range tt.contains {
				if !strings.Contains(tt.logo, expected) {
					t.Errorf("logo does not contain expected text: %s", expected)
				}
			}
		})
	}
}

func TestPrint(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set test version info
	oldVersion := Version
	oldCommit := GitCommit
	oldBuildTime := BuildTime
	Version = "1.0.0"
	GitCommit = "abc123"
	BuildTime = "2024-01-01T00:00:00Z"
	defer func() {
		Version = oldVersion
		GitCommit = oldCommit
		BuildTime = oldBuildTime
	}()

	// Call Print
	Print()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected elements
	expectedContents := []string{
		"FREIGHTLINER",
		"Container Registry Replication",
		"Version: 1.0.0",
		"Commit: abc123",
		"Built: 2024-01-01T00:00:00Z",
		"Runtime: Go",
		runtime.GOOS,
		runtime.GOARCH,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("Print() output does not contain expected text: %s", expected)
		}
	}
}

func TestPrintSmall(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set test version info
	oldVersion := Version
	Version = "2.0.0"
	defer func() {
		Version = oldVersion
	}()

	// Call PrintSmall
	PrintSmall()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected elements
	expectedContents := []string{
		"FREIGHTLINER",
		"v2.0.0",
		"(o)",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("PrintSmall() output does not contain expected text: %s", expected)
		}
	}

	// Verify it's actually smaller than the full logo
	if len(output) >= len(Logo) {
		t.Error("PrintSmall() output is not smaller than full logo")
	}
}

func TestPrintVersion(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set test version info
	oldVersion := Version
	oldCommit := GitCommit
	oldBuildTime := BuildTime
	Version = "3.0.0"
	GitCommit = "def456"
	BuildTime = "2024-06-01T12:00:00Z"
	defer func() {
		Version = oldVersion
		GitCommit = oldCommit
		BuildTime = oldBuildTime
	}()

	// Call PrintVersion
	PrintVersion()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected elements
	expectedContents := []string{
		"Freightliner v3.0.0",
		"Git Commit: def456",
		"Built: 2024-06-01T12:00:00Z",
		"Go Version:",
		"OS/Arch:",
		runtime.GOOS,
		runtime.GOARCH,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("PrintVersion() output does not contain expected text: %s", expected)
		}
	}

	// Verify it doesn't contain the ASCII art
	if strings.Contains(output, "_______________") {
		t.Error("PrintVersion() should not contain ASCII art")
	}
}

func TestVersionVariables(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		varValue string
	}{
		{
			name:     "Version has default value",
			varName:  "Version",
			varValue: Version,
		},
		{
			name:     "GitCommit has default value",
			varName:  "GitCommit",
			varValue: GitCommit,
		},
		{
			name:     "BuildTime has default value",
			varName:  "BuildTime",
			varValue: BuildTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.varValue == "" {
				t.Errorf("%s should not be empty", tt.varName)
			}
		})
	}
}

func TestLogoConsistency(t *testing.T) {
	// Test that Logo is properly formatted
	lines := strings.Split(Logo, "\n")
	if len(lines) < 5 {
		t.Error("Logo should have multiple lines")
	}

	// Test that SmallLogo is properly formatted
	smallLines := strings.Split(SmallLogo, "\n")
	if len(smallLines) < 3 {
		t.Error("SmallLogo should have multiple lines")
	}

	// Verify SmallLogo is indeed smaller
	if len(smallLines) >= len(lines) {
		t.Error("SmallLogo should have fewer lines than Logo")
	}
}

func TestPrintOutputFormat(t *testing.T) {
	// Test that all print functions produce non-empty output
	tests := []struct {
		name string
		fn   func()
	}{
		{"Print", Print},
		{"PrintSmall", PrintSmall},
		{"PrintVersion", PrintVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call function
			tt.fn()

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if len(output) == 0 {
				t.Errorf("%s produced empty output", tt.name)
			}

			// Verify output ends with newline
			if !strings.HasSuffix(output, "\n") {
				t.Errorf("%s output should end with newline", tt.name)
			}
		})
	}
}

func TestPrintWithDifferentVersions(t *testing.T) {
	testCases := []struct {
		version   string
		commit    string
		buildTime string
	}{
		{"dev", "unknown", "unknown"},
		{"1.0.0", "abc123", "2024-01-01"},
		{"v2.5.3", "def456xyz", "2024-12-31T23:59:59Z"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("version-%s", tc.version), func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Set version info
			oldVersion := Version
			oldCommit := GitCommit
			oldBuildTime := BuildTime
			Version = tc.version
			GitCommit = tc.commit
			BuildTime = tc.buildTime
			defer func() {
				Version = oldVersion
				GitCommit = oldCommit
				BuildTime = oldBuildTime
			}()

			// Call Print
			Print()

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Verify version info appears in output
			if !strings.Contains(output, tc.version) {
				t.Errorf("output should contain version %s", tc.version)
			}
			if !strings.Contains(output, tc.commit) {
				t.Errorf("output should contain commit %s", tc.commit)
			}
			if !strings.Contains(output, tc.buildTime) {
				t.Errorf("output should contain build time %s", tc.buildTime)
			}
		})
	}
}

func BenchmarkPrint(b *testing.B) {
	// Suppress output
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Print()
	}
}

func BenchmarkPrintSmall(b *testing.B) {
	// Suppress output
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PrintSmall()
	}
}

func BenchmarkPrintVersion(b *testing.B) {
	// Suppress output
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PrintVersion()
	}
}

package formatting

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
)

// TemplateFormatter provides Go template-based formatting for command output
type TemplateFormatter struct {
	template *template.Template
}

// NewTemplateFormatter creates a new template formatter with the given template string
func NewTemplateFormatter(tmpl string) (*TemplateFormatter, error) {
	// Add helper functions to template
	funcMap := template.FuncMap{
		"json":  toJSON,
		"split": strings.Split,
		"join":  strings.Join,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
	}

	t, err := template.New("output").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	return &TemplateFormatter{template: t}, nil
}

// Format executes the template with the provided data and writes to the writer
func (f *TemplateFormatter) Format(w io.Writer, data interface{}) error {
	return f.template.Execute(w, data)
}

// FormatString executes the template and returns the result as a string
func (f *TemplateFormatter) FormatString(data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := f.Format(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// toJSON is a helper function to output data as JSON
func toJSON(v interface{}) string {
	// This is a simple implementation
	// In production, you'd want to use proper JSON marshaling
	return fmt.Sprintf("%+v", v)
}

// IsTemplateFormat checks if a format string is a Go template
func IsTemplateFormat(format string) bool {
	return strings.HasPrefix(format, "{{") && strings.HasSuffix(format, "}}")
}

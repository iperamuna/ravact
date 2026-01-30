package stubs

import (
	"embed"
	"strings"
)

//go:embed templates/*.stub
var templatesFS embed.FS

// LoadAndReplace loads a stub from the embedded filesystem and replaces placeholders
func LoadAndReplace(name string, replacements map[string]string) (string, error) {
	content, err := templatesFS.ReadFile("templates/" + name + ".stub")
	if err != nil {
		return "", err
	}

	result := string(content)
	for key, value := range replacements {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}

	return result, nil
}

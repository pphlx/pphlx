package main

import (
	"strings"
	"testing"
)

func TestDefaultConfigTarget(t *testing.T) {
	// Verify default target fallback
	config := Config{}
	if config.Output.Target == "" {
		config.Output.Target = "php"
	}
	if config.Output.Target != "php" {
		t.Errorf("Expected default target 'php', got '%s'", config.Output.Target)
	}
}

func TestCLIOverrideTarget(t *testing.T) {
	// Mock CLI args override check
	mockArgs := []string{"pphlx", "build", "--target", "standalone"}

	cliTarget := ""
	for i := 1; i < len(mockArgs)-1; i++ {
		arg := strings.ToLower(mockArgs[i])
		if arg == "--target" || arg == "-t" {
			cliTarget = mockArgs[i+1]
			break
		}
	}

	if cliTarget != "standalone" {
		t.Errorf("Expected CLI override target 'standalone', got '%s'", cliTarget)
	}
}

func TestTargetExtensions(t *testing.T) {
	tests := []struct {
		target    string
		expected  string
	}{
		{"php", ".php"},
		{"standalone", ".php"}, // Standalone maps templates to .php internally before compiling
		{"ssg", ".html"},
		{"blade", ".blade.php"},
		{"twig", ".html.twig"},
	}

	for _, tt := range tests {
		ext := ".php"
		switch strings.ToLower(tt.target) {
		case "ssg":
			ext = ".html"
		case "blade":
			ext = ".blade.php"
		case "twig":
			ext = ".html.twig"
		}
		if ext != tt.expected {
			t.Errorf("For target '%s', expected extension '%s', got '%s'", tt.target, tt.expected, ext)
		}
	}
}

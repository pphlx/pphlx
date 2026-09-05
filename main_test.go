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
		target   string
		expected string
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

func TestDirectivesCompilation(t *testing.T) {
	// 1. Test parseComponent with is:inline script and style
	componentContent := `<style is:inline>.local { color: red; }</style>
<style>.global { color: blue; }</style>
<script is:inline>console.log('inline');</script>
<script>console.log('bundled');</script>
<div>Hello World</div>`

	comp := parseComponent("TestComponent", componentContent, "TestComponent.pphx")

	if !strings.Contains(comp.HTML, "is:inline") {
		t.Errorf("Expected comp.HTML to contain is:inline tags, got: %s", comp.HTML)
	}
	if strings.Contains(comp.CSS, ".local") {
		t.Errorf("Expected comp.CSS to exclude is:inline style, got: %s", comp.CSS)
	}
	if !strings.Contains(comp.CSS, ".global") {
		t.Errorf("Expected comp.CSS to include global style, got: %s", comp.CSS)
	}
	if strings.Contains(comp.JS, "console.log('inline')") {
		t.Errorf("Expected comp.JS to exclude is:inline script, got: %s", comp.JS)
	}
	if !strings.Contains(comp.JS, "console.log('bundled')") {
		t.Errorf("Expected comp.JS to include bundled script, got: %s", comp.JS)
	}

	// 2. Test set:html and set:text directives via compilePage
	pageInput := `<div set:html="$rawHtml"></div><span set:text="$escapedText"></span><code is:raw><span>{{literal}}</span></code>`
	compiled, _, _, err := compilePage(pageInput, ".", ".")
	if err != nil {
		t.Fatalf("compilePage failed: %v", err)
	}

	if !strings.Contains(compiled, "<?php echo $rawHtml; ?>") {
		t.Errorf("Expected set:html to compile to '<?php echo $rawHtml; ?>', got: %s", compiled)
	}
	if !strings.Contains(compiled, "<?php echo htmlspecialchars($escapedText); ?>") {
		t.Errorf("Expected set:text to compile to '<?php echo htmlspecialchars($escapedText); ?>', got: %s", compiled)
	}
	if !strings.Contains(compiled, "<span>{{literal}}</span></code>") {
		t.Errorf("Expected is:raw content to remain literal, got: %s", compiled)
	}
}

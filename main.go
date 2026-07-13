package main

import (
	"archive/zip"
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/fsnotify/fsnotify"
)

// Config holds project configuration
type Config struct {
	SrcDir  string `json:"srcDir"`
	OutDir  string `json:"outDir"`
	CssOut  string `json:"cssOut"`
	JsOut   string `json:"jsOut"`
}

// Component represents a parsed .pphx or JS/React component
type Component struct {
	Name          string
	Path          string
	HTML          string
	CSS           string
	JS            string
	IsLayout      bool
	IsJsComponent bool
}

//go:embed mcp/*
var mcpFS embed.FS

var (
	importRegex      = regexp.MustCompile(`(?m)^@import\s+(\w+)\s+from\s+'([^']+)'\s*\r?\n?`)
	styleRegex       = regexp.MustCompile(`(?s)<style>(.*?)</style>`)
	scriptRegex      = regexp.MustCompile(`(?s)<script>(.*?)</script>`)
	templateRegex    = regexp.MustCompile(`(?s)<template[^>]*>(.*?)</template>`)
	attrRegex        = regexp.MustCompile(`(\w+[\w:-]*)(?:="([^"]*)")?`)
	phpShortTagRegex = regexp.MustCompile(`\{\{\s*(\w+)(?:\s*\?\?\s*'(.*?)')?\s*\}\}`)
	viteComponents   = make(map[string]string)
)

func main() {
	if len(os.Args) > 1 {
		cmd := strings.ToLower(os.Args[1])
		if cmd == "add" {
			if len(os.Args) < 3 {
				fmt.Println("Error: pphlx add requires a repository URL.")
				fmt.Println("Usage: pphlx add github.com/username/repo[@version]")
				os.Exit(1)
			}
			repo := os.Args[2]
			
			// Find pphlx.json directory
			projectDir := "."
			if _, err := os.Stat("pphlx.json"); os.IsNotExist(err) {
				if _, err := os.Stat("test_project/pphlx.json"); err == nil {
					projectDir = "test_project"
				}
			}
			
			err := addDependency(repo, projectDir)
			if err != nil {
				fmt.Printf("Error adding dependency: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Dependency added successfully!")
			os.Exit(0)
		}
		if cmd == "mcp" {
			if len(os.Args) > 2 && strings.ToLower(os.Args[2]) == "install" {
				err := installMCPServer()
				if err != nil {
					fmt.Printf("Error installing MCP server: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("PPHLX MCP server registered successfully!")
				os.Exit(0)
			}
			runMCPServer()
			os.Exit(0)
		}
	}

	fmt.Println("PPHLX Compiler Starting...")

	// Default command is "build", check if user passed "dev" or "watch"
	mode := "build"
	if len(os.Args) > 1 {
		cmd := strings.ToLower(os.Args[1])
		if cmd == "dev" || cmd == "watch" {
			mode = "dev"
		}
	}

	// 1. Read config (looking for pphlx.config.mjs first, then fallback to json)
	configPath := "pphlx.config.mjs"
	isMjs := true
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "pphlx.config.json"
		isMjs = false
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			configPath = "pphlx.json"
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				// Fallback to test_project if run from compiler root
				configPath = filepath.Join("test_project", "pphlx.config.mjs")
				isMjs = true
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					configPath = filepath.Join("test_project", "pphlx.config.json")
					isMjs = false
				}
			}
		}
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	var config Config
	if isMjs {
		// Parse .mjs configuration using regex mapping
		configStr := string(configData)
		config.SrcDir = parseMjsField(configStr, "srcDir")
		config.OutDir = parseMjsField(configStr, "outDir")
		config.CssOut = parseMjsField(configStr, "cssOut")
		config.JsOut = parseMjsField(configStr, "jsOut")
	} else {
		if err := json.Unmarshal(configData, &config); err != nil {
			fmt.Printf("Error parsing JSON config: %v\n", err)
			os.Exit(1)
		}
	}

	projectDir := filepath.Dir(configPath)

	// Run initial compilation
	compileAll(config, projectDir)

	if mode == "dev" {
		startWatcher(config, projectDir)
	}
}

// compileAll executes the main compilation loop for all templates
func compileAll(config Config, projectDir string) {
	fmt.Printf("[%s] Rebuilding templates...\n", time.Now().Format("15:04:05"))
	
	srcDir := filepath.Join(projectDir, config.SrcDir)
	outDir := filepath.Join(projectDir, config.OutDir)
	cssOut := filepath.Join(projectDir, config.CssOut)
	jsOut := filepath.Join(projectDir, config.JsOut)

	// Reset Vite components map for this compile run
	viteComponents = make(map[string]string)

	// Bundles
	var globalCSS strings.Builder
	var globalJS strings.Builder

	// Prepend require and process shims at the very beginning of the global JS bundle
	globalJS.WriteString(`
// Require shim for browser CDN compatibility
window.require = window.require || function(mod) {
  if (mod === 'react') return window.React;
  if (mod === 'react-dom') return window.ReactDOM;
  return undefined;
};

// Process environment shim for Vue/Svelte compatibility
window.process = window.process || { env: { NODE_ENV: 'production' } };
`)

	compiledScripts := make(map[string]bool)
	compiledStyles := make(map[string]bool)

	// Walk entire src directory recursively
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(outDir, relPath)

		if info.IsDir() {
			// Skip components and layouts directories from being copied directly
			dirName := info.Name()
			if dirName == "components" || dirName == "layouts" {
				return filepath.SkipDir
			}
			// Recreate the directory structure in output
			return os.MkdirAll(outPath, 0755)
		}

		// Process files
		if strings.HasSuffix(info.Name(), ".pphx") {
			// Compile .pphx to .php
			phpOutPath := strings.TrimSuffix(outPath, ".pphx") + ".php"
			
			pageContent, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			compiledPage, css, js, err := compilePage(string(pageContent), filepath.Dir(path), srcDir)
			if err != nil {
				return fmt.Errorf("error compiling %s: %v", info.Name(), err)
			}

			// Append styling/scripts to global bundles if unique
			for _, style := range css {
				cleanStyle := strings.TrimSpace(style)
				if cleanStyle != "" && !compiledStyles[cleanStyle] {
					compiledStyles[cleanStyle] = true
					globalCSS.WriteString(cleanStyle + "\n")
				}
			}
			for _, script := range js {
				cleanScript := strings.TrimSpace(script)
				if cleanScript != "" && !compiledScripts[cleanScript] {
					compiledScripts[cleanScript] = true
					globalJS.WriteString(cleanScript + "\n")
				}
			}

			// Inject css and js links (calculating correct relative paths)
			cssRelPath := getRelativePath(phpOutPath, cssOut)
			jsRelPath := getRelativePath(phpOutPath, jsOut)
			
			cssTag := fmt.Sprintf(`<link rel="stylesheet" href="%s">`, cssRelPath)
			jsTag := fmt.Sprintf(`<script src="%s"></script>`, jsRelPath)

			compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_CSS}}", cssTag)
			compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_JS}}", jsTag)

			os.MkdirAll(filepath.Dir(phpOutPath), 0755)
			err = ioutil.WriteFile(phpOutPath, []byte(strings.TrimLeft(compiledPage, " \t\r\n")), 0644)
			if err != nil {
				return err
			}
		} else {
			// Smart copy non-.pphx files
			err = copyFileIfNewer(path, outPath)
			if err != nil {
				return fmt.Errorf("error copying file %s to %s: %v", path, outPath, err)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Build Error: %v\n", err)
		return
	}

	// Write global CSS & JS
	if globalCSS.Len() > 0 {
		os.MkdirAll(filepath.Dir(cssOut), 0755)
		ioutil.WriteFile(cssOut, []byte(globalCSS.String()), 0644)
	}
	
	// Inject lightweight PPHLX Islands hydration runtime
	runtimeScript := `
// PPHLX Islands Runtime
document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll(".pphlx-island").forEach(island => {
    const compName = island.getAttribute("data-component");
    const islandId = island.id;
    const props = window.pphlxProps ? window.pphlxProps[islandId] : {};
    
    if (window[compName]) {
      const ComponentModule = window[compName];
      const Component = ComponentModule.default || ComponentModule;
      
      if (window.ReactDOM && window.ReactDOM.createRoot) {
        // React 18+ Mount
        const root = window.ReactDOM.createRoot(island);
        root.render(window.React.createElement(Component, props));
      } else if (window.Vue && window.Vue.createApp) {
        // Vue 3 Mount
        window.Vue.createApp(Component, props).mount(island);
      } else if (typeof Component === "function" && Component.prototype && Component.prototype.$destroy) {
        // Svelte Mount
        new Component({ target: island, props: props });
      } else if (window.SolidJS && window.SolidJS.render) {
        // SolidJS Mount
        window.SolidJS.render(() => Component(props), island);
      } else if (window.preact && window.preact.render) {
        // Preact Mount
        window.preact.render(window.preact.h(Component, props), island);
      } else if (Component.render) {
        Component.render(island, props);
      } else {
        console.warn("No runtime renderer found to mount component " + compName);
      }
    } else {
      console.error("Component " + compName + " not found in window scope.");
    }
  });
});
`
	globalJS.WriteString(runtimeScript + "\n")

	if globalJS.Len() > 0 {
		os.MkdirAll(filepath.Dir(jsOut), 0755)
		ioutil.WriteFile(jsOut, []byte(globalJS.String()), 0644)
	}

	// Trigger local Vite compilation if Svelte/Vue/Angular components are found
	if len(viteComponents) > 0 {
		err = runViteBuild(config, projectDir)
		if err != nil {
			fmt.Printf("Vite Build Error: %v\n", err)
			return
		}
	}

	fmt.Printf("[%s] Build complete successfully!\n", time.Now().Format("15:04:05"))
}

// copyFileIfNewer copies a file if the destination is missing or older than source
func copyFileIfNewer(src, dst string) error {
	srcStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	dstStat, err := os.Stat(dst)
	if err == nil {
		// File exists, check modification times
		if !srcStat.ModTime().After(dstStat.ModTime()) {
			return nil // Source is not newer, skip copy
		}
	}

	// Copy process
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	os.MkdirAll(filepath.Dir(dst), 0755)
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcStat.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

// startWatcher initializes the fsnotify file-watcher loop
func startWatcher(config Config, projectDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	srcDir := filepath.Join(projectDir, config.SrcDir)
	fmt.Printf("Watching source directory recursively: %s\n", srcDir)

	// Watch recursively by adding subfolders
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip component/layouts directories from watcher triggers
			dirName := info.Name()
			if dirName == "components" || dirName == "layouts" {
				return filepath.SkipDir
			}
			err = watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error setting up folder watcher: %v\n", err)
		return
	}

	// Channel to handle events
	done := make(chan bool)
	go func() {
		// Coalesce rapid multiple updates (debouncing)
		var timer *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Watch for modifications or creations
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					// If a new folder is created, watch it recursively
					if event.Has(fsnotify.Create) {
						fi, err := os.Stat(event.Name)
						if err == nil && fi.IsDir() {
							watcher.Add(event.Name)
							fmt.Printf("Watching new folder: %s\n", event.Name)
						}
					}

					// Rebuild on file changes (excluding temporary/IDE files)
					baseName := filepath.Base(event.Name)
					if !strings.HasPrefix(baseName, ".") && !strings.HasSuffix(baseName, "~") {
						fmt.Printf("File changed: %s\n", baseName)
						if timer != nil {
							timer.Stop()
						}
						timer = time.AfterFunc(100*time.Millisecond, func() {
							compileAll(config, projectDir)
						})
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Watcher error: %v\n", err)
			}
		}
	}()

	<-done
}

// compilePage parses imports, layouts, and expands components
func compilePage(content string, currentDir string, srcDir string) (string, []string, []string, error) {
	// Pre-process custom bracket syntax to standard PHP tags first
	content = parsePphlxBrackets(content)

	var cssBundles []string
	var jsBundles []string

	// 1. Extract imports
	matches := importRegex.FindAllStringSubmatch(content, -1)
	imports := make(map[string]Component)

	for _, match := range matches {
		compName := match[1]
		relPath := match[2]

		// Resolve component file path
		var compPath string
		var isJsComponent bool

		ext := filepath.Ext(relPath)
		isPackage := strings.HasPrefix(relPath, "github.com/") || strings.HasPrefix(relPath, "gitlab.com/")

		if ext == "" {
			if (isPackage) {
				compPath = filepath.Clean(filepath.Join(srcDir, "../.pphlx/packages", relPath+".pphx"))
			} else {
				compPath = filepath.Clean(filepath.Join(currentDir, relPath+".pphx"))
				if _, err := os.Stat(compPath); os.IsNotExist(err) {
					compPath = filepath.Clean(filepath.Join(srcDir, relPath+".pphx"))
				}
			}
			
			if _, err := os.Stat(compPath); os.IsNotExist(err) {
				// Fallback to JS/Vue/Svelte/TS/Solid extensions
				for _, jsExt := range []string{".jsx", ".tsx", ".js", ".vue", ".svelte", ".solid.jsx", ".solid.tsx", ".ts"} {
					var testPath string
					if (isPackage) {
						testPath = filepath.Clean(filepath.Join(srcDir, "../.pphlx/packages", relPath+jsExt))
					} else {
						testPath = filepath.Clean(filepath.Join(currentDir, relPath+jsExt))
					}
					
					if _, err := os.Stat(testPath); err == nil {
						compPath = testPath
						isJsComponent = true
						break
					}
					if !isPackage {
						testPath = filepath.Clean(filepath.Join(srcDir, relPath+jsExt))
						if _, err := os.Stat(testPath); err == nil {
							compPath = testPath
							isJsComponent = true
							break
						}
					}
				}
			}
		} else {
			if isPackage {
				compPath = filepath.Clean(filepath.Join(srcDir, "../.pphlx/packages", relPath))
			} else {
				compPath = filepath.Clean(filepath.Join(currentDir, relPath))
				if _, err := os.Stat(compPath); os.IsNotExist(err) {
					compPath = filepath.Clean(filepath.Join(srcDir, relPath))
				}
			}
			if ext == ".jsx" || ext == ".tsx" || ext == ".js" || ext == ".vue" || ext == ".svelte" || ext == ".ts" || strings.HasSuffix(relPath, ".solid.jsx") || strings.HasSuffix(relPath, ".solid.tsx") {
				isJsComponent = true
			}
		}

		var compObj Component
		if isJsComponent {
			// Svelte, Vue, SolidJS, TS, and TSX files containing Angular are routed to Vite
			isVite := ext == ".vue" || ext == ".svelte" || strings.HasSuffix(compPath, ".ts") || strings.HasSuffix(compPath, ".tsx") || strings.HasSuffix(compPath, ".vue") || strings.HasSuffix(compPath, ".svelte") || strings.HasSuffix(compPath, ".solid.jsx") || strings.HasSuffix(compPath, ".solid.tsx")
			
			if isVite {
				viteComponents[compName] = compPath
				compObj = Component{
					Name:          compName,
					Path:          compPath,
					IsJsComponent: true,
				}
			} else {
				// Compile JS/JSX/TSX component with native esbuild
				jsCode, err := compileJSComponent(compName, compPath)
				if err != nil {
					return "", nil, nil, fmt.Errorf("failed to compile JS component %s: %v", compName, err)
				}
				compObj = Component{
					Name:          compName,
					Path:          compPath,
					JS:            jsCode,
					IsJsComponent: true,
				}
			}
		} else {
			compContent, err := ioutil.ReadFile(compPath)
			if err != nil {
				return "", nil, nil, fmt.Errorf("failed to read imported component %s: %v", compName, err)
			}
			processedCompContent := parsePphlxBrackets(string(compContent))
			compObj = parseComponent(compName, processedCompContent, compPath)
		}

		imports[compName] = compObj

		if compObj.CSS != "" {
			cssBundles = append(cssBundles, compObj.CSS)
		}
		if compObj.JS != "" {
			jsBundles = append(jsBundles, compObj.JS)
		}
	}

	// Clean import lines from page content
	cleanedContent := importRegex.ReplaceAllString(content, "")

	// 2. Recursively expand components
	expandedContent := cleanedContent
	for name, comp := range imports {
		expandedContent = expandComponent(expandedContent, name, comp)
	}

	return expandedContent, cssBundles, jsBundles, nil
}

// compileJSComponent compiles JS/React/JSX components using native esbuild
func compileJSComponent(compName string, compPath string) (string, error) {
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{compPath},
		Bundle:      true,
		Write:       false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Format:      api.FormatIIFE,
		GlobalName:  compName,
		External:    []string{"react", "react-dom", "preact", "preact/hooks", "solid-js"},
		Loader: map[string]api.Loader{
			".js":  api.LoaderJS,
			".jsx": api.LoaderJSX,
			".tsx": api.LoaderTSX,
			".ts":  api.LoaderTS,
		},
		JSX: api.JSXTransform,
	})

	if len(result.Errors) > 0 {
		var errs []string
		for _, err := range result.Errors {
			errs = append(errs, fmt.Sprintf("%s:%d: %s", err.Location.File, err.Location.Line, err.Text))
		}
		return "", fmt.Errorf("esbuild errors: %s", strings.Join(errs, "; "))
	}

	if len(result.OutputFiles) > 0 {
		return string(result.OutputFiles[0].Contents), nil
	}

	return "", fmt.Errorf("no output files generated by esbuild")
}

// parseComponent extracts templates, style, and scripts from component files
func parseComponent(name string, content string, path string) Component {
	var html, css, js string
	isLayout := true

	// Check if it's a layout (layouts usually don't have <template> blocks)
	tmplMatch := templateRegex.FindStringSubmatch(content)
	if len(tmplMatch) > 1 {
		html = tmplMatch[1]
		isLayout = false
	} else {
		// It's a layout, html is everything outside style/script
		html = styleRegex.ReplaceAllString(content, "")
		html = scriptRegex.ReplaceAllString(html, "")
	}

	styleMatch := styleRegex.FindStringSubmatch(content)
	if len(styleMatch) > 1 {
		css = styleMatch[1]
	}

	scriptMatch := scriptRegex.FindStringSubmatch(content)
	if len(scriptMatch) > 1 {
		js = scriptMatch[1]
	}

	return Component{
		Name:     name,
		Path:     path,
		HTML:     html,
		CSS:      css,
		JS:       js,
		IsLayout: isLayout,
	}
}

// expandComponent processes slots and properties of components inside templates
func expandComponent(content string, compName string, comp Component) string {
	// Match attributes safely, skipping over quoted strings and PHP blocks <?php ... ?>
	attrPattern := `(\s+(?:[^>"]|"[^"]*"|'[^']*'|<\?php.*?\?>)*)`
	
	// Pattern for self-closing components: <Card title="xx" />
	selfClosingRegex := regexp.MustCompile(fmt.Sprintf(`(?s)<%s%s?/>`, compName, attrPattern))
	// Pattern for components with children: <Card title="xx">body</Card>
	blockRegex := regexp.MustCompile(fmt.Sprintf(`(?s)<%s%s?>(.*?)</%s>`, compName, attrPattern, compName))

	// Resolve block tags first
	content = blockRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := blockRegex.FindStringSubmatch(match)
		attrs := ""
		if len(submatches) > 1 {
			attrs = submatches[1]
		}
		slot := ""
		if len(submatches) > 2 {
			slot = submatches[2]
		}

		if comp.IsJsComponent {
			return renderJSComponent(comp, attrs, slot)
		}
		return renderTemplate(comp, attrs, slot)
	})

	// Resolve self-closing tags
	content = selfClosingRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := selfClosingRegex.FindStringSubmatch(match)
		attrs := ""
		if len(submatches) > 1 {
			attrs = submatches[1]
		}

		if comp.IsJsComponent {
			return renderJSComponent(comp, attrs, "")
		}
		return renderTemplate(comp, attrs, "")
	})

	return content
}

// renderJSComponent outputs the HTML placeholder and serializes PHP props in script tags
func renderJSComponent(comp Component, attrs string, slot string) string {
	props := make(map[string]string)
	attrMatches := attrRegex.FindAllStringSubmatch(attrs, -1)
	
	hydrate := ""
	for _, match := range attrMatches {
		name := match[1]
		val := ""
		if len(match) > 2 {
			val = match[2]
		}
		if strings.HasPrefix(name, "client:") {
			hydrate = strings.TrimPrefix(name, "client:")
		} else {
			props[name] = val
		}
	}
	
	islandId := fmt.Sprintf("pphlx-%s-%d", strings.ToLower(comp.Name), time.Now().UnixNano())
	
	var propsBuilder strings.Builder
	propsBuilder.WriteString("{")
	i := 0
	for k, v := range props {
		if i > 0 {
			propsBuilder.WriteString(",")
		}
		if strings.Contains(v, "<?php") {
			if strings.Contains(v, "json_encode") {
				propsBuilder.WriteString(fmt.Sprintf("%q: %s", k, v))
			} else {
				propsBuilder.WriteString(fmt.Sprintf("%q: %q", k, v))
			}
		} else {
			propsBuilder.WriteString(fmt.Sprintf("%q: %q", k, v))
		}
		i++
	}
	propsBuilder.WriteString("}")

	var result strings.Builder
	result.WriteString(fmt.Sprintf(`<div id="%s" class="pphlx-island" data-component="%s" data-hydrate="%s"></div>`, islandId, comp.Name, hydrate))
	result.WriteString("\n<script>\n")
	result.WriteString(fmt.Sprintf("  window.pphlxProps = window.pphlxProps || {};\n"))
	result.WriteString(fmt.Sprintf("  window.pphlxProps[%q] = %s;\n", islandId, propsBuilder.String()))
	result.WriteString("</script>\n")
	
	return result.String()
}

// renderTemplate interpolates properties and slots into a template string
func renderTemplate(comp Component, attrs string, slot string) string {
	result := comp.HTML

	// Extract attributes
	props := make(map[string]string)
	attrMatches := attrRegex.FindAllStringSubmatch(attrs, -1)
	for _, match := range attrMatches {
		props[match[1]] = match[2]
	}

	// 1. Replace slot
	result = strings.ReplaceAll(result, "{{slot}}", slot)

	// 2. Replace variables: {{title}} or {{type ?? 'default'}}
	result = phpShortTagRegex.ReplaceAllStringFunc(result, func(match string) string {
		submatches := phpShortTagRegex.FindStringSubmatch(match)
		varName := submatches[1]
		
		// Skip system placeholders so they can be injected in the final build stage
		if varName == "PPHLX_CSS" || varName == "PPHLX_JS" {
			return match
		}

		defaultVal := ""
		if len(submatches) > 2 {
			defaultVal = submatches[2]
		}

		if val, exists := props[varName]; exists {
			return val
		}
		return defaultVal
	})

	return result
}

// getRelativePath calculates the relative asset path from page to asset file
func getRelativePath(fromPathPath, toPathPath string) string {
	fromDir := filepath.Dir(fromPathPath)
	rel, err := filepath.Rel(fromDir, toPathPath)
	if err != nil {
		return toPathPath
	}
	return filepath.ToSlash(rel)
}

// parseMjsField extracts a configuration string property from JavaScript .mjs config using regex
func parseMjsField(content string, fieldName string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?m)%s\s*:\s*["']([^"']+)["']`, fieldName))
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parsePphlxBrackets converts custom {|= } and {| } tags to standard PHP echo and code blocks
func parsePphlxBrackets(content string) string {
	phpBracketEchoRegex := regexp.MustCompile(`(?s)\{\|=\s*(.*?)\s*\|\}`)
	phpBracketStatementRegex := regexp.MustCompile(`(?s)\{\|\s*(.*?)\s*\|\}`)

	// 1. Convert Echo tags first (so they don't get matched by the general statement pattern)
	content = phpBracketEchoRegex.ReplaceAllString(content, "<?php echo $1; ?>")

	// 2. Convert remaining Statement tags
	content = phpBracketStatementRegex.ReplaceAllString(content, "<?php $1 ?>")

	return content
}

// runViteBuild generates a temporary entry file, compiles Vue/Svelte components, and appends the result
func runViteBuild(config Config, projectDir string) error {
	entryPath := filepath.Join(projectDir, "src", ".pphlx_entry.js")
	var entryContent strings.Builder

	for name, path := range viteComponents {
		rel, err := filepath.Rel(filepath.Join(projectDir, "src"), path)
		if err != nil {
			rel = path
		}
		relPath := filepath.ToSlash(rel)
		entryContent.WriteString(fmt.Sprintf("import %s from './%s';\n", name, relPath))
		entryContent.WriteString(fmt.Sprintf("window.%s = %s;\n", name, name))
	}

	// Expose SolidJS render helper
	entryContent.WriteString("import { render as solidRender } from 'solid-js/web';\n")
	entryContent.WriteString("window.SolidJS = { render: solidRender };\n")

	err := ioutil.WriteFile(entryPath, []byte(entryContent.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write Vite entry file: %v", err)
	}

	viteConfigPath := filepath.Join(projectDir, "pphlx.vite.config.mjs")
	viteConfig := `
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import solidPlugin from 'vite-plugin-solid';

export default defineConfig({
  plugins: [vue(), svelte(), solidPlugin()],
  build: {
    lib: {
      entry: 'src/.pphlx_entry.js',
      formats: ['iife'],
      name: 'PphlxViteComponents',
      fileName: () => 'pphlx_vite.js'
    },
    rollupOptions: {
      external: ['vue'],
      output: {
        globals: {
          vue: 'Vue'
        }
      }
    },
    outDir: 'dist/assets/js',
    emptyOutDir: false,
    minify: true
  }
});
`
	ioutil.WriteFile(viteConfigPath, []byte(viteConfig), 0644)

	fmt.Println("Running local Vite compilation for Vue/Svelte components...")
	
	// Use powershell shell exec to spawn npx command on Windows safely
	cmd := exec.Command("cmd", "/c", "npx vite build --config pphlx.vite.config.mjs")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("vite build failed: %v", err)
	}

	// Append compiled bundles to global JS
	viteBundlePath := filepath.Join(projectDir, "dist", "assets", "js", "pphlx_vite.js")
	viteJS, err := ioutil.ReadFile(viteBundlePath)
	if err == nil {
		jsOut := filepath.Join(projectDir, config.JsOut)
		f, err := os.OpenFile(jsOut, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			f.Write([]byte("\n" + string(viteJS) + "\n"))
		}
		// Remove temporary vite asset file
		os.Remove(viteBundlePath)
	}

	// Clean up temporary entry file
	os.Remove(entryPath)

	return nil
}

// Add package dependencies zip downloader and extractor
func addDependency(repoURL string, projectDir string) error {
	version := "main"
	repoClean := repoURL
	if strings.Contains(repoURL, "@") {
		parts := strings.Split(repoURL, "@")
		repoClean = parts[0]
		version = parts[1]
	}

	parts := strings.Split(repoClean, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid repository URL format. Must be like: github.com/username/repo")
	}
	domain := parts[0]
	username := parts[1]
	repo := parts[2]

	if domain != "github.com" {
		return fmt.Errorf("only github.com is supported at this time")
	}

	// 1. Build GitHub download URL
	downloadURL := fmt.Sprintf("https://github.com/%s/%s/archive/%s.zip", username, repo, version)
	fmt.Printf("Downloading package from: %s\n", downloadURL)

	// 2. Fetch the ZIP file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to make download request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download package. Server returned status: %s", resp.Status)
	}

	// Create temporary directory for ZIP download
	tempDir, err := ioutil.TempDir("", "pphlx-pkg-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	zipFilePath := filepath.Join(tempDir, "package.zip")
	out, err := os.Create(zipFilePath)
	if err != nil {
		return fmt.Errorf("failed to create temp zip file: %v", err)
	}
	
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return fmt.Errorf("failed to write zip file contents: %v", err)
	}

	// 3. Extract the ZIP
	r, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return fmt.Errorf("failed to open downloaded zip file: %v", err)
	}
	defer r.Close()

	destDir := filepath.Join(projectDir, ".pphlx", "packages", domain, username, repo)
	// Clean previous package directory if exists
	os.RemoveAll(destDir)
	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory %s: %v", destDir, err)
	}

	// The ZIP contains a top-level directory like "repo-version" or "repo-commit-hash".
	// We want to skip this first folder level so the package root sits directly under destDir.
	var topLevelDir string
	if len(r.File) > 0 {
		firstFile := r.File[0].Name
		idx := strings.Index(firstFile, "/")
		if idx != -1 {
			topLevelDir = firstFile[:idx+1]
		}
	}

	for _, f := range r.File {
		// Skip files outside our top-level wrapper or directories
		if topLevelDir != "" && !strings.HasPrefix(f.Name, topLevelDir) {
			continue
		}
		
		relPath := f.Name
		if topLevelDir != "" {
			relPath = strings.TrimPrefix(f.Name, topLevelDir)
		}
		if relPath == "" {
			continue
		}

		fpath := filepath.Join(destDir, relPath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	// 4. Update pphlx.json dependencies
	manifestPath := filepath.Join(projectDir, "pphlx.json")
	var manifest map[string]interface{}
	
	if manifestData, err := ioutil.ReadFile(manifestPath); err == nil {
		json.Unmarshal(manifestData, &manifest)
	}
	if manifest == nil {
		manifest = make(map[string]interface{})
	}

	depsRaw, ok := manifest["dependencies"]
	var deps map[string]interface{}
	if ok {
		deps, _ = depsRaw.(map[string]interface{})
	}
	if deps == nil {
		deps = make(map[string]interface{})
	}

	deps[repoClean] = version
	manifest["dependencies"] = deps

	newManifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err == nil {
		ioutil.WriteFile(manifestPath, newManifestData, 0644)
	}

	return nil
}

// MCP JSON-RPC Protocol Structs
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type ToolListResult struct {
	Tools []Tool `json:"tools"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolCallResult struct {
	Content []Content `json:"content"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolCallArguments struct {
	Query          string `json:"query,omitempty"`
	Framework      string `json:"framework,omitempty"`
	ComponentName  string `json:"componentName,omitempty"`
	TargetPagePath string `json:"targetPagePath,omitempty"`
	Topic          string `json:"topic,omitempty"`
}

func runMCPServer() {
	// Write initialization log to stderr (standard stdout is reserved for JSON-RPC)
	fmt.Fprintln(os.Stderr, "PPHLX MCP server starting on stdio...")
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(req.ID, -32700, "Parse error: invalid JSON received")
			continue
		}

		switch req.Method {
		case "initialize":
			// Acknowledge connection and capabilities
			res := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"capabilities": map[string]interface{}{
						"tools": map[string]interface{}{},
					},
					"serverInfo": map[string]string{
						"name":    "pphlx-mcp",
						"version": "1.0.0",
					},
				},
			}
			sendResponse(res)

		case "notifications/initialized":
			// Client initialized handshake, no action needed

		case "tools/list":
			res := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: ToolListResult{
					Tools: []Tool{
						{
							Name:        "pphlx/search_docs",
							Description: "Search the PPHLX documentation database.",
							InputSchema: InputSchema{
								Type: "object",
								Properties: map[string]Property{
									"query": {Type: "string", Description: "Search keyword"},
								},
								Required: []string{"query"},
							},
						},
						{
							Name:        "pphlx/generate_island",
							Description: "Generate a component template file and append the @import statement inside a .pphx template.",
							InputSchema: InputSchema{
								Type: "object",
								Properties: map[string]Property{
									"framework": {
										Type:        "string",
										Description: "The UI framework layout style",
										Enum:        []string{"react", "vue", "svelte", "solid", "preact"},
									},
									"componentName":  {Type: "string", Description: "PascalCase component name (e.g. ShoppingCart)"},
									"targetPagePath": {Type: "string", Description: "Absolute path to the target .pphx file"},
								},
								Required: []string{"framework", "componentName", "targetPagePath"},
							},
						},
						{
							Name:        "pphlx/get_best_practices",
							Description: "Get best practice guidelines for PPHLX development topics.",
							InputSchema: InputSchema{
								Type: "object",
								Properties: map[string]Property{
									"topic": {
										Type:        "string",
										Description: "Target guideline topic",
										Enum:        []string{"state-sharing", "styling", "routing", "php-variables"},
									},
								},
								Required: []string{"topic"},
							},
						},
					},
				},
			}
			sendResponse(res)

		case "tools/call":
			var callArgs struct {
				Name      string            `json:"name"`
				Arguments ToolCallArguments `json:"arguments"`
			}
			if err := json.Unmarshal(req.Params, &callArgs); err != nil {
				sendError(req.ID, -32602, "Invalid params")
				continue
			}

			result, err := handleToolCall(callArgs.Name, callArgs.Arguments)
			if err != nil {
				sendError(req.ID, -32603, err.Error())
			} else {
				res := JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result:  result,
				}
				sendResponse(res)
			}

		default:
			sendError(req.ID, -32601, "Method not found")
		}
	}
}

func handleToolCall(toolName string, args ToolCallArguments) (interface{}, error) {
	switch toolName {
	case "pphlx/search_docs":
		query := strings.ToLower(args.Query)
		docs := fmt.Sprintf("No exact matches found for \"%s\".\n\nPPHLX standard syntax:\n- Templates use the .pphx extension.\n- Framework integration components are imported using @import (e.g. @import Button from '../components/ThemeLayout').", query)

		if strings.Contains(query, "import") || strings.Contains(query, "extension") {
			docs = "Importing components in PPHLX:\n- React/JSX components can omit extensions: @import Header from './Header'\n- Vue, Svelte, Solid, and Preact components must include their extensions: @import Counter from './Counter.vue' or @import Card from './Card.svelte'"
		} else if strings.Contains(query, "state") || strings.Contains(query, "share") {
			docs = "State Sharing in PPHLX:\n- Expose state globally or dispatch custom events between framework islands: window.dispatchEvent(new CustomEvent('pphlx-state-update', { detail: data }))\n- Svelte or Vue islands can listen and react in real-time."
		}

		return ToolCallResult{
			Content: []Content{
				{Type: "text", Text: docs},
			},
		}, nil

	case "pphlx/generate_island":
		// 1. Verify target page exists
		targetPath := args.TargetPagePath
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("target template file does not exist at: %s", targetPath)
		}

		// 2. Find nearest project directory (contain pphlx.json) to locate src/components
		dir := filepath.Dir(targetPath)
		projectDir := ""
		for {
			if _, err := os.Stat(filepath.Join(dir, "pphlx.json")); err == nil {
				projectDir = dir
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		if projectDir == "" {
			projectDir = filepath.Dir(targetPath)
		}

		// Read pphlx.json to resolve srcDir, defaulting to "src"
		srcDirName := "src"
		manifestPath := filepath.Join(projectDir, "pphlx.json")
		if manifestData, err := ioutil.ReadFile(manifestPath); err == nil {
			var manifest map[string]interface{}
			if err := json.Unmarshal(manifestData, &manifest); err == nil {
				if val, ok := manifest["srcDir"]; ok {
					if srcStr, isStr := val.(string); isStr {
						srcDirName = srcStr
					}
				}
			}
		}

		componentsDir := filepath.Join(projectDir, srcDirName, "components")
		os.MkdirAll(componentsDir, 0755)

		fileExtension := "jsx"
		boilerplateCode := ""

		switch args.Framework {
		case "react":
			fileExtension = "jsx"
			boilerplateCode = fmt.Sprintf("export default function %s() {\n  return (\n    <button className=\"px-4 py-2 bg-blue-600 text-white rounded\">\n      %s Island\n    </button>\n  );\n}\n", args.ComponentName, args.ComponentName)
		case "vue":
			fileExtension = "vue"
			boilerplateCode = fmt.Sprintf("<template>\n  <button class=\"px-4 py-2 bg-emerald-600 text-white rounded\">\n    {{ label }}\n  </button>\n</template>\n\n<script>\nexport default {\n  data() {\n    return {\n      label: '%s Island'\n    };\n  }\n};\n</script>\n", args.ComponentName)
		case "svelte":
			fileExtension = "svelte"
			boilerplateCode = fmt.Sprintf("<script>\n  let label = '%s Island';\n</script>\n\n<button class=\"px-4 py-2 bg-orange-600 text-white rounded\">\n  {label}\n</button>\n", args.ComponentName)
		case "solid":
			fileExtension = "solid.jsx"
			boilerplateCode = fmt.Sprintf("import { createSignal } from \"solid-js\";\n\nexport default function %s() {\n  const [label] = createSignal(\"%s Island\");\n  return (\n    <button class=\"px-4 py-2 bg-cyan-600 text-white rounded\">\n      {label()}\n    </button>\n  );\n}\n", args.ComponentName, args.ComponentName)
		case "preact":
			fileExtension = "js"
			boilerplateCode = fmt.Sprintf("export default function %s() {\n  return (\n    <button class=\"px-4 py-2 bg-indigo-600 text-white rounded\">\n      %s Island\n    </button>\n  );\n}\n", args.ComponentName, args.ComponentName)
		}

		componentFilename := fmt.Sprintf("%s.%s", args.ComponentName, fileExtension)
		componentFullPath := filepath.Join(componentsDir, componentFilename)

		// Write component file
		if err := ioutil.WriteFile(componentFullPath, []byte(boilerplateCode), 0644); err != nil {
			return nil, fmt.Errorf("failed to write component file: %v", err)
		}

		// Inject @import at the top of the template file
		pageData, err := ioutil.ReadFile(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read template page: %v", err)
		}
		pageContent := string(pageData)

		extImport := fmt.Sprintf(".%s", fileExtension)
		if args.Framework == "react" {
			extImport = ""
		}
		importStatement := fmt.Sprintf("@import %s from '../components/%s%s'\n", args.ComponentName, args.ComponentName, extImport)

		if !strings.Contains(pageContent, fmt.Sprintf("@import %s", args.ComponentName)) {
			newContent := importStatement + pageContent
			ioutil.WriteFile(targetPath, []byte(newContent), 0644)
		}

		successMessage := fmt.Sprintf("[Success] Generated %s component: %s\nUpdated template: %s", args.Framework, componentFullPath, targetPath)
		return ToolCallResult{
			Content: []Content{
				{Type: "text", Text: successMessage},
			},
		}, nil

	case "pphlx/get_best_practices":
		topic := args.Topic
		responseText := ""

		if topic == "state-sharing" {
			responseText = "Best Practice: Share reactive state across separate framework islands using custom window events or micro-stores, avoiding heavy framework-specific context APIs."
		} else if topic == "styling" {
			responseText = "Best Practice: PPHLX supports tailwind out of the box. Scope vanilla CSS styles inside components to avoid global stylesheet conflicts."
		} else {
			responseText = fmt.Sprintf("Guidelines for topic: %s. Always preserve Smarty Variable key structures when referencing PHP properties inside template blocks.", topic)
		}

		return ToolCallResult{
			Content: []Content{
				{Type: "text", Text: responseText},
			},
		}, nil

	default:
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
}

func sendResponse(res JSONRPCResponse) {
	data, _ := json.Marshal(res)
	fmt.Println(string(data))
}

func sendError(id interface{}, code int, message string) {
	res := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	sendResponse(res)
}

func installMCPServer() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	mcpDestDir := filepath.Join(homeDir, ".gemini", "antigravity-ide", "mcp", "pphlx")
	err = os.MkdirAll(mcpDestDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	files := []string{
		"mcp/search_docs.json",
		"mcp/generate_island.json",
		"mcp/get_best_practices.json",
		"mcp/instructions.md",
	}

	for _, file := range files {
		data, err := mcpFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %v", file, err)
		}

		baseName := filepath.Base(file)
		destPath := filepath.Join(mcpDestDir, baseName)

		if baseName == "instructions.md" {
			execPath, err := os.Executable()
			if err == nil {
				execPath = filepath.Clean(execPath)
				content := string(data)
				content = strings.Replace(content, `F:\VS CODE\GO\PPHLX\pphlx.exe`, execPath, 1)
				data = []byte(content)
			}
		}

		err = ioutil.WriteFile(destPath, data, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s to %s: %v", baseName, destPath, err)
		}
		fmt.Printf("Registered tool file: %s\n", destPath)
	}

	return nil
}

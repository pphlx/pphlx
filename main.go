package main

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/fsnotify/fsnotify"
)

// Single source of truth for PPHLX compiler version
const Version = "1.1.6"

// OutputConfig holds compilation output options
type OutputConfig struct {
	Target string `json:"target"`
	Goos   string `json:"goos"`
	Goarch string `json:"goarch"`
}

type ConfigProfile struct {
	OutDir string       `json:"outDir"`
	CssOut string       `json:"cssOut"`
	JsOut  string       `json:"jsOut"`
	Output OutputConfig `json:"output"`
}

// Config holds project configuration
type Config struct {
	SrcDir       string                   `json:"srcDir"`
	OutDir       string                   `json:"outDir"`
	CssOut       string                   `json:"cssOut"`
	JsOut        string                   `json:"jsOut"`
	Site         string                   `json:"site"`
	Base         string                   `json:"base"`
	Sitemap      bool                     `json:"sitemap"`
	Default      string                   `json:"default"`
	Output       OutputConfig             `json:"output"`
	Environments map[string]ConfigProfile `json:"environments"`
}

type TelemetryConfig struct {
	Target      string `json:"target"`
	HasProfiles bool   `json:"hasProfiles"`
	Goos        string `json:"goos"`
	Goarch      string `json:"goarch"`
}

type TelemetryPayload struct {
	CliCommand string          `json:"cliCommand"`
	Version    string          `json:"version"`
	OS         string          `json:"os"`
	Arch       string          `json:"arch"`
	Config     TelemetryConfig `json:"config"`
	Nonce      string          `json:"powNonce"`
}

// solvePoW generates a SHA-256 Proof-of-Work nonce starting with "0000"
func solvePoW(base string) string {
	var nonce int
	for {
		candidate := fmt.Sprintf("%s:%d", base, nonce)
		hash := sha256.Sum256([]byte(candidate))
		hexHash := hex.EncodeToString(hash[:])
		if strings.HasPrefix(hexHash, "0000") {
			return fmt.Sprintf("%d", nonce)
		}
		nonce++
		if nonce > 2000000 {
			// Safety fallback
			return "0"
		}
	}
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
	Framework     string
}

//go:embed mcp/*
var mcpFS embed.FS

var (
	importRegex          = regexp.MustCompile(`(?m)^@import\s+(\w+)\s+from\s+['"]([^'"]+)['"]\s*\r?\n?`)
	styleRegex           = regexp.MustCompile(`(?s)<style>(.*?)</style>`)
	scriptRegex          = regexp.MustCompile(`(?s)<script>(.*?)</script>`)
	templateRegex        = regexp.MustCompile(`(?s)<template[^>]*>(.*?)</template>`)
	attrRegex            = regexp.MustCompile(`(\w+[\w:-]*)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|\{([^}]+)\}|([^\s>]+)))?`)
	phpShortTagRegex     = regexp.MustCompile(`\{\{\s*(\w+)(?:\s*\?\?\s*'(.*?)')?\s*\}\}`)
	viteComponents       = make(map[string]string)
	sitemapOverrideRegex = regexp.MustCompile(`(?i)@pphlx-sitemap:\s*(true|false)`)
)

// VirtualFiles stores project files map when running in WebAssembly
var VirtualFiles map[string]string

var activeConfig Config
var activeMode string
var activeDevOutDir string

func readProjectFile(filePath string) ([]byte, error) {
	if VirtualFiles != nil {
		normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
		if content, ok := VirtualFiles[normalizedPath]; ok {
			return []byte(content), nil
		}
		cleanPath := strings.TrimPrefix(normalizedPath, "./")
		if content, ok := VirtualFiles[cleanPath]; ok {
			return []byte(content), nil
		}
		return nil, fmt.Errorf("file not found in VFS: %s", filePath)
	}
	return os.ReadFile(filePath)
}

// Diagnostic represents an AST or compiler lint error
type Diagnostic struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Snippet string `json:"snippet"`
}

func (d Diagnostic) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n\033[1;31m[PPHLX ERROR] %s\033[0m\n", d.Code))
	sb.WriteString(fmt.Sprintf("File: \033[36m%s:%d:%d\033[0m\n", d.File, d.Line, d.Column))
	if d.Snippet != "" {
		sb.WriteString(fmt.Sprintf("  \033[90m%4d |\033[0m %s\n", d.Line, d.Snippet))
	}
	sb.WriteString(fmt.Sprintf("  \033[1;33mReason:\033[0m %s\n", d.Message))
	return sb.String()
}

var (
	imgSrcRegex         = regexp.MustCompile(`(?i)<img[^>]+src=["']([^"']+)["']`)
	linkHrefRegex       = regexp.MustCompile(`(?i)<link[^>]+href=["']([^"']+)["']`)
	sourceSrcsetRegex   = regexp.MustCompile(`(?i)<source[^>]+srcset=["']([^"']+)["']`)
	componentOpenRegex  = regexp.MustCompile(`<([A-Z]\w*)(?:\s+[^>]*)?>`)
	componentCloseRegex = regexp.MustCompile(`</([A-Z]\w*)>`)
)

// RunDiagnostics performs static AST linting, asset existence checks, and tag integrity validation
func RunDiagnostics(srcDir string, projectDir string) []Diagnostic {
	var diagnostics []Diagnostic
	if !projectFileExists(srcDir) {
		return diagnostics
	}

	_ = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".pphx") {
			return nil
		}

		contentBytes, err := readProjectFile(path)
		if err != nil {
			return nil
		}
		content := string(contentBytes)
		lines := strings.Split(content, "\n")

		relPath, _ := filepath.Rel(projectDir, path)
		if relPath == "" {
			relPath = path
		}

		// 1. Check for broken image/asset paths (<img src> and <link href>)
		matches := append(imgSrcRegex.FindAllStringSubmatchIndex(content, -1), linkHrefRegex.FindAllStringSubmatchIndex(content, -1)...)
		for _, m := range matches {
			if len(m) >= 4 {
				srcVal := content[m[2]:m[3]]
				// Skip dynamic PHP expressions, HTTP URLs, external CSS stylesheets, data URIs, or absolute web paths
				if strings.HasPrefix(srcVal, "http://") || strings.HasPrefix(srcVal, "https://") ||
					strings.HasPrefix(srcVal, "//") || strings.HasPrefix(srcVal, "data:") ||
					strings.HasPrefix(srcVal, "<?") || strings.HasPrefix(srcVal, "{|") ||
					strings.HasSuffix(srcVal, ".css") {
					continue
				}

				// Resolve asset relative to srcDir, projectDir, publicDir, or file dir
				assetPath1 := filepath.Join(filepath.Dir(path), srcVal)
				assetPath2 := filepath.Join(srcDir, strings.TrimPrefix(srcVal, "/"))
				assetPath3 := filepath.Join(projectDir, strings.TrimPrefix(srcVal, "/"))
				assetPath4 := filepath.Join(projectDir, "public", strings.TrimPrefix(srcVal, "/"))
				assetPath5 := filepath.Join(srcDir, "assets", strings.TrimPrefix(srcVal, "/"))
				assetPath6 := filepath.Join(projectDir, "public", "assets", strings.TrimPrefix(srcVal, "/"))

				if !projectFileExists(assetPath1) && !projectFileExists(assetPath2) && !projectFileExists(assetPath3) && !projectFileExists(assetPath4) && !projectFileExists(assetPath5) && !projectFileExists(assetPath6) {
					lineNo := 1
					for i, line := range lines {
						if strings.Contains(line, srcVal) {
							lineNo = i + 1
							break
						}
					}
					snippet := ""
					if lineNo <= len(lines) {
						snippet = strings.TrimSpace(lines[lineNo-1])
					}
					diagnostics = append(diagnostics, Diagnostic{
						File:    relPath,
						Line:    lineNo,
						Column:  1,
						Code:    "Asset Resolution Warning",
						Message: fmt.Sprintf("Asset not found in local workspace paths: '%s'", srcVal),
						Snippet: snippet,
					})
				}
			}
		}

		// 2. Check for unclosed brace-pipe expression delimiters (file-level check for multi-line blocks)
		totalOpen := strings.Count(content, "{|")
		totalClose := strings.Count(content, "|}")
		if totalOpen > totalClose {
			lineNo := 1
			for i, l := range lines {
				if strings.Contains(l, "{|") {
					lineNo = i + 1
					break
				}
			}
			diagnostics = append(diagnostics, Diagnostic{
				File:    relPath,
				Line:    lineNo,
				Column:  1,
				Code:    "Unclosed Expression Delimiter",
				Message: fmt.Sprintf("Unbalanced '{|' delimiters in file (%d opening vs %d closing)", totalOpen, totalClose),
				Snippet: strings.TrimSpace(lines[lineNo-1]),
			})
		}

		// 3. Check for unimported components (Capitalized tags without @import directive)
		importedComps := make(map[string]bool)
		for _, m := range importRegex.FindAllStringSubmatch(content, -1) {
			if len(m) > 1 {
				importedComps[m[1]] = true
			}
		}

		for lineIdx, line := range lines {
			lineNo := lineIdx + 1
			opens := componentOpenRegex.FindAllStringSubmatch(line, -1)
			for _, op := range opens {
				if len(op) > 1 {
					compName := op[1]
					if !importedComps[compName] {
						diagnostics = append(diagnostics, Diagnostic{
							File:    relPath,
							Line:    lineNo,
							Column:  strings.Index(line, "<"+compName) + 1,
							Code:    "Unimported Component",
							Message: fmt.Sprintf("Component '<%s>' is used without an @import directive at top of file", compName),
							Snippet: strings.TrimSpace(line),
						})
					}
				}
			}
		}

		return nil
	})

	return diagnostics
}

func projectFileExists(filePath string) bool {
	if VirtualFiles != nil {
		normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
		if _, ok := VirtualFiles[normalizedPath]; ok {
			return true
		}
		cleanPath := strings.TrimPrefix(normalizedPath, "./")
		if _, ok := VirtualFiles[cleanPath]; ok {
			return true
		}
		return false
	}
	_, err := os.Stat(filePath)
	return err == nil
}

// resolveSourceRootAndEntry determines the source root directory and exact entry template file based on config.SrcDir
func resolveSourceRootAndEntry(config Config, projectDir string) (srcRootDir string, entryFile string) {
	srcVal := strings.TrimSpace(config.SrcDir)
	if srcVal == "" {
		srcVal = "src"
	}

	fullPath := filepath.Join(projectDir, srcVal)
	fi, err := os.Stat(fullPath)
	if err == nil && !fi.IsDir() {
		// srcDir points directly to a file e.g. "src/index.pphx" or "src/demo.pphx"
		return filepath.Dir(fullPath), fullPath
	}

	// If fullPath + ".pphx" exists (e.g., "src/demo.pphx" when srcDir is "src/demo")
	if _, err := os.Stat(fullPath + ".pphx"); err == nil {
		return filepath.Dir(fullPath), fullPath + ".pphx"
	}

	// Check if index.pphx exists inside directory
	indexFile := filepath.Join(fullPath, "index.pphx")
	if _, err := os.Stat(indexFile); err == nil {
		return fullPath, indexFile
	}

	// Default fallback
	return fullPath, indexFile
}

// pruneEmptyDirs removes any empty subdirectories inside dirPath
func pruneEmptyDirs(dirPath string) {
	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || path == dirPath {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err == nil && len(entries) == 0 {
			_ = os.Remove(path)
		}
		return nil
	})
}

// wipeDirContents removes all files and subdirectories inside dirPath without removing dirPath itself
func wipeDirContents(dirPath string) error {
	if !projectFileExists(dirPath) {
		return os.MkdirAll(dirPath, 0755)
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		itemPath := filepath.Join(dirPath, entry.Name())
		_ = os.RemoveAll(itemPath)
	}
	return nil
}

// loadPphlxIgnore reads .pphlxignore file from projectDir if present
func loadPphlxIgnore(projectDir string) []string {
	ignoreFile := filepath.Join(projectDir, ".pphlxignore")
	contentBytes, err := readProjectFile(ignoreFile)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(contentBytes), "\n")
	var patterns []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

// isPphlxIgnored checks if relative path matches any pattern in .pphlxignore
func isPphlxIgnored(relPath string, ignorePatterns []string) bool {
	if len(ignorePatterns) == 0 {
		return false
	}
	cleanRel := strings.ReplaceAll(relPath, "\\", "/")
	cleanRel = strings.TrimPrefix(cleanRel, "./")

	for _, pattern := range ignorePatterns {
		cleanPat := strings.ReplaceAll(pattern, "\\", "/")
		cleanPat = strings.TrimPrefix(cleanPat, "./")
		cleanPat = strings.TrimSuffix(cleanPat, "/")

		if cleanPat == "" {
			continue
		}

		if cleanRel == cleanPat || strings.HasPrefix(cleanRel, cleanPat+"/") {
			return true
		}

		baseName := filepath.Base(cleanRel)
		if matched, _ := filepath.Match(cleanPat, baseName); matched {
			return true
		}
		if matched, _ := filepath.Match(cleanPat, cleanRel); matched {
			return true
		}
	}
	return false
}

// FrameworkSourceExtensions defines supported UI framework source extensions that must be omitted from standalone dist/ copying if not compiled
var FrameworkSourceExtensions = map[string]bool{
	".jsx":       true,
	".tsx":       true,
	".vue":       true,
	".svelte":    true,
	".solid.jsx": true,
	".solid.tsx": true,
	".ts":        true,
	".mts":       true,
	".cts":       true,
	".marko":     true,
	".astro":     true,
}

// isFrameworkSourceFile checks if a path ends with any supported UI framework source extension
func isFrameworkSourceFile(filePath string) bool {
	lowerPath := strings.ToLower(filePath)
	if strings.HasSuffix(lowerPath, ".solid.jsx") || strings.HasSuffix(lowerPath, ".solid.tsx") {
		return true
	}
	ext := filepath.Ext(lowerPath)
	return FrameworkSourceExtensions[ext]
}

// buildDependencyGraph scans @import statements across srcDir to find imported components/helpers
func buildDependencyGraph(srcDir string, ignorePatterns []string) map[string]bool {
	importedAsComponent := make(map[string]bool)
	if !projectFileExists(srcDir) {
		return importedAsComponent
	}

	// Recursive scanner helper for nested component dependencies
	var scanFileDependencies func(filePath string)
	scanFileDependencies = func(filePath string) {
		relPath, errRel := filepath.Rel(srcDir, filePath)
		if errRel == nil && isPphlxIgnored(relPath, ignorePatterns) {
			return
		}

		contentBytes, err := readProjectFile(filePath)
		if err != nil {
			return
		}
		dir := filepath.Dir(filePath)
		matches := importRegex.FindAllStringSubmatch(string(contentBytes), -1)
		for _, m := range matches {
			if len(m) > 2 {
				importPath := m[2]
				resolvedPath := filepath.Clean(filepath.Join(dir, importPath))
				cleanSlash := strings.ReplaceAll(resolvedPath, "\\", "/")

				childRel, childErr := filepath.Rel(srcDir, resolvedPath)
				if childErr == nil && isPphlxIgnored(childRel, ignorePatterns) {
					continue
				}

				if !importedAsComponent[resolvedPath] {
					importedAsComponent[resolvedPath] = true
					importedAsComponent[cleanSlash] = true
					if projectFileExists(resolvedPath) {
						scanFileDependencies(resolvedPath)
					}
				}
			}
		}
	}

	_ = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err == nil && isPphlxIgnored(relPath, ignorePatterns) {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".pphx") || isFrameworkSourceFile(path) {
			scanFileDependencies(path)
		}
		return nil
	})

	return importedAsComponent
}

// formatDuration formats a time.Duration into dynamic human-readable units (ns, µs, ms, s, m, h, d, mo, y)
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000.0)
	}
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		secs := d.Seconds() - float64(mins*60)
		return fmt.Sprintf("%dm %.1fs", mins, secs)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	days := int(d.Hours() / 24)
	if days < 30 {
		hours := int(d.Hours()) % 24
		mins := int(d.Minutes()) % 60
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	months := days / 30
	remDays := days % 30
	if months < 12 {
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%dmo %dd %dh", months, remDays, hours)
	}
	years := months / 12
	remMonths := months % 12
	return fmt.Sprintf("%dy %dmo %dd", years, remMonths, remDays)
}

// compileAll executes the main compilation loop for all templates
func compileAll(config Config, projectDir string) {
	startTime := time.Now()
	activeConfig = config
	targetLower := strings.ToLower(config.Output.Target)
	fmt.Printf("[%s] Rebuilding templates...\n", time.Now().Format("15:04:05"))

	srcRootDir, entryFile := resolveSourceRootAndEntry(config, projectDir)
	srcDir := srcRootDir
	outDir := filepath.Join(projectDir, config.OutDir)
	if config.OutDir == "" {
		outDir = filepath.Join(projectDir, "dist")
	}

	cssOutVal := config.CssOut
	if cssOutVal == "" {
		cssOutVal = filepath.Join(config.OutDir, "assets", "css", "styles.css")
		if config.OutDir == "" {
			cssOutVal = filepath.Join("dist", "assets", "css", "styles.css")
		}
	}
	cssOut := filepath.Join(projectDir, cssOutVal)

	jsOutVal := config.JsOut
	if jsOutVal == "" {
		jsOutVal = filepath.Join(config.OutDir, "assets", "js", "bundle.js")
		if config.OutDir == "" {
			jsOutVal = filepath.Join("dist", "assets", "js", "bundle.js")
		}
	}
	jsOut := filepath.Join(projectDir, jsOutVal)

	// Run AST Diagnostic & Lint Engine
	diagnostics := RunDiagnostics(srcDir, projectDir)
	if len(diagnostics) > 0 {
		for _, d := range diagnostics {
			fmt.Printf("\033[33m[PPHLX DIAGNOSTIC WARNING]\033[0m %s\n", d.String())
		}
	}

	targetOutDir := outDir
	_ = wipeDirContents(targetOutDir)

	// Load .pphlxignore patterns and build component dependency graph
	ignorePatterns := loadPphlxIgnore(projectDir)
	importedAsComponent := buildDependencyGraph(srcDir, ignorePatterns)

	// Reset Vite components map for this compile run
	viteComponents = make(map[string]string)

	// Bundles
	var globalCSS strings.Builder
	var globalJS strings.Builder

	// Prepend require, process shims and pphlx.desktop API helper at the very beginning of the global JS bundle
	globalJS.WriteString(`
// Require shim for browser CDN compatibility
window.require = window.require || function(mod) {
  if (mod === 'react') return window.React;
  if (mod === 'react-dom') return window.ReactDOM;
  return undefined;
};

// Process environment shim for Vue/Svelte compatibility
window.process = window.process || { env: { NODE_ENV: 'production' } };

// PPHLX Desktop App API Bridge
window.pphlx = window.pphlx || {};
window.pphlx.desktop = {
  openFileDialog: async function(options) {
    if (window.pphlxDesktopOpenFile) {
      return await window.pphlxDesktopOpenFile();
    }
    console.warn("pphlx.desktop.openFileDialog is only available in native desktop target.");
    return null;
  },
  saveFileDialog: async function(options) {
    if (window.pphlxDesktopSaveFile) {
      return await window.pphlxDesktopSaveFile();
    }
    console.warn("pphlx.desktop.saveFileDialog is only available in native desktop target.");
    return null;
  },
  showNotification: function(title, message) {
    if (window.pphlxDesktopShowNotification) {
      window.pphlxDesktopShowNotification(title, message);
    } else {
      alert(title + ": " + message);
    }
  },
  window: {
    minimize: function() {
      if (window.pphlxDesktopMinimize) {
        window.pphlxDesktopMinimize();
      } else {
        console.warn("window.minimize is only available in native desktop target.");
      }
    },
    maximize: function() {
      if (window.pphlxDesktopMaximize) {
        window.pphlxDesktopMaximize();
      } else {
        console.warn("window.maximize is only available in native desktop target.");
      }
    },
    close: function() {
      if (window.pphlxDesktopClose) {
        window.pphlxDesktopClose();
      } else {
        window.close();
      }
    }
  }
};
`)

	hasPartytown := false
	compiledScripts := make(map[string]bool)
	compiledStyles := make(map[string]bool)
	var sitemapURLs []string
	sitemapAdded := make(map[string]bool)

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

		// Check .pphlxignore
		if isPphlxIgnored(relPath, ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		outPath := filepath.Join(targetOutDir, relPath)

		if info.IsDir() {
			dirName := info.Name()
			if relPath != "." && (dirName == "node_modules" || dirName == "dist" || dirName == ".git" || dirName == ".vscode" || dirName == "storage" || dirName == ".antigravity" || strings.HasPrefix(dirName, ".")) {
				return filepath.SkipDir
			}
			return nil
		}

		// Process files
		cleanRel := strings.ReplaceAll(relPath, "\\", "/")
		isInsidePages := strings.HasPrefix(cleanRel, "pages/") || cleanRel == "pages"
		canonicalPath := filepath.Clean(path)
		canonicalPathSlash := strings.ReplaceAll(canonicalPath, "\\", "/")

		// Universal Component Suppression Rule:
		// If ANY file (.pphx, .js, .jsx, .vue, .svelte, .solid, .ts, .php, .html)
		// was imported as a component dependency and is NOT inside pages/, suppress standalone emission/copying in dist/!
		if (importedAsComponent[canonicalPath] || importedAsComponent[canonicalPathSlash]) && !isInsidePages {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".pphx") {
			ext := ".php"
			switch targetLower {
			case "ssg", "android", "ios":
				ext = ".html"
			case "blade":
				ext = ".blade.php"
			case "twig":
				ext = ".html.twig"
			}
			phpOutPath := strings.TrimSuffix(outPath, ".pphx") + ext

			pageContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Check sitemap override tags
			includeInSitemap := false
			matches := sitemapOverrideRegex.FindStringSubmatch(string(pageContent))
			if len(matches) > 1 {
				includeInSitemap = strings.ToLower(matches[1]) == "true"
			} else {
				// Fallback to default
				defaultMode := strings.ToLower(config.Default)
				if defaultMode == "" || defaultMode == "public" {
					includeInSitemap = true
				}
			}

			if includeInSitemap {
				rel, err := filepath.Rel(srcDir, path)
				if err == nil {
					cleanPath := strings.ReplaceAll(rel, "\\", "/")
					cleanPath = strings.TrimSuffix(cleanPath, ".pphx")
					if cleanPath == "index" {
						cleanPath = ""
					} else if strings.HasSuffix(cleanPath, "/index") {
						cleanPath = strings.TrimSuffix(cleanPath, "/index")
					}

					// Define sitemapAdded map in compileAll to track unique paths
					if !sitemapAdded[cleanPath] {
						sitemapAdded[cleanPath] = true
						sitemapURLs = append(sitemapURLs, cleanPath)
					}
				}
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

			// Autonomous Conditional Asset Tag Injection
			hasCssContent := strings.TrimSpace(globalCSS.String()) != ""
			hasJsContent := strings.TrimSpace(globalJS.String()) != "" || len(viteComponents) > 0

			hasCssPlaceholder := strings.Contains(compiledPage, "{{PPHLX_CSS}}")
			hasJsPlaceholder := strings.Contains(compiledPage, "{{PPHLX_JS}}")

			basePrefix := "/"
			if strings.TrimSpace(config.Base) != "" {
				basePrefix = "/" + strings.Trim(config.Base, "/")
				if basePrefix != "/" {
					basePrefix += "/"
				}
			}

			if hasCssContent {
				cssOutForRel := cssOut
				if strings.TrimSpace(config.CssOut) == "" {
					cssOutForRel = filepath.Join(targetOutDir, "assets", "css", "styles.css")
				}
				relToOut, err := filepath.Rel(targetOutDir, cssOutForRel)
				if err != nil || strings.HasPrefix(relToOut, "..") {
					relToOut = "assets/css/styles.css"
				}
				cssRelPath := basePrefix + strings.TrimPrefix(filepath.ToSlash(relToOut), "/")
				cssTag := fmt.Sprintf(`<link rel="stylesheet" href="%s">`, cssRelPath)
				if hasCssPlaceholder {
					compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_CSS}}", cssTag)
				} else if strings.Contains(compiledPage, "</head>") {
					compiledPage = strings.ReplaceAll(compiledPage, "</head>", cssTag+"\n</head>")
				}
			} else if hasCssPlaceholder {
				compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_CSS}}", "")
			}

			if hasJsContent {
				jsTargetForRel := jsOut
				if strings.TrimSpace(config.JsOut) == "" {
					jsTargetForRel = filepath.Join(targetOutDir, "assets", "js", "bundle.js")
				}
				relToOut, err := filepath.Rel(targetOutDir, jsTargetForRel)
				if err != nil || strings.HasPrefix(relToOut, "..") {
					relToOut = "assets/js/bundle.js"
				}
				jsRelPath := basePrefix + strings.TrimPrefix(filepath.ToSlash(relToOut), "/")
				if jsRelPath != "" {
					jsTag := fmt.Sprintf(`<script src="%s"></script>`, jsRelPath)
					if hasJsPlaceholder {
						compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_JS}}", jsTag)
					} else if strings.Contains(compiledPage, "</body>") {
						compiledPage = strings.ReplaceAll(compiledPage, "</body>", jsTag+"\n</body>")
					} else if strings.Contains(compiledPage, "</head>") {
						compiledPage = strings.ReplaceAll(compiledPage, "</head>", jsTag+"\n</head>")
					}
				} else if hasJsPlaceholder {
					compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_JS}}", "")
				}
			} else if hasJsPlaceholder {
				compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_JS}}", "")
			}

			if strings.Contains(compiledPage, `type="text/partytown"`) || strings.Contains(compiledPage, `type='text/partytown'`) {
				hasPartytown = true
				partytownSnippet := `<script>!(function(w,p,f,c){c=w[p]=w[p]||{};c.lib='/~partytown/';var s=w.document.createElement('script');s.src=c.lib+'partytown.js';s.defer=true;w.document.head.appendChild(s);})(window,'partytown');</script>`
				if strings.Contains(compiledPage, "</head>") {
					compiledPage = strings.ReplaceAll(compiledPage, "</head>", partytownSnippet+"\n</head>")
				}
			}

			pageBytes := []byte(strings.TrimLeft(compiledPage, " \t\r\n"))
			if targetLower == "ssg" || targetLower == "android" || targetLower == "ios" {
				// Evaluate PHP to static HTML
				tempFile := phpOutPath + ".tmp.php"
				err = os.WriteFile(tempFile, pageBytes, 0644)
				if err == nil {
					cmd := exec.Command("php", tempFile)
					outBytes, cmdErr := cmd.Output()
					os.Remove(tempFile)
					if cmdErr == nil {
						pageBytes = outBytes
					} else {
						fmt.Printf("[Warning] Failed to run PHP compiler for SSG on %s: %v. Outputting raw template.\n", phpOutPath, cmdErr)
					}
				}
			}

			os.MkdirAll(filepath.Dir(phpOutPath), 0755)
			err = os.WriteFile(phpOutPath, pageBytes, 0644)
			if err != nil {
				return err
			}
			if entryFile != "" && filepath.Clean(path) == filepath.Clean(entryFile) {
				entryOutPath := filepath.Join(targetOutDir, "index"+ext)
				if entryOutPath != phpOutPath {
					_ = os.WriteFile(entryOutPath, pageBytes, 0644)
				}
			}
		} else {
			// Skip project manifests, configs, lockfiles, and unabsorbed framework source files from being copied as static assets into dist/
			baseName := info.Name()
			if baseName == "package.json" || baseName == "package-lock.json" || baseName == "pphlx.json" || baseName == "pphlx.config.json" || baseName == "pphlx.config.mjs" || baseName == "pphlx.vite.config.mjs" || baseName == "go.mod" || baseName == "go.sum" || strings.HasPrefix(baseName, ".") {
				return nil
			}

			// Omit unabsorbed framework source files (.jsx, .vue, .svelte, etc.) from dist/ copying
			if isFrameworkSourceFile(path) {
				return nil
			}

			// Smart copy static asset files as-is with exact tree hierarchy
			err = copyFileIfNewer(path, outPath)
			if err != nil {
				return fmt.Errorf("error copying file %s to %s: %v", path, outPath, err)
			}

			// If it's a php file, check sitemap inclusion
			if strings.HasSuffix(info.Name(), ".php") {
				pageContent, err := os.ReadFile(path)
				if err == nil {
					includeInSitemap := false
					matches := sitemapOverrideRegex.FindStringSubmatch(string(pageContent))
					if len(matches) > 1 {
						includeInSitemap = strings.ToLower(matches[1]) == "true"
					} else {
						// Fallback to default
						defaultMode := strings.ToLower(config.Default)
						if defaultMode == "" || defaultMode == "public" {
							includeInSitemap = true
						}
					}

					if includeInSitemap {
						rel, err := filepath.Rel(srcDir, path)
						if err == nil {
							cleanPath := strings.ReplaceAll(rel, "\\", "/")
							cleanPath = strings.TrimSuffix(cleanPath, ".php")
							if cleanPath == "index" {
								cleanPath = ""
							} else if strings.HasSuffix(cleanPath, "/index") {
								cleanPath = strings.TrimSuffix(cleanPath, "/index")
							}
							if !sitemapAdded[cleanPath] {
								sitemapAdded[cleanPath] = true
								sitemapURLs = append(sitemapURLs, cleanPath)
							}
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Build Error: %v\n", err)
		return
	}

	// Copy all static assets from public/ to outDir
	publicDir := filepath.Join(projectDir, "public")
	if projectFileExists(publicDir) {
		_ = filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || path == publicDir {
				return err
			}
			relPath, err := filepath.Rel(publicDir, path)
			if err != nil {
				return err
			}
			destPath := filepath.Join(targetOutDir, relPath)
			if info.IsDir() {
				return os.MkdirAll(destPath, 0755)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(destPath, data, 0644)
		})
	}

	targetCssOut := cssOut
	if strings.TrimSpace(config.CssOut) == "" {
		targetCssOut = filepath.Join(targetOutDir, "assets", "css", "styles.css")
	}

	targetJsOut := jsOut
	if strings.TrimSpace(config.JsOut) == "" {
		targetJsOut = filepath.Join(targetOutDir, "assets", "js", "bundle.js")
	}

	if activeMode == "dev" {
		targetCssOut = filepath.Join(targetOutDir, "css", "styles.css")
		targetJsOut = filepath.Join(targetOutDir, "js", "bundle.js")
	}

	// Write global CSS bundle ONLY if non-empty CSS content exists
	if strings.TrimSpace(globalCSS.String()) != "" {
		os.MkdirAll(filepath.Dir(targetCssOut), 0755)
		_ = os.WriteFile(targetCssOut, []byte(globalCSS.String()), 0644)
	}

	// Inject lightweight PPHLX Islands hydration runtime
	runtimeScript := `
// PPHLX Islands Runtime
document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll(".pphlx-island").forEach(island => {
    const compName = island.getAttribute("data-component");
    const framework = island.getAttribute("data-framework") || "";
    const islandId = island.id;
    const props = window.pphlxProps ? window.pphlxProps[islandId] : {};
    
    let ComponentModule = window[compName];
    if (!ComponentModule && window.PphlxViteComponents && window.PphlxViteComponents[compName]) {
      ComponentModule = window.PphlxViteComponents[compName];
    }
    if (ComponentModule) {
      const Component = ComponentModule.default || ComponentModule;
      
      if (framework === "react" && window.ReactDOM && window.ReactDOM.createRoot) {
        // React 18+ Mount
        const root = window.ReactDOM.createRoot(island);
        root.render(window.React.createElement(Component, props));
      } else if (framework === "vue" && window.Vue && window.Vue.createApp) {
        // Vue 3 Mount
        window.Vue.createApp(Component, props).mount(island);
      } else if (framework === "svelte") {
        // Support Svelte 4 classes and Svelte 5 functions
        if (Component.prototype && Component.prototype.$destroy) {
          new Component({ target: island, props: props });
        } else if (window.Svelte && window.Svelte.mount) {
          window.Svelte.mount(Component, { target: island, props: props });
        } else if (typeof Component === "function") {
          Component(island, props);
        }
      } else if (framework === "solid" && window.SolidJS && window.SolidJS.render) {
        // SolidJS Mount
        window.SolidJS.render(() => Component(props), island);
      } else if (framework === "preact" && window.preact && window.preact.render) {
        // Preact Mount
        window.preact.render(window.preact.h(Component, props), island);
      } else if (Component.render) {
        Component.render(island, props);
      } else {
        // Backwards compatibility fallback if data-framework not set
        if (window.ReactDOM && window.ReactDOM.createRoot) {
          const root = window.ReactDOM.createRoot(island);
          root.render(window.React.createElement(Component, props));
        } else if (window.Vue && window.Vue.createApp) {
          window.Vue.createApp(Component, props).mount(island);
        } else if (typeof Component === "function") {
          if (Component.prototype && Component.prototype.$destroy) {
            new Component({ target: island, props: props });
          } else if (window.Svelte && window.Svelte.mount) {
            window.Svelte.mount(Component, { target: island, props: props });
          } else {
            Component(island, props);
          }
        } else {
          console.warn("No runtime renderer found for component " + compName);
        }
      }
    } else {
      console.error("Component " + compName + " not found in window scope.");
    }
  });
});
`
	globalJS.WriteString(runtimeScript + "\n")

	if globalJS.Len() > 0 {
		os.MkdirAll(filepath.Dir(targetJsOut), 0755)
		os.WriteFile(targetJsOut, []byte(globalJS.String()), 0644)
	}

	// Trigger local Vite compilation if Svelte/Vue/Angular components are found
	if len(viteComponents) > 0 {
		err = runViteBuild(config, projectDir)
		if err != nil {
			fmt.Printf("Vite Build Error: %v\n", err)
			return
		}
		// Append compiled Vite bundle into targetJsOut so all Vue/Svelte/Solid components are loaded
		viteBundlePath := filepath.Join(projectDir, config.OutDir, "assets", "js", "pphlx_vite.js")
		if config.OutDir == "" {
			viteBundlePath = filepath.Join(projectDir, "dist", "assets", "js", "pphlx_vite.js")
		}
		if viteJS, errRead := os.ReadFile(viteBundlePath); errRead == nil {
			existingJS, _ := os.ReadFile(targetJsOut)
			mergedJS := string(existingJS) + "\n" + string(viteJS) + "\n"
			_ = os.WriteFile(targetJsOut, []byte(mergedJS), 0644)
		}
	}

	// Generate Partytown files natively if script tags were found
	if hasPartytown {
		partytownDir := filepath.Join(outDir, "~partytown")
		os.MkdirAll(partytownDir, 0755)

		partytownJS := `/* Partytown core runtime stub */
(function() {
  console.log("Natively initialized Partytown Service Worker loader...");
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/~partytown/partytown-sw.js', { scope: '/' });
  }
})();`
		partytownSW := `/* Partytown Service Worker stub */
self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', () => self.clients.claim());
self.addEventListener('fetch', (event) => {
  // Proxy partytown fetch event handler
});`
		os.WriteFile(filepath.Join(partytownDir, "partytown.js"), []byte(partytownJS), 0644)
		os.WriteFile(filepath.Join(partytownDir, "partytown-sw.js"), []byte(partytownSW), 0644)
		fmt.Println("Generated native Partytown scripts in ~partytown/")
	}

	// Generate Sitemap natively if enabled in configuration
	if config.Sitemap {
		var sitemapBuilder strings.Builder
		sitemapBuilder.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
		sitemapBuilder.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")

		siteURL := config.Site
		if siteURL == "" {
			siteURL = "http://localhost/"
		}
		if !strings.HasSuffix(siteURL, "/") {
			siteURL += "/"
		}

		for _, urlPath := range sitemapURLs {
			sitemapBuilder.WriteString(fmt.Sprintf("  <url>\n    <loc>%s%s</loc>\n  </url>\n", siteURL, urlPath))
		}

		sitemapBuilder.WriteString("</urlset>\n")
		sitemapPath := filepath.Join(outDir, "sitemap.xml")
		os.WriteFile(sitemapPath, []byte(sitemapBuilder.String()), 0644)
		fmt.Println("Generated sitemap.xml natively.")
	}

	pruneEmptyDirs(targetOutDir)

	fmt.Printf("✓ Built in %s\n", formatDuration(time.Since(startTime)))

	if targetLower == "standalone" && activeMode == "build" {
		fmt.Println("Compiling Standalone Go Binary...")
		standaloneFile := filepath.Join(projectDir, "standalone_main.go")

		goos := config.Output.Goos
		if goos == "" {
			goos = runtime.GOOS
		}
		binaryName := "app"
		if goos == "windows" {
			binaryName = "app.exe"
		}
		binaryPath := filepath.Join(outDir, binaryName)

		standaloneSource := fmt.Sprintf(`package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

//go:embed %s/*
var embedFS embed.FS

func main() {
	subFS, err := fs.Sub(embedFS, "%s")
	if err != nil {
		fmt.Printf("Error accessing embedded folder: %%v\n", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("PPHLX Standalone Go server starting on http://localhost:%%s\n", port)
	fileServer := http.FileServer(http.FS(subFS))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := r.URL.Path
		if reqPath == "/" {
			reqPath = "/index.php"
		}

		if strings.HasSuffix(reqPath, ".php") {
			cleanPath := strings.TrimPrefix(reqPath, "/")
			data, readErr := fs.ReadFile(subFS, cleanPath)
			if readErr == nil {
				tmpFile, tmpErr := os.CreateTemp("", "pphlx-standalone-*.php")
				if tmpErr == nil {
					tmpPath := tmpFile.Name()
					_, _ = tmpFile.Write(data)
					_ = tmpFile.Close()
					defer os.Remove(tmpPath)

					cmd := exec.Command("php", tmpPath)
					out, execErr := cmd.Output()
					if execErr == nil {
						w.Header().Set("Content-Type", "text/html; charset=utf-8")
						_, _ = w.Write(out)
						return
					}
				}
			}
		}

		fileServer.ServeHTTP(w, r)
	})

	err = http.ListenAndServe(":"+port, handler)
	if err != nil {
		fmt.Printf("Server error: %%v\n", err)
	}
}
`, config.OutDir, config.OutDir)

		err = os.WriteFile(standaloneFile, []byte(standaloneSource), 0644)
		if err == nil {
			goStart := time.Now()
			cmd := exec.Command("go", "build", "-o", binaryPath, standaloneFile)
			cmd.Dir = projectDir

			// Inject GOOS and GOARCH environment variables if configured
			if config.Output.Goos != "" || config.Output.Goarch != "" {
				cmd.Env = os.Environ()
				if config.Output.Goos != "" {
					cmd.Env = append(cmd.Env, "GOOS="+config.Output.Goos)
				}
				if config.Output.Goarch != "" {
					cmd.Env = append(cmd.Env, "GOARCH="+config.Output.Goarch)
				}
			}

			output, cmdErr := cmd.CombinedOutput()
			os.Remove(standaloneFile)
			goElapsed := time.Since(goStart).Round(time.Millisecond)
			if cmdErr != nil {
				fmt.Printf("[Error] Failed to compile Standalone Go Binary: %v\nOutput: %s\n", cmdErr, string(output))
			} else {
				fmt.Printf("✓ Standalone Go Binary compiled successfully in %v: %s (Target OS: %s, Arch: %s)\n", goElapsed, binaryPath, goos, config.Output.Goarch)
			}
		} else {
			fmt.Printf("[Error] Failed to write standalone build file: %v\n", err)
		}
	}

	if targetLower == "desktop" && activeMode == "build" {
		fmt.Println("Compiling Desktop Native Application...")
		desktopFile := filepath.Join(projectDir, "desktop_main.go")

		goos := config.Output.Goos
		if goos == "" {
			goos = runtime.GOOS
		}
		binaryName := "app"
		if goos == "windows" {
			binaryName = "app.exe"
		}
		binaryPath := filepath.Join(outDir, binaryName)

		var desktopSource string
		if goos == "windows" {
			desktopSource = fmt.Sprintf(`package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"github.com/jchv/go-webview2"
)

//go:embed %s/*
var embedFS embed.FS

type DesktopWindow interface {
	Bind(name string, f interface{}) error
}

var extensionRegistrators []func(DesktopWindow)

func RegisterExtension(reg func(DesktopWindow)) {
	extensionRegistrators = append(extensionRegistrators, reg)
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("Error starting server: %%v\n", err)
		os.Exit(1)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	subFS, err := fs.Sub(embedFS, "%s")
	if err != nil {
		os.Exit(1)
	}

	go http.Serve(listener, http.FileServer(http.FS(subFS)))

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     false,
		WindowOptions: webview2.WindowOptions{
			Title:  "PPHLX Desktop Application",
			Width:  1024,
			Height: 768,
		},
	})
	if w == nil {
		os.Exit(1)
	}
	defer w.Destroy()

	// Bind Core Standard Drivers
	w.Bind("pphlxDesktopOpenFile", func() string {
		cmd := exec.Command("powershell", "-NoProfile", "-Command", "Add-Type -AssemblyName System.Windows.Forms; $d = New-Object System.Windows.Forms.OpenFileDialog; if ($d.ShowDialog() -eq 'OK') { $d.FileName }")
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	})

	w.Bind("pphlxDesktopSaveFile", func() string {
		cmd := exec.Command("powershell", "-NoProfile", "-Command", "Add-Type -AssemblyName System.Windows.Forms; $d = New-Object System.Windows.Forms.SaveFileDialog; if ($d.ShowDialog() -eq 'OK') { $d.FileName }")
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	})

	w.Bind("pphlxDesktopShowNotification", func(title, message string) {
		escapedTitle := strings.ReplaceAll(title, "'", "''")
		escapedMsg := strings.ReplaceAll(message, "'", "''")
		cmd := exec.Command("powershell", "-NoProfile", "-Command", "Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show('" + escapedMsg + "', '" + escapedTitle + "')")
		cmd.Run()
	})

	w.Bind("pphlxDesktopClose", func() {
		w.Terminate()
	})

	// Run all custom developer-defined extensions
	for _, reg := range extensionRegistrators {
		reg(w)
	}

	w.Navigate(fmt.Sprintf("http://127.0.0.1:%%d", port))
	w.Run()
}
`, config.OutDir, config.OutDir)
		} else {
			desktopSource = fmt.Sprintf(`package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"github.com/webview/webview_go"
)

//go:embed %s/*
var embedFS embed.FS

type DesktopWindow interface {
	Bind(name string, f interface{}) error
}

var extensionRegistrators []func(DesktopWindow)

func RegisterExtension(reg func(DesktopWindow)) {
	extensionRegistrators = append(extensionRegistrators, reg)
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("Error starting server: %%v\n", err)
		os.Exit(1)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	subFS, err := fs.Sub(embedFS, "%s")
	if err != nil {
		os.Exit(1)
	}

	go http.Serve(listener, http.FileServer(http.FS(subFS)))

	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("PPHLX Desktop Application")
	w.SetSize(1024, 768, webview.HintNone)

	// Bind Core Standard Drivers
	w.Bind("pphlxDesktopOpenFile", func() string {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("powershell", "-NoProfile", "-Command", "Add-Type -AssemblyName System.Windows.Forms; $d = New-Object System.Windows.Forms.OpenFileDialog; if ($d.ShowDialog() -eq 'OK') { $d.FileName }")
		case "darwin":
			cmd = exec.Command("osascript", "-e", "POSIX path of (choose file)")
		default:
			cmd = exec.Command("zenity", "--file-selection")
		}
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	})

	w.Bind("pphlxDesktopSaveFile", func() string {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("powershell", "-NoProfile", "-Command", "Add-Type -AssemblyName System.Windows.Forms; $d = New-Object System.Windows.Forms.SaveFileDialog; if ($d.ShowDialog() -eq 'OK') { $d.FileName }")
		case "darwin":
			cmd = exec.Command("osascript", "-e", "POSIX path of (choose file name)")
		default:
			cmd = exec.Command("zenity", "--file-selection", "--save")
		}
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	})

	w.Bind("pphlxDesktopShowNotification", func(title, message string) {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			escapedTitle := strings.ReplaceAll(title, "'", "''")
			escapedMsg := strings.ReplaceAll(message, "'", "''")
			cmd = exec.Command("powershell", "-NoProfile", "-Command", "Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show('" + escapedMsg + "', '" + escapedTitle + "')")
		case "darwin":
			cmd = exec.Command("osascript", "-e", fmt.Sprintf("display notification \"%%s\" with title \"%%s\"", message, title))
		default:
			cmd = exec.Command("notify-send", title, message)
		}
		cmd.Run()
	})

	w.Bind("pphlxDesktopClose", func() {
		w.Terminate()
	})

	// Run all custom developer-defined extensions
	for _, reg := range extensionRegistrators {
		reg(w)
	}

	w.Navigate(fmt.Sprintf("http://127.0.0.1:%%d", port))
	w.Run()
}
`, config.OutDir, config.OutDir)
		}

		// Check for custom bridge files in src/desktop/
		var copiedBridges []string
		desktopSrcDir := filepath.Join(projectDir, config.SrcDir, "desktop")
		if _, err := os.Stat(desktopSrcDir); err == nil {
			files, _ := os.ReadDir(desktopSrcDir)
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".go") {
					srcFilePath := filepath.Join(desktopSrcDir, file.Name())
					dstFilePath := filepath.Join(projectDir, "temp_bridge_"+file.Name())
					input, err := os.ReadFile(srcFilePath)
					if err == nil {
						err = os.WriteFile(dstFilePath, input, 0644)
						if err == nil {
							copiedBridges = append(copiedBridges, dstFilePath)
						}
					}
				}
			}
		}

		err = os.WriteFile(desktopFile, []byte(desktopSource), 0644)
		if err == nil {
			buildFiles := []string{desktopFile}
			buildFiles = append(buildFiles, copiedBridges...)

			args := []string{"build", "-o", binaryPath}
			args = append(args, buildFiles...)

			cmd := exec.Command("go", args...)
			cmd.Dir = projectDir

			if config.Output.Goos != "" || config.Output.Goarch != "" {
				cmd.Env = os.Environ()
				if config.Output.Goos != "" {
					cmd.Env = append(cmd.Env, "GOOS="+config.Output.Goos)
				}
				if config.Output.Goarch != "" {
					cmd.Env = append(cmd.Env, "GOARCH="+config.Output.Goarch)
				}
			}

			output, cmdErr := cmd.CombinedOutput()

			// Clean up all temporary files
			os.Remove(desktopFile)
			for _, bridge := range copiedBridges {
				os.Remove(bridge)
			}

			if cmdErr != nil {
				fmt.Printf("[Error] Failed to compile Desktop Native Application: %v\nOutput: %s\n", cmdErr, string(output))
			} else {
				fmt.Printf("✓ Desktop Native Application compiled successfully: %s (Target OS: %s, Arch: %s)\n", binaryPath, goos, config.Output.Goarch)
			}
		} else {
			fmt.Printf("[Error] Failed to write desktop build file: %v\n", err)
		}
	}

	if targetLower == "android" && activeMode == "build" {
		fmt.Println("Scaffolding Android Mobile Project...")
		androidDir := filepath.Join(outDir, "android")
		os.MkdirAll(filepath.Join(androidDir, "app/src/main/java/org/pphlx/app"), 0755)

		// Write MainActivity.java
		javaSource := `package org.pphlx.app;

import android.app.Activity;
import android.os.Bundle;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

public class MainActivity extends Activity {
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        WebView webView = new WebView(this);
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        webView.setWebViewClient(new WebViewClient());
        webView.loadUrl("file:///android_asset/index.html");
        setContentView(webView);
    }
}`
		os.WriteFile(filepath.Join(androidDir, "app/src/main/java/org/pphlx/app/MainActivity.java"), []byte(javaSource), 0644)

		// Write AndroidManifest.xml
		manifestSource := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="org.pphlx.app">
    <uses-permission android:name="android.permission.INTERNET" />
    <application
        android:allowBackup="true"
        android:label="PPHLX Mobile App"
        android:theme="@android:style/Theme.NoTitleBar.Fullscreen">
        <activity
            android:name=".MainActivity"
            android:exported="true">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>
    </application>
</manifest>`
		os.WriteFile(filepath.Join(androidDir, "app/src/main/AndroidManifest.xml"), []byte(manifestSource), 0644)

		// Write build.gradle files
		gradleRootSource := `// Top-level build file
buildscript {
    repositories {
        google()
        mavenCentral()
    }
    dependencies {
        classpath 'com.android.tools.build:gradle:8.2.2'
    }
}
allprojects {
    repositories {
        google()
        mavenCentral()
    }
}`
		os.WriteFile(filepath.Join(androidDir, "build.gradle"), []byte(gradleRootSource), 0644)

		gradleAppSource := `plugins {
    id 'com.android.application'
}
android {
    namespace 'org.pphlx.app'
    compileSdk 34
    defaultConfig {
        applicationId "org.pphlx.app"
        minSdk 21
        targetSdk 34
        versionCode 1
        versionName "1.0"
    }
}`
		os.WriteFile(filepath.Join(androidDir, "app/build.gradle"), []byte(gradleAppSource), 0644)

		// Copy web assets into Android assets
		assetsDir := filepath.Join(androidDir, "app/src/main/assets")
		os.MkdirAll(assetsDir, 0755)

		filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path == androidDir || strings.HasPrefix(path, androidDir+string(filepath.Separator)) {
				return nil
			}
			rel, err := filepath.Rel(outDir, path)
			if err == nil && rel != "." {
				dst := filepath.Join(assetsDir, rel)
				if info.IsDir() {
					os.MkdirAll(dst, 0755)
				} else {
					copyFileIfNewer(path, dst)
				}
			}
			return nil
		})
		fmt.Printf("✓ Android Mobile Project scaffolded successfully: %s\n", androidDir)
	}

	if targetLower == "ios" && activeMode == "build" {
		fmt.Println("Scaffolding iOS Xcode Project...")
		iosDir := filepath.Join(outDir, "ios")
		os.MkdirAll(filepath.Join(iosDir, "PPHLXApp"), 0755)

		// Write ViewController.swift
		swiftSource := `import UIKit
import WebKit

class ViewController: UIViewController, WKUIDelegate {
    var webView: WKWebView!
    
    override func loadView() {
        let webConfiguration = WKWebViewConfiguration()
        webView = WKWebView(frame: .zero, configuration: webConfiguration)
        webView.uiDelegate = self
        view = webView
    }
    
    override func viewDidLoad() {
        super.viewDidLoad()
        if let indexURL = Bundle.main.url(forResource: "index", withExtension: "html", subdirectory: "www") {
            webView.loadFileURL(indexURL, allowingReadAccessTo: indexURL.deletingLastPathComponent())
        }
    }
}`
		os.WriteFile(filepath.Join(iosDir, "PPHLXApp/ViewController.swift"), []byte(swiftSource), 0644)

		// Write AppDelegate.swift
		appDelegateSource := `import UIKit

@main
class AppDelegate: UIResponder, UIApplicationDelegate {
    var window: UIWindow?
    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        window = UIWindow(frame: UIScreen.main.bounds)
        let viewController = ViewController()
        window?.rootViewController = viewController
        window?.makeKeyAndVisible()
        return true
    }
}`
		os.WriteFile(filepath.Join(iosDir, "PPHLXApp/AppDelegate.swift"), []byte(appDelegateSource), 0644)

		// Copy web assets into iOS www folder
		wwwDir := filepath.Join(iosDir, "PPHLXApp/www")
		os.MkdirAll(wwwDir, 0755)

		filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path == iosDir || strings.HasPrefix(path, iosDir+string(filepath.Separator)) {
				return nil
			}
			rel, err := filepath.Rel(outDir, path)
			if err == nil && rel != "." {
				dst := filepath.Join(wwwDir, rel)
				if info.IsDir() {
					os.MkdirAll(dst, 0755)
				} else {
					copyFileIfNewer(path, dst)
				}
			}
			return nil
		})
		fmt.Printf("✓ iOS Xcode Project scaffolded successfully: %s\n", iosDir)
	}

	totalElapsed := time.Since(startTime).Round(time.Millisecond)
	fmt.Printf("[%s] Build complete successfully in %v!\n", time.Now().Format("15:04:05"), totalElapsed)
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

func isPortAvailable(port int) bool {
	// 1. Check IPv4 loopback
	ln1, err1 := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err1 != nil {
		return false
	}
	ln1.Close()

	// 2. Check IPv6 loopback
	ln2, err2 := net.Listen("tcp6", fmt.Sprintf("[::1]:%d", port))
	if err2 != nil {
		return false
	}
	ln2.Close()

	// 3. Check wildcard address
	ln3, err3 := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err3 != nil {
		return false
	}
	ln3.Close()

	return true
}

func findFreePort(startPort int) int {
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return startPort
}

func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}

// getInMemoryGlobalCSS compiles global CSS rules in memory
func getInMemoryGlobalCSS(config Config, projectDir string) string {
	var cssBuilder strings.Builder
	srcRootDir, _ := resolveSourceRootAndEntry(config, projectDir)
	_ = filepath.Walk(srcRootDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".pphx") {
			content, err := os.ReadFile(path)
			if err == nil {
				_, css, _, _ := compilePage(string(content), filepath.Dir(path), srcRootDir)
				for _, style := range css {
					cleanStyle := strings.TrimSpace(style)
					if cleanStyle != "" {
						cssBuilder.WriteString(cleanStyle + "\n")
					}
				}
			}
		}
		return nil
	})
	return cssBuilder.String()
}

// getInMemoryGlobalJS returns the lightweight PPHLX Islands hydration script directly from RAM
func getInMemoryGlobalJS() string {
	return `
// PPHLX Islands Runtime
(function() {
  function initPphlxIslands() {
    document.querySelectorAll(".pphlx-island").forEach(island => {
      const compName = island.getAttribute("data-component");
      const framework = island.getAttribute("data-framework") || "";
      const islandId = island.id;
      const props = window.pphlxProps ? window.pphlxProps[islandId] : {};
      
      let ComponentModule = window[compName];
      if (!ComponentModule && window.PphlxViteComponents && window.PphlxViteComponents[compName]) {
        ComponentModule = window.PphlxViteComponents[compName];
      }
      if (!ComponentModule && typeof window !== "undefined") {
        try { ComponentModule = eval(compName); } catch(e) {}
      }
      
      if (ComponentModule) {
        const Component = ComponentModule.default || ComponentModule;
        if (framework === "react" && window.ReactDOM && window.ReactDOM.createRoot) {
          const root = window.ReactDOM.createRoot(island);
          root.render(window.React.createElement(Component, props));
        } else if (framework === "vue" && window.Vue && window.Vue.createApp) {
          window.Vue.createApp(Component, props).mount(island);
        } else if (framework === "svelte") {
          if (Component.prototype && Component.prototype.$destroy) {
            new Component({ target: island, props: props });
          } else if (window.Svelte && window.Svelte.mount) {
            window.Svelte.mount(Component, { target: island, props: props });
          } else if (typeof Component === "function") {
            Component(island, props);
          }
        } else if (framework === "solid" && window.SolidJS && window.SolidJS.render) {
          window.SolidJS.render(() => Component(props), island);
        } else if (framework === "preact" && window.preact && window.preact.render) {
          window.preact.render(window.preact.h(Component, props), island);
        } else if (Component.render) {
          Component.render(island, props);
        }
      }
    });
  }

  if (document.readyState === "complete" || document.readyState === "interactive") {
    initPphlxIslands();
  } else {
    document.addEventListener("DOMContentLoaded", initPphlxIslands);
  }
})();
`
}

type LogEvent struct {
	Timestamp string
	Status    int
	Path      string
	Duration  time.Duration
	Category  string
}

var logChan = make(chan LogEvent, 1024)
var logWorkerOnce sync.Once

func startLogWorker() {
	logWorkerOnce.Do(func() {
		go func() {
			for ev := range logChan {
				var statusStr string
				switch {
				case ev.Status >= 200 && ev.Status < 300:
					statusStr = fmt.Sprintf("\033[32m[%d]\033[0m", ev.Status)
				case ev.Status >= 300 && ev.Status < 400:
					statusStr = fmt.Sprintf("\033[33m[%d]\033[0m", ev.Status)
				case ev.Status >= 400 && ev.Status < 500:
					statusStr = fmt.Sprintf("\033[1;33m[%d]\033[0m", ev.Status)
				default:
					statusStr = fmt.Sprintf("\033[1;31m[%d]\033[0m", ev.Status)
				}

				var durStr string
				if ev.Duration < time.Millisecond {
					durStr = fmt.Sprintf("%dµs", ev.Duration.Microseconds())
				} else {
					durStr = fmt.Sprintf("%dms", ev.Duration.Milliseconds())
				}

				catStr := ""
				if ev.Category != "" {
					catStr = fmt.Sprintf(" \033[90m(%s)\033[0m", ev.Category)
				}

				fmt.Printf("%s %s %s %s%s\n", ev.Timestamp, statusStr, ev.Path, durStr, catStr)
			}
		}()
	})
}

type devResponseLogger struct {
	http.ResponseWriter
	statusCode int
}

func (rl *devResponseLogger) WriteHeader(code int) {
	rl.statusCode = code
	rl.ResponseWriter.WriteHeader(code)
}

var devLoggerPool = sync.Pool{
	New: func() interface{} {
		return &devResponseLogger{statusCode: http.StatusOK}
	},
}

func startDevServerAndWatcher(config Config, projectDir string, mode string) {
	startLogWorker()
	serverStartTime := time.Now()
	hasHost := false
	for _, arg := range os.Args {
		if arg == "--host" || arg == "-h" || strings.HasPrefix(arg, "--host=") {
			hasHost = true
			break
		}
	}

	port := findFreePort(6321)
	srcRootDir, entryFile := resolveSourceRootAndEntry(config, projectDir)

	outDir := filepath.Join(projectDir, config.OutDir)
	if config.OutDir == "" {
		outDir = filepath.Join(projectDir, "dist")
	}

	hostAddr := "127.0.0.1"
	if hasHost {
		hostAddr = "0.0.0.0"
	}

	if mode == "preview" {
		currentPort := port
		readyElapsedMs := time.Since(serverStartTime).Milliseconds()

		fmt.Printf("\n  \x1b[30m\x1b[42m pphlx \x1b[0m \x1b[32mv%s\x1b[0m \x1b[90mpreview ready in\x1b[0m \x1b[37m%d\x1b[0m \x1b[90mms\x1b[0m\n\n", Version, readyElapsedMs)
		fmt.Printf("  \x1b[90m┃\x1b[0m \x1b[1mLocal\x1b[0m    \x1b[36mhttp://localhost:%d/\x1b[0m\n", currentPort)
		if hasHost {
			ips := getLocalIPs()
			for _, ip := range ips {
				fmt.Printf("  \x1b[90m┃\x1b[0m \x1b[1mNetwork\x1b[0m  \x1b[36mhttp://%s:%d/\x1b[0m\n", ip, currentPort)
			}
		}
		fmt.Printf("  \x1b[90m┃\x1b[0m \x1b[90mServing directory: %s\x1b[0m\n\n", outDir)

		fs := http.FileServer(http.Dir(outDir))
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rl := devLoggerPool.Get().(*devResponseLogger)
			rl.ResponseWriter = w
			rl.statusCode = http.StatusOK

			reqPath := filepath.Clean(r.URL.Path)
			targetFile := filepath.Join(outDir, reqPath)

			if reqPath == "/" || reqPath == "\\" || reqPath == "." {
				targetFile = filepath.Join(outDir, "index.php")
				if _, err := os.Stat(targetFile); os.IsNotExist(err) {
					targetFile = filepath.Join(outDir, "index.html")
				}
			} else if fi, err := os.Stat(targetFile); err != nil || fi.IsDir() {
				if _, err := os.Stat(targetFile + ".php"); err == nil {
					targetFile = targetFile + ".php"
				} else if _, err := os.Stat(targetFile + ".html"); err == nil {
					targetFile = targetFile + ".html"
				} else if _, err := os.Stat(filepath.Join(targetFile, "index.php")); err == nil {
					targetFile = filepath.Join(targetFile, "index.php")
				} else if _, err := os.Stat(filepath.Join(targetFile, "index.html")); err == nil {
					targetFile = filepath.Join(targetFile, "index.html")
				}
			}

			if strings.HasSuffix(targetFile, ".php") {
				if phpContent, err := os.ReadFile(targetFile); err == nil {
					finalHtml := string(phpContent)
					_, phpErr := exec.LookPath("php")
					if phpErr == nil {
						cmd := exec.Command("php", "-r", "eval('?>'.file_get_contents('php://stdin'));")
						cmd.Stdin = strings.NewReader(finalHtml)
						outBytes, cmdErr := cmd.CombinedOutput()
						if cmdErr == nil {
							finalHtml = string(outBytes)
						} else {
							fmt.Printf("\033[1;31m[PHP Preview Server Evaluation Error]\033[0m %v\nOutput: %s\n", cmdErr, string(outBytes))
						}
					}
					if strings.Contains(finalHtml, "<?php") {
						phpCodeBlockRegex := regexp.MustCompile(`(?s)<\?php.*?\?>`)
						finalHtml = phpCodeBlockRegex.ReplaceAllString(finalHtml, "")
					}
					rl.Header().Set("Content-Type", "text/html; charset=utf-8")
					rl.Write([]byte(finalHtml))
				} else {
					fs.ServeHTTP(rl, r)
				}
			} else if strings.HasSuffix(targetFile, ".html") {
				rl.Header().Set("Content-Type", "text/html; charset=utf-8")
				http.ServeFile(rl, r, targetFile)
			} else {
				fs.ServeHTTP(rl, r)
			}

			dur := time.Since(start)
			category := "page"
			if strings.HasSuffix(reqPath, ".css") || strings.HasSuffix(reqPath, ".js") || strings.HasSuffix(reqPath, ".webp") || strings.HasSuffix(reqPath, ".png") || strings.HasSuffix(reqPath, ".ico") || strings.HasSuffix(reqPath, ".svg") {
				category = "asset"
			}
			if rl.statusCode >= 400 {
				category = "missing"
			}
			select {
			case logChan <- LogEvent{
				Timestamp: time.Now().Format("15:04:05"),
				Status:    rl.statusCode,
				Path:      r.URL.Path,
				Duration:  dur,
				Category:  category,
			}:
			default:
			}

			devLoggerPool.Put(rl)
		})
		_ = http.ListenAndServe(fmt.Sprintf("%s:%d", hostAddr, currentPort), handler)
		return
	} else {
		// Pure 100% In-Memory Go HTTP Server Engine (Zero Disk Writes)
		currentPort := port
		var (
			devCacheMutex sync.RWMutex
			devCSSCache   = make(map[string]bool)
			devJSCache    = make(map[string]bool)
			devCSSBuffer  strings.Builder
			devJSBuffer   strings.Builder
		)

		// Build component dependency graph for dev mode direct request protections
		ignorePatterns := loadPphlxIgnore(projectDir)
		importedAsComponent := buildDependencyGraph(srcRootDir, ignorePatterns)

		// Pre-compile entry page once at startup to populate initial in-memory dev buffers
		if pageBytes, err := os.ReadFile(entryFile); err == nil {
			_, css, js, _ := compilePage(string(pageBytes), filepath.Dir(entryFile), srcRootDir)
			for _, style := range css {
				cleanStyle := strings.TrimSpace(style)
				if cleanStyle != "" && !devCSSCache[cleanStyle] {
					devCSSCache[cleanStyle] = true
					devCSSBuffer.WriteString(cleanStyle)
					devCSSBuffer.WriteByte('\n')
				}
			}
			for _, script := range js {
				cleanScript := strings.TrimSpace(script)
				if cleanScript != "" && !devJSCache[cleanScript] {
					devJSCache[cleanScript] = true
					devJSBuffer.WriteString(cleanScript)
					devJSBuffer.WriteByte('\n')
				}
			}
		}

		// Non-blocking background worker for Vue/Svelte/SolidJS Vite components
		if len(viteComponents) > 0 {
			go runViteBuild(config, projectDir)
		}

		go func() {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				rl := devLoggerPool.Get().(*devResponseLogger)
				rl.ResponseWriter = w
				rl.statusCode = http.StatusOK
				category := "page"

				defer func() {
					dur := time.Since(start)
					if rl.statusCode >= 400 {
						category = "missing"
					}
					select {
					case logChan <- LogEvent{
						Timestamp: time.Now().Format("15:04:05"),
						Status:    rl.statusCode,
						Path:      r.URL.Path,
						Duration:  dur,
						Category:  category,
					}:
					default:
					}
					devLoggerPool.Put(rl)
				}()

				reqPath := filepath.Clean(r.URL.Path)

				// 1. Static asset fallback from public/ (e.g. /assets/pphlx.svg)
				pubAssetPath := filepath.Join(projectDir, "public", reqPath)
				if fi, err := os.Stat(pubAssetPath); err == nil && !fi.IsDir() {
					category = "asset"
					http.ServeFile(rl, r, pubAssetPath)
					return
				}

				// 2. Static asset fallback from src/ (non-.pphx files e.g. /assets/logo.png)
				srcAssetPath := filepath.Join(srcRootDir, reqPath)
				if fi, err := os.Stat(srcAssetPath); err == nil && !fi.IsDir() && !strings.HasSuffix(reqPath, ".pphx") {
					category = "asset"
					http.ServeFile(rl, r, srcAssetPath)
					return
				}

				// Calculate dynamic CSS and JS asset paths from user's pphlx.config.json
				devCssPath := "/css/styles.css"
				if config.CssOut != "" {
					relCss, err := filepath.Rel(outDir, filepath.Join(projectDir, config.CssOut))
					if err == nil {
						devCssPath = "/" + filepath.ToSlash(relCss)
					}
				}

				devJsPath := "/js/bundle.js"
				if config.JsOut != "" {
					relJs, err := filepath.Rel(outDir, filepath.Join(projectDir, config.JsOut))
					if err == nil {
						devJsPath = "/" + filepath.ToSlash(relJs)
					}
				}

				cleanReq := filepath.ToSlash(reqPath)

				// 3. Virtual global CSS & JS bundle endpoints
				if cleanReq == devCssPath || cleanReq == "/css/styles.css" || reqPath == "/css/styles.css" || reqPath == "\\css\\styles.css" {
					category = "virtual"
					rl.Header().Set("Content-Type", "text/css; charset=utf-8")
					rl.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
					devCacheMutex.RLock()
					cssOut := getInMemoryGlobalCSS(config, projectDir) + "\n" + devCSSBuffer.String()
					devCacheMutex.RUnlock()
					rl.Write([]byte(cssOut))
					return
				}
				if cleanReq == devJsPath || cleanReq == "/js/bundle.js" || reqPath == "/js/bundle.js" || reqPath == "\\js\\bundle.js" {
					category = "virtual"
					rl.Header().Set("Content-Type", "application/javascript; charset=utf-8")
					rl.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
					devCacheMutex.RLock()
					jsOut := devJSBuffer.String()
					outDirName := config.OutDir
					if outDirName == "" {
						outDirName = "dist"
					}
					if activeMode == "dev" {
						outDirName = ".pphlx/cache"
					}
					viteBundlePath := filepath.Join(projectDir, outDirName, "assets", "js", "pphlx_vite.js")
					if viteJS, err := os.ReadFile(viteBundlePath); err == nil {
						jsOut += "\n" + string(viteJS) + "\n"
					}
					jsOut += "\n" + getInMemoryGlobalJS() + "\n"
					devCacheMutex.RUnlock()
					rl.Write([]byte(jsOut))
					return
				}

				// 4. Dev Server Direct Request Protection:
				// Return a clean 404 Developer Safeguard Page if a browser directly requests an absorbed component module or framework source file
				targetPphxFile := entryFile
				if reqPath != "/" && reqPath != "\\" && reqPath != "/index.php" && reqPath != "/index.html" {
					candidatePphx := filepath.Join(srcRootDir, reqPath)
					if !strings.HasSuffix(candidatePphx, ".pphx") {
						candidatePphx += ".pphx"
					}

					cleanCand := filepath.Clean(candidatePphx)
					cleanCandSlash := strings.ReplaceAll(cleanCand, "\\", "/")
					rawReqPath := filepath.Clean(filepath.Join(srcRootDir, reqPath))
					rawReqSlash := strings.ReplaceAll(rawReqPath, "\\", "/")

					// Check .pphlxignore patterns in dev server mode
					reqRel, _ := filepath.Rel(srcRootDir, rawReqPath)
					if isPphlxIgnored(reqRel, ignorePatterns) {
						rl.WriteHeader(http.StatusNotFound)
						rl.Header().Set("Content-Type", "text/html; charset=utf-8")
						rl.Write(fmt.Appendf(nil, `
							<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;padding:3rem;background:#0b0d11;color:#f0f6fc;min-height:100vh;box-sizing:border-box;">
								<div style="max-w-2xl:margin:0 auto;background:#13161c;border:1px solid #232731;padding:2rem;border-radius:12px;box-shadow:0 10px 25px rgba(0,0,0,0.5);">
									<span style="background:#e03e3e;color:#fff;font-weight:bold;padding:3px 8px;border-radius:4px;font-size:12px;font-family:monospace;">404 PPHLXIGNORED ROUTE</span>
									<h2 style="margin-top:1rem;color:#ffffff;font-size:22px;">Ignored Route File</h2>
									<p style="color:#9da5b4;font-size:14px;line-height:1.6;">The requested route <code style="color:#4bf3c8;background:#1c2029;padding:2px 6px;border-radius:4px;">%s</code> matches a pattern defined in <code style="color:#4bf3c8;">.pphlxignore</code> and is excluded from compilation and dev server routes.</p>
								</div>
							</div>
						`, reqPath))
						return
					}

					// Check if requested route is an absorbed component module
					if importedAsComponent[cleanCand] || importedAsComponent[cleanCandSlash] || importedAsComponent[rawReqPath] || importedAsComponent[rawReqSlash] {
						rl.WriteHeader(http.StatusNotFound)
						rl.Header().Set("Content-Type", "text/html; charset=utf-8")
						rl.Write(fmt.Appendf(nil, `
							<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;padding:3rem;background:#0b0d11;color:#f0f6fc;min-height:100vh;box-sizing:border-box;">
								<div style="max-w-2xl:margin:0 auto;background:#13161c;border:1px solid #232731;padding:2rem;border-radius:12px;box-shadow:0 10px 25px rgba(0,0,0,0.5);">
									<span style="background:#ff4d4d;color:#fff;font-weight:bold;padding:3px 8px;border-radius:4px;font-size:12px;font-family:monospace;">404 ABSORBED COMPONENT MODULE</span>
									<h2 style="margin-top:1rem;color:#ffffff;font-size:22px;">Direct HTTP Access Prohibited</h2>
									<p style="color:#9da5b4;font-size:14px;line-height:1.6;">The module <code style="color:#4bf3c8;background:#1c2029;padding:2px 6px;border-radius:4px;">%s</code> is an inlined component module absorbed by another page template and cannot be requested directly as a standalone page route.</p>
								</div>
							</div>
						`, reqPath))
						return
					}

					// Check if requested route is an unattached framework source file (.jsx, .vue, .svelte, etc.)
					if isFrameworkSourceFile(rawReqPath) {
						rl.WriteHeader(http.StatusNotFound)
						rl.Header().Set("Content-Type", "text/html; charset=utf-8")
						rl.Write(fmt.Appendf(nil, `
							<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;padding:3rem;background:#0b0d11;color:#f0f6fc;min-height:100vh;box-sizing:border-box;">
								<div style="max-w-2xl:margin:0 auto;background:#13161c;border:1px solid #232731;padding:2rem;border-radius:12px;box-shadow:0 10px 25px rgba(0,0,0,0.5);">
									<span style="background:#ff9800;color:#fff;font-weight:bold;padding:3px 8px;border-radius:4px;font-size:12px;font-family:monospace;">404 UNATTACHED FRAMEWORK SOURCE</span>
									<h2 style="margin-top:1rem;color:#ffffff;font-size:22px;">Framework Source File</h2>
									<p style="color:#9da5b4;font-size:14px;line-height:1.6;">The source file <code style="color:#4bf3c8;background:#1c2029;padding:2px 6px;border-radius:4px;">%s</code> is a UI framework component that must be imported inside a <code style="color:#4bf3c8;">.pphx</code> template page to render.</p>
								</div>
							</div>
						`, reqPath))
						return
					}

					if _, err := os.Stat(candidatePphx); err == nil {
						targetPphxFile = candidatePphx
					} else {
						candidateIndex := filepath.Join(srcRootDir, reqPath, "index.pphx")
						if _, err := os.Stat(candidateIndex); err == nil {
							targetPphxFile = candidateIndex
						}
					}
				}

				pageContent, err := os.ReadFile(targetPphxFile)
				if err != nil {
					rl.WriteHeader(http.StatusNotFound)
					rl.Write([]byte("404 Not Found"))
					return
				}

				compiledPage, css, js, err := compilePage(string(pageContent), filepath.Dir(targetPphxFile), srcRootDir)
				if err != nil {
					rl.WriteHeader(http.StatusInternalServerError)
					rl.Write([]byte(fmt.Sprintf("PPHLX Dev Compile Error: %v", err)))
					return
				}

				devCacheMutex.Lock()
				for _, style := range css {
					cleanStyle := strings.TrimSpace(style)
					if cleanStyle != "" && !devCSSCache[cleanStyle] {
						devCSSCache[cleanStyle] = true
						devCSSBuffer.WriteString(cleanStyle)
						devCSSBuffer.WriteByte('\n')
					}
				}
				for _, script := range js {
					cleanScript := strings.TrimSpace(script)
					if cleanScript != "" && !devJSCache[cleanScript] {
						devJSCache[cleanScript] = true
						devJSBuffer.WriteString(cleanScript)
						devJSBuffer.WriteByte('\n')
					}
				}
				devCacheMutex.Unlock()

				hasDevCss := devCSSBuffer.Len() > 0
				hasDevJs := devJSBuffer.Len() > 0 || len(viteComponents) > 0

				cssTag := fmt.Sprintf(`<link rel="stylesheet" href="%s">`, devCssPath)
				jsTag := fmt.Sprintf(`<script src="%s"></script>`, devJsPath)

				if hasDevCss {
					if strings.Contains(compiledPage, "{{PPHLX_CSS}}") {
						compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_CSS}}", cssTag)
					} else if strings.Contains(compiledPage, "</head>") {
						compiledPage = strings.ReplaceAll(compiledPage, "</head>", cssTag+"\n</head>")
					}
				} else {
					compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_CSS}}", "")
				}

				if hasDevJs {
					if strings.Contains(compiledPage, "{{PPHLX_JS}}") {
						compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_JS}}", jsTag)
					} else if strings.Contains(compiledPage, "</body>") {
						compiledPage = strings.ReplaceAll(compiledPage, "</body>", jsTag+"\n</body>")
					} else if strings.Contains(compiledPage, "</head>") {
						compiledPage = strings.ReplaceAll(compiledPage, "</head>", jsTag+"\n</head>")
					}
				} else {
					compiledPage = strings.ReplaceAll(compiledPage, "{{PPHLX_JS}}", "")
				}

				finalHtml := strings.TrimLeft(compiledPage, " \t\r\n")

				// Optional PHP CLI stream evaluation in RAM if PHP code blocks exist
				_, phpErr := exec.LookPath("php")
				if phpErr == nil && (strings.Contains(finalHtml, "<?php") || strings.Contains(finalHtml, "{|")) {
					cmd := exec.Command("php", "-r", "eval('?>'.file_get_contents('php://stdin'));")
					cmd.Stdin = strings.NewReader(finalHtml)
					outBytes, cmdErr := cmd.CombinedOutput()
					if cmdErr == nil {
						finalHtml = string(outBytes)
					} else {
						fmt.Printf("\033[1;31m[PHP Dev Server Evaluation Error]\033[0m %v\nOutput: %s\n", cmdErr, string(outBytes))
					}
				}

				rl.Header().Set("Content-Type", "text/html; charset=utf-8")
				rl.Write([]byte(finalHtml))
			})

			_ = http.ListenAndServe(fmt.Sprintf("%s:%d", hostAddr, currentPort), handler)
		}()
	}

	version := Version
	readyElapsedMs := time.Since(serverStartTime).Milliseconds()
	if mode == "preview" {
		fmt.Printf("\n  \x1b[30m\x1b[42m pphlx \x1b[0m \x1b[32mv%s\x1b[0m \x1b[90mpreview ready in\x1b[0m \x1b[37m%d\x1b[0m \x1b[90mms\x1b[0m\n\n", version, readyElapsedMs)
		fmt.Printf("  \x1b[90m┃\x1b[0m \x1b[1mLocal\x1b[0m    \x1b[36mhttp://localhost:%d/\x1b[0m\n", port)
		if hasHost {
			ips := getLocalIPs()
			for _, ip := range ips {
				fmt.Printf("  \x1b[90m┃\x1b[0m \x1b[1mNetwork\x1b[0m  \x1b[36mhttp://%s:%d/\x1b[0m\n", ip, port)
			}
		}
		fmt.Printf("  \x1b[90m┃\x1b[0m \x1b[90mServing directory: %s\x1b[0m\n\n", outDir)
	} else {
		fmt.Printf("\n  \x1b[30m\x1b[42m pphlx \x1b[0m \x1b[32mv%s\x1b[0m \x1b[90mready in\x1b[0m \x1b[37m%d\x1b[0m \x1b[90mms\x1b[0m\n\n", version, readyElapsedMs)
		fmt.Printf("  \x1b[90m┃\x1b[0m \x1b[1mLocal\x1b[0m    \x1b[36mhttp://localhost:%d/\x1b[0m\n", port)
		if hasHost {
			ips := getLocalIPs()
			for _, ip := range ips {
				fmt.Printf("  \x1b[90m┃\x1b[0m \x1b[1mNetwork\x1b[0m  \x1b[36mhttp://%s:%d/\x1b[0m\n", ip, port)
			}
			fmt.Println()
		} else {
			fmt.Printf("  \x1b[90m┃\x1b[0m \x1b[90mNetwork  use --host to expose\x1b[0m\n\n")
		}
		startWatcher(config, projectDir)
	}
}

// startWatcher initializes the fsnotify file-watcher loop
func startWatcher(config Config, projectDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	srcRootDir, _ := resolveSourceRootAndEntry(config, projectDir)
	srcDir := srcRootDir
	outDirName := filepath.Clean(config.OutDir)
	fmt.Printf("  watching for file changes...\n\n")

	// Watch recursively by adding subfolders
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirName := info.Name()
			rel, _ := filepath.Rel(projectDir, path)
			if rel != "." && (dirName == "dist" || dirName == outDirName || dirName == "node_modules" || dirName == ".git" || dirName == ".vscode" || dirName == "storage" || dirName == ".antigravity" || strings.HasPrefix(dirName, ".")) {
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
					relPath, _ := filepath.Rel(projectDir, event.Name)
					firstSegment := strings.Split(filepath.ToSlash(relPath), "/")[0]
					if firstSegment == "dist" || firstSegment == outDirName || firstSegment == "node_modules" || firstSegment == ".git" || firstSegment == ".vscode" || firstSegment == "storage" || strings.HasPrefix(firstSegment, ".") {
						continue
					}

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
							if activeMode == "dev" {
								fmt.Printf("[%s] File changed: %s (in-memory updated)\n", time.Now().Format("15:04:05"), baseName)
								if len(viteComponents) > 0 {
									go runViteBuild(config, projectDir)
								}
							} else {
								compileAll(config, projectDir)
							}
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

// CompilePageWithAssets compiles a page and applies autonomous asset tag injection (CSS & JS scripts).
// This is the single source of truth helper used across native CLI builds and Go WASM browser compilation.
func CompilePageWithAssets(content string, currentDir string, srcDir string) (string, []string, []string, error) {
	compiledPHP, css, jsList, err := compilePage(content, currentDir, srcDir)
	if err != nil {
		return compiledPHP, css, jsList, err
	}

	hasCssContent := len(css) > 0
	hasJsContent := len(jsList) > 0 || len(viteComponents) > 0

	hasCssPlaceholder := strings.Contains(compiledPHP, "{{PPHLX_CSS}}")
	hasJsPlaceholder := strings.Contains(compiledPHP, "{{PPHLX_JS}}")

	if hasCssContent {
		cssTag := `<link rel="stylesheet" href="/assets/css/styles.css">`
		if hasCssPlaceholder {
			compiledPHP = strings.ReplaceAll(compiledPHP, "{{PPHLX_CSS}}", cssTag)
		} else if strings.Contains(compiledPHP, "</head>") {
			compiledPHP = strings.ReplaceAll(compiledPHP, "</head>", cssTag+"\n</head>")
		}
	} else if hasCssPlaceholder {
		compiledPHP = strings.ReplaceAll(compiledPHP, "{{PPHLX_CSS}}", "")
	}

	if hasJsContent {
		jsTag := `<script src="/assets/js/bundle.js"></script>`
		if hasJsPlaceholder {
			compiledPHP = strings.ReplaceAll(compiledPHP, "{{PPHLX_JS}}", jsTag)
		} else if strings.Contains(compiledPHP, "</body>") {
			compiledPHP = strings.ReplaceAll(compiledPHP, "</body>", jsTag+"\n</body>")
		} else if strings.Contains(compiledPHP, "</head>") {
			compiledPHP = strings.ReplaceAll(compiledPHP, "</head>", jsTag+"\n</head>")
		}
	} else if hasJsPlaceholder {
		compiledPHP = strings.ReplaceAll(compiledPHP, "{{PPHLX_JS}}", "")
	}

	return compiledPHP, css, jsList, nil
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

		if isPackage {
			compPath = filepath.Clean(filepath.Join(srcDir, "../.pphlx/packages", relPath))
		} else {
			compPath = filepath.Clean(filepath.Join(currentDir, relPath))
			if !projectFileExists(compPath) {
				compPath = filepath.Clean(filepath.Join(srcDir, relPath))
			}
		}
		if ext == ".jsx" || ext == ".tsx" || ext == ".js" || ext == ".vue" || ext == ".svelte" || ext == ".ts" || strings.HasSuffix(relPath, ".solid.jsx") || strings.HasSuffix(relPath, ".solid.tsx") {
			isJsComponent = true
		}

		var compObj Component
		if isJsComponent {
			framework := "react"
			if ext == ".vue" {
				framework = "vue"
			} else if ext == ".svelte" {
				framework = "svelte"
			} else if strings.Contains(compPath, ".solid.") {
				framework = "solid"
			} else {
				fileBytes, err := readProjectFile(compPath)
				if err == nil && (strings.Contains(string(fileBytes), "preact") || strings.Contains(string(fileBytes), "preact.Component")) {
					framework = "preact"
				}
			}

			// Svelte, Vue, SolidJS, TS, and TSX files containing Angular are routed to Vite
			isVite := ext == ".vue" || ext == ".svelte" || strings.HasSuffix(compPath, ".ts") || strings.HasSuffix(compPath, ".tsx") || strings.HasSuffix(compPath, ".vue") || strings.HasSuffix(compPath, ".svelte") || strings.HasSuffix(compPath, ".solid.jsx") || strings.HasSuffix(compPath, ".solid.tsx")

			if isVite {
				viteComponents[compName] = compPath
				compObj = Component{
					Name:          compName,
					Path:          compPath,
					JS:            "",
					IsJsComponent: true,
					Framework:     framework,
				}
			} else {
				// Compile JS/JSX/TSX component with native esbuild
				var jsCode string
				var err error
				if VirtualFiles != nil {
					// In browser WASM, just read raw file as JS code without running esbuild CLI
					fileBytes, readErr := readProjectFile(compPath)
					if readErr == nil {
						jsCode = string(fileBytes)
					}
				} else {
					jsCode, err = compileJSComponent(compName, compPath)
				}
				if err != nil {
					return "", nil, nil, fmt.Errorf("failed to compile JS component %s: %v", compName, err)
				}
				compObj = Component{
					Name:          compName,
					Path:          compPath,
					JS:            jsCode,
					IsJsComponent: true,
					Framework:     framework,
				}
			}
		} else {
			compContent, err := readProjectFile(compPath)
			if err != nil {
				return "", nil, nil, fmt.Errorf("failed to read imported component %s: %v", compName, err)
			}
			// Recursively compile nested template components (e.g. Layout importing Head)
			nestedHtml, nestedCss, nestedJs, compileErr := compilePage(string(compContent), filepath.Dir(compPath), srcDir)
			if compileErr == nil {
				compObj = parseComponent(compName, nestedHtml, compPath)
				cssBundles = append(cssBundles, nestedCss...)
				jsBundles = append(jsBundles, nestedJs...)
			} else {
				processedCompContent := parsePphlxBrackets(string(compContent))
				compObj = parseComponent(compName, processedCompContent, compPath)
			}
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
		EntryPoints:       []string{compPath},
		Bundle:            true,
		Write:             false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Format:            api.FormatIIFE,
		GlobalName:        compName,
		External:          []string{"react", "react-dom", "preact", "preact/hooks", "solid-js"},
		Banner: map[string]string{
			"js": `var require = typeof require !== "undefined" ? require : function(m) { if (m === "react") return window.React; if (m === "react-dom") return window.ReactDOM; if (m === "preact") return window.preact; if (m === "solid-js") return window.SolidJS; return window[m]; };`,
		},
		Loader: map[string]api.Loader{
			".js":  api.LoaderJSX,
			".jsx": api.LoaderJSX,
			".tsx": api.LoaderTSX,
			".ts":  api.LoaderTS,
		},
		JSX: api.JSXTransform,
	})

	if len(result.Errors) > 0 {
		var errs []string
		for _, err := range result.Errors {
			if err.Location != nil {
				errs = append(errs, fmt.Sprintf("%s:%d: %s", err.Location.File, err.Location.Line, err.Text))
			} else {
				errs = append(errs, err.Text)
			}
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
		html = importRegex.ReplaceAllString(html, "")
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

	target := strings.ToLower(activeConfig.Output.Target)
	if target == "blade" {
		bladeName := strings.ToLower(compName)
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
			return fmt.Sprintf("<x-%s%s>%s</x-%s>", bladeName, attrs, slot, bladeName)
		})
		content = selfClosingRegex.ReplaceAllStringFunc(content, func(match string) string {
			submatches := selfClosingRegex.FindStringSubmatch(match)
			attrs := ""
			if len(submatches) > 1 {
				attrs = submatches[1]
			}
			return fmt.Sprintf("<x-%s%s />", bladeName, attrs)
		})
		return content
	}

	if target == "twig" {
		twigName := strings.ToLower(compName)
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
			twigAttrs := attributesToTwig(attrs)
			if slot != "" {
				return fmt.Sprintf("{%% set slot %%}%s{%% endset %%}\n{%% include 'components/%s.html.twig' with { %s } %%}", slot, twigName, twigAttrs)
			}
			return fmt.Sprintf("{%% include 'components/%s.html.twig' with { %s } %%}", twigName, twigAttrs)
		})
		content = selfClosingRegex.ReplaceAllStringFunc(content, func(match string) string {
			submatches := selfClosingRegex.FindStringSubmatch(match)
			attrs := ""
			if len(submatches) > 1 {
				attrs = submatches[1]
			}
			twigAttrs := attributesToTwig(attrs)
			return fmt.Sprintf("{%% include 'components/%s.html.twig' with { %s } %%}", twigName, twigAttrs)
		})
		return content
	}

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
		if len(match) > 2 && match[2] != "" {
			val = match[2]
		} else if len(match) > 3 && match[3] != "" {
			val = match[3]
		} else if len(match) > 4 && match[4] != "" {
			val = match[4]
		} else if len(match) > 5 && match[5] != "" {
			val = match[5]
		}
		if strings.HasPrefix(name, "client:") {
			hydrate = strings.TrimPrefix(name, "client:")
		} else {
			props[name] = val
		}
	}

	h := fnv.New32a()
	h.Write([]byte(comp.Name + attrs))
	islandId := fmt.Sprintf("pphlx-%s-%x", strings.ToLower(comp.Name), h.Sum32())

	var propsBuilder strings.Builder
	propsBuilder.WriteString("{")
	i := 0
	for k, v := range props {
		if i > 0 {
			propsBuilder.WriteString(",")
		}
		cleanV := strings.TrimSpace(v)
		if strings.HasPrefix(cleanV, "{|=") && strings.HasSuffix(cleanV, "|}") {
			cleanV = strings.TrimSpace(cleanV[3 : len(cleanV)-2])
		} else if strings.HasPrefix(cleanV, "{") && strings.HasSuffix(cleanV, "}") {
			cleanV = strings.TrimSpace(cleanV[1 : len(cleanV)-1])
		}

		phpEchoRegex := regexp.MustCompile(`<\?php\s+echo\s+(.*?);?\s*\?>`)

		if strings.HasPrefix(cleanV, "$") {
			propsBuilder.WriteString(fmt.Sprintf("%q: <?php echo json_encode(%s); ?>", k, cleanV))
		} else if m := phpEchoRegex.FindStringSubmatch(v); len(m) > 1 {
			expr := strings.TrimSpace(m[1])
			propsBuilder.WriteString(fmt.Sprintf("%q: <?php echo json_encode(%s); ?>", k, expr))
		} else if strings.Contains(v, "<?php") {
			if strings.Contains(v, "json_encode") {
				propsBuilder.WriteString(fmt.Sprintf("%q: %s", k, v))
			} else {
				propsBuilder.WriteString(fmt.Sprintf("%q: %s", k, v))
			}
		} else {
			propsBuilder.WriteString(fmt.Sprintf("%q: %q", k, v))
		}
		i++
	}
	propsBuilder.WriteString("}")

	var result strings.Builder
	result.WriteString(fmt.Sprintf(`<div id="%s" class="pphlx-island" data-component="%s" data-framework="%s" data-hydrate="%s"></div>`, islandId, comp.Name, comp.Framework, hydrate))
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
		val := ""
		if len(match) > 2 && match[2] != "" {
			val = match[2]
		} else if len(match) > 3 && match[3] != "" {
			val = match[3]
		} else if len(match) > 4 && match[4] != "" {
			val = match[4]
		} else if len(match) > 5 && match[5] != "" {
			val = match[5]
		}
		props[match[1]] = val
	}

	// 1. Replace slot
	result = strings.ReplaceAll(result, "{{slot}}", slot)
	result = strings.ReplaceAll(result, "{|= $slot; |}", slot)
	result = strings.ReplaceAll(result, "{|= $slot |}", slot)
	result = strings.ReplaceAll(result, "<?php echo $slot; ?>", slot)
	result = strings.ReplaceAll(result, "<?php echo $slot; ?>;", slot)

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
			if strings.HasPrefix(val, "$") {
				return fmt.Sprintf("<?php echo %s; ?>", val)
			}
			return val
		}
		return defaultVal
	})

	return result
}

// getRelativePath calculates the relative asset path from page to asset file
func getRelativePath(fromPathPath, toPathPath string) string {
	if strings.TrimSpace(toPathPath) == "" {
		return ""
	}
	fromDir := filepath.Dir(fromPathPath)
	rel, err := filepath.Rel(fromDir, toPathPath)
	if err != nil {
		return filepath.ToSlash(toPathPath)
	}
	cleanRel := filepath.ToSlash(rel)
	if cleanRel == "." || cleanRel == "./" {
		return ""
	}
	return cleanRel
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

// parseMjsBool extracts a configuration boolean property from JavaScript .mjs config using regex
func parseMjsBool(content string, fieldName string) bool {
	re := regexp.MustCompile(fmt.Sprintf(`(?m)%s\s*:\s*(true|false)`, fieldName))
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1] == "true"
	}
	return false
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

// detectVitePlugins dynamically inspects package.json dependencies to import only installed framework plugins
func detectVitePlugins(projectDir string) (string, string) {
	pkgPath := filepath.Join(projectDir, "package.json")
	if !projectFileExists(pkgPath) {
		return "", "[]"
	}
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return "", "[]"
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	_ = json.Unmarshal(data, &pkg)

	allDeps := make(map[string]bool)
	for k := range pkg.Dependencies {
		allDeps[k] = true
	}
	for k := range pkg.DevDependencies {
		allDeps[k] = true
	}

	imports := []string{}
	plugins := []string{}

	if allDeps["@vitejs/plugin-vue"] || allDeps["vue"] {
		imports = append(imports, "import vue from '@vitejs/plugin-vue';")
		plugins = append(plugins, "vue()")
	}
	if allDeps["@sveltejs/vite-plugin-svelte"] || allDeps["svelte"] {
		imports = append(imports, "import { svelte } from '@sveltejs/vite-plugin-svelte';")
		plugins = append(plugins, "svelte()")
	}
	if allDeps["vite-plugin-solid"] || allDeps["solid-js"] {
		imports = append(imports, "import solidPlugin from 'vite-plugin-solid';")
		plugins = append(plugins, "solidPlugin()")
	}
	if allDeps["@vitejs/plugin-react"] || allDeps["react"] {
		imports = append(imports, "import react from '@vitejs/plugin-react';")
		plugins = append(plugins, "react()")
	}
	if allDeps["@preact/preset-vite"] || allDeps["preact"] {
		imports = append(imports, "import preact from '@preact/preset-vite';")
		plugins = append(plugins, "preact()")
	}

	importStr := strings.Join(imports, "\n")
	pluginStr := "[" + strings.Join(plugins, ", ") + "]"

	return importStr, pluginStr
}

// runViteBuild generates a temporary entry file, compiles Vue/Svelte components, and appends the result
func runViteBuild(config Config, projectDir string) error {
	entryDir := filepath.Join(projectDir, config.SrcDir)
	if config.SrcDir == "." || config.SrcDir == "" {
		entryDir = projectDir
	}
	os.MkdirAll(entryDir, 0755)
	entryPath := filepath.Join(entryDir, ".pphlx_entry.js")

	var entryContent strings.Builder

	if len(viteComponents) == 0 {
		_ = filepath.Walk(entryDir, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(p))
				if ext == ".vue" || ext == ".svelte" || strings.HasSuffix(p, ".solid.jsx") || strings.HasSuffix(p, ".solid.tsx") {
					baseName := filepath.Base(p)
					name := strings.TrimSuffix(baseName, ext)
					name = strings.TrimSuffix(name, ".solid")
					if name != "" {
						viteComponents[name] = p
					}
				}
			}
			return nil
		})
	}

	var exportNames []string
	for name, path := range viteComponents {
		rel, err := filepath.Rel(entryDir, path)
		if err != nil {
			rel = path
		}
		relPath := filepath.ToSlash(rel)
		if !strings.HasPrefix(relPath, ".") && !strings.HasPrefix(relPath, "/") {
			relPath = "./" + relPath
		}
		entryContent.WriteString(fmt.Sprintf("import %s from '%s';\n", name, relPath))
		entryContent.WriteString(fmt.Sprintf("window.%s = %s;\n", name, name))
		exportNames = append(exportNames, name)
	}

	// Expose SolidJS render helper
	entryContent.WriteString("import { render as solidRender } from 'solid-js/web';\n")
	entryContent.WriteString("window.SolidJS = { render: solidRender };\n")

	// Expose Svelte 5 mount helper
	entryContent.WriteString("import { mount as svelteMount } from 'svelte';\n")
	entryContent.WriteString("window.Svelte = { mount: svelteMount };\n")

	if len(exportNames) > 0 {
		entryContent.WriteString(fmt.Sprintf("export { %s };\n", strings.Join(exportNames, ", ")))
	}

	err := os.WriteFile(entryPath, []byte(entryContent.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write Vite entry file: %v", err)
	}

	relEntry, _ := filepath.Rel(projectDir, entryPath)
	relEntryPath := filepath.ToSlash(relEntry)

	targetViteOutDir := config.OutDir
	if targetViteOutDir == "" {
		targetViteOutDir = "dist"
	}
	if activeMode == "dev" {
		targetViteOutDir = ".pphlx/cache"
	}

	outJsDir := filepath.ToSlash(filepath.Join(targetViteOutDir, "assets", "js"))

	configFile := "pphlx.vite.config.mjs"
	mainMjs := filepath.Join(projectDir, "pphlx.config.mjs")
	mainCjs := filepath.Join(projectDir, "pphlx.config.cjs")

	useMainConfig := false
	if projectFileExists(mainMjs) {
		content, _ := os.ReadFile(mainMjs)
		if strings.Contains(string(content), "plugins") || strings.Contains(string(content), "build") {
			configFile = "pphlx.config.mjs"
			useMainConfig = true
		}
	} else if projectFileExists(mainCjs) {
		content, _ := os.ReadFile(mainCjs)
		if strings.Contains(string(content), "plugins") || strings.Contains(string(content), "build") {
			configFile = "pphlx.config.cjs"
			useMainConfig = true
		}
	}

	if !useMainConfig {
		pphlxDir := filepath.Join(projectDir, ".pphlx")
		_ = os.MkdirAll(pphlxDir, 0755)
		viteConfigPath := filepath.Join(pphlxDir, "pphlx.vite.config.mjs")
		importsStr, pluginsArrayStr := detectVitePlugins(projectDir)
		viteConfig := fmt.Sprintf(`
import { defineConfig } from 'vite';
%s

export default defineConfig({
  plugins: %s,
  publicDir: false,
  build: {
    lib: {
      entry: '%s',
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
    outDir: '%s',
    emptyOutDir: false,
    minify: true
  }
});
`, importsStr, pluginsArrayStr, relEntryPath, outJsDir)
		os.WriteFile(viteConfigPath, []byte(viteConfig), 0644)
		configFile = filepath.ToSlash(filepath.Join(".pphlx", "pphlx.vite.config.mjs"))
	}

	fmt.Println("Running local Vite compilation for Vue/Svelte components...")

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", fmt.Sprintf("npx vite build --config %s", configFile))
	} else {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("npx vite build --config %s", configFile))
	}
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("vite build failed: %v", err)
	}

	// Append compiled bundles to global JS
	viteBundlePath := filepath.Join(projectDir, targetViteOutDir, "assets", "js", "pphlx_vite.js")
	viteJS, err := os.ReadFile(viteBundlePath)
	if err == nil {
		jsOut := filepath.Join(projectDir, config.JsOut)
		f, err := os.OpenFile(jsOut, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			f.Write([]byte("\n" + string(viteJS) + "\n"))
		}
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
	tempDir, err := os.MkdirTemp("", "pphlx-pkg-")
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

	if manifestData, err := os.ReadFile(manifestPath); err == nil {
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
		os.WriteFile(manifestPath, newManifestData, 0644)
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
			if _, err1 := os.Stat(filepath.Join(dir, "pphlx.config.mjs")); err1 == nil ||
				projectFileExists(filepath.Join(dir, "pphlx.config.cjs")) ||
				projectFileExists(filepath.Join(dir, "pphlx.config.json")) ||
				projectFileExists(filepath.Join(dir, "pphlx.json")) ||
				projectFileExists(filepath.Join(dir, "package.json")) {
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

		// Read config to resolve srcDir, defaulting to "src"
		srcDirName := "src"
		if activeConfig.SrcDir != "" {
			srcDirName = activeConfig.SrcDir
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
		if err := os.WriteFile(componentFullPath, []byte(boilerplateCode), 0644); err != nil {
			return nil, fmt.Errorf("failed to write component file: %v", err)
		}

		// Inject @import at the top of the template file
		pageData, err := os.ReadFile(targetPath)
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
			os.WriteFile(targetPath, []byte(newContent), 0644)
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

		err = os.WriteFile(destPath, data, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s to %s: %v", baseName, destPath, err)
		}
		fmt.Printf("Registered tool file: %s\n", destPath)
	}

	return nil
}

func attributesToTwig(attrs string) string {
	matches := attrRegex.FindAllStringSubmatch(attrs, -1)
	var pairs []string
	for _, match := range matches {
		name := match[1]
		val := ""
		if match[2] != "" {
			val = fmt.Sprintf("'%s'", match[2])
		} else if match[3] != "" {
			val = fmt.Sprintf("'%s'", match[3])
		} else if match[4] != "" {
			val = match[4]
		} else if match[5] != "" {
			val = match[5]
		}
		if val != "" {
			pairs = append(pairs, fmt.Sprintf("%s: %s", name, val))
		}
	}
	return strings.Join(pairs, ", ")
}

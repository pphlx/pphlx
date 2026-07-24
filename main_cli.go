//go:build !js || !wasm

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

func main() {
	if len(os.Args) > 1 {
		cmd := strings.ToLower(os.Args[1])
		if cmd == "version" || cmd == "--version" || cmd == "-v" || cmd == "-version" {
			fmt.Println("pphlx v1.1.0")
			os.Exit(0)
		}
		if cmd == "init" {
			runInitCLI()
			os.Exit(0)
		}
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
		if cmd == "telemetry" {
			if len(os.Args) > 2 {
				subCmd := strings.ToLower(os.Args[2])
				homeDir, _ := os.UserHomeDir()
				configDir := filepath.Join(homeDir, ".pphlx")
				os.MkdirAll(configDir, 0755)
				configFile := filepath.Join(configDir, "telemetry.json")

				if subCmd == "disable" {
					ioutil.WriteFile(configFile, []byte(`{"disabled": true}`), 0644)
					fmt.Println("✓ PPHLX telemetry has been disabled globally on this machine.")
					os.Exit(0)
				}
				if subCmd == "enable" {
					ioutil.WriteFile(configFile, []byte(`{"disabled": false}`), 0644)
					fmt.Println("✓ PPHLX telemetry has been enabled globally on this machine.")
					os.Exit(0)
				}
			}
			fmt.Println("Usage: pphlx telemetry <enable|disable>")
			os.Exit(0)
		}
		if cmd == "check" {
			fmt.Println("Checking PPHLX project files...")
			srcDir := "src"
			if _, err := os.Stat("src"); os.IsNotExist(err) {
				if _, err := os.Stat("test_project/src"); err == nil {
					srcDir = "test_project/src"
				}
			}
			projectDir := "."
			if srcDir == "test_project/src" {
				projectDir = "test_project"
			}
			diagnostics := RunDiagnostics(srcDir, projectDir)
			if len(diagnostics) > 0 {
				for _, d := range diagnostics {
					fmt.Print(d.String())
				}
				fmt.Printf("\n\033[1;31mPPHLX diagnostic check failed: %d error(s) found.\033[0m\n", len(diagnostics))
				os.Exit(1)
			}
			fmt.Println("✓ Template syntax validation: OK")
			fmt.Println("✓ Component import resolution: OK")
			fmt.Println("✓ Asset path & image resolution: OK")
			fmt.Println("\033[1;32mPPHLX diagnostic check: 0 errors, 0 warnings.\033[0m")
			os.Exit(0)
		}
	}

	fmt.Println("PPHLX Compiler Starting...")

	// Default command is "build", check if user passed "dev" or "watch"
	activeMode = "build"
	if len(os.Args) > 1 {
		cmd := strings.ToLower(os.Args[1])
		if cmd == "dev" || cmd == "watch" {
			activeMode = "dev"
		}
	}

	// 1. Read config (looking for pphlx.config.mjs first, then pphlx.config.json, then pphlx.json)
	configPath := "./pphlx.config.mjs"
	if runtime.GOOS == "wasip1" {
		configPath = "/pphlx.config.mjs"
	}
	isMjs := true
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if runtime.GOOS == "wasip1" {
			configPath = "/pphlx.config.json"
		} else {
			configPath = "./pphlx.config.json"
		}
		isMjs = false
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if runtime.GOOS == "wasip1" {
				configPath = "/pphlx.json"
			} else {
				configPath = "./pphlx.json"
			}
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				// Only fallback to test_project if developing inside PPHLX root compiler repository
				if _, err := os.Stat("pphlx-core"); err == nil {
					configPath = "./test_project/pphlx.config.json"
					isMjs = false
				}
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					fmt.Println("\033[1;31mError: No PPHLX project configuration (pphlx.config.json or pphlx.json) found in current directory.\033[0m")
					fmt.Println("👉 Run 'npx pphlx init' to initialize a new PPHLX project automatically.")
					os.Exit(1)
				}
			}
		}
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading config '%s': %v\n", configPath, err)
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
		config.Site = parseMjsField(configStr, "site")
		config.Sitemap = parseMjsBool(configStr, "sitemap")
		config.Default = parseMjsField(configStr, "default")
	} else {
		if err := json.Unmarshal(configData, &config); err != nil {
			fmt.Printf("Error parsing JSON config: %v\n", err)
			os.Exit(1)
		}
	}

	// 2. Parse flags: --env / -e, --all, and --target / -t
	cliEnv := ""
	cliAll := false
	cliTarget := ""

	for i := 1; i < len(os.Args)-1; i++ {
		arg := strings.ToLower(os.Args[i])
		if arg == "--env" || arg == "-e" {
			cliEnv = os.Args[i+1]
		}
		if arg == "--target" || arg == "-t" {
			cliTarget = os.Args[i+1]
		}
	}
	for i := 1; i < len(os.Args); i++ {
		arg := strings.ToLower(os.Args[i])
		if arg == "--all" {
			cliAll = true
		}
		if strings.HasPrefix(arg, "--env=") {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				cliEnv = parts[1]
			}
		}
		if strings.HasPrefix(arg, "--target=") {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				cliTarget = parts[1]
			}
		}
	}

	projectDir := filepath.Dir(configPath)

	// If --all is passed, compile all environments in parallel
	if cliAll && len(config.Environments) > 0 {
		var wg sync.WaitGroup
		fmt.Printf("Compiling all %d environments in parallel...\n", len(config.Environments))
		for envName, envProfile := range config.Environments {
			wg.Add(1)
			go func(name string, profile ConfigProfile) {
				defer wg.Done()
				envCfg := config
				if profile.OutDir != "" {
					envCfg.OutDir = profile.OutDir
				}
				if profile.CssOut != "" {
					envCfg.CssOut = profile.CssOut
				}
				if profile.JsOut != "" {
					envCfg.JsOut = profile.JsOut
				}
				if profile.Output.Target != "" {
					envCfg.Output.Target = profile.Output.Target
				}
				if profile.Output.Goos != "" {
					envCfg.Output.Goos = profile.Output.Goos
				}
				if profile.Output.Goarch != "" {
					envCfg.Output.Goarch = profile.Output.Goarch
				}
				if envCfg.Output.Target == "" {
					envCfg.Output.Target = "php"
				}
				fmt.Printf("[%s] Compiling target '%s' to '%s'...\n", name, envCfg.Output.Target, envCfg.OutDir)
				compileAll(envCfg, projectDir)
			}(envName, envProfile)
		}
		wg.Wait()
		fmt.Println("✓ All environments compiled successfully!")
		if activeMode == "dev" {
			fmt.Println("[Warning] dev/watch mode is only supported for a single environment. Watching the default environment profile...")
			startWatcher(config, projectDir)
		}
		os.Exit(0)
	}

	// Overlay environmental overrides if --env is specified
	if cliEnv != "" {
		profile, exists := config.Environments[cliEnv]
		if !exists {
			fmt.Printf("Error: Environment profile '%s' not found in configuration.\n", cliEnv)
			os.Exit(1)
		}
		if profile.OutDir != "" {
			config.OutDir = profile.OutDir
		}
		if profile.CssOut != "" {
			config.CssOut = profile.CssOut
		}
		if profile.JsOut != "" {
			config.JsOut = profile.JsOut
		}
		if profile.Output.Target != "" {
			config.Output.Target = profile.Output.Target
		}
		if profile.Output.Goos != "" {
			config.Output.Goos = profile.Output.Goos
		}
		if profile.Output.Goarch != "" {
			config.Output.Goarch = profile.Output.Goarch
		}
		fmt.Printf("Loaded environment profile: %s (Target: %s, OutDir: %s)\n", cliEnv, config.Output.Target, config.OutDir)
	}

	if cliTarget != "" {
		config.Output.Target = strings.ToLower(cliTarget)
	}

	if config.Output.Target == "" {
		config.Output.Target = "php"
	}

	// Run initial compilation
	compileAll(config, projectDir)

	// Non-blocking background telemetry dispatch
	sendTelemetryAsync(config, activeMode)

	if activeMode == "dev" || activeMode == "preview" {
		startDevServerAndWatcher(config, projectDir, activeMode)
	}
}

func isTelemetryDisabled() bool {
	if os.Getenv("PPHLX_TELEMETRY_DISABLED") == "1" || os.Getenv("PPHLX_TELEMETRY_DISABLED") == "true" {
		return true
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	configFile := filepath.Join(homeDir, ".pphlx", "telemetry.json")
	data, err := ioutil.ReadFile(configFile)
	if err == nil && strings.Contains(string(data), `"disabled": true`) {
		return true
	}
	return false
}

func sendTelemetryAsync(config Config, cmdName string) {
	if isTelemetryDisabled() {
		return
	}
	// Run in background goroutine so compiler main thread never blocks
	go func() {
		defer func() { recover() }() // Catch background panics safely

		base := fmt.Sprintf("%s:%s:%s:%s", cmdName, "1.0.8", runtime.GOOS, runtime.GOARCH)
		nonce := solvePoW(base)

		payload := TelemetryPayload{
			CliCommand: cmdName,
			Version:    "1.1.0",
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
			Config: TelemetryConfig{
				Target:      config.Output.Target,
				HasProfiles: len(config.Environments) > 0,
				Goos:        config.Output.Goos,
				Goarch:      config.Output.Goarch,
			},
			Nonce: nonce,
		}

		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return
		}

		client := &http.Client{Timeout: 2 * time.Second}
		req, err := http.NewRequest("POST", "http://localhost:4321/api/stats.json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}()
}

func runInitCLI() {
	fmt.Println("Initializing PPHLX project configuration...")

	currDir, _ := filepath.Abs(".")
	projectName := filepath.Base(currDir)

	// 1. Create or update package.json
	pkgPath := "package.json"
	var pkgData map[string]interface{}
	if _, err := os.Stat(pkgPath); err == nil {
		content, _ := ioutil.ReadFile(pkgPath)
		json.Unmarshal(content, &pkgData)
	}
	if pkgData == nil {
		pkgData = make(map[string]interface{})
		pkgData["name"] = projectName
		pkgData["version"] = "1.0.0"
		pkgData["private"] = true
	}
	scripts, ok := pkgData["scripts"].(map[string]interface{})
	if !ok || scripts == nil {
		scripts = make(map[string]interface{})
	}
	scripts["build"] = "pphlx"
	scripts["dev"] = "pphlx dev"
	scripts["watch"] = "pphlx watch"
	scripts["start"] = "pphlx dev"
	scripts["preview"] = "pphlx preview"
	scripts["check"] = "pphlx check"
	pkgData["scripts"] = scripts

	pkgBytes, _ := json.MarshalIndent(pkgData, "", "  ")
	ioutil.WriteFile(pkgPath, pkgBytes, 0644)
	fmt.Println("✓ Configured package.json scripts")

	// 2. Create pphlx.json
	pphlxJsonPath := "pphlx.json"
	if _, err := os.Stat(pphlxJsonPath); os.IsNotExist(err) {
		pphlxJson := map[string]interface{}{
			"name":        projectName,
			"version":     "1.0.0",
			"description": "PPHLX Monolithic Application",
			"scripts": map[string]string{
				"build": "pphlx build",
				"dev":   "pphlx dev",
				"watch": "pphlx watch",
			},
			"dependencies": map[string]string{
				"pphlx": "^1.1.0",
			},
		}
		pBytes, _ := json.MarshalIndent(pphlxJson, "", "  ")
		ioutil.WriteFile(pphlxJsonPath, pBytes, 0644)
		fmt.Println("✓ Created pphlx.json (Project Manifest)")
	}

	// 3. Create pphlx.config.json
	configJsonPath := "pphlx.config.json"
	if _, err := os.Stat(configJsonPath); os.IsNotExist(err) {
		configJson := map[string]interface{}{
			"srcDir":   "src",
			"outDir":   "dist",
			"cssOut":   "dist/css/styles.css",
			"jsOut":    "dist/js/bundle.js",
			"output": map[string]string{
				"target": "php",
			},
		}
		cBytes, _ := json.MarshalIndent(configJson, "", "  ")
		ioutil.WriteFile(configJsonPath, cBytes, 0644)
		fmt.Println("✓ Created pphlx.config.json (Compiler Config)")
	}

	// 4. Create pphlx.vite.config.mjs
	viteConfigPath := "pphlx.vite.config.mjs"
	if _, err := os.Stat(viteConfigPath); os.IsNotExist(err) {
		viteContent := `import { defineConfig } from 'vite';
import pphlx from 'pphlx/vite';

export default defineConfig({
  plugins: [pphlx()],
  build: {
    outDir: 'dist',
    emptyOutDir: true
  }
});
`
		ioutil.WriteFile(viteConfigPath, []byte(viteContent), 0644)
		fmt.Println("✓ Created pphlx.vite.config.mjs (Vite Integration Config)")
	}

	// 5. Scaffold src/ directory structure (src/index.pphx, src/layouts/Layout.pphx, src/assets/pphlx.svg, public/favicon.svg)
	os.MkdirAll("src/layouts", 0755)
	os.MkdirAll("src/components", 0755)
	os.MkdirAll("src/assets", 0755)
	os.MkdirAll("public", 0755)

	// Layout.pphx
	if _, err := os.Stat("src/layouts/Layout.pphx"); os.IsNotExist(err) {
		layoutContent := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to PPHLX</title>
    <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: #0a0a0a;
            --text-color: #ededed;
            --border-color: #333333;
            --hover-bg: #1a1a1a;
            --accent: #ffffff;
        }
        
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: 'Inter', system-ui, -apple-system, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-color);
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            overflow-x: hidden;
            -webkit-font-smoothing: antialiased;
        }

        main {
            flex: 1;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            width: 100%;
            padding: 2rem;
        }
    </style>
    {{PPHLX_CSS}}
</head>
<body>
    <main>
        {{slot}}
    </main>
    {{PPHLX_JS}}
</body>
</html>
`
		ioutil.WriteFile("src/layouts/Layout.pphx", []byte(layoutContent), 0644)
		fmt.Println("✓ Created src/layouts/Layout.pphx")
	}

	// index.pphx
	if _, err := os.Stat("src/index.pphx"); os.IsNotExist(err) {
		indexContent := `@import Layout from './layouts/Layout.pphx'

<Layout title="Welcome to PPHLX!">
    <div class="container">
        <div class="logo-wrapper">
            <img src="./assets/pphlx.svg" width="160" height="auto" alt="PPHLX logo" class="logo-svg" />
        </div>

        <div class="get-started">
            <p>Get started by editing</p>
            <code class="code-block">src/index.pphx</code>
        </div>

        <div class="grid">
            <a href="https://pphlx.org" target="_blank" class="card">
                <h2>Docs <span>&rarr;</span></h2>
                <p>Find in-depth information about PPHLX features and API.</p>
            </a>
            <a href="http://pphlx.org/on/discord" target="_blank" class="card">
                <h2>Discord <span>&rarr;</span></h2>
                <p>Join the community and chat with other PPHLX developers.</p>
            </a>
            <a href="https://github.com/pphlx/pphlx" target="_blank" class="card">
                <h2>GitHub <span>&rarr;</span></h2>
                <p>Contribute to the compiler, report issues, and star the repo.</p>
            </a>
            <a href="https://pphlx.org" target="_blank" class="card">
                <h2>Deploy <span>&rarr;</span></h2>
                <p>Instantly deploy your standalone binaries or PHP pages.</p>
            </a>
        </div>
    </div>

    <style>
        .container {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            max-width: 1000px;
            width: 100%;
            margin: 0 auto;
            font-family: 'Inter', system-ui, -apple-system, sans-serif;
        }

        .logo-wrapper {
            display: flex;
            justify-content: center;
            align-items: center;
            margin-bottom: 4rem;
            position: relative;
        }

        .logo-wrapper::before {
            content: '';
            position: absolute;
            width: 240px;
            height: 120px;
            background: rgba(255, 255, 255, 0.05);
            filter: blur(40px);
            z-index: -1;
            border-radius: 50%;
        }

        .get-started {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            margin-bottom: 5rem;
            font-size: 1rem;
            color: #888;
        }

        .code-block {
            background-color: rgba(255, 255, 255, 0.1);
            padding: 0.3rem 0.6rem;
            border-radius: 6px;
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
            font-size: 0.9rem;
            color: #ededed;
            border: 1px solid rgba(255, 255, 255, 0.08);
        }

        .grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 1.5rem;
            width: 100%;
        }

        .card {
            padding: 1.5rem 2rem;
            border-radius: 8px;
            background: rgba(255, 255, 255, 0.02);
            border: 1px solid rgba(255, 255, 255, 0.08);
            text-decoration: none;
            color: var(--text-color);
            transition: all 0.2s ease;
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }

        .card:hover {
            background: rgba(255, 255, 255, 0.04);
            border-color: rgba(255, 255, 255, 0.2);
            transform: translateY(-2px);
        }

        .card h2 {
            font-size: 1.2rem;
            font-weight: 600;
            display: flex;
            align-items: center;
            gap: 0.4rem;
        }

        .card h2 span {
            transition: transform 0.2s ease;
        }

        .card:hover h2 span {
            transform: translateX(4px);
        }

        .card p {
            margin: 0;
            font-size: 0.95rem;
            color: #888;
            line-height: 1.5;
        }

        @media (max-width: 768px) {
            .grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</Layout>
`
		ioutil.WriteFile("src/index.pphx", []byte(indexContent), 0644)
		fmt.Println("✓ Created src/index.pphx")
	}

	// pphlx.svg
	if _, err := os.Stat("src/assets/pphlx.svg"); os.IsNotExist(err) {
		svgContent := `<svg width="115" height="40" viewBox="0 0 460 160" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true" class="logo-svg">
  <path d="M161.736 139.478V60.6216H178.289L178.634 73.6109L177.255 73.0362C178.787 68.5914 181.279 65.1812 184.726 62.8056C188.175 60.43 192.199 59.2422 196.796 59.2422C202.543 59.2422 207.334 60.6599 211.165 63.4953C214.997 66.3307 217.87 70.2007 219.786 75.1053C221.779 79.9331 222.775 85.3742 222.775 91.4282C222.775 97.4056 221.779 102.847 219.786 107.751C217.87 112.656 214.959 116.526 211.05 119.361C207.218 122.197 202.429 123.615 196.682 123.615C193.693 123.615 190.895 123.077 188.29 122.005C185.684 120.855 183.424 119.284 181.509 117.292C179.669 115.3 178.328 112.924 177.485 110.165L178.979 109.245V139.478H161.736ZM191.969 110.395C196.03 110.395 199.21 108.709 201.509 105.337C203.885 101.965 205.073 97.329 205.073 91.4282C205.073 85.5275 203.885 80.8911 201.509 77.5192C199.21 74.1474 196.03 72.4614 191.969 72.4614C189.286 72.4614 186.948 73.1895 184.956 74.6455C183.041 76.0248 181.547 78.1323 180.473 80.9677C179.477 83.8032 178.979 87.29 178.979 91.4282C178.979 95.5665 179.477 99.0533 180.473 101.889C181.547 104.724 183.041 106.87 184.956 108.326C186.948 109.705 189.286 110.395 191.969 110.395Z" fill="#FFFFFF"></path>
  <path d="M231.169 139.478V60.6216H247.722L248.067 73.6109L246.688 73.0362C248.22 68.5914 250.711 65.1812 254.159 62.8056C257.608 60.43 261.631 59.2422 266.229 59.2422C271.977 59.2422 276.766 60.6599 280.598 63.4953C284.429 66.3307 287.303 70.2007 289.218 75.1053C291.212 79.9331 292.208 85.3742 292.208 91.4282C292.208 97.4056 291.212 102.847 289.218 107.751C287.303 112.656 284.391 116.526 280.482 119.361C276.651 122.197 271.861 123.615 266.114 123.615C263.125 123.615 260.328 123.077 257.722 122.005C255.117 120.855 252.857 119.284 250.941 117.292C249.101 115.3 247.76 112.924 246.918 110.165L248.412 109.245V139.478H231.169ZM261.401 110.395C265.463 110.395 268.642 108.709 270.942 105.337C273.317 101.965 274.505 97.329 274.505 91.4282C274.505 85.5275 273.317 80.8911 270.942 77.5192C268.642 74.1474 265.463 72.4614 261.401 72.4614C258.719 72.4614 256.382 73.1895 254.389 74.6455C252.473 76.0248 250.979 78.1323 249.905 80.9677C248.91 83.8032 248.412 87.29 248.412 91.4282C248.412 95.5665 248.91 99.0533 249.905 101.889C250.979 104.724 252.473 106.87 254.389 108.326C256.382 109.705 258.719 110.395 261.401 110.395Z" fill="#FFFFFF"></path>
  <path d="M300.601 122.236V40.6211H317.844V74.9913H315.546C316.158 71.3895 317.346 68.4391 319.109 66.1401C320.947 63.8411 323.209 62.1168 325.89 60.9673C328.65 59.8178 331.639 59.2431 334.857 59.2431C339.456 59.2431 343.287 60.2393 346.352 62.2318C349.418 64.2243 351.716 66.9831 353.249 70.5082C354.859 74.0333 355.663 78.0566 355.663 82.578V122.236H338.42V86.7162C338.42 82.1182 337.692 78.6697 336.236 76.3707C334.857 74.0717 332.481 72.9221 329.11 72.9221C325.584 72.9221 322.825 74.1099 320.833 76.4856C318.84 78.8612 317.844 82.4248 317.844 87.176V122.236H300.601Z" fill="#FFFFFF"></path>
  <path d="M383.771 122.236C378.712 122.236 374.765 120.972 371.931 118.443C369.095 115.913 367.678 111.814 367.678 106.143V40.6211H384.919V104.304C384.919 106.143 385.341 107.446 386.184 108.212C387.027 108.978 388.253 109.361 389.862 109.361H394.001V122.236H383.771Z" fill="#FFFFFF"></path>
  <path d="M395.858 122.231L417.813 90.964L396.433 60.6172H414.71L427.929 80.3886L440.688 60.6172H459.426L438.16 91.079L460 122.231H441.723L428.159 101.31L414.48 122.231H395.858Z" fill="#FFFFFF"></path>
  <path d="M75.8798 50.4495C60.855 50.9676 47.685 58.742 43.0304 63.022C40.6942 64.7341 34.5193 71.7417 31.5278 80.1422C26.2313 95.0152 32.5978 110.637 38.9644 119.572C43.726 126.255 57.5942 140.578 77.3238 142.095C93.321 143.327 109.153 135.8 115.63 130.219C120.66 125.885 133.5 113.526 134.409 92.8217C135.318 72.1171 121.41 62.4234 113.812 57.886C106.375 53.4454 97.6005 49.7004 75.8798 50.4495Z" fill="#FFFFFF"></path>
  <path d="M96.8516 43.9678C96.4771 43.7537 97.226 42.131 97.6544 41.3998C107.819 22.5677 118.574 17.3245 119.054 17.1106C119.439 16.9394 119.781 16.9744 119.964 17.1641C136.175 34.6586 137.538 65.9628 136.71 68.3105C136.446 69.0595 134.944 68.792 134.569 67.8825C122.264 47.4988 97.3811 44.2705 96.8516 43.9678Z" fill="#FFFFFF"></path>
  <path d="M47.9524 48.7344C47.7812 48.7344 46.6666 48.6553 45.0098 48.7344C18.0992 50.0184 7.45266 59.6485 0.497599 65.801C-0.160157 66.3829 -0.00169498 67.0672 0.0696518 67.3525C4.02867 82.6002 19.9182 97.3127 21.6838 98.2222C23.0963 98.9499 23.378 97.7764 23.3422 97.0987C21.6838 71.1511 33.5553 62.0093 38.4294 57.5619C42.7094 53.6564 47.9971 51.1757 48.5944 50.3393C49.3969 49.2159 48.1664 48.7344 47.9524 48.7344Z" fill="#FFFFFF"></path>
  <path d="M93.6416 120.59C93.0858 120.247 93.4813 119.448 93.6416 119.253C101.043 113.225 115.813 101.266 116.54 100.367C117.268 99.4681 117.557 99.4932 117.61 99.618C119.032 102.935 117.076 112.084 113.223 116.363C109.371 120.644 103.754 121.393 100.864 121.393C97.975 121.393 94.338 121.019 93.6416 120.59Z" fill="#13151A"></path>
  <path d="M46.829 99.5551C46.3154 99.2555 46.187 99.9297 46.187 100.304C45.9195 102.23 45.759 106.35 48.0595 112.181C50.3601 118.012 56.7266 120.848 59.9366 121.276C63.1465 121.704 69.46 121.009 70.4231 120.635C71.3862 120.26 71.172 119.724 70.7976 119.404C70.4231 119.083 47.471 99.9297 46.829 99.5551Z" fill="#13151A"></path>
</svg>
`
		ioutil.WriteFile("src/assets/pphlx.svg", []byte(svgContent), 0644)
		fmt.Println("✓ Created src/assets/pphlx.svg")
	}

	// favicon.svg
	if _, err := os.Stat("public/favicon.svg"); os.IsNotExist(err) {
		favContent := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
  <rect width="100" height="100" rx="20" fill="#0A0A0A"/>
  <path d="M25 75V25H45C55 25 60 30 60 38C60 46 55 51 45 51H37V75H25ZM37 40H44C48 40 50 38 50 35C50 32 48 30 44 30H37V40Z" fill="#FFFFFF"/>
</svg>`
		ioutil.WriteFile("public/favicon.svg", []byte(favContent), 0644)
		fmt.Println("✓ Created public/favicon.svg")
	}

	// favicon.ico (Base64 Binary)
	if _, err := os.Stat("public/favicon.ico"); os.IsNotExist(err) {
		icoB64 := "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mMzMABR4ZF48eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF+gdFhYjAAAAAAAAAAAAAAABHRoVYR4ZF/seGRf/HhkX/x0XFw0AAAAAHRkXex4ZF/0eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkW2iIaGh4AAAAAAAAAABwYGDYfGRfpHhkX/x4ZF/8eGRf/HhkX/xwYGEkAAAAAHBkWWx0ZF/MeGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRbBIxcXFgAAAAAnFBQNHxkXwB4ZF/8eGRf/HhkX/x4ZF/8eGRf/HRgWdAAAAAAgGxUwHhkX5B4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x0ZF50aGhoKAAAAAR4ZGV0eGRfyHhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGhewAAAAAAAAAAQdGRe/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf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AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAeGBVfHhkX9B4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/ceGRZnAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAbGw0THRgXvx4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF+QbFRUwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIBcXOB0YFvseGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/x4ZF7lAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB0ZF4UdGRf7HhkX/x4ZF/8eGRf/HhkX/x4ZF/8eGRf/HhkX/RwZFlIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABcXFxYdGRa3HhkX/h4ZF/8eGRf/HhkX/x4ZF/8eGRf/HRkXyiQAAAcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEcGBg2HhkY4x4ZF/8eGRf/HhkX/x4ZF/8eGhf6GxgYSwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAHBgVSR4ZF+seGRf/HhkX/x4ZF/4dGRewHh4eEQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB4XF0weGRfTHhkX/h4ZF+geFhZEAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMeGRkzHRkWwR4XFW9AAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="
		icoBytes, err := base64.StdEncoding.DecodeString(icoB64)
		if err == nil {
			ioutil.WriteFile("public/favicon.ico", icoBytes, 0644)
			fmt.Println("✓ Created public/favicon.ico")
		}
	}

	// README.md
	if _, err := os.Stat("README.md"); os.IsNotExist(err) {
		readmeContent := `# 🚀 Welcome to PPHLX

PPHLX is a modern, high-performance web framework. It combines the developer experience of modern JavaScript tooling (like Vite hot-reloading) with the raw performance of native PHP binaries and zero Node.js runtime in production. 

This is your official PPHLX minimal starter template, designed with clean aesthetics and modern component-driven architecture to get you building high-performance PHP applications immediately.

## 📂 Project Structure

Inside of your PPHLX project, you'll see the following folders and files:

` + "```text" + `
/
├── public/
│   ├── favicon.ico
│   └── favicon.svg
├── src/
│   ├── assets/
│   │   └── pphlx.svg
│   ├── layouts/
│   │   └── Layout.pphx
│   └── index.pphx
├── pphlx.config.json
├── pphlx.vite.config.mjs
├── package.json
└── README.md
` + "```" + `

PPHLX acts as a strict 1:1 Static Site Generator (SSG) mirroring compiler. Any .pphx file inside the src directory will be directly compiled to the dist directory.
- src/index.pphx compiles to dist/index.html
- src/layouts/ and src/assets/ contain your layout wrappers and static internal assets.

Static assets that do not need compilation (like your favicon) can be placed in the public/ directory.

## 🧞 Commands

All commands are run from the root of the project, from a terminal:

| Command                   | Action                                           |
| :------------------------ | :----------------------------------------------- |
| npm install             | Installs dependencies                            |
| npm run dev             | Starts local dev server at localhost:6321      |
| npm run build           | Compiles your .pphx files into the dist/ dir |

## 📚 Learn More

- **Documentation:** [Read our docs](https://pphlx.org)
- **Community:** [Join our Discord](http://pphlx.org/on/discord)
- **Repository:** [GitHub](https://github.com/pphlx/pphlx)
`
		ioutil.WriteFile("README.md", []byte(readmeContent), 0644)
		fmt.Println("✓ Created README.md")
	}

	fmt.Println("\033[1;32m✓ PPHLX project initialized successfully with starter design!\033[0m")
}



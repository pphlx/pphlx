//go:build !js || !wasm

package main

import (
	"bytes"
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
			fmt.Println("pphlx v" + Version)
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

	// Run execution mode handler
	switch activeMode {
	case "dev":
		devConfig := config
		devConfig.OutDir = config.SrcDir
		if devConfig.OutDir == "" {
			devConfig.OutDir = "."
		}
		compileAll(devConfig, projectDir)
		startDevServerAndWatcher(config, projectDir, activeMode)
	case "build":
		compileAll(config, projectDir)
	case "preview":
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
			Version:    Version,
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
	fmt.Println("Initializing PPHLX configuration...")

	// 1. Update package.json scripts if package.json exists
	pkgPath := "package.json"
	if _, err := os.Stat(pkgPath); err == nil {
		content, err := ioutil.ReadFile(pkgPath)
		if err == nil {
			var pkgData map[string]interface{}
			if err := json.Unmarshal(content, &pkgData); err == nil && pkgData != nil {
				scripts, ok := pkgData["scripts"].(map[string]interface{})
				if !ok || scripts == nil {
					scripts = make(map[string]interface{})
				}
				if _, exists := scripts["build"]; !exists {
					scripts["build"] = "pphlx"
				}
				if _, exists := scripts["dev"]; !exists {
					scripts["dev"] = "pphlx dev"
				}
				if _, exists := scripts["watch"]; !exists {
					scripts["watch"] = "pphlx watch"
				}
				if _, exists := scripts["start"]; !exists {
					scripts["start"] = "pphlx dev"
				}
				if _, exists := scripts["preview"]; !exists {
					scripts["preview"] = "pphlx preview"
				}
				if _, exists := scripts["check"]; !exists {
					scripts["check"] = "pphlx check"
				}
				pkgData["scripts"] = scripts

				pkgBytes, _ := json.MarshalIndent(pkgData, "", "  ")
				_ = ioutil.WriteFile(pkgPath, pkgBytes, 0644)
				fmt.Println("✓ Configured package.json scripts")
			}
		}
	}

	// 2. Create minimal pphlx.config.json if not present
	configJsonPath := "pphlx.config.json"
	if _, err := os.Stat(configJsonPath); os.IsNotExist(err) {
		configJson := map[string]interface{}{
			"srcDir": "src",
			"outDir": "dist",
			"output": map[string]string{
				"target": "php",
			},
		}
		cBytes, _ := json.MarshalIndent(configJson, "", "  ")
		_ = ioutil.WriteFile(configJsonPath, cBytes, 0644)
		fmt.Println("✓ Created pphlx.config.json (Compiler Config)")
	}
}



# PPHLX MCP Server Instructions

To run the PPHLX MCP server in Antigravity IDE:

### Execution Configuration
*   **Command**: `F:\VS CODE\GO\PPHLX\pphlx.exe`
*   **Arguments**: `mcp`
*   **Transport**: `stdio`

### Best Practices
*   When generating islands, the model should verify the target `.pphx` template exists.
*   Framework imports require explicit file extensions (except for standard React/JSX).
*   For cross-platform compilation targets, support `-t / --target` flag overrides (e.g. `standalone`, `desktop`, `android`, `ios`).

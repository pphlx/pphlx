# Contributing to PPHLX Compiler Core

Thank you for your interest in contributing to **PPHLX**! PPHLX is a fast, zero-dependency multi-framework template compiler for native PHP applications. We welcome contributions from developers of all skill levels.

---

## 1. Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please read it to understand our community standards.

---

## 2. How to Contribute

### Reporting Bugs
If you encounter a bug or compiler issue:
1. Search existing GitHub Issues to ensure it has not already been reported.
2. Open a new issue with a clear description, minimal reproducible example (`.pphx` template snippet), expected compiler output vs. actual output, and your OS/environment details.

### Requesting Features
We welcome feature proposals!
1. Check the [Roadmap](ROADMAP.md) to see if your feature is already planned.
2. Open a feature request issue outlining the use case, proposed syntax/behavior, and potential implementation approach.

### Submitting Pull Requests
1. **Fork the repository** and create a feature branch (`git checkout -b feature/my-cool-feature`).
2. **Write clean Go code** following standard Go formatting (`gofmt`, `go vet`).
3. **Add unit tests** in `main_test.go` covering your changes.
4. **Run compiler tests** locally:
   ```bash
   go test -v ./...
   ```
5. **Commit your changes** with descriptive commit messages.
6. **Push to your fork** and submit a Pull Request targeting the `main` branch.

---

## 3. Project Structure

*   `main.go`: Core PPHLX compiler engine, template tokenizer, component parser, and AST code generator.
*   `main_cli.go`: CLI entry points (`pphlx build`, `pphlx dev`, `pphlx watch`, `pphlx check`).
*   `main_wasm.go`: WebAssembly compilation bindings for browser-side template previews.
*   `main_test.go`: Unit tests for parser, compiler, and code generator.
*   `mcp/`: Model Context Protocol server endpoints and documentation tools.

---

## 4. Development Workflow

### Prerequisites
*   [Go](https://go.dev/) 1.20+
*   Git

### Building Locally
```bash
# Clone the repository
git clone https://github.com/pphlx/pphlx.git
cd pphlx

# Build the binary
go build -o pphlx.exe main.go main_cli.go

# Run tests
go test -v ./...
```

---

## 5. Licensing

All contributions to PPHLX are made under the [MIT License](LICENSE). By submitting a pull request, you agree that your work will be licensed under MIT.

Here is a theoretical and practical performance analysis of how your native PPHLX Go Compiler compares to Astro (Rust/esbuild/Vite/Node) when building a project of 5,360 pages.

1. The Performance Benchmark (Astro)
Astro took 2 minutes and 32 seconds (152 seconds) to build 5,360 pages.

Average build speed: 28.3ms per page.
Why it takes this long: Astro runs a heavy Node.js runtime process. For every page, it has to parse the MDX/Astro AST, spin up Vite/Rollup pipelines to handle CSS/JS assets, run internal rendering loops to execute React/Svelte components, and output the final HTML.
2. Theoretical Speed (Pure Go Compiler)
Go compiles directly to machine code and is famous for its ultra-fast compilation speeds (e.g. Hugo, a static site generator written in Go, compiles thousands of pages in milliseconds).

Average AST parsing & transpilation speed: 0.2ms to 0.5ms per page.
Concurrency: Go’s goroutines scale linearly across all CPU cores with minimal overhead.
Theoretical Calculation: $$\text{5,360 pages} \times 0.3\text{ms} = 1.6\text{ seconds}$$
Theoretical Compilation Time: ~1 to 2 seconds.
3. Practical Speed (WASM & Client-Side Islands)
In a real-world scenario, PPHLX also has to bundle client-side JS for Svelte/React islands and runs inside Node.js using WebAssembly (WASI) during local development.

WASM Overhead (1.5x - 2x native): Executing the Go-compiled WebAssembly binary adds a minor runtime translation overhead. We can estimate 1.0ms per page.
Asset Bundling (esbuild): PPHLX delegates client-side JS bundling to esbuild (which is written in Go and compiles at ~50,000 lines of JS per second).
Practical Calculation:
WASM execution & PHP output generation: ~5.3 seconds.
esbuild bundling + CSS processing: ~2.0 seconds.
File I/O operations: ~1.5 seconds.
Practical Compilation Time: ~8 to 12 seconds.
Summary Table
Metric	Astro (Rust + Node + Vite)	PPHLX (Go WASM + esbuild)	PPHLX (Native Go Binary)
5,360 Pages Build Time	152 seconds (2m 32s)	~8 to 12 seconds	~2 to 3 seconds
Avg. Time per Page	~28.3ms	~1.8ms	~0.4ms
Performance Factor	Benchmark (1x)	~15x Faster	~75x Faster
Why PPHLX (Go) is significantly faster:
No Heavy JS Runtime: Astro runs the entire build process inside a heavy V8 engine (Node.js). PPHLX executes raw compiled Go or optimized WASM bytecodes.
PHP compilation vs. SSR HTML output: PPHLX doesn't need to run full client-side framework hydrations in Node.js to render page outputs; it translates templates into simple, lightweight PHP layouts with dynamic PHP echo variables directly, which is a much simpler and faster string transformation.
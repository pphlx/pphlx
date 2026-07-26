# PPHLX Compiler Architecture: Current Execution Flow & DAG Invariants

This document outlines the current compilation pipeline in `pphlx-core/main.go` and specifies the architectural transition from **Linear Directory File Walking** to **DAG Page-Rooted Dependency Graph Traversal**.

---

## 📊 Current Go Compiler Pipeline (`pphlx-core/main.go`)

The diagram below illustrates the exact current logic executed during `pphlx build` / `pphlx dev`:

```mermaid
graph TD
    Start["Start: pphlx build / pphlx dev"] --> Config["Load pphlx.config.json & Resolve Source Root"]
    Config --> Pass1Start["Pass 1: Pre-Scan Direct Page Imports"]

    subgraph Pass1["Pass 1: Direct Import Discovery"]
        Pass1Start --> WalkPass1["Walk srcDir for .pphx and .php files"]
        WalkPass1 --> CheckImports{"Extract @import Specifiers"}
        CheckImports -->|"Find Specifier"| RegisterImport["Register Specifier Path in importedAsComponent Map"]
        CheckImports -->|"No Specifier"| NextPass1["Continue Pass 1 Directory Walk"]
    end

    Pass1 --> Pass2Start["Pass 2: Linear Directory Walk & Emission"]

    subgraph Pass2["Pass 2: File Compilation & Output Emission"]
        Pass2Start --> WalkPass2["filepath.Walk srcDir"]
        WalkPass2 --> IsDir{"Is Directory or Config File?"}
        IsDir -->|"Yes"| SkipDir["Skip / Create Output Directory"]
        IsDir -->|"No"| CheckSuppression{"Check Component Suppression Rule"}

        CheckSuppression --> IsSuppressed{"Path in importedAsComponent AND Not inside pages/?"}
        IsSuppressed -->|"True"| SuppressFile["Omit File from dist/ Emission"]
        IsSuppressed -->|"False"| CheckExt{"File Extension?"}

        CheckExt -->|"*.pphx File"| CompilePageEngine["compilePage Engine"]
        CheckExt -->|"Static Asset"| CopyAsset["copyFileIfNewer to dist/"]

        subgraph CompilePage["compilePage: Expansion & Asset Extraction"]
            CompilePageEngine --> ParseBrackets["Parse Brackets {|= ... |}"]
            ParseBrackets --> ExtractImports["Extract @import Declarations"]
            ExtractImports --> IsJsComp{"Is JS Component? React / Vue / Svelte"}
            IsJsComp -->|"Yes"| RenderJS["renderJSComponent: Island Placeholder & Props"]
            IsJsComp -->|"No"| RenderTmpl["renderTemplate: Inline Component HTML"]
            RenderTmpl --> ExpandChild["Recursively Expand Child Components"]
            ExpandChild --> AccumulateAssets["Accumulate Inline Style & Script into Global Buffers"]
        end

        CompilePage --> InjectAnchors["Inject {{PPHLX_CSS}} & {{PPHLX_JS}} Anchors"]
        InjectAnchors --> WriteFile["Write Compiled Output to dist/filename.php"]
    end

    Pass2 --> PostProcess["Post-Processing & Scaffolding"]
    PostProcess --> BundleAssets["Bundle Global CSS & JS Files"]
    BundleAssets --> ViteBuild["Trigger Vite Build if Vue / Svelte Present"]
    ViteBuild --> Complete["End Build: Complete"]
```

---

## ⚠️ Known Limitation in Current Pass 1 Pre-Scanner

```
[src/index.pphx]  ──(@import)──>  [src/layouts/Layout.pphx]  ──(@import)──>  [src/components/Head.pphx]
      │                                    │                                      │
   Scanned                              Scanned                             NOT SCANNED
 (Direct Import)                    (Direct Import)                       (Nested Child)
      │                                    │                                      │
   Marked in                            Marked in                              UNMARKED in
importedAsComponent                  importedAsComponent                  importedAsComponent
```

### 🔴 The Leak Symptom:
1. `index.pphx` imports `Layout.pphx`.
2. Pass 1 marks `Layout.pphx` in `importedAsComponent`.
3. `Layout.pphx` imports `Head.pphx`. Pass 1 **does not recursively inspect imports inside child components**.
4. During Pass 2 (`filepath.Walk`), `Head.pphx` is encountered. Because `importedAsComponent["Head.pphx"]` is `false`, the compiler misidentifies `Head.pphx` as an independent page route and writes `dist/components/Head.php`.

---

## 🎯 Proposed DAG Page-Rooted Graph Traversal Specification

To mirror the component suppression model used by Astro, Vite, and Next.js:

```mermaid
graph LR
    subgraph PageDiscovery["1. Router Entry Discovery"]
        R1["Scan src/pages/ or Entry Route"] --> P1["index.pphx (Page Root)"]
    end

    subgraph DAGTraversal["2. Recursive Graph Traversal (DAG)"]
        P1 -->|"@import"| C1["Layout.pphx (Component Node)"]
        C1 -->|"@import"| C2["Head.pphx (Component Node)"]
        C1 -->|"@import"| C3["Navbar.jsx (Island Node)"]
    end

    subgraph OutputFiltering["3. Graph Node Output Classification"]
        P1 -->|"Is Page Route = True"| OUT1["Emit dist/index.php"]
        C1 -->|"Is Page Route = False"| SUPP1["Inlined into Parent - Suppress Output"]
        C2 -->|"Is Page Route = False"| SUPP2["Inlined into Parent - Suppress Output"]
        C3 -->|"Is Page Route = False"| SUPP3["Bundled into JS - Suppress Output"]
    end
```

### 📐 Invariant Specification Rules:
1. **Entry Graph Traversal**: Recursively follow all `@import` paths starting strictly from Page Routes (`src/pages/` or `pphlx.config.json` entry file).
2. **Component Marking**: Every specifier encountered during graph traversal is marked as an **Inlined Component Node**.
3. **Strict Standalone Omission**: Files that are reachable ONLY as imported component nodes within page graphs are **100% omitted** from `dist/` standalone file output.

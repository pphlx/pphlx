# PPHLX Model Context Protocol (MCP) Specification

This specification defines the Model Context Protocol (MCP) server for the PPHLX compilation framework. It allows AI models (such as Claude, Gemini, or ChatGPT running in Cursor/VS Code) to interact with PPHLX syntax, tools, compilation states, and best practices.

---

## 1. Architectural Overview

The PPHLX MCP Server runs as a lightweight Node.js/TypeScript CLI (or built-in to the Rust/Go compiler CLI) communicating over **Standard Input/Output (stdio)** JSON-RPC 2.0.

```
+------------------+                   +--------------------+
|  AI Assistant    | <--- stdio --->   | PPHLX MCP Server   |
|  (Cursor/IDE)    |    JSON-RPC       | (Node/Go/Rust CLI) |
+------------------+                   +--------------------+
                                                 |
                                                 v
                                        +--------------------+
                                        | PPHLX Workspace    |
                                        | (Compilers & Docs) |
                                        +--------------------+
```

---

## 2. Tool Reference Registry

The server registers and exposes the following tools to the AI client:

### `pphlx/search_docs`
*   **Description**: Searches the PPHLX documentation database for syntax, lifecycle details, and routing specifications.
*   **Arguments**:
*   `query` (string, required): The search keyword or phrase (e.g., "how to import svelte component").

### `pphlx/generate_island`
*   **Description**: Generates component boilerplate code and inserts the matching `@import` statement inside a target `.pphx` template page.
*   **Arguments**:
*   `framework` (string, required): `'react' | 'vue' | 'svelte' | 'solid' | 'preact'`
*   `componentName` (string, required): PascalCase name of the component (e.g., `ShoppingBag`).
*   `targetPagePath` (string, required): Absolute path to the `.pphx` file where the import should be placed.

### `pphlx/get_best_practices`
*   **Description**: Returns official patterns for styling, cross-framework state sharing, and template variables.
*   **Arguments**:
*   `topic` (string, required): `'state-sharing' | 'styling' | 'routing' | 'php-variables'`

### `pphlx/check_compilation`
*   **Description**: Invokes the local PPHLX compiler watcher on the workspace and returns diagnostic logs to help the AI self-debug compiler errors.
*   **Arguments**:
*   `workspacePath` (string, required): Path to the PPHLX project root.

---

## 3. Boilerplate TypeScript Implementation

Below is a complete starter template for `src/index.ts` using the official `@modelcontextprotocol/sdk`.

### `package.json`
```json
{
  "name": "pphlx-mcp",
  "version": "1.0.0",
  "description": "Model Context Protocol Server for PPHLX",
  "main": "build/index.js",
  "type": "module",
  "scripts": {
    "build": "tsc",
    "start": "node build/index.js"
  },
  "dependencies": {
    "@modelcontextprotocol/sdk": "^0.6.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0"
  }
}
```

### `tsconfig.json`
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "outDir": "./build",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src/**/*"]
}
```

### `src/index.ts`
```typescript
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";

// Initialize server
const server = new Server(
  {
    name: "pphlx-mcp",
    version: "1.0.0",
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// Register Available Tools
server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools: [
      {
        name: "pphlx/search_docs",
        description: "Search the PPHLX documentation database.",
        inputSchema: {
          type: "object",
          properties: {
            query: { type: "string", description: "Search keyword" },
          },
          required: ["query"],
        },
      },
      {
        name: "pphlx/generate_island",
        description: "Generate component boilerplate and add the import to a .pphx template.",
        inputSchema: {
          type: "object",
          properties: {
            framework: { 
              type: "string", 
              enum: ["react", "vue", "svelte", "solid", "preact"] 
            },
            componentName: { type: "string", description: "PascalCase component name" },
            targetPagePath: { type: "string", description: "Absolute path to the target .pphx file" },
          },
          required: ["framework", "componentName", "targetPagePath"],
        },
      },
      {
        name: "pphlx/get_best_practices",
        description: "Get best practices guidelines on core PPHLX development topics.",
        inputSchema: {
          type: "object",
          properties: {
            topic: { 
              type: "string", 
              enum: ["state-sharing", "styling", "routing", "php-variables"] 
            },
          },
          required: ["topic"],
        },
      }
    ],
  };
});

// Handle Tool Executions
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  try {
    switch (name) {
      case "pphlx/search_docs": {
        const query = args?.query as string;
        // Mock search implementation
        return {
          content: [
            {
              type: "text",
              text: `Results for query: "${query}"\n\n1. PPHLX Templates support mixing Svelte (.svelte) and React (.jsx) components inside the same routing page.\n2. Non-React components require explicit extensions in imports (e.g., @import Counter from './Counter.vue').`
            }
          ]
        };
      }

      case "pphlx/generate_island": {
        const { framework, componentName, targetPagePath } = args as {
          framework: string;
          componentName: string;
          targetPagePath: string;
        };

        // 1. Generate Svelte/Vue/React file code...
        // 2. Append `@import ComponentName from '../components/ComponentName'` at the top of the .pphx file...

        return {
          content: [
            {
              type: "text",
              text: `Successfully created ${framework} component "${componentName}" and appended import statement to "${targetPagePath}".`
            }
          ]
        };
      }

      case "pphlx/get_best_practices": {
        const topic = args?.topic as string;
        let responseText = "";

        if (topic === "state-sharing") {
          responseText = "Best Practice: Share reactive state across separate framework islands using custom window events or micro-stores, avoiding heavy framework-specific context APIs.";
        } else if (topic === "styling") {
          responseText = "Best Practice: PPHLX supports tailwind out of the box. Scope vanilla CSS styles inside components to avoid global stylesheet conflicts.";
        } else {
          responseText = `Guidelines for topic: ${topic}. Always preserve Smarty Variable key structures when referencing PHP properties inside template blocks.`;
        }

        return {
          content: [{ type: "text", text: responseText }]
        };
      }

      default:
        throw new Error(`Tool not found: ${name}`);
    }
  } catch (error: any) {
    return {
      isError: true,
      content: [{ type: "text", text: error.message || "An error occurred" }]
    };
  }
});

// Start the server using stdio transport
async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("PPHLX MCP server running on stdio");
}

main().catch((error) => {
  console.error("Fatal error starting server:", error);
  process.exit(1);
});
```

# 🚀 Welcome to PPHLX

PPHLX is a modern, high-performance web framework. It combines the developer experience of modern JavaScript tooling (like Vite hot-reloading) with the raw performance of native PHP binaries and zero Node.js runtime in production. 

This is your minimal starter template, heavily inspired by clean modern aesthetics, designed to get you building immediately.

## 📂 Project Structure

Inside of your PPHLX project, you'll see the following folders and files:

```text
/
├── public/
│   ├── assets/
│   │   └── pphlx.svg
│   ├── favicon.ico
│   └── favicon.svg
├── layouts/
│   └── Layout.pphx
├── src/
│   └── index.pphx
├── package.json
├── pphlx.config.json
├── pphlx.json
└── README.md
```

PPHLX acts as a strict 1:1 Static Site Generator (SSG) mirroring compiler. Any `.pphx` file inside the `src` directory will be directly compiled to the `dist` directory.
- `src/index.pphx` compiles to `dist/index.html`
- `src/layouts/` and `src/assets/` contain your layout wrappers and static internal assets.

Static assets that do not need compilation (like your favicon) can be placed in the `public/` directory.

## 🧞 Commands

All commands are run from the root of the project, from a terminal:

| Command                   | Action                                           |
| :------------------------ | :----------------------------------------------- |
| `npm install`             | Installs dependencies                            |
| `npm run dev`             | Starts local dev server at `localhost:6322`      |
| `npm run build`           | Compiles your `.pphx` files into the `dist/` dir |

## 📚 Learn More

- **Documentation:** [Read our docs](https://pphlx.org)
- **Community:** [Join our Discord](http://pphlx.org/on/discord)
- **Repository:** [GitHub](https://github.com/pphlx/pphlx)

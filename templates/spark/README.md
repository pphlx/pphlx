# Spark - The PPHLX Monolith Template

Spark is a beautifully designed, high-performance monolith template built for [PPHLX](https://pphlx.org/). It provides a modern, dark-themed starting point for your next PHP application, featuring built-in Tailwind CSS, responsive components, and a modern full-stack aesthetic.

## Features

- **PPHLX Powered**: Pure, optimized PHP output with no vendor lock-in.
- **Tailwind CSS**: Pre-configured utility-first styling natively injected.
- **Modern UI**: High-end dark mode aesthetics with beautiful micro-animations, glassmorphism, and gradients.
- **Responsive Layouts**: Fully responsive navigation, hero sections, interactive features cards, and a massive immersive footer.
- **Ready to Deploy**: Designed to develop WordPress Plugins, WHMCS addons, Modern SaaS Products, Web Apps, and more.
- **SEO Optimized**: Pre-configured Open Graph tags, Twitter Cards, and Favicons.

## Quick Start

### 1. Install Dependencies

Ensure you have Node.js installed, then run:

```bash
npm install
```

### 2. Start the Development Server

Start the local PPHLX compiler and development server:

```bash
npm run dev
```
The server will start at `http://localhost:6321`.

### 3. Build for Production

To compile your `.pphx` files into pure PHP/HTML for production deployment:

```bash
npm run build
```
Your compiled assets will be available in the `dist` folder.

## Scripts

- `npm run dev`: Starts the development server with automatic file watching.
- `npm run build`: Compiles the project for production.
- `npm run preview`: Previews the production build locally.
- `npm run check`: Runs the PPHLX compiler checks.

## Project Structure

- `src/` - Contains all `.pphx` template files.
  - `layouts/` - Base wrappers and HTML head configuration (e.g., `Layout.pphx`).
  - `components/` - Reusable UI components (`Hero`, `Navbar`, `Features`, `Footer`, etc.).
  - `index.pphx` - The main entry point.
- `public/` - Static assets (favicons, social sharing graphics) served at the root `/`.
- `dist/` - The compiled output directory (generated after running build).

## Learn More

To learn more about the PPHLX framework and how to build monolithic applications, visit the [PPHLX Documentation](https://pphlx.org/).

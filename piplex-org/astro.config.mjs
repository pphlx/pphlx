// @ts-check

import mdx from '@astrojs/mdx';
import sitemap from '@astrojs/sitemap';
import tailwind from '@astrojs/tailwind';
import { defineConfig } from 'astro/config';
import astroExpressiveCode from 'astro-expressive-code';
import icon from 'astro-icon';
import fs from 'node:fs';
import houston from './houston.theme.json';

const piplexGrammar = JSON.parse(
	fs.readFileSync(
		new URL('./src/syntaxes/piplex.tmLanguage.json', import.meta.url),
		'utf-8'
	)
);
const pphxLanguage = {
	...piplexGrammar,
	name: 'pphx',
	scopeName: 'text.html.pphx',
};

/* https://docs.netlify.com/configure-builds/environment-variables/#read-only-variables */
const NETLIFY_PREVIEW_SITE = process.env.CONTEXT !== 'production' && process.env.DEPLOY_PRIME_URL;

// https://astro.build/config
export default defineConfig({
	site: NETLIFY_PREVIEW_SITE || 'https://piplex.org',
	prefetch: true,
	integrations: [
		tailwind({
			applyBaseStyles: false,
		}),
		astroExpressiveCode({
			themes: [houston],
			shiki: {
				langs: [pphxLanguage],
			},
			styleOverrides: {
				borderRadius: '0.375rem',
				borderColor: 'rgb(84 88 100)',
			},
			defaultProps: {
				overridesByLang: {
					'bash,sh,shell': {
						frame: 'none',
					},
				},
			},
		}),
		icon({
			svgoOptions: {
				plugins: [
					{ name: 'preset-default' },
					{
						name: 'prefixIds',
						// Ensure IDs used in SVGs are unique to avoid clashes between inline SVGs.
						params: { prefix: () => Math.round(Math.random() * 1_000_000_000).toString(36) },
					},
				],
			},
		}),
		mdx({ optimize: true }),
		sitemap(),
	],
	image: {
		domains: ['v1.screenshot.11ty.dev', 'storage.googleapis.com', 'avatars.githubusercontent.com'],
	},
	experimental: {
		contentIntellisense: true,
	},
});

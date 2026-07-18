import type { ThemeAndAuthor, ThemeCategory, ThemeTool } from '../_types/index.ts';

export const themeCategories: ThemeCategory[] = [
	{ id: 1, value: 'blog', name: 'Blogs' },
	{ id: 2, value: 'ecommerce', name: 'E-commerce' },
	{ id: 3, value: 'landing-page', name: 'Landing pages' },
	{ id: 4, value: 'portfolio', name: 'Portfolios' },
	{ id: 5, value: 'docs', name: 'Documentation' }
];

export const themeTools: ThemeTool[] = [
	{ id: 1, value: 'react', name: 'React' },
	{ id: 2, value: 'vue', name: 'Vue' },
	{ id: 3, value: 'svelte', name: 'Svelte' },
	{ id: 4, value: 'solid', name: 'Solid' },
	{ id: 5, value: 'preact', name: 'Preact' },
	{ id: 6, value: 'tailwind', name: 'Tailwind' }
];

export const allThemes: ThemeAndAuthor[] = [
	{
		Theme: {
			id: 1,
			slug: 'pphlx-blog-theme',
			title: 'PPHLX Blog Starter',
			featured: false,
			description: 'A modern, clean blog theme for PPHLX.',
			fullDescription: 'A modern, clean blog theme for PPHLX built with Svelte and Tailwind CSS.',
			body: 'This is a sample blog theme designed to work out-of-the-box with PPHLX.',
			image: '/assets/themes/blog.webp',
			images: [],
			authorId: 1,
			paid: false,
			stars: 12,
			publishDate: new Date('2026-07-20'),
			approved: true,
			denied: false,
			hidden: true, // Sample theme - hidden from public view
			price: 0,
			sellingThroughPortal: false,
			links: []
		},
		Author: {
			id: 1,
			name: 'PPHLX Team',
			role: 'maintainer',
			githubId: 12345,
			username: 'pphlx',
			updatedAt: new Date('2026-07-20'),
			createdAt: new Date('2026-07-20')
		}
	},
	{
		Theme: {
			id: 2,
			slug: 'pphlx-ecommerce-theme',
			title: 'PPHLX E-commerce Starter',
			featured: false,
			description: 'A high-performance e-commerce theme for PPHLX.',
			fullDescription: 'A high-performance e-commerce theme built with React and Tailwind CSS.',
			body: 'This is a sample e-commerce theme designed to work out-of-the-box with PPHLX.',
			image: '/assets/themes/ecommerce.webp',
			images: [],
			authorId: 1,
			paid: true,
			stars: 25,
			publishDate: new Date('2026-07-20'),
			approved: true,
			denied: false,
			hidden: true, // Sample theme - hidden from public view
			price: 2900,
			sellingThroughPortal: false,
			links: []
		},
		Author: {
			id: 1,
			name: 'PPHLX Team',
			role: 'maintainer',
			githubId: 12345,
			username: 'pphlx',
			updatedAt: new Date('2026-07-20'),
			createdAt: new Date('2026-07-20')
		}
	}
];

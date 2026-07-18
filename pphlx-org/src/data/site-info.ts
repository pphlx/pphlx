export type SocialLink = {
	/** Longer descriptive label, e.g. `"Join the PPHLX community on Discord"` */
	text: string;
	/** Short label with the name of the platform, e.g. `"Discord"`*/
	label: string;
	/** Icon name for use with `astro-icon`, e.g. `"social/discord"`. */
	icon: string;
	/** URL for our profile on the external platform. */
	href: string;
	/** Platform ID, e.g. `"discord"`. Used for `pphlx.org/on/PLATFORM` redirects. */
	platform: string;
	/** Whether this platform should be linked in the site header */
	showInHeader?: boolean;
};

type SiteInfo = {
	name: string;
	title: string;
	description: string;
	image: {
		src: string;
		alt: string;
	};
	socialLinks: SocialLink[];
};

const siteInfo: SiteInfo = {
	name: 'PPHLX',
	title: 'Fast, Zero-Dependency Multi-Framework PHP Template Compiler',
	description:
		'PPHLX compiles modern, component-based PHP templates (.pphx) into standard, standalone, production-ready .php files.',
	image: {
		src: '/og/social.jpg',
		alt: 'Build the web you want',
	},
	socialLinks: [
		{
			platform: 'bluesky',
			icon: 'social/bluesky',
			label: 'Bluesky',
			text: 'Follow PPHLX on Bluesky',
			href: 'https://bsky.app/profile/pphlx.org',
		},
		{
			platform: 'discord',
			href: '/chat',
			icon: 'social/discord',
			label: 'Discord',
			text: 'Join the PPHLX community on Discord',
		},
		{
			platform: 'github',
			icon: 'social/github',
			label: 'GitHub',
			text: "Go to PPHLX's GitHub repo",
			href: 'https://github.com/pphlx/pphlx',
			showInHeader: true,
		},
		{
			platform: 'linkedin',
			icon: 'social/linkedin',
			label: 'LinkedIn',
			text: 'Follow PPHLX on LinkedIn',
			href: 'https://www.linkedin.com/company/pphlx',
		},
		{
			platform: 'mastodon',
			icon: 'social/mastodon',
			label: 'Mastodon',
			text: 'Follow PPHLX on Mastodon',
			href: 'https://mastodon.social/@pphlx',
		},
		{
			platform: 'reddit',
			icon: 'social/reddit',
			label: 'Reddit',
			text: 'Join the official PPHLX community on Reddit',
			href: 'https://www.reddit.com/r/PPHLX/',
		},
		{
			platform: 'twitter',
			icon: 'social/twitter',
			href: 'https://x.com/pphlxlabs',
			label: 'X.com',
			text: 'Follow PPHLX on x.com (formerly Twitter)',
		},
		{
			platform: 'youtube',
			icon: 'social/youtube',
			href: 'https://www.youtube.com/@pphlxlabs',
			label: 'YouTube',
			text: 'Follow PPHLX on YouTube',
		},
	],
};

export default siteInfo;

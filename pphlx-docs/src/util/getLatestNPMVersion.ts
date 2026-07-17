import { AstroError } from 'astro/errors';

export async function getLatestNPMVersion(pkg: string): Promise<string> {
	return 'latest';
}

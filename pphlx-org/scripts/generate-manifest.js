import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const playbooksDir = path.resolve(__dirname, '../public/playbooks');
const manifestPath = path.join(playbooksDir, 'manifest.json');

function getFilesRecursively(dir, baseDir) {
	let results = [];
	const list = fs.readdirSync(dir);
	for (const file of list) {
		const filePath = path.join(dir, file);
		const stat = fs.statSync(filePath);
		if (stat && stat.isDirectory()) {
			if (file !== 'node_modules') {
				results = results.concat(getFilesRecursively(filePath, baseDir));
			}
		} else {
			// Skip playbook.json and system/hidden files
			if (file !== 'playbook.json' && !file.startsWith('.')) {
				const relative = path.relative(baseDir, filePath).replace(/\\/g, '/');
				if (!relative.includes('dist/assets/js/assets/') && file !== 'package.json' && file !== 'package-lock.json') {
					results.push(relative);
				}
			}
		}
	}
	return results;
}

try {
	console.log('Generating playbooks manifest.json...');
	const items = fs.readdirSync(playbooksDir);
	const playbooks = [];

	for (const item of items) {
		const itemPath = path.join(playbooksDir, item);
		if (fs.statSync(itemPath).isDirectory()) {
			const configPath = path.join(itemPath, 'playbook.json');
			if (fs.existsSync(configPath)) {
				const config = JSON.parse(fs.readFileSync(configPath, 'utf-8'));
				const files = getFilesRecursively(itemPath, itemPath);
				
				playbooks.push({
					id: item,
					name: config.name || item,
					description: config.description || '',
					defaultFile: config.defaultFile || 'index.pphx',
					files: files
				});
			}
		}
	}

	const manifest = { playbooks };
	fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2), 'utf-8');
	console.log(`Successfully generated ${manifestPath} with ${playbooks.length} playbooks!`);
} catch (error) {
	console.error('Failed to generate playbooks manifest:', error);
	process.exit(1);
}

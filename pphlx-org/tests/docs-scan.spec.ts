import { test, expect } from '@playwright/test';

test('Scan all documentation pages for leftover Astro references', async ({ page }) => {
  test.setTimeout(120000); // Allow up to 2 minutes to crawl 180 pages
  const docsBaseURL = 'http://localhost:4322';
  
  // 1. Go to the starting documentation page
  await page.goto(`${docsBaseURL}/en/getting-started/`);
  
  // 2. Extract all English documentation links from the sidebar
  const links = await page.evaluate((baseUrl) => {
    const anchors = Array.from(document.querySelectorAll('a'));
    return anchors
      .map(a => a.getAttribute('href'))
      .filter((href): href is string => !!href && href.startsWith('/en/'))
      .map(href => {
        // Resolve absolute URL
        return href.startsWith('http') ? href : baseUrl + href;
      });
  }, docsBaseURL);
  
  // De-duplicate links
  const uniqueLinks = Array.from(new Set(links));
  console.log(`\nFound ${uniqueLinks.length} unique English documentation pages to scan.`);
  
  const failures: { url: string; match: string; context: string }[] = [];
  
  // 3. Visit each page and scan for Astro references
  for (const url of uniqueLinks) {
    try {
      await page.goto(url, { waitUntil: 'domcontentloaded' });
    } catch (e) {
      console.warn(`Could not load page: ${url}`);
      continue;
    }
    
    // Get text content of the main article/content area
    const contentArea = page.locator('main').first();
    if (await contentArea.count() === 0) continue;
    
    const text = await contentArea.innerText();
    
    // 1. check for "@astrojs" (ignoring starlight import instructions if they are required)
    if (text.includes('@astrojs') && !text.includes('@astrojs/starlight') && !text.includes('@astrojs/sitemap')) {
      failures.push({ url, match: '@astrojs', context: '@astrojs package reference found' });
    }
    // 2. check for "getViteConfig"
    if (text.includes('getViteConfig')) {
      failures.push({ url, match: 'getViteConfig', context: 'getViteConfig function reference found' });
    }
    // 3. check for "AstroContainer"
    if (text.includes('AstroContainer')) {
      failures.push({ url, match: 'AstroContainer', context: 'AstroContainer reference found' });
    }
    // 4. check for "astro.config"
    if (text.includes('astro.config')) {
      failures.push({ url, match: 'astro.config', context: 'astro.config reference found' });
    }
    
    // 5. check for general "astro" mentions in code blocks (e.g. `import ... from 'astro'`)
    const codeBlocks = await page.locator('pre, code').allInnerTexts();
    for (const code of codeBlocks) {
      // Find 'astro' as a whole word, excluding starlight imports or config mentions we allowed
      const hasAstroImport = /(?:\bfrom\s+['"]astro['"]|\bimport\s+['"]astro['"])/i.test(code);
      const hasAstroGlobal = /\bAstro\.(?:generator|glob|props|request|redirect|resolve|slots|self|site|locals)\b/.test(code);
      
      if (hasAstroImport || hasAstroGlobal) {
        failures.push({ url, match: 'Astro API/Import in code block', context: code.slice(0, 150) });
      }
    }
  }
  
  if (failures.length > 0) {
    console.error('\n--- FAILURES DETECTED IN DOCUMENTATION PAGES ---');
    failures.forEach(f => {
      console.error(`- Page: ${f.url}\n  Match: ${f.match}\n  Context: ${f.context}\n`);
    });
    expect(failures.length).toBe(0);
  } else {
    console.log('\nAll scanned documentation pages are clean of leftover Astro code blocks and imports!');
  }
});

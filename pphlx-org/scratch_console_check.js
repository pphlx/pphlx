const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  
  page.on('console', msg => {
    console.log(`[BROWSER CONSOLE] ${msg.type()}: ${msg.text()}`);
  });
  
  page.on('pageerror', err => {
    console.log(`[BROWSER ERROR] ${err.message}`);
  });

  try {
    console.log("Navigating to http://localhost:8000/ ...");
    await page.goto('http://localhost:8000/', { waitUntil: 'networkidle' });
    console.log("Navigation complete. Waiting 3 seconds for hydration...");
    await new Promise(resolve => setTimeout(resolve, 3000));
  } catch (e) {
    console.error("Navigation error:", e);
  } finally {
    await browser.close();
  }
})();

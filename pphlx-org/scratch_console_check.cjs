const { chromium } = require('playwright');
const path = require('path');

(async () => {
  const browser = await chromium.launch({ 
    headless: true,
    channel: 'chrome',
    viewport: { width: 1280, height: 800 }
  });
  const page = await browser.newPage();
  
  page.on('console', msg => {
    console.log(`[MAIN CONSOLE] ${msg.type()}: ${msg.text()}`);
  });
  
  page.on('pageerror', err => {
    console.log(`[MAIN ERROR] ${err.message}`);
  });

  try {
    console.log("Navigating to http://localhost:4321/play/ ...");
    await page.goto('http://localhost:4321/play/', { waitUntil: 'domcontentloaded' });
    console.log("Waiting 5 seconds for page assets to load...");
    await new Promise(resolve => setTimeout(resolve, 5000));

    console.log("Selecting 'multiframe-dashboard' playbook...");
    await page.selectOption('#playbook-select', 'multiframe-dashboard');
    console.log("Waiting 3 seconds for playbook VFS to load...");
    await new Promise(resolve => setTimeout(resolve, 3000));

    console.log("Clicking Reset Code...");
    await page.click('#btn-reset-code');
    console.log("Waiting 1 second for custom confirm modal to appear...");
    await new Promise(resolve => setTimeout(resolve, 1000));

    console.log("Clicking Confirm button inside the custom modal...");
    await page.click('#ide-modal-btn-confirm');
    console.log("Waiting 3 seconds after VFS reset...");
    await new Promise(resolve => setTimeout(resolve, 3000));

    console.log("Clicking HTML Preview Tab...");
    await page.click('#tab-preview');
    console.log("Waiting 6 seconds for iframe rendering and hydration...");
    await new Promise(resolve => setTimeout(resolve, 6000));

    // Take screenshot and save to artifacts directory
    const screenshotPath = 'C:\\Users\\KillerTyzon\\.gemini\\antigravity-ide\\brain\\fea78b21-b284-4aa1-8fe3-d2738e64934a\\playbook_preview.png';
    await page.screenshot({ path: screenshotPath });
    console.log(`Screenshot saved to: ${screenshotPath}`);

  } catch (e) {
    console.error("Test error:", e);
  } finally {
    await browser.close();
  }
})();

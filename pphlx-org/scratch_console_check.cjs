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
    console.log(`[MAIN ERROR] ${err.stack || err.message}`);
  });

  try {
    console.log("Navigating to http://localhost:4321/play/ ...");
    await page.goto('http://localhost:4321/play/', { waitUntil: 'domcontentloaded' });
    await new Promise(resolve => setTimeout(resolve, 5000));

    console.log("Selecting 'ecommerce' playbook...");
    await page.selectOption('#playbook-select', 'ecommerce');
    await new Promise(resolve => setTimeout(resolve, 3000));

    console.log("Clicking Reset Code...");
    await page.click('#btn-reset-code');
    await new Promise(resolve => setTimeout(resolve, 1000));
    await page.click('#ide-modal-btn-confirm');
    await new Promise(resolve => setTimeout(resolve, 3000));

    // Get the Monaco output and compiled preview HTML
    const compilerData = await page.evaluate(() => {
      const iframe = document.getElementById('preview-iframe');
      const doc = iframe.contentDocument || iframe.contentWindow.document;
      return {
        phpOutput: window.inputEditor ? window.inputEditor.getValue() : 'NO EDITOR',
        previewHtml: doc ? doc.documentElement.outerHTML : 'NO DOC'
      };
    });

    console.log("--- MONACO INPUT EDITOR ---");
    console.log(compilerData.phpOutput.substring(0, 1000));
    console.log("--- PROCESSED PREVIEW IFRAME HTML ---");
    // Find all script tags containing pphlxProps
    const scriptRegex = /<script>([\s\S]*?)<\/script>/g;
    let match;
    while ((match = scriptRegex.exec(compilerData.previewHtml)) !== null) {
      if (match[1].includes('pphlxProps')) {
        console.log("Found Script Tag:\n", match[1].trim());
      }
    }

  } catch (e) {
    console.error("Test error:", e);
  } finally {
    await browser.close();
  }
})();

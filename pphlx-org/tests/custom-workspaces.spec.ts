import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test('local workspace and copy path workflow', async ({ page, context }) => {
  // Grant clipboard read/write permissions
  await context.grantPermissions(['clipboard-read', 'clipboard-write']);

  // Go to playground page
  await page.goto('/play/');
  page.on('console', msg => console.log('PAGE LOG 1:', msg.text()));
  page.on('pageerror', err => console.log('PAGE ERROR 1:', err.message));

  // 1. Verify three-dots "More Actions..." button is visible
  const menuBtn = page.locator('#btn-playbook-menu');
  await expect(menuBtn).toBeVisible({ timeout: 30000 });
  await menuBtn.click();

  // 2. Verify context menu opens with "New Local Workspace..." and "Delete Local Workspace" is hidden
  const newWsBtn = page.locator('#pb-ctx-new');
  await expect(newWsBtn).toBeVisible();
  
  const deleteWsBtn = page.locator('#pb-ctx-delete');
  await expect(deleteWsBtn).toBeHidden();

  // 3. Click "New Local Workspace...", enter name in the inline select input, and press Enter
  await newWsBtn.click();
  const selectInput = page.locator('#playbook-select-input');
  await expect(selectInput).toBeVisible();
  await selectInput.fill('Demo App');
  await selectInput.press('Enter');

  // 4. Verify dropdown selects "Demo App"
  const select = page.locator('#playbook-select');
  await expect(select).toHaveValue(/custom_ws_/);

  // 5. Verify file tree loads index.pphx and Layout.pphx
  const indexFile = page.locator('.font-mono:has-text("index.pphx")').first();
  await expect(indexFile).toBeVisible();

  // 6. Right click on index.pphx to show context menu and click Copy Path
  await indexFile.click({ button: 'right' });
  const copyPathBtn = page.locator('#ctx-copy-path');
  await expect(copyPathBtn).toBeVisible();
  await copyPathBtn.click();

  // 7. Verify dynamic copied path
  const clipboardText = await page.evaluate(async () => {
    return await navigator.clipboard.readText();
  });
  console.log('Copied Path from VFS:', clipboardText);
  expect(clipboardText).toContain('browser\\local_workspaces\\Demo App\\index.pphx');

  // 8. Delete local workspace
  await menuBtn.click();
  await expect(deleteWsBtn).toBeVisible();
  await deleteWsBtn.click();

  // Confirm delete
  await page.locator('#ide-modal-btn-confirm').click();

  // 9. Verify fallback to ecommerce
  await expect(select).toHaveValue('ecommerce');
});

test('sidebar resizer and detached preview sync workflow', async ({ page, context }) => {
  await page.goto('/play/');
  page.on('console', msg => console.log('PAGE LOG 2:', msg.text()));
  page.on('pageerror', err => console.log('PAGE ERROR 2:', err.message));

  // Wait for initial playbook compilation to complete
  await expect(page.locator('#compile-time')).not.toHaveText('Compiled in 0ms', { timeout: 30000 });

  // 1. Test star rating click functionality inside the main preview iframe (if CDNs are accessible)
  // Switch to preview tab
  await page.locator('#tab-preview').click();
  const mainIframe = page.frameLocator('#preview-iframe');
  
  // Check if SolidJS component is successfully hydrated (meaning buttons are rendered)
  const isHydrated = await mainIframe.locator('button').first().waitFor({ state: 'attached', timeout: 2000 }).then(() => true).catch(() => false);

  if (isHydrated) {
    const mainStarButtons = mainIframe.locator('.my-3 button');
    await expect(mainStarButtons).toHaveCount(5, { timeout: 10000 });
    await mainStarButtons.nth(2).click();
    const mainThirdStarSvg = mainStarButtons.nth(2).locator('svg');
    const mainFourthStarSvg = mainStarButtons.nth(3).locator('svg');
    await expect(mainThirdStarSvg).toHaveAttribute('fill', '#4bf3c8');
    await expect(mainFourthStarSvg).toHaveAttribute('fill', 'none');
  } else {
    console.log('Skipping main preview star click tests because unpkg.com CDN is blocked/offline in the test environment.');
  }

  // 2. Test detached live preview window sync and click functionality
  // Set up promise to wait for popout page opening
  const pagePromise = context.waitForEvent('page');
  await page.locator('#sim-btn-popout').click();
  const popoutPage = await pagePromise;
  await popoutPage.waitForLoadState();

  // Verify the popout url has ?detached=true
  expect(popoutPage.url()).toContain('detached=true');

  // Verify header and workspace elements are hidden in detached layout
  await expect(popoutPage.locator('#ide-header')).toBeHidden();
  await expect(popoutPage.locator('#activity-bar-parent')).toBeHidden();
  await expect(popoutPage.locator('#explorer-sidebar')).toBeHidden();
  await expect(popoutPage.locator('#editor-workspace')).toBeHidden();
  await expect(popoutPage.locator('#preview-pane')).toBeVisible();

  // Verify sync worked - the iframe inside popout should have content
  const popoutIframe = popoutPage.frameLocator('#preview-iframe');
  const popoutBody = popoutIframe.locator('body');
  await expect(popoutBody).toBeVisible();

  if (isHydrated) {
    // Verify star rating click works inside the popout page iframe
    const popoutStarButtons = popoutIframe.locator('.my-3 button');
    await expect(popoutStarButtons).toHaveCount(5, { timeout: 10000 });
    await popoutStarButtons.nth(1).click();
    const popoutSecondStarSvg = popoutStarButtons.nth(1).locator('svg');
    const popoutThirdStarSvg = popoutStarButtons.nth(2).locator('svg');
    await expect(popoutSecondStarSvg).toHaveAttribute('fill', '#4bf3c8');
    await expect(popoutThirdStarSvg).toHaveAttribute('fill', 'none');
  } else {
    console.log('Skipping popout preview star click tests because unpkg.com CDN is blocked/offline in the test environment.');
  }

  await popoutPage.close();

  // 3. Verify resizer elements are visible and perform dragging
  const sidebarResizer = page.locator('#sidebar-resizer');
  await expect(sidebarResizer).toBeVisible();

  const workspaceResizer = page.locator('#workspace-resizer');
  await expect(workspaceResizer).toBeVisible();

  const sidebar = page.locator('#explorer-sidebar');
  const initialBoundingBox = await sidebar.boundingBox();
  expect(initialBoundingBox).not.toBeNull();
  const initialWidth = initialBoundingBox!.width;

  // Drag the resizer right by 50 pixels
  const resizerBox = await sidebarResizer.boundingBox();
  expect(resizerBox).not.toBeNull();
  await page.mouse.move(resizerBox!.x + resizerBox!.width / 2, resizerBox!.y + resizerBox!.height / 2);
  await page.mouse.down();
  await page.mouse.move(resizerBox!.x + 50, resizerBox!.y + resizerBox!.height / 2);
  await page.mouse.up();

  const newBoundingBox = await sidebar.boundingBox();
  expect(newBoundingBox).not.toBeNull();
  console.log(`Resized Sidebar from ${initialWidth}px to ${newBoundingBox!.width}px`);
  expect(newBoundingBox!.width).toBeGreaterThan(initialWidth);
});


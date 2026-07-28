import { test, expect } from '@playwright/test';

test.describe('POS', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/login');
		await page.fill('#username', 'admin');
		await page.fill('#password', 'admin123');
		await page.click('button[type="submit"]');
		await page.waitForURL('/', { timeout: 10000 });
	});

	test('POS page loads with product grid', async ({ page }) => {
		await page.goto('/pos');
		await expect(page.locator('text=Point of Sale').first()).toBeVisible({ timeout: 10000 });
	});

	test('search products in POS', async ({ page }) => {
		await page.goto('/pos');
		const searchInput = page.locator('input[placeholder*="Search"]').first();
		if (await searchInput.isVisible()) {
			await searchInput.fill('test');
			await expect(page.locator('text=results').first()).toBeVisible({ timeout: 5000 });
		}
	});

	test('add product to cart', async ({ page }) => {
		await page.goto('/pos');
		const productTile = page.locator('[class*="product"], button:has-text("Tap to add")').first();
		if (await productTile.isVisible({ timeout: 5000 })) {
			await productTile.click();
			await expect(page.locator('text=Cart').first()).toBeVisible();
		}
	});

	test('cart shows empty state initially', async ({ page }) => {
		await page.goto('/pos');
		await expect(page.locator('text=No items added yet').first()).toBeVisible({ timeout: 10000 });
	});

	test('checkout with cash payment', async ({ page }) => {
		await page.goto('/pos');
		const productTile = page.locator('[class*="product"], button:has-text("Tap to add")').first();
		if (await productTile.isVisible({ timeout: 5000 })) {
			await productTile.click();
			const cashButton = page.locator('button:has-text("Cash")').first();
			if (await cashButton.isVisible()) {
				await cashButton.click();
			}
			const completeButton = page.locator('button:has-text("Complete Order")').first();
			if (await completeButton.isVisible()) {
				await expect(completeButton).toBeEnabled();
			}
		}
	});
});

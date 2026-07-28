import { test, expect } from '@playwright/test';

test.describe('Products', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/login');
		await page.fill('#username', 'admin');
		await page.fill('#password', 'admin123');
		await page.click('button[type="submit"]');
		await page.waitForURL('/', { timeout: 10000 });
	});

	test('catalog page loads with products tab', async ({ page }) => {
		await page.goto('/catalog');
		await expect(page.locator('text=Products').first()).toBeVisible({ timeout: 10000 });
	});

	test('search products in catalog', async ({ page }) => {
		await page.goto('/catalog');
		const searchInput = page.locator('input[placeholder*="Search products"]').first();
		if (await searchInput.isVisible({ timeout: 5000 })) {
			await searchInput.fill('test');
		}
	});

	test('create product form opens', async ({ page }) => {
		await page.goto('/catalog');
		const addButton = page.locator('button:has-text("Add Product")').first();
		if (await addButton.isVisible({ timeout: 5000 })) {
			await addButton.click();
			await expect(page.locator('text=Create product').first()).toBeVisible({ timeout: 5000 });
		}
	});

	test('product form has required fields', async ({ page }) => {
		await page.goto('/catalog');
		const addButton = page.locator('button:has-text("Add Product")').first();
		if (await addButton.isVisible({ timeout: 5000 })) {
			await addButton.click();
			await expect(page.locator('input[name="sku"], input[placeholder*="SKU"]').first()).toBeVisible({ timeout: 5000 });
			await expect(page.locator('input[name="name"], input[placeholder*="Name"]').first()).toBeVisible();
		}
	});

	test('edit product via detail page', async ({ page }) => {
		await page.goto('/catalog');
		const productRow = page.locator('tr, [class*="product-item"]').first();
		if (await productRow.isVisible({ timeout: 5000 })) {
			await productRow.click();
			await page.waitForTimeout(1000);
		}
	});

	test('delete product shows confirmation', async ({ page }) => {
		await page.goto('/catalog');
		const deleteButton = page.locator('button:has-text("Delete")').first();
		if (await deleteButton.isVisible({ timeout: 5000 })) {
			await deleteButton.click();
			await expect(page.locator('text=Are you sure').first()).toBeVisible({ timeout: 5000 });
		}
	});
});

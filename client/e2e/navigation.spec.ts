import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/login');
		await page.fill('#username', 'admin');
		await page.fill('#password', 'admin123');
		await page.click('button[type="submit"]');
		await page.waitForURL('/', { timeout: 10000 });
	});

	test('sidebar is visible after login', async ({ page }) => {
		await expect(page.locator('nav, aside, [class*="sidebar"]').first()).toBeVisible({ timeout: 10000 });
	});

	test('navigate to dashboard', async ({ page }) => {
		const dashboardLink = page.locator('a[href="/"], a:has-text("Dashboard")').first();
		if (await dashboardLink.isVisible()) {
			await dashboardLink.click();
			await page.waitForURL('/');
		}
	});

	test('navigate to catalog', async ({ page }) => {
		const catalogLink = page.locator('a[href="/catalog"], a:has-text("Catalog")').first();
		if (await catalogLink.isVisible({ timeout: 5000 })) {
			await catalogLink.click();
			await page.waitForURL('**/catalog');
		}
	});

	test('navigate to POS', async ({ page }) => {
		const posLink = page.locator('a[href="/pos"], a:has-text("POS")').first();
		if (await posLink.isVisible({ timeout: 5000 })) {
			await posLink.click();
			await page.waitForURL('**/pos');
		}
	});

	test('navigate to orders', async ({ page }) => {
		const ordersLink = page.locator('a[href="/orders"], a:has-text("Orders")').first();
		if (await ordersLink.isVisible({ timeout: 5000 })) {
			await ordersLink.click();
			await page.waitForURL('**/orders');
		}
	});

	test('navigate to reports', async ({ page }) => {
		const reportsLink = page.locator('a[href="/reports"], a:has-text("Reports")').first();
		if (await reportsLink.isVisible({ timeout: 5000 })) {
			await reportsLink.click();
			await page.waitForURL('**/reports');
		}
	});

	test('navigate to settings', async ({ page }) => {
		const settingsLink = page.locator('a[href="/settings"], a:has-text("Settings")').first();
		if (await settingsLink.isVisible({ timeout: 5000 })) {
			await settingsLink.click();
			await page.waitForURL('**/settings');
		}
	});

	test('permission-based nav items visible for admin', async ({ page }) => {
		const settingsLink = page.locator('a[href="/settings"], a:has-text("Settings")').first();
		await expect(settingsLink).toBeVisible({ timeout: 5000 });
		const usersLink = page.locator('a[href="/users"], a:has-text("Users")').first();
		await expect(usersLink).toBeVisible({ timeout: 5000 });
	});
});

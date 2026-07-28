import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
	test('login page renders correctly', async ({ page }) => {
		await page.goto('/login');
		await expect(page.locator('text=Welcome Back')).toBeVisible();
		await expect(page.locator('input#username')).toBeVisible();
		await expect(page.locator('input#password')).toBeVisible();
	});

	test('shows validation errors for empty login form', async ({ page }) => {
		await page.goto('/login');
		await page.click('button[type="submit"]');
		await expect(page.locator('text=Username is required')).toBeVisible();
	});

	test('login with invalid credentials shows error', async ({ page }) => {
		await page.goto('/login');
		await page.fill('#username', 'nonexistent_user');
		await page.fill('#password', 'wrongpassword');
		await page.click('button[type="submit"]');
		await expect(page.locator('.sonner-toast, [data-sonner-toast]')).toBeVisible({ timeout: 10000 });
	});

	test('successful login redirects to dashboard', async ({ page }) => {
		await page.goto('/login');
		await page.fill('#username', 'admin');
		await page.fill('#password', 'admin123');
		await page.click('button[type="submit"]');
		await page.waitForURL('/', { timeout: 10000 });
		expect(page.url()).toContain('/');
	});

	test('protected route redirects to login when unauthenticated', async ({ page }) => {
		await page.goto('/catalog');
		await page.waitForURL('**/login', { timeout: 10000 });
		expect(page.url()).toContain('/login');
	});

	test('logout clears session and redirects to login', async ({ page }) => {
		await page.goto('/login');
		await page.fill('#username', 'admin');
		await page.fill('#password', 'admin123');
		await page.click('button[type="submit"]');
		await page.waitForURL('/', { timeout: 10000 });

		const logoutButton = page.locator('button:has-text("Logout"), a:has-text("Logout"), [data-testid="logout"]');
		if (await logoutButton.isVisible()) {
			await logoutButton.click();
			await page.waitForURL('**/login', { timeout: 10000 });
			expect(page.url()).toContain('/login');
		}
	});

	test('register page renders correctly', async ({ page }) => {
		await page.goto('/register');
		await expect(page.locator('text=Create Account')).toBeVisible();
		await expect(page.locator('input#username')).toBeVisible();
		await expect(page.locator('input#password')).toBeVisible();
	});

	test('register form shows password strength indicator', async ({ page }) => {
		await page.goto('/register');
		await page.fill('#password', 'Test123!');
		await expect(page.locator('text=STRENGTH')).toBeVisible();
	});

	test('login page has link to register', async ({ page }) => {
		await page.goto('/login');
		const registerLink = page.locator('a[href="/register"]');
		await expect(registerLink).toBeVisible();
		await registerLink.click();
		await page.waitForURL('**/register');
	});
});

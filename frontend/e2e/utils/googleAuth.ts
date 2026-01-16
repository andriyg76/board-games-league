import { Page, expect } from '@playwright/test';

type Credentials = {
  email: string;
  password: string;
};

const clickIfVisible = async (page: Page, label: string) => {
  const button = page.getByRole('button', { name: label });
  if (await button.isVisible().catch(() => false)) {
    await button.click();
    return true;
  }
  return false;
};

export const loginWithGoogle = async (
  page: Page,
  baseURL: string,
  credentials: Credentials
) => {
  const origin = new URL(baseURL).origin;

  await page.goto(baseURL, { waitUntil: 'domcontentloaded' });

  const alreadyLoggedIn = await page
    .getByTestId('auth-logout-button')
    .isVisible()
    .catch(() => false);
  if (alreadyLoggedIn) {
    return;
  }

  await page.getByTestId('auth-login-button').click();
  await page.getByText('Login with Google', { exact: true }).click();

  await page.waitForURL(/accounts\.google\.com/);

  const useAnotherAccount = page.getByText('Use another account');
  if (await useAnotherAccount.isVisible().catch(() => false)) {
    await useAnotherAccount.click();
  }

  const emailInput = page.locator('input[type="email"]');
  await emailInput.waitFor({ timeout: 60000 });
  await emailInput.fill(credentials.email);
  await page.getByRole('button', { name: /Next/i }).click();

  const passwordInput = page.locator('input[type="password"]');
  await passwordInput.waitFor({ timeout: 60000 });
  await passwordInput.fill(credentials.password);
  await page.getByRole('button', { name: /Next/i }).click();

  await clickIfVisible(page, 'Continue');
  await clickIfVisible(page, 'Allow');

  await page.waitForURL(
    (url) => url.origin === origin && url.pathname.startsWith('/ui'),
    { timeout: 120000 }
  );

  await expect(page.getByTestId('auth-logout-button')).toBeVisible({
    timeout: 60000,
  });
};

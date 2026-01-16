import fs from 'fs';
import { chromium, request, type FullConfig } from '@playwright/test';
import { authDir, adminStatePath, inviteeStatePath } from './utils/paths';
import {
  getAdminCredentials,
  getBaseURL,
  getInviteeCredentials,
  getInviteeEmail,
} from './utils/env';
import { loginWithGoogle } from './utils/googleAuth';

const ensureInviteeUser = async (baseURL: string, inviteeEmail: string) => {
  const apiContext = await request.newContext({
    baseURL,
    storageState: adminStatePath,
  });

  const response = await apiContext.put('/api/admin/user/create', {
    data: { external_ids: [inviteeEmail] },
  });

  if (![200, 201].includes(response.status())) {
    throw new Error(
      `Invitee user create failed: ${response.status()} ${response.statusText()}`
    );
  }

  await apiContext.dispose();
};

const globalSetup = async (config: FullConfig) => {
  const baseURL = process.env.E2E_BASE_URL || getBaseURL();
  const adminCredentials = getAdminCredentials();
  const inviteeCredentials = getInviteeCredentials();
  const inviteeEmail = getInviteeEmail();

  fs.mkdirSync(authDir, { recursive: true });

  const browser = await chromium.launch();

  const adminContext = await browser.newContext();
  const adminPage = await adminContext.newPage();
  await loginWithGoogle(adminPage, baseURL, adminCredentials);
  await adminContext.storageState({ path: adminStatePath });

  if (inviteeEmail) {
    await ensureInviteeUser(baseURL, inviteeEmail);
  }

  if (inviteeCredentials) {
    const inviteeContext = await browser.newContext();
    const inviteePage = await inviteeContext.newPage();
    await loginWithGoogle(inviteePage, baseURL, inviteeCredentials);
    await inviteeContext.storageState({ path: inviteeStatePath });
    await inviteeContext.close();
  }

  await adminContext.close();
  await browser.close();
};

export default globalSetup;

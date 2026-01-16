import fs from 'fs';
import { test, expect } from '@playwright/test';
import { adminStatePath, inviteeStatePath } from '../utils/paths';
import { buildName, buildRunId } from '../utils/testData';

test.describe.serial('UI backend E2E flow', () => {
  test.use({ storageState: adminStatePath });

  const runId = buildRunId();
  const leagueName = buildName('E2E League', runId);
  const inviteeAlias = `E2EInvitee-${runId}`;
  const virtualAlias = `E2EVirtual-${runId}`;
  const roundName = buildName('E2E Round', runId);

  let leagueCode = '';
  let invitationLink = '';

  test('admin creates league and invitation', async ({ page }) => {
    await page.goto('/ui/leagues');
    await page.getByTestId('create-league-button').click();
    await page.getByTestId('create-league-name-input').fill(leagueName);
    await page.getByRole('button', { name: 'Create' }).click();

    await page.waitForURL(/\/ui\/leagues\/[^/]+$/);
    await expect(page.getByText(leagueName)).toBeVisible();

    const leagueUrl = new URL(page.url());
    leagueCode = leagueUrl.pathname.split('/').pop() || '';
    expect(leagueCode).not.toBe('');

    await page.getByTestId('league-invitation-tab').click();
    await page.getByTestId('create-invitation-button').click();
    await page.getByTestId('invitation-alias-input').fill(inviteeAlias);
    await page.getByTestId('create-invitation-confirm-button').click();

    const linkInput = page.getByTestId('invitation-link-input');
    await expect(linkInput).toBeVisible();
    invitationLink = await linkInput.inputValue();
    expect(invitationLink).toContain('/ui/leagues/join/');

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('invitee accepts invitation', async ({ browser }) => {
    if (!fs.existsSync(inviteeStatePath)) {
      test.skip(true, 'Invitee credentials not configured.');
    }
    if (!invitationLink) {
      test.fail(true, 'Invitation link missing from setup step.');
    }

    const context = await browser.newContext({
      storageState: inviteeStatePath,
    });
    const page = await context.newPage();

    await page.goto(invitationLink);
    await expect(page.getByText('Congratulations')).toBeVisible({
      timeout: 60000,
    });

    await page.getByRole('button', { name: 'Go to League' }).click();
    await page.getByTestId('league-members-tab').click();
    await expect(page.getByText(inviteeAlias)).toBeVisible();

    await context.close();
  });

  test('admin creates and finishes game round', async ({ page }) => {
    if (!leagueCode) {
      test.fail(true, 'League code missing from setup step.');
    }

    await page.goto(`/ui/game-rounds/new?league=${leagueCode}`);

    await page.getByText('Ticket to Ride', { exact: true }).click();
    await page.getByTestId('game-type-next-button').click();

    await page.getByTestId('add-virtual-player-button').click();
    await page.getByTestId('virtual-player-alias-input').fill(virtualAlias);
    await page.getByTestId('create-virtual-player-button').click();
    await expect(page.getByText(virtualAlias)).toBeVisible();

    await page.getByTestId('select-players-next-button').click();

    await page.getByTestId('round-name-input').fill(roundName);
    await page.getByTestId('configure-round-next-button').click();

    await page.getByTestId('finish-game-button').click();
    await page.waitForURL(/\/ui\/game-rounds/);
    await expect(page.getByText('Game Rounds')).toBeVisible();

    await page.goto(`/ui/leagues/${leagueCode}`);
    await page.getByTestId('league-standings-tab').click();
    await expect(page.getByText(virtualAlias)).toBeVisible();
  });
});

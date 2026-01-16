# UI-Backend E2E Testing (Playwright)

## Scope

The E2E suite drives the UI and validates real backend behavior in a production-like environment.

## Requirements

- Isolated production-like instance (UI + API + MongoDB).
- UI and API on the same domain with HTTPS (required for Secure cookies).
- Backend environment variables set:
  - `MONGODB_URI`
  - `GOOGLE_CLIENT_ID`
  - `GOOGLE_CLIENT_SECRET`
  - `JWT_SECRET`
  - `SESSION_SECRET`
  - `SUPERADMINS`
  - `IPINFO_TOKEN`
- Google OAuth test accounts with 2FA disabled (automation-friendly).

## E2E Environment Variables

- `E2E_BASE_URL` (optional, default `http://localhost:5173`)
- `E2E_GOOGLE_EMAIL` (required, admin user)
- `E2E_GOOGLE_PASSWORD` (required, admin user)
- `E2E_INVITEE_EMAIL` (optional, invitee user)
- `E2E_INVITEE_PASSWORD` (optional, invitee user)

Notes:
- The admin account must be included in `SUPERADMINS`.
- If invitee credentials are provided, the suite logs in the invitee to accept invitations.
- If only `E2E_INVITEE_EMAIL` is provided, the admin creates the invitee user via `/api/admin/user/create`.

## Install Playwright

From the `frontend` directory:

- `npm install`
- `npx playwright install --with-deps`

## Run Tests

From the `frontend` directory:

- `E2E_BASE_URL=https://your-env.example.com E2E_GOOGLE_EMAIL=... E2E_GOOGLE_PASSWORD=... npm run e2e`

Optional invitee login:

- `E2E_INVITEE_EMAIL=... E2E_INVITEE_PASSWORD=... npm run e2e`

## Data Strategy

- Tests create unique names per run (no global cleanup).
- If you need cleanup, reset the test database between runs.

## Debugging

- `npm run e2e:ui` for interactive debugging.
- Playwright artifacts (report, screenshots, video) are stored under `frontend/playwright-report` and `frontend/test-results`.

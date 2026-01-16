export type GoogleCredentials = {
  email: string;
  password: string;
};

const requireEnv = (name: string): string => {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
};

export const getBaseURL = (): string =>
  process.env.E2E_BASE_URL || 'http://localhost:5173';

export const getAdminCredentials = (): GoogleCredentials => ({
  email: requireEnv('E2E_GOOGLE_EMAIL'),
  password: requireEnv('E2E_GOOGLE_PASSWORD'),
});

export const getInviteeCredentials = (): GoogleCredentials | null => {
  const email = process.env.E2E_INVITEE_EMAIL;
  const password = process.env.E2E_INVITEE_PASSWORD;

  if (!email && !password) {
    return null;
  }

  if (!email || !password) {
    throw new Error(
      'Both E2E_INVITEE_EMAIL and E2E_INVITEE_PASSWORD must be set together.'
    );
  }

  return { email, password };
};

export const getInviteeEmail = (): string | null => {
  return process.env.E2E_INVITEE_EMAIL || null;
};

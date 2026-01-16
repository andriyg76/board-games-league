export const buildRunId = (): string => {
  const random = Math.random().toString(36).slice(2, 8);
  return `${Date.now()}-${random}`;
};

export const buildName = (prefix: string, runId: string): string => {
  return `${prefix} ${runId}`;
};

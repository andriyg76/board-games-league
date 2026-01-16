import path from 'path';
import { fileURLToPath } from 'url';

const filename = fileURLToPath(import.meta.url);
const dirname = path.dirname(filename);

export const e2eRoot = path.resolve(dirname, '..');
export const authDir = path.join(e2eRoot, '.auth');
export const adminStatePath = path.join(authDir, 'admin.json');
export const inviteeStatePath = path.join(authDir, 'invitee.json');

# Electron UI Support TODO

## Auth callback and deep-link handling

### Context
Electron does not automatically intercept `https://bgl.andriydc.eu/ui/auth-callback`
links unless OS-level app-link/universal-link support is configured. For MVP,
the practical approach is to use a custom scheme (e.g. `bgl://auth-callback`)
or a loopback redirect.

### TODO
1) **Custom scheme callback**
   - Add `bgl://auth-callback` to OAuth redirect URIs.
   - Ensure Electron registers the custom scheme.
2) **Login request flag**
   - Add a login query flag (e.g. `client=electron`).
3) **Backend callback selection**
   - When `client=electron`, use the Electron callback URL.
   - Otherwise use the standard web callback (`/ui/auth-callback`).
   - Note: current backend uses a single callback URL cached by `sync.Once`,
     so this needs refactoring to allow per-request callback selection.
4) **Frontend flow**
   - After receiving the Electron deep link, route to the auth callback screen
     and continue the existing session handling.

### Open question
- Do we keep a single OAuth client and add multiple redirect URIs,
  or register a dedicated Electron OAuth client?

## CORS allowlist considerations
- Adding CORS allowlist means returning:
  `Access-Control-Allow-Origin` (exact origin),
  `Access-Control-Allow-Credentials: true`,
  and `Vary: Origin`.
- This only works for **valid HTTP/HTTPS origins**.
  - Examples: `https://bgl.andriydc.eu`, `http://localhost:5173`.
- For `file://` or `Origin: null`, CORS with credentials does not work.
- Custom schemes (e.g., `app://bgl`) require a standard/secure scheme and must be
  explicitly allowed by the backend.
- Current backend has **no CORS middleware**; `TRUSTED_ORIGINS` is not used
  for CORS headers.

## Main-process proxy (Electron)
- Proxy lives in Electron **main process** (Node.js) and handles HTTP requests.
- Renderer communicates via IPC; no browser CORS limitations apply.
- Requires cookie management (Node fetch has no cookie jar by default).
- Pros: works with `file://` and custom schemes, avoids CORS.
- Cons: more plumbing and state handling in main process.

# Mobile UI Plan (MVP) and Routing

## Goals
- Provide a mobile-first UI for the primary game workflow.
- Keep the desktop UI intact during the transition.
- Minimize accidental cross-impact between desktop and mobile.

## MVP Scope
- Login / signup (mobile entry)
- Accept invitation (public, no login required to view)
- League selection (only when no selection and >1 available)
- League main screen (current league info + games + start game)
- Game flow (multi-step, mobile-first)

## Out of Scope (MVP)
- Invitation creation (send invites)
- League administration and settings
- Desktop navigation and admin dashboards

## Routing Strategy
Use a dedicated mobile route prefix to keep the surface separate:

```
/m                     -> MobileEntry (routing decision)
/m/login               -> MobileLogin
/m/accept-invite/:token -> MobileAcceptInvite
/m/league/select       -> MobileLeagueSelect
/m/league              -> MobileLeagueHome
/m/game/start          -> MobileGameStart (placeholder)
/m/game/:code          -> MobileGameFlow
```

This allows parallel development without breaking desktop routes.

## Entry Logic
Mobile entry should resolve to the correct screen:

1) If user is not logged in -> redirect to `/m/login`
2) Load leagues (if not loaded yet)
3) If saved league exists and active -> set it -> `/m/league`
4) If no saved league:
   - If active leagues > 1 -> `/m/league/select`
   - If exactly 1 -> auto-select -> `/m/league`
5) If no leagues -> show empty state message

## Session UI Mode Detection
We store the UI preference in sessionStorage:
- `ui_mode = mobile | desktop`

Auto-detection uses a device heuristic (screen width, UA, touch).

Rules:
1) If user enters a mode-specific path (`/m` or `/ui`), that is treated as
   the current mode.
2) If the current mode does not fit the device, we show a confirm:
   "It looks like the {mode} version fits your device better. Switch?"
   - Yes -> switch and store preferred mode
   - No  -> keep current mode and store it
3) If the current mode fits the device, store it and continue without prompt.
4) If no mode is stored and user opens a neutral path (e.g. `/`), we prompt
   only when the device prefers mobile; otherwise we default to desktop.

Auth callback (`/ui/auth-callback`) is excluded from auto-redirect to avoid
breaking the login flow.

## Accept Invitation Flow
- The accept page is public and should show preview details.
- User can tap "Accept" even without login.
- If not logged in, redirect to login; after login auto-accept invitation.
- After accept: set current league and go to `/m/league`.

## League Selection
Only place where the user can see the list of leagues.
It shows only when:
- no league is selected AND
- user has more than one available league.

## League Main Screen
Shows current league name, list of active games, and "Start game" CTA.
No league selection or switching here.

## Game Flow
Mobile game flow is a sequence of screens (wizard).
Exit should confirm save/discard depending on state.

## Path Parity With Desktop Routes
Options:

1) **Keep separate mobile prefix** (current approach)
   - Pros: no collision with existing desktop routes
   - Cons: links differ between desktop and mobile

2) **Share same paths with layout switching**
   - Pros: single URL for desktop and mobile
   - Cons: higher risk of breaking desktop; more complex rules

Recommendation (MVP):
Keep `/m` prefix and add targeted redirects later if needed.
We can add mobile-only aliases for deep links (invites) without breaking `/ui`.

## Open Questions
- Do we want auto-redirect from `/` to `/m` on mobile user agents?
- Should invite links be updated to `/m/accept-invite/:token` or keep `/ui` for now?
- Should we add a dedicated "Exit game" confirmation component?

## Electron Wrapper Considerations
Decisions (confirmed):
- Electron should always use the mobile UI.
- OAuth login should open in the system browser.
- API base URL should be configurable (env).

Routing recommendation:
- Prefer hash routing for Electron (`#/m/...`) to avoid deep-link issues
  when running without a local server.
- Keep history routing for web.

Implications:
1) **Force mobile mode in Electron**
   - Set `ui_mode = mobile` at startup (before router guard),
     and skip the confirm prompt.
2) **API base URL**
   - Introduce `VITE_API_BASE_URL` and use it in apiClient.
3) **OAuth callback**
   - Use a custom scheme (e.g., `app://auth-callback`) or a deep-link handler
     to bring users back into the Electron app after system browser login.
4) **Invitation links**
   - Use a configurable public web base URL for shareable invites
     instead of `window.location.origin`.


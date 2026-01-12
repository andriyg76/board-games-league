# League System Test Plan

## Overview

This document outlines comprehensive testing procedures for the Board Games League system. Tests should be performed in the order listed to ensure proper system functionality.

## Prerequisites

- Backend server running with MongoDB connection
- Frontend development server running
- At least 2 test user accounts (1 superadmin, 1 regular user)
- Clean database state or test data

## Test Environment Setup

### 1. Create Test Users

Create the following test accounts:

**Superadmin User:**
- Email: `admin@test.com`
- Name: `Admin User`
- Superadmin: `true`

**Regular Users:**
- Email: `user1@test.com`, Name: `User One`
- Email: `user2@test.com`, Name: `User Two`
- Email: `user3@test.com`, Name: `User Three`

### 2. Verify Authentication

- [ ] Login as each user successfully
- [ ] Verify JWT tokens are issued
- [ ] Verify tokens work for authenticated endpoints

---

## Phase 1: League Creation (Superadmin Only)

### Test 1.1: Create League as Superadmin

**Preconditions:**
- Logged in as superadmin user

**Steps:**
1. Navigate to `/ui/leagues`
2. Click "Create League" button
3. Enter league details:
   - Name: "Test League Alpha"
   - Description: "Testing league creation and management"
4. Click "Create"

**Expected Results:**
- [✓] League is created successfully
- [✓] Success message is displayed
- [✓] Redirected to league details page
- [✓] League appears in leagues list
- [✓] Creator is automatically added as member

**API Verification:**
```bash
GET /api/leagues
# Should return the newly created league
```

### Test 1.2: Attempt League Creation as Regular User

**Preconditions:**
- Logged in as regular user (user1@test.com)

**Steps:**
1. Navigate to `/ui/leagues`
2. Verify "Create League" button is not visible
3. Attempt direct API call: `POST /api/leagues`

**Expected Results:**
- [✓] Create button is hidden for non-superadmin
- [✓] API returns 403 Forbidden
- [✓] Error message: "Superadmin privileges required"

---

## Phase 2: Invitation Flow

### Test 2.1: Create Invitation Link

**Preconditions:**
- Logged in as superadmin (who is member of Test League Alpha)
- On league details page

**Steps:**
1. Navigate to "Invitation" tab
2. Click "Create Invitation"
3. Wait for link to be generated
4. Click "Copy Link" button

**Expected Results:**
- [✓] Loading indicator appears during generation
- [✓] Invitation link is displayed
- [✓] Expiry date shows 7 days from now
- [✓] Success message: "Link copied to clipboard"
- [✓] Link format: `https://.../ui/leagues/join/{token}`

**API Verification:**
```bash
POST /api/leagues/{code}/invitations
# Response should include token and invitation_link
```

### Test 2.2: Accept Invitation (New Member)

**Preconditions:**
- Have valid invitation link from Test 2.1
- Logged in as user2@test.com (not yet a member)

**Steps:**
1. Paste invitation link in browser
2. Navigate to the link
3. Wait for automatic acceptance

**Expected Results:**
- [✓] Loading screen appears
- [✓] Success message: "Successfully joined Test League Alpha"
- [✓] Shows "What's next" guidance
- [✓] "Go to League" button appears
- [✓] User is redirected to league details when clicking button

**API Verification:**
```bash
POST /api/leagues/invitations/{token}/accept
# Should return league details
# User should now appear in members list
```

### Test 2.3: Accept Already-Used Invitation

**Preconditions:**
- Use same invitation link from Test 2.1
- Logged in as user3@test.com (not a member)

**Steps:**
1. Navigate to the used invitation link

**Expected Results:**
- [✓] Error message: "Invitation not found or already used"
- [✓] "Go to Home" button displayed
- [✓] User is NOT added to league

### Test 2.4: Member Attempts to Use Invitation

**Preconditions:**
- Create new invitation as superadmin
- Logged in as user2@test.com (already a member)

**Steps:**
1. Navigate to new invitation link

**Expected Results:**
- [✓] Error message indicating user is already a member
- [✓] No changes to membership

### Test 2.5: Expired Invitation

**Preconditions:**
- Create invitation
- Manually update `expires_at` in database to past date (or wait 7 days)

**Steps:**
1. Attempt to use expired invitation link

**Expected Results:**
- [✓] Error message: "Invitation expired (valid for 7 days)"
- [✓] User is NOT added to league

---

## Phase 3: League Membership Management

### Test 3.1: View Members List

**Preconditions:**
- Logged in as any league member
- On league details page

**Steps:**
1. Click "Members" tab
2. View members list

**Expected Results:**
- [✓] All league members are displayed
- [✓] Shows user avatar, name, status
- [✓] Shows join date for each member
- [✓] Regular users do NOT see ban/unban buttons
- [✓] Superadmin sees ban/unban actions

### Test 3.2: Ban Member (Superadmin)

**Preconditions:**
- Logged in as superadmin
- On league members tab
- user2@test.com is active member

**Steps:**
1. Locate user2@test.com in members list
2. Click "Ban" button
3. Confirm action in dialog

**Expected Results:**
- [✓] Confirmation dialog appears
- [✓] User status changes to "banned"
- [✓] Badge/chip shows "Banned" status
- [✓] Success notification displayed

**API Verification:**
```bash
PUT /api/leagues/{code}/members/{userId}/status
Body: { "status": "banned" }
```

### Test 3.3: Verify Banned Member Cannot Play Games

**Preconditions:**
- user2@test.com is banned from league
- Attempt to create game round with league_id

**Steps:**
1. Try to add user2@test.com to a game in this league

**Expected Results:**
- [✓] Error preventing banned user from participating
- [✓] Game creation fails with appropriate message

### Test 3.4: Unban Member (Superadmin)

**Preconditions:**
- user2@test.com is currently banned

**Steps:**
1. Click "Unban" button for user2@test.com
2. Confirm action

**Expected Results:**
- [✓] User status returns to "active"
- [✓] User can now participate in games again

### Test 3.5: Leave League

**Preconditions:**
- Logged in as user2@test.com
- Member of Test League Alpha

**Steps:**
1. On league details page
2. Click "Leave League" button
3. Confirm action in dialog

**Expected Results:**
- [✓] Confirmation dialog appears: "Are you sure you want to leave?"
- [✓] User is removed from league
- [✓] Redirected to leagues list
- [✓] League no longer appears in user's leagues
- [✓] User cannot access league details page

**API Verification:**
```bash
DELETE /api/leagues/{code}/members/me
# Should return 204 No Content
```

---

## Phase 4: Game Rounds with League Context

### Test 4.1: Create Game Round in League

**Preconditions:**
- Logged in as superadmin
- Test League Alpha exists with 3+ active members
- Game type "Carcassonne" exists

**Steps:**
1. Navigate to "Start New Game"
2. Select game type: "Carcassonne"
3. **Select league: "Test League Alpha"**
4. Add players:
   - Admin User (moderator)
   - User One
   - User Three
5. Assign positions and teams
6. Click "Start Game"

**Expected Results:**
- [✓] League dropdown shows all user's leagues
- [✓] Only active league members can be selected
- [✓] Banned members do NOT appear in player selection
- [✓] Game is created with `league_id` field set
- [✓] Game appears in league's game history

**API Verification:**
```bash
POST /api/games
Body: {
  "league_id": "...",
  "name": "Game 1",
  ...
}
# Response should include league_id
```

### Test 4.2: Create Game Round Without League

**Steps:**
1. Start new game
2. Leave league field empty/null
3. Add any players
4. Create game

**Expected Results:**
- [✓] Game is created successfully
- [✓] Game has no league_id
- [✓] Game does NOT affect any league standings

### Test 4.3: Finalize Game and Verify Points

**Preconditions:**
- Game round created in Test 4.1
- Not yet finalized

**Steps:**
1. Enter final scores for all players:
   - Admin User: 80 points (1st place)
   - User One: 60 points (2nd place)
   - User Three: 40 points (3rd place)
2. Click "Finalize Game"

**Expected Results:**
- [✓] Game is marked as finalized
- [✓] Player positions are calculated from scores
- [✓] League standings are automatically updated

**Expected Points Calculation:**

**Admin User:**
- Participation: 2 points
- Position (1st): 10 points
- Moderation: 1 point
- **Total: 13 points**

**User One:**
- Participation: 2 points
- Position (2nd): 6 points
- **Total: 8 points**

**User Three:**
- Participation: 2 points
- Position (3rd): 3 points
- **Total: 5 points**

---

## Phase 5: Standings Calculation

### Test 5.1: View League Standings After One Game

**Preconditions:**
- One game finalized (Test 4.3)

**Steps:**
1. Navigate to league details
2. View "Standings" tab

**Expected Results:**
- [✓] Standings table is displayed
- [✓] Rankings:
  - 1st: Admin User - 13 points
  - 2nd: User One - 8 points
  - 3rd: User Three - 5 points
- [✓] Games played: 1 for each
- [✓] Podium counts shown correctly
- [✓] Medal icons for top 3 positions
- [✓] Point breakdown visible in details

### Test 5.2: Play Multiple Games and Verify Rankings

**Steps:**
1. Create and finalize 3 more games with varying results
2. Ensure different players win different games
3. View updated standings

**Expected Results:**
- [✓] Standings update after each game finalization
- [✓] Point totals are cumulative
- [✓] Games played counter increases
- [✓] Podium counts update correctly
- [✓] Rankings change based on total points
- [✓] Tie-breaker works (fewer games played ranks higher)

### Test 5.3: Player Details Modal

**Steps:**
1. On standings tab
2. Click "Details" (info icon) for a player

**Expected Results:**
- [✓] Modal opens with detailed statistics
- [✓] Shows total points breakdown:
  - Participation points
  - Position points
  - Moderation points
- [✓] Shows games played count
- [✓] Shows games moderated count
- [✓] Shows podium statistics (1st/2nd/3rd place counts)
- [✓] All numbers match calculation

### Test 5.4: Empty Standings (No Games)

**Preconditions:**
- Create new league with members but no games

**Steps:**
1. View standings tab

**Expected Results:**
- [✓] Message: "No standings data yet. Play your first games!"
- [✓] No table displayed
- [✓] No errors

---

## Phase 6: League Management (Superadmin)

### Test 6.1: Update League Details

**Preconditions:**
- Logged in as superadmin

**Steps:**
1. Navigate to league details
2. Click "Edit League" button
3. Update name: "Test League Alpha - Updated"
4. Update description: "Updated description"
5. Click "Save"

**Expected Results:**
- [✓] League name is updated
- [✓] League description is updated
- [✓] Success notification displayed
- [✓] Changes visible immediately

**API Verification:**
```bash
PUT /api/leagues/{code}
Body: { "name": "...", "description": "..." }
```

### Test 6.2: Archive League

**Steps:**
1. Click "Archive League" button
2. Confirm action

**Expected Results:**
- [✓] Confirmation dialog appears
- [✓] League status changes to "archived"
- [✓] Badge shows "Archived"
- [✓] League no longer appears in active leagues list
- [✓] Cannot create new games in archived league
- [✓] Can still view standings and history

### Test 6.3: Unarchive League

**Steps:**
1. View archived league
2. Click "Unarchive" button
3. Confirm action

**Expected Results:**
- [✓] League status returns to "active"
- [✓] League appears in active leagues list
- [✓] Can create new games again

### Test 6.4: Attempt Update as Non-Superadmin

**Preconditions:**
- Logged in as regular user

**Steps:**
1. Navigate to league details
2. Verify no "Edit" or "Archive" buttons visible
3. Attempt direct API call

**Expected Results:**
- [✓] Management buttons hidden
- [✓] API returns 403 Forbidden

---

## Phase 7: UI/UX and Edge Cases

### Test 7.1: Loading States

**Test all loading states:**
- [ ] League list loading
- [ ] League details loading
- [ ] Standings loading
- [ ] Members loading
- [ ] Invitation generation loading
- [ ] Invitation acceptance loading

**Expected:**
- [✓] Skeleton loaders or spinners displayed
- [✓] Smooth transitions
- [✓] No layout shift

### Test 7.2: Error Handling

**Test error scenarios:**
- [ ] Network error during API call
- [ ] Invalid league code in URL
- [ ] Accessing league as non-member
- [ ] Creating invitation when not a member
- [ ] Server error (500)

**Expected:**
- [✓] User-friendly error messages
- [✓] No cryptic technical errors
- [✓] Option to retry or go back
- [✓] Errors don't crash the app

### Test 7.3: Responsive Design

**Test on different screen sizes:**
- [ ] Desktop (1920x1080)
- [ ] Tablet (768x1024)
- [ ] Mobile (375x667)

**Expected:**
- [✓] All components render properly
- [✓] Tables adapt to small screens
- [✓] Buttons accessible
- [✓] No horizontal scrolling (except tables)

### Test 7.4: Internationalization

**Test language switching:**
- [ ] English (en)
- [ ] Ukrainian (uk)
- [ ] Estonian (et)

**Expected:**
- [✓] All league UI text translated
- [✓] No missing translation keys
- [✓] Proper formatting for dates/numbers
- [✓] Language persists across navigation

### Test 7.5: Navigation and Routing

**Test all routes:**
- [ ] `/ui/leagues` - League list
- [ ] `/ui/leagues/{code}` - League details
- [ ] `/ui/leagues/join/{token}` - Accept invitation
- [ ] Back button functionality
- [ ] Breadcrumbs (if applicable)

**Expected:**
- [✓] All routes load correctly
- [✓] 404 for invalid codes/tokens
- [✓] Proper redirects after actions
- [✓] Browser back/forward work

### Test 7.6: Empty States

**Test empty states:**
- [ ] No leagues exist
- [ ] League has no members (besides creator)
- [ ] League has no games
- [ ] League has no standings

**Expected:**
- [✓] Helpful empty state messages
- [✓] Call-to-action buttons where appropriate
- [✓] Icons or illustrations
- [✓] No errors

---

## Phase 8: Performance Testing

### Test 8.1: Large Data Sets

**Create test data:**
- 50+ league members
- 100+ game rounds in a league

**Test:**
- [ ] Standings calculation speed
- [ ] Members list loading
- [ ] Game history loading

**Expected:**
- [✓] Page loads in < 2 seconds
- [✓] Standings calculation in < 1 second
- [✓] Smooth scrolling
- [✓] Pagination works (if implemented)

### Test 8.2: Concurrent Users

**Simulate:**
- Multiple users creating games simultaneously
- Multiple invitation acceptances
- Simultaneous standings views

**Expected:**
- [✓] No race conditions
- [✓] Data consistency maintained
- [✓] No duplicate memberships
- [✓] Accurate point calculations

---

## Phase 9: Security Testing

### Test 9.1: Authorization Checks

**Attempt unauthorized actions:**
- [ ] View league details as non-member
- [ ] Create league as regular user
- [ ] Ban member as regular user
- [ ] Update league as regular user

**Expected:**
- [✓] All unauthorized actions blocked
- [✓] Appropriate error messages
- [✓] No data leakage

### Test 9.2: Token Security

**Test invitation tokens:**
- [ ] Cannot guess tokens
- [ ] Expired tokens rejected
- [ ] Used tokens rejected
- [ ] Invalid tokens rejected

**Expected:**
- [✓] Tokens are cryptographically secure
- [✓] All validation works correctly

### Test 9.3: Input Validation

**Test with malicious input:**
- [ ] XSS attempts in league name/description
- [ ] SQL injection attempts
- [ ] Overly long strings
- [ ] Special characters

**Expected:**
- [✓] All input sanitized
- [✓] No XSS vulnerabilities
- [✓] No injection vulnerabilities
- [✓] Proper validation errors

---

## Regression Testing Checklist

After any changes to league system, verify:

- [ ] Existing leagues still accessible
- [ ] Points calculation still accurate
- [ ] Invitations still work
- [ ] All API endpoints respond correctly
- [ ] Frontend components render properly
- [ ] No console errors
- [ ] No broken links
- [ ] Backwards compatibility maintained

---

## Test Automation

### Unit Tests to Verify

**Backend:**
- [ ] `LeagueService` unit tests
- [ ] Points calculation tests
- [ ] Standings sorting tests
- [ ] Repository tests with mocks

**Frontend:**
- [ ] Component tests (Vitest)
- [ ] Store tests (Pinia)
- [ ] API client tests

### Integration Tests

- [ ] API endpoint tests (end-to-end)
- [ ] Database operations
- [ ] Authentication flow
- [ ] Invitation flow

---

## Test Sign-Off

| Test Phase | Tester | Date | Status | Notes |
|------------|--------|------|--------|-------|
| Phase 1: League Creation | | | ⬜ | |
| Phase 2: Invitation Flow | | | ⬜ | |
| Phase 3: Membership Mgmt | | | ⬜ | |
| Phase 4: Game Rounds | | | ⬜ | |
| Phase 5: Standings | | | ⬜ | |
| Phase 6: League Management | | | ⬜ | |
| Phase 7: UI/UX | | | ⬜ | |
| Phase 8: Performance | | | ⬜ | |
| Phase 9: Security | | | ⬜ | |

---

## Reporting Issues

When reporting bugs found during testing, include:

1. **Test phase and number** (e.g., Test 2.3)
2. **Preconditions** (user role, data state)
3. **Steps to reproduce**
4. **Expected result**
5. **Actual result**
6. **Screenshots/logs** (if applicable)
7. **Browser/environment details**

---

## Testing Tools

- **Backend:** Go test framework, MongoDB shell
- **Frontend:** Browser DevTools, Vue DevTools
- **API:** Postman, curl, Thunder Client
- **Network:** Browser Network tab
- **Database:** MongoDB Compass

---

## Success Criteria

The league system is considered ready for production when:

- ✅ All test phases pass 100%
- ✅ No critical or high-priority bugs
- ✅ Performance benchmarks met
- ✅ Security audit passed
- ✅ Code review approved
- ✅ Documentation complete
- ✅ User acceptance testing passed

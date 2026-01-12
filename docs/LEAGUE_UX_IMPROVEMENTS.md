# League System UX Improvements

## Overview

This document outlines User Experience (UX) improvements for the League System, including error handling, loading states, confirmation dialogs, and feedback mechanisms.

---

## Current State Assessment

### ✅ Well-Implemented Features

**LeagueList.vue:**
- ✓ Loading state with progress indicator
- ✓ Error display with alert
- ✓ Empty state message
- ✓ Basic error handling in try-catch blocks

**LeagueDetails.vue:**
- ✓ Loading state with spinner
- ✓ Error display with alert
- ✓ Error handling in API calls
- ✓ Native confirmation dialog for ban action

**AcceptInvitation.vue:**
- ✓ Excellent loading state
- ✓ Comprehensive error handling with specific messages
- ✓ Success state with guidance
- ✓ Multiple error scenarios handled

**LeagueInvitation.vue:**
- ✓ Loading state during generation
- ✓ Error display
- ✓ Success feedback when copied
- ✓ Expiry date display

**LeagueStandings.vue:**
- ✓ Empty state message
- ✓ Details modal for player stats

---

## Recommended Improvements

### Priority 1: Critical UX Issues

#### 1.1 User-Facing Error Messages

**Current Issue:**
```typescript
catch (error) {
  console.error('Error creating league:', error);  // User doesn't see this!
}
```

**Recommendation:**
Add toast notifications or inline error alerts for failed actions.

**Implementation:**
```typescript
// Using Vuetify's snackbar or v-alert
const errorMessage = ref<string | null>(null);

try {
  await leagueStore.createLeague(newLeagueName.value);
} catch (error) {
  errorMessage.value = error instanceof Error
    ? error.message
    : 'Не вдалося створити лігу';
  console.error('Error creating league:', error);
}
```

**Affected Components:**
- `LeagueList.vue` - createLeague action
- `LeagueDetails.vue` - archiveLeague, unarchiveLeague, banMember actions
- All components with API calls

#### 1.2 Loading States for Actions

**Current Issue:**
Archive/unarchive and ban/unban actions have no loading indicators.

**Recommendation:**
Add loading states to action buttons.

**Implementation in LeagueDetails.vue:**
```typescript
const archiving = ref(false);

const archiveLeague = async () => {
  archiving.value = true;
  try {
    await leagueStore.archiveLeague(currentLeague.value.code);
    showSuccess('Лігу успішно архівовано');
  } catch (error) {
    showError('Помилка архівування ліги');
  } finally {
    archiving.value = false;
  }
};
```

**Affected Components:**
- `LeagueDetails.vue` - archive, unarchive, ban actions
- `LeagueInvitation.vue` - already has loading state ✓

#### 1.3 Better Confirmation Dialogs

**Current Issue:**
Native `confirm()` dialogs are not styled and don't match the app's design.

**Recommendation:**
Use Vuetify dialogs for confirmations.

**Implementation:**
```vue
<template>
  <!-- Confirmation Dialog -->
  <v-dialog v-model="showConfirmDialog" max-width="400">
    <v-card>
      <v-card-title>{{ confirmTitle }}</v-card-title>
      <v-card-text>{{ confirmMessage }}</v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="showConfirmDialog = false">Скасувати</v-btn>
        <v-btn
          color="error"
          :loading="actionInProgress"
          @click="confirmAction"
        >
          {{ confirmButtonText }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
const showConfirmDialog = ref(false);
const confirmTitle = ref('');
const confirmMessage = ref('');
const confirmAction = ref<(() => void) | null>(null);

const requestBan = (member: LeagueMember) => {
  confirmTitle.value = 'Заблокувати користувача?';
  confirmMessage.value = `Ви впевнені, що хочете заблокувати ${member.user_name}? Користувач не зможе брати участь в іграх цієї ліги.`;
  confirmButtonText.value = 'Заблокувати';
  confirmAction.value = () => banMember(member);
  showConfirmDialog.value = true;
};
</script>
```

**Affected Components:**
- `LeagueDetails.vue` - ban/unban actions, archive/unarchive actions
- `LeagueList.vue` - leave league action (to be implemented)

---

### Priority 2: Enhanced Feedback

#### 2.1 Toast Notifications

**Recommendation:**
Add a global toast/snackbar component for success/error messages.

**Implementation:**

Create `composables/useToast.ts`:
```typescript
import { ref } from 'vue';

const message = ref('');
const type = ref<'success' | 'error' | 'info' | 'warning'>('info');
const show = ref(false);

export function useToast() {
  const showSuccess = (msg: string) => {
    message.value = msg;
    type.value = 'success';
    show.value = true;
  };

  const showError = (msg: string) => {
    message.value = msg;
    type.value = 'error';
    show.value = true;
  };

  const showInfo = (msg: string) => {
    message.value = msg;
    type.value = 'info';
    show.value = true;
  };

  return {
    message,
    type,
    show,
    showSuccess,
    showError,
    showInfo,
  };
}
```

Add to `App.vue`:
```vue
<template>
  <v-app>
    <!-- ... app content ... -->

    <v-snackbar
      v-model="toast.show"
      :color="toast.type"
      :timeout="3000"
      location="top"
    >
      {{ toast.message }}
      <template v-slot:actions>
        <v-btn icon="mdi-close" @click="toast.show = false" />
      </template>
    </v-snackbar>
  </v-app>
</template>

<script setup>
import { useToast } from '@/composables/useToast';
const toast = useToast();
</script>
```

**Usage in Components:**
```typescript
import { useToast } from '@/composables/useToast';

const { showSuccess, showError } = useToast();

try {
  await leagueStore.createLeague(name);
  showSuccess('Лігу успішно створено!');
} catch (error) {
  showError('Не вдалося створити лігу');
}
```

#### 2.2 Optimistic UI Updates

**Recommendation:**
Update UI immediately for better perceived performance, then rollback if API fails.

**Example - Ban Member:**
```typescript
const banMember = async (member: LeagueMember) => {
  // Optimistically update UI
  member.status = 'banned';

  try {
    await leagueStore.banUser(currentLeague.value.code, member.user_id);
    showSuccess(`${member.user_name} заблоковано`);
  } catch (error) {
    // Rollback on error
    member.status = 'active';
    showError('Не вдалося заблокувати користувача');
  }
};
```

---

### Priority 3: Accessibility & Polish

#### 3.1 Keyboard Navigation

**Recommendation:**
Ensure all interactive elements are keyboard accessible.

**Checklist:**
- [ ] All buttons focusable
- [ ] Tab order logical
- [ ] Enter key works on focused buttons
- [ ] Escape closes dialogs
- [ ] ARIA labels on icon buttons

**Implementation:**
```vue
<v-btn
  icon="mdi-dots-vertical"
  aria-label="Керування лігою"
  @click="showManageMenu = !showManageMenu"
/>
```

#### 3.2 Loading Skeletons

**Recommendation:**
Replace spinners with skeleton loaders for better UX.

**Implementation:**
```vue
<v-skeleton-loader
  v-if="loading"
  type="list-item-avatar-three-line@3"
/>
```

**Affected Components:**
- `LeagueList.vue` - league cards
- `LeagueDetails.vue` - standings, members
- `LeagueStandings.vue` - table rows

#### 3.3 Transition Animations

**Recommendation:**
Add smooth transitions for better visual feedback.

**Implementation:**
```vue
<transition-group name="list" tag="div">
  <league-card
    v-for="league in activeLeagues"
    :key="league.code"
    :league="league"
  />
</transition-group>

<style scoped>
.list-move,
.list-enter-active,
.list-leave-active {
  transition: all 0.3s ease;
}

.list-enter-from {
  opacity: 0;
  transform: translateY(-30px);
}

.list-leave-to {
  opacity: 0;
  transform: translateY(30px);
}
</style>
```

#### 3.4 Empty State Improvements

**Current:**
Simple text message.

**Recommendation:**
Add illustrations and call-to-action.

**Implementation:**
```vue
<v-card-text v-if="activeLeagues.length === 0" class="text-center py-8">
  <v-icon size="80" color="grey-lighten-1" class="mb-4">
    mdi-trophy-outline
  </v-icon>
  <div class="text-h6 mb-2">Немає активних ліг</div>
  <div class="text-body-2 text-medium-emphasis mb-4">
    Створіть нову лігу або приєднайтесь через запрошення
  </div>
  <div class="d-flex justify-center gap-2">
    <v-btn
      v-if="canCreateLeague"
      color="primary"
      @click="showCreateDialog = true"
    >
      Створити лігу
    </v-btn>
  </div>
</v-card-text>
```

---

### Priority 4: Advanced Features

#### 4.1 Search and Filtering

**Recommendation:**
Add search for leagues and members lists.

**Implementation:**
```vue
<v-text-field
  v-model="searchQuery"
  prepend-inner-icon="mdi-magnify"
  label="Пошук ліг"
  clearable
  hide-details
  class="mb-4"
/>

<script setup>
const searchQuery = ref('');

const filteredLeagues = computed(() => {
  if (!searchQuery.value) return activeLeagues.value;

  return activeLeagues.value.filter(league =>
    league.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  );
});
</script>
```

#### 4.2 Pagination for Large Lists

**Recommendation:**
Add pagination when members/standings exceed 50 items.

**Implementation:**
```vue
<v-data-table
  :items="standings"
  :headers="headers"
  :items-per-page="25"
  :page.sync="page"
/>
```

#### 4.3 Real-Time Updates

**Recommendation:**
Add WebSocket or polling for real-time standings updates.

**Implementation (Polling):**
```typescript
import { useIntervalFn } from '@vueuse/core';

onMounted(() => {
  // Poll standings every 30 seconds
  useIntervalFn(() => {
    if (activeTab.value === 'standings') {
      leagueStore.refreshStandings(currentLeague.value.code);
    }
  }, 30000);
});
```

#### 4.4 Undo Actions

**Recommendation:**
Allow undo for critical actions like ban/leave.

**Implementation:**
```typescript
const showUndoBan = ref(false);
let undoTimeout: ReturnType<typeof setTimeout> | null = null;

const banMember = async (member: LeagueMember) => {
  try {
    await leagueStore.banUser(code, member.user_id);

    showUndoBan.value = true;
    undoTimeout = setTimeout(() => {
      showUndoBan.value = false;
    }, 5000);

    showSuccess('Користувача заблоковано');
  } catch (error) {
    showError('Помилка блокування');
  }
};

const undoBan = async (member: LeagueMember) => {
  if (undoTimeout) clearTimeout(undoTimeout);
  showUndoBan.value = false;

  try {
    await leagueStore.unbanUser(code, member.user_id);
    showSuccess('Блокування скасовано');
  } catch (error) {
    showError('Помилка скасування');
  }
};
```

---

## Form Validation Improvements

### LeagueList.vue - Create Dialog

**Current:**
Basic required validation.

**Recommendation:**
Add comprehensive validation.

**Implementation:**
```typescript
const nameRules = [
  (v: string) => !!v || 'Назва обов\'язкова',
  (v: string) => v.length >= 3 || 'Мінімум 3 символи',
  (v: string) => v.length <= 50 || 'Максимум 50 символів',
  (v: string) => /^[а-яА-ЯёЁіІїЇєЄa-zA-Z0-9\s-]+$/.test(v) || 'Недопустимі символи',
];

const descriptionRules = [
  (v: string) => !v || v.length <= 200 || 'Максимум 200 символів',
];
```

---

## Error Message Guidelines

### User-Friendly Error Messages

**Bad:**
```
Error: MongoError: connection refused
```

**Good:**
```
Не вдалося підключитися до сервера. Перевірте з'єднання та спробуйте ще раз.
```

### Error Message Mapping

Create a utility for mapping API errors to user-friendly messages:

```typescript
// utils/errorMessages.ts
export function getUserFriendlyError(error: Error): string {
  const errorMap: Record<string, string> = {
    'Network Error': 'Помилка мережі. Перевірте з\'єднання.',
    'unauthorized': 'Необхідна авторизація',
    'forbidden': 'Недостатньо прав доступу',
    'not found': 'Ресурс не знайдено',
    'already exists': 'Така ліга вже існує',
    'invalid token': 'Невірне запрошення',
    'token expired': 'Запрошення прострочене',
  };

  const message = error.message.toLowerCase();

  for (const [key, value] of Object.entries(errorMap)) {
    if (message.includes(key)) {
      return value;
    }
  }

  return 'Сталася помилка. Спробуйте ще раз.';
}
```

**Usage:**
```typescript
import { getUserFriendlyError } from '@/utils/errorMessages';

catch (error) {
  const message = error instanceof Error
    ? getUserFriendlyError(error)
    : 'Невідома помилка';
  showError(message);
}
```

---

## Implementation Checklist

### Phase 1: Critical Fixes
- [ ] Add toast notification system
- [ ] Add error messages to all action handlers
- [ ] Add loading states to action buttons
- [ ] Replace native confirm() with Vuetify dialogs

### Phase 2: Enhanced Feedback
- [ ] Add success messages for all actions
- [ ] Implement optimistic UI updates
- [ ] Add user-friendly error message mapping

### Phase 3: Polish
- [ ] Add skeleton loaders
- [ ] Add transition animations
- [ ] Improve empty states
- [ ] Add ARIA labels for accessibility

### Phase 4: Advanced (Optional)
- [ ] Add search and filtering
- [ ] Add pagination
- [ ] Implement real-time updates
- [ ] Add undo functionality

---

## Testing After Implementation

For each improved component, test:

1. **Error Scenarios:**
   - Network failure
   - Invalid input
   - Unauthorized access
   - Server errors

2. **Loading States:**
   - Verify spinner appears
   - Verify buttons disable during loading
   - Verify smooth transitions

3. **Success Feedback:**
   - Verify success messages appear
   - Verify UI updates correctly
   - Verify redirects work

4. **Accessibility:**
   - Tab through all interactive elements
   - Test with screen reader
   - Test keyboard shortcuts

---

## Conclusion

These improvements will significantly enhance the user experience of the League System by:

1. **Providing clear feedback** for all user actions
2. **Reducing frustration** with better error messages
3. **Improving perceived performance** with loading states
4. **Ensuring accessibility** for all users
5. **Adding polish** with animations and transitions

Implement in phases based on priority, testing thoroughly after each phase.

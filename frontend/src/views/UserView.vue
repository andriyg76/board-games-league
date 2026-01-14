<template>
  <div>
    <template v-if="userStore.loggedIn">
      <h2 style="margin-bottom: 24px;">{{ t('user.yourProfile') }}</h2>
      <n-form @submit.prevent="updateUserProfile">
        <n-grid :cols="24" :x-gap="16">
          <n-gi :span="24" :responsive="{ m: 12 }">
            <n-card style="margin-bottom: 16px;">
              <template #header>
                <div style="font-size: 1rem; font-weight: 500;">{{ t('user.profileInfo') }}</div>
              </template>
              <p>
                {{ t('user.currentName') }}: <strong>{{ userStore.user.name }}</strong>
              </p>
              <p>
                {{ t('user.currentAlias') }}: <strong>{{ userStore.user.alias }}</strong>
                <span v-if="isAliasUnique !== null" style="margin-left: 8px;">
                  <span v-if="isAliasUnique">✔️ ({{ t('user.unique') }})</span>
                  <span v-else>❌ ({{ t('user.notUnique') }})</span>
                </span>
              </p>
              <img v-if="userStore.user.avatar" :src="userStore.user.avatar" :alt="`${userStore.user.name}'s avatar`" height="64" width="64" style="margin: 12px 0; border-radius: 50%; object-fit: cover;" />
            </n-card>
          </n-gi>

          <n-gi :span="24" :responsive="{ m: 12 }">
            <n-card>
              <template #header>
                <div style="font-size: 1rem; font-weight: 500;">{{ t('user.editProfile') }}</div>
              </template>
              <n-form-item :label="t('user.yourAlias')" :validation-status="isAliasUnique === false ? 'error' : undefined" :feedback="isAliasUnique === false ? t('user.aliasNotUnique') : undefined">
                <n-input
                    v-model:value="userStore.user.alias"
                    @update:value="checkAliasUniqueness"
                    clearable
                />
              </n-form-item>

              <n-form-item :label="t('user.selectName')" style="margin-top: 16px;">
                <n-select
                    v-model:value="userStore.user.name"
                    :options="nameOptions"
                />
              </n-form-item>

              <n-form-item :label="t('user.selectAvatar')" style="margin-top: 16px;">
                <n-select
                    v-model:value="userStore.user.avatar"
                    :options="avatarOptions"
                    :render-label="renderAvatarLabel"
                    :render-tag="renderAvatarTag"
                />
              </n-form-item>
            </n-card>
          </n-gi>
        </n-grid>

        <n-button
            type="primary"
            @click="updateUserProfile"
            :disabled="!isAliasUnique && userStore.user.alias !== initialAlias"
            style="margin-top: 16px;"
        >
          {{ t('user.saveProfile') }}
        </n-button>
      </n-form>

      <n-grid :cols="24" :x-gap="16" style="margin-top: 24px;">
        <n-gi :span="24">
          <n-card>
            <template #header>
              <div style="font-size: 1rem; font-weight: 500;">{{ t('user.activeSessions') }}</div>
            </template>
            <n-skeleton v-if="loadingSessions" height="200px" />
            <n-data-table
              v-else-if="sessions.length > 0"
              :columns="sessionColumns"
              :data="sessions"
              :row-class-name="getRowClassName"
            />
            <p v-else>{{ t('user.noActiveSessions') }}</p>
          </n-card>
        </n-gi>
      </n-grid>
    </template>
    <template v-else>
      <p>{{ t('user.pleaseLogin') }}</p>
    </template>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, watch, computed, h } from 'vue';
import { NGrid, NGi, NCard, NForm, NFormItem, NInput, NSelect, NButton, NDataTable, NSkeleton, NAvatar, DataTableColumns } from 'naive-ui';
import UserApi, { User, SessionInfo } from "@/api/UserApi";
import { useUserStore } from '@/store/user';
import { useI18n } from 'vue-i18n';

const { t, locale } = useI18n();
const userStore = useUserStore();
const isAliasUnique = ref<boolean | null>(null);
const initialAlias = ref<string>('');
const initialName = ref<string>('');
const initialAvatar = ref<string>('');
const sessions = ref<SessionInfo[]>([]);
const loadingSessions = ref(false);

const nameOptions = computed(() => 
  userStore.user.names?.map(name => ({ label: name, value: name })) || []
);

const avatarOptions = computed(() => 
  userStore.user.avatars?.map(avatar => ({ label: avatar, value: avatar })) || []
);

const renderAvatarLabel = (option: { label: string; value: string }) => {
  return h('div', { style: 'display: flex; align-items: center; gap: 8px;' }, [
    h(NAvatar, { src: option.value, size: 24, round: true }),
    h('span', t('common.select'))
  ]);
};

const renderAvatarTag = ({ option }: { option: { label: string; value: string } }) => {
  return h(NAvatar, { src: option.value, size: 24, round: true }, { default: () => null });
};

const sessionColumns: DataTableColumns<SessionInfo> = [
  { title: t('user.location'), key: 'location' },
  { title: t('user.ipAddress'), key: 'ip_address' },
  { title: t('user.userAgent'), key: 'user_agent', ellipsis: { tooltip: true } },
  { title: t('user.created'), key: 'created_at' },
  { title: t('user.lastActivity'), key: 'updated_at' },
  { 
    title: t('user.status'), 
    key: 'status',
    render: (row: SessionInfo) => {
      return h('span', row.is_current ? t('common.current') : t('common.active'));
    }
  },
];

const getRowClassName = (row: SessionInfo) => {
  return row.is_current ? 'current-session' : '';
};

async function checkAliasUniqueness() {
  if (userStore.user.alias && userStore.user.alias.length >= 3) {
    try {
      const response = await UserApi.checkAlias(userStore.user.alias);
      isAliasUnique.value = response.isUnique;
    } catch (e) {
      console.error("Error checking alias uniqueness: ", e);
      isAliasUnique.value = false;
    }
  } else {
    isAliasUnique.value = null; // Reset if alias is too short
  }
}

async function updateUserProfile() {
  // Only update if the alias is unique or hasn't changed from the initial value
  if ((isAliasUnique.value || userStore.user.alias === initialAlias.value) && userStore.user.alias) {
    try {
      await UserApi.updateUser(userStore.user as User);
      console.log('Profile updated successfully!');
      initialAlias.value = userStore.user.alias;
      initialAvatar.value = userStore.user.avatar;
      initialName.value = userStore.user.name;
    } catch (e) {
      console.error("Error updating profile: ", e);
    }
  } else {
    console.warn("Cannot save: Alias is not unique or invalid.");
  }
}

// Watch for changes in the user object's alias to trigger uniqueness check
watch(() => userStore.user.alias, (newAlias, oldAlias) => {
  if (newAlias !== oldAlias && newAlias !== initialAlias.value) {
    checkAliasUniqueness();
  } else if (newAlias === initialAlias.value) {
    isAliasUnique.value = true; // If alias reverts to original, it's unique
  }
});

function formatDate(dateString: string): string {
  if (!dateString) return '';
  const localeMap: Record<string, string> = { 'uk': 'uk-UA', 'en': 'en-US', 'et': 'et-EE' };
  const date = new Date(dateString);
  return date.toLocaleString(localeMap[locale.value] || 'en-US');
}

async function loadSessions() {
  loadingSessions.value = true;
  try {
    // Get rotate token from localStorage if available
    const rotateToken = localStorage.getItem('rotateToken') || undefined;
    sessions.value = await UserApi.getUserSessions(rotateToken || undefined);
    // Format dates for display
    sessions.value = sessions.value.map(session => ({
      ...session,
      location: session.geo_info 
        ? `${session.geo_info.city || ''}${session.geo_info.city && session.geo_info.country ? ', ' : ''}${session.geo_info.country || ''}`
        : t('common.unknown'),
      created_at: formatDate(session.created_at || ''),
      updated_at: formatDate(session.updated_at || ''),
    }));
  } catch (e) {
    console.error("Error loading sessions:", e);
    sessions.value = [];
  } finally {
    loadingSessions.value = false;
  }
}

onMounted(async () => {
  // Fetch the current user data if not already loaded in the store
  if (!userStore.loggedIn) {
    try {
      const user = await UserApi.getUser();
      if (user) {
        userStore.setUser(user);
      }
    } catch (e) {
      console.error("Error fetching current user:", e);
    }
  }
  initialAlias.value = userStore.user.alias || '';
  initialName.value = userStore.user.name || '';
  initialAvatar.value = userStore.user.avatar || '';
  
  // Load sessions
  if (userStore.loggedIn) {
    await loadSessions();
  }
});
</script>

<style scoped>
:deep(.current-session) {
  background-color: rgba(32, 128, 240, 0.1);
}
</style>

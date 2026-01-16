<template>
  <div>
    <n-spin v-if="loading" size="large" style="display: flex; justify-content: center; padding: 64px;" />

    <n-alert v-else-if="error" type="error" style="margin-bottom: 16px;">
      {{ error }}
    </n-alert>

    <template v-else-if="currentLeague">
      <n-grid :cols="24" :x-gap="16">
        <n-gi :span="24">
          <n-card>
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <div style="display: flex; align-items: center; gap: 12px;">
                  <n-icon size="32" color="#18a058">
                    <TrophyIcon />
                  </n-icon>
                  <div>
                    <div style="font-size: 1.5rem; font-weight: 500;">{{ currentLeague.name }}</div>
                    <n-tag :type="currentLeague.status === 'active' ? 'success' : 'default'" size="small" style="margin-top: 8px;">
                      {{ currentLeague.status === 'active' ? t('leagues.active') : t('leagues.archived') }}
                    </n-tag>
                  </div>
                </div>
                <n-dropdown
                  v-if="canManageLeague"
                  :options="manageOptions"
                  trigger="click"
                  @select="handleManageAction"
                >
                  <n-button quaternary circle>
                    <template #icon>
                      <n-icon><MoreVerticalIcon /></n-icon>
                    </template>
                  </n-button>
                </n-dropdown>
              </div>
            </template>

            <n-tabs v-model:value="activeTab" type="line">
              <n-tab name="standings" data-testid="league-standings-tab">
                <template #tab>
                  <n-icon style="margin-right: 4px; vertical-align: middle;"><ChartLineIcon /></n-icon>
                  {{ t('leagues.standings') }}
                </template>
                <league-standings :standings="standings" />
              </n-tab>

              <n-tab name="members" data-testid="league-members-tab">
                <template #tab>
                  <n-icon style="margin-right: 4px; vertical-align: middle;"><PeopleIcon /></n-icon>
                  {{ t('leagues.members') }} ({{ members.length }})
                </template>
                <n-list>
                  <n-list-item
                    v-for="member in members"
                    :key="member.code"
                  >
                    <template #prefix>
                      <n-avatar v-if="member.user_avatar" :src="member.user_avatar" round />
                      <n-avatar v-else color="#f0a020" round>
                        <n-icon><PersonClockIcon /></n-icon>
                      </n-avatar>
                    </template>

                    <div>
                      <div style="font-weight: 500;">
                        {{ member.alias || member.user_name }}
                        <span v-if="member.alias && member.user_name && member.alias !== member.user_name" style="font-size: 0.875rem; opacity: 0.7; margin-left: 8px;">
                          ({{ member.user_name }})
                        </span>
                      </div>
                      <div style="font-size: 0.875rem; opacity: 0.7; margin-top: 4px;">
                        <template v-if="member.status === 'pending'">
                          {{ t('leagues.awaitingJoin') }}
                        </template>
                        <template v-else>
                          {{ t('leagues.joined') }} {{ formatDate(member.joined_at) }}
                        </template>
                      </div>
                    </div>

                    <template #suffix>
                      <n-space :size="8">
                        <n-tag :type="getMemberStatusTagType(member.status)" size="small">
                          {{ getMemberStatusText(member.status) }}
                        </n-tag>
                        <!-- Invitation link for virtual/pending members -->
                        <n-button
                          v-if="(member.status === 'virtual' || member.status === 'pending') && member.invitation_token"
                          type="primary"
                          quaternary
                          size="small"
                          :loading="extendingToken === member.invitation_token"
                          @click="handleInvitationAction(member)"
                        >
                          <template #icon>
                            <n-icon><LinkIcon /></n-icon>
                          </template>
                          {{ t('leagues.openInvitation') }}
                        </n-button>
                        <n-button
                          v-if="canManageLeague && member.status === 'active'"
                          quaternary
                          circle
                          size="small"
                          @click="banMember(member)"
                        >
                          <template #icon>
                            <n-icon><BlockIcon /></n-icon>
                          </template>
                        </n-button>
                        <n-button
                          v-if="canManageLeague && member.status === 'banned'"
                          quaternary
                          circle
                          size="small"
                          @click="unbanMember(member)"
                        >
                          <template #icon>
                            <n-icon><CheckmarkCircleIcon /></n-icon>
                          </template>
                        </n-button>
                      </n-space>
                    </template>
                  </n-list-item>
                </n-list>
              </n-tab>

              <n-tab name="invitation" data-testid="league-invitation-tab">
                <template #tab>
                  <n-icon style="margin-right: 4px; vertical-align: middle;"><PersonAddIcon /></n-icon>
                  {{ t('leagues.invitation') }}
                </template>
                <league-invitation-component :league-code="currentLeague.code" />
              </n-tab>
            </n-tabs>
          </n-card>
        </n-gi>
      </n-grid>

      <!-- Invitation Details Dialog -->
      <invitation-details-dialog
        :model-value="showInvitationDialog"
        :invitation="selectedInvitation"
        @update:model-value="showInvitationDialog = $event"
        @cancel="handleCancelInvitation"
      />
    </template>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, h } from 'vue';
import { NGrid, NGi, NCard, NIcon, NTabs, NTab, NTag, NButton, NDropdown, NList, NListItem, NAvatar, NSpace, NSpin, NAlert } from 'naive-ui';
import { 
  Trophy as TrophyIcon,
  EllipsisVertical as MoreVerticalIcon,
  TrendingUp as ChartLineIcon,
  People as PeopleIcon,
  PersonAdd as PersonAddIcon,
  Archive as ArchiveIcon,
  ArchiveOutline as ArchiveArrowUpIcon,
  Time as PersonClockIcon,
  Ban as BlockIcon,
  CheckmarkCircle as CheckmarkCircleIcon,
  Link as LinkIcon
} from '@vicons/ionicons5';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useLeagueStore } from '@/store/league';
import { useUserStore } from '@/store/user';
import LeagueStandings from '@/components/league/LeagueStandings.vue';
import LeagueInvitationComponent from '@/components/league/LeagueInvitation.vue';
import InvitationDetailsDialog from '@/components/league/InvitationDetailsDialog.vue';
import type { LeagueMember, LeagueInvitation } from '@/api/LeagueApi';

const { t, locale } = useI18n();
const route = useRoute();
const leagueStore = useLeagueStore();
const userStore = useUserStore();

const activeTab = ref('standings');
const extendingToken = ref<string | null>(null);
const showInvitationDialog = ref(false);
const selectedInvitation = ref<LeagueInvitation | null>(null);

const canManageLeague = computed(() => userStore.isSuperAdmin);

const loading = computed(() => leagueStore.isLoading);
const error = computed(() => leagueStore.errorMessage);
const currentLeague = computed(() => leagueStore.currentLeague);
const members = computed(() => leagueStore.currentLeagueMembers);
const standings = computed(() => leagueStore.currentLeagueStandings);

const manageOptions = computed(() => {
  if (!currentLeague.value) return [];
  
  if (currentLeague.value.status === 'active') {
    return [
      {
        label: t('leagues.archive'),
        key: 'archive',
        icon: () => h(NIcon, null, { default: () => h(ArchiveIcon) }),
      },
    ];
  } else {
    return [
      {
        label: t('leagues.unarchive'),
        key: 'unarchive',
        icon: () => h(NIcon, null, { default: () => h(ArchiveArrowUpIcon) }),
      },
    ];
  }
});

const formatDate = (dateStr: string) => {
  const localeMap: Record<string, string> = { 'uk': 'uk-UA', 'en': 'en-US', 'et': 'et-EE' };
  return new Date(dateStr).toLocaleDateString(localeMap[locale.value] || 'en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
};

const getMemberStatusTagType = (status: string): 'success' | 'warning' | 'error' | 'default' => {
  switch (status) {
    case 'active': return 'success';
    case 'pending': return 'warning';
    case 'banned': return 'error';
    default: return 'default';
  }
};

const getMemberStatusText = (status: string) => {
  switch (status) {
    case 'active': return t('leagues.memberActive');
    case 'pending': return t('leagues.pendingMember');
    case 'banned': return t('leagues.memberBanned');
    default: return status;
  }
};

const handleManageAction = (key: string) => {
  if (key === 'archive') {
    archiveLeague();
  } else if (key === 'unarchive') {
    unarchiveLeague();
  }
};

const archiveLeague = async () => {
  if (!currentLeague.value) return;

  try {
    await leagueStore.archiveLeague(currentLeague.value.code);
  } catch (error) {
    console.error('Error archiving league:', error);
  }
};

const unarchiveLeague = async () => {
  if (!currentLeague.value) return;

  try {
    await leagueStore.unarchiveLeague(currentLeague.value.code);
  } catch (error) {
    console.error('Error unarchiving league:', error);
  }
};

const banMember = async (member: LeagueMember) => {
  if (!currentLeague.value) return;

  const confirmed = confirm(`${t('leagues.confirmBan')} ${member.user_name}?`);
  if (!confirmed) return;

  try {
    await leagueStore.banUser(currentLeague.value.code, member.user_id);
  } catch (error) {
    console.error('Error banning member:', error);
  }
};

const unbanMember = async (member: LeagueMember) => {
  if (!currentLeague.value) return;

  const confirmed = confirm(`${t('leagues.confirmUnban')} ${member.user_name}?`);
  if (!confirmed) return;

  try {
    await leagueStore.unbanUser(currentLeague.value.code, member.user_id);
  } catch (error) {
    console.error('Error unbanning member:', error);
  }
};

const handleCancelInvitation = async (token: string) => {
  try {
    await leagueStore.cancelInvitation(token);
    // Reload members to update invitation_token
    await leagueStore.loadCurrentLeagueMembers();
    showInvitationDialog.value = false;
    selectedInvitation.value = null;
  } catch (error) {
    console.error('Error cancelling invitation:', error);
    alert(error instanceof Error ? error.message : t('leagues.error'));
  }
};

const handleInvitationAction = async (member: LeagueMember) => {
  if (!member.invitation_token || !currentLeague.value) return;

  try {
    // Try to preview invitation to check if it's expired
    const preview = await leagueStore.previewInvitation(member.invitation_token);
    
    if (preview.status === 'expired') {
      // Offer to extend the invitation
      const confirmed = confirm(
        `${t('leagues.invitationExpired')} ${member.alias || member.user_name}. ${t('leagues.extendInvitationQuestion')}`
      );
      if (confirmed) {
        extendingToken.value = member.invitation_token;
        try {
          const extended = await leagueStore.extendInvitation(member.invitation_token);
          selectedInvitation.value = extended;
          showInvitationDialog.value = true;
        } catch (error) {
          console.error('Error extending invitation:', error);
          alert(error instanceof Error ? error.message : t('leagues.error'));
        } finally {
          extendingToken.value = null;
        }
      }
    } else {
      // Open invitation details - find invitation from my invitations list
      const myInvitations = await leagueStore.listMyInvitations();
      const invitation = myInvitations.find(inv => inv.token === member.invitation_token);
      if (invitation) {
        selectedInvitation.value = invitation;
        showInvitationDialog.value = true;
      } else {
        // If not found in active, check expired
        const expiredInvitations = await leagueStore.listMyExpiredInvitations();
        const expiredInv = expiredInvitations.find(inv => inv.token === member.invitation_token);
        if (expiredInv) {
          selectedInvitation.value = expiredInv;
          showInvitationDialog.value = true;
        }
      }
    }
  } catch (error) {
    console.error('Error handling invitation:', error);
    // If preview fails, try to find invitation from lists
    try {
      const [myInvitations, expiredInvitations] = await Promise.all([
        leagueStore.listMyInvitations(),
        leagueStore.listMyExpiredInvitations()
      ]);
      const invitation = myInvitations.find(inv => inv.token === member.invitation_token) ||
                        expiredInvitations.find(inv => inv.token === member.invitation_token);
      if (invitation) {
        selectedInvitation.value = invitation;
        showInvitationDialog.value = true;
      } else {
        alert(error instanceof Error ? error.message : t('leagues.error'));
      }
    } catch (err) {
      console.error('Error getting invitation:', err);
      alert(error instanceof Error ? error.message : t('leagues.error'));
    }
  }
};

onMounted(async () => {
  const leagueCode = route.params.code as string;
  if (leagueCode) {
    try {
      await leagueStore.setCurrentLeague(leagueCode);
    } catch (error) {
      console.error('Error loading league:', error);
    }
  }
});
</script>

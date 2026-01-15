<template>
  <div>
    <n-card style="margin-bottom: 16px; background: rgba(32, 128, 240, 0.1);">
      <div style="display: flex; align-items: center; gap: 12px;">
        <n-icon size="24" color="#2080f0">
          <InformationIcon />
        </n-icon>
        <div>
          <div style="font-size: 1rem; font-weight: 500;">{{ t('leagues.howToInvite') }}</div>
          <div style="font-size: 0.875rem; opacity: 0.7;">
            {{ t('leagues.invitationInfo') }}
          </div>
        </div>
      </div>
    </n-card>

    <n-card>
      <div style="text-align: center; padding: 32px 0;">
        <n-button
            type="primary"
            size="large"
            @click="showCreateDialog = true"
        >
          <template #icon>
            <n-icon><LinkPlusIcon /></n-icon>
          </template>
          {{ t('leagues.createInvitation') }}
        </n-button>
      </div>

      <n-alert
        v-if="error"
        type="error"
        closable
        @close="error = null"
        style="margin-top: 16px;"
      >
        {{ error }}
      </n-alert>
    </n-card>

    <!-- Active Invitations List -->
    <n-card v-if="invitations.length > 0" style="margin-top: 16px;">
      <template #header>
        <div style="display: flex; align-items: center; gap: 8px;">
          <n-icon><LinkVariantIcon /></n-icon>
          {{ t('leagues.activeInvitations') }}
        </div>
      </template>
      <n-list>
        <n-list-item
          v-for="invitation in invitations"
          :key="invitation.token"
          clickable
          @click="openInvitationDetails(invitation)"
        >
          <template #prefix>
            <n-icon color="#2080f0"><PersonClockIcon /></n-icon>
          </template>

          <div>
            <div style="font-weight: 500;">{{ invitation.player_alias }}</div>
            <div style="font-size: 0.875rem; opacity: 0.7;">
              {{ t('leagues.validUntil') }} {{ formatDate(invitation.expires_at) }}
            </div>
          </div>

          <template #suffix>
            <n-icon color="#808080"><ChevronForwardIcon /></n-icon>
          </template>
        </n-list-item>
      </n-list>
    </n-card>

    <!-- Expired Invitations List -->
    <n-card v-if="expiredInvitations.length > 0" style="margin-top: 16px;">
      <template #header>
        <div style="display: flex; align-items: center; gap: 8px; color: #f0a020;">
          <n-icon color="#f0a020"><ClockAlertIcon /></n-icon>
          {{ t('leagues.expiredInvitations') }}
        </div>
      </template>
      <n-list>
        <n-list-item
          v-for="invitation in expiredInvitations"
          :key="invitation.token"
        >
          <template #prefix>
            <n-icon color="#f0a020"><PersonClockOutlineIcon /></n-icon>
          </template>

          <div>
            <div style="font-weight: 500; opacity: 0.7;">{{ invitation.player_alias }}</div>
            <div style="font-size: 0.875rem; opacity: 0.7;">
              {{ t('leagues.expiredAt') }} {{ formatDate(invitation.expires_at) }}
            </div>
          </div>

          <template #suffix>
            <n-button
              type="primary"
              quaternary
              size="small"
              :loading="extendingToken === invitation.token"
              @click.stop="handleExtendInvitation(invitation)"
            >
              <template #icon>
                <n-icon><RefreshIcon /></n-icon>
              </template>
              {{ t('leagues.extendInvitation') }}
            </n-button>
          </template>
        </n-list-item>
      </n-list>
    </n-card>

    <!-- Loading indicator for invitations list -->
    <n-card v-if="loadingList" style="margin-top: 16px;">
      <div style="text-align: center; padding: 32px 0;">
        <n-spin size="small" />
      </div>
    </n-card>

    <!-- Create Invitation Dialog -->
    <n-modal v-model:show="showCreateDialog" preset="card" :title="t('leagues.createInvitation')" style="max-width: 400px;">
      <n-form-item :label="t('leagues.playerAlias')">
        <n-input
          v-model:value="newPlayerAlias"
          :placeholder="t('leagues.playerAliasHint')"
          autofocus
          @keyup.enter="generateInvitation"
        />
      </n-form-item>
      <template #action>
        <div style="display: flex; justify-content: flex-end; gap: 8px; width: 100%;">
          <n-button @click="showCreateDialog = false">
            {{ t('common.cancel') }}
          </n-button>
          <n-button
            type="primary"
            :loading="generating"
            :disabled="!newPlayerAlias.trim()"
            @click="generateInvitation"
          >
            {{ t('leagues.createInvitation') }}
          </n-button>
        </div>
      </template>
    </n-modal>

    <!-- Invitation Details Dialog -->
    <invitation-details-dialog
      :model-value="showDialog"
      :invitation="selectedInvitation"
      @update:model-value="showDialog = $event"
      @cancel="handleCancelInvitation"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { NCard, NButton, NIcon, NAlert, NList, NListItem, NModal, NFormItem, NInput, NSpin } from 'naive-ui';
import { 
  Information as InformationIcon,
  Add as LinkPlusIcon,
  Link as LinkVariantIcon,
  Time as PersonClockIcon,
  ChevronForward as ChevronForwardIcon,
  TimeOutline as ClockAlertIcon,
  PersonOutline as PersonClockOutlineIcon,
  Refresh as RefreshIcon
} from '@vicons/ionicons5';
import { useI18n } from 'vue-i18n';
import { useLeagueStore } from '@/store/league';
import InvitationDetailsDialog from './InvitationDetailsDialog.vue';
import type { LeagueInvitation } from '@/api/LeagueApi';

interface Props {
  leagueCode: string;
}

const { t, locale } = useI18n();
defineProps<Props>();

const leagueStore = useLeagueStore();

const generating = ref(false);
const loadingList = ref(false);
const invitations = ref<LeagueInvitation[]>([]);
const expiredInvitations = ref<LeagueInvitation[]>([]);
const selectedInvitation = ref<LeagueInvitation | null>(null);
const showDialog = ref(false);
const showCreateDialog = ref(false);
const newPlayerAlias = ref('');
const extendingToken = ref<string | null>(null);
const error = ref<string | null>(null);

const loadInvitations = async () => {
  loadingList.value = true;
  try {
    const [active, expired] = await Promise.all([
      leagueStore.listMyInvitations(),
      leagueStore.listMyExpiredInvitations()
    ]);
    invitations.value = active;
    expiredInvitations.value = expired;
  } catch (err) {
    console.error('Error loading invitations:', err);
  } finally {
    loadingList.value = false;
  }
};

const generateInvitation = async () => {
  if (!newPlayerAlias.value.trim()) return;

  generating.value = true;
  error.value = null;

  try {
    const invitation = await leagueStore.createInvitation(newPlayerAlias.value.trim());
    // Add new invitation to the list
    invitations.value.unshift(invitation);
    // Close create dialog and open details dialog
    showCreateDialog.value = false;
    newPlayerAlias.value = '';
    selectedInvitation.value = invitation;
    showDialog.value = true;
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('leagues.error');
    console.error('Error generating invitation:', err);
  } finally {
    generating.value = false;
  }
};

const openInvitationDetails = (invitation: LeagueInvitation) => {
  selectedInvitation.value = invitation;
  showDialog.value = true;
};

const handleCancelInvitation = async (token: string) => {
  try {
    await leagueStore.cancelInvitation(token);
    // Find the invitation to move it to expired
    const cancelledInvitation = invitations.value.find(inv => inv.token === token);
    if (cancelledInvitation) {
      // Remove from active list
      invitations.value = invitations.value.filter(inv => inv.token !== token);
      // Add to expired list (mark as expired by updating expires_at to now)
      const expiredInvitation = {
        ...cancelledInvitation,
        expires_at: new Date().toISOString()
      };
      expiredInvitations.value.push(expiredInvitation);
    }
    showDialog.value = false;
    selectedInvitation.value = null;
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('leagues.error');
    console.error('Error cancelling invitation:', err);
  }
};

const handleExtendInvitation = async (invitation: LeagueInvitation) => {
  extendingToken.value = invitation.token;
  error.value = null;

  try {
    const extended = await leagueStore.extendInvitation(invitation.token);
    // Move from expired to active
    expiredInvitations.value = expiredInvitations.value.filter(inv => inv.token !== invitation.token);
    invitations.value.unshift(extended);
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('leagues.error');
    console.error('Error extending invitation:', err);
  } finally {
    extendingToken.value = null;
  }
};

const formatDate = (dateStr: string) => {
  const localeMap: Record<string, string> = { 'uk': 'uk-UA', 'en': 'en-US', 'et': 'et-EE' };
  return new Date(dateStr).toLocaleString(localeMap[locale.value] || 'en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

onMounted(() => {
  loadInvitations();
});
</script>

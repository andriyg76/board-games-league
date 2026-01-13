<template>
  <div>
    <v-card variant="tonal" color="primary" class="mb-4">
      <v-card-text>
        <div class="d-flex align-center">
          <v-icon start size="large">mdi-information</v-icon>
          <div>
            <div class="text-subtitle-1">{{ t('leagues.howToInvite') }}</div>
            <div class="text-caption">
              {{ t('leagues.invitationInfo') }}
            </div>
          </div>
        </div>
      </v-card-text>
    </v-card>

    <v-card elevation="2">
      <v-card-text>
        <div class="text-center py-4">
          <v-btn
            color="primary"
            size="large"
            @click="showCreateDialog = true"
          >
            <v-icon start>mdi-link-plus</v-icon>
            {{ t('leagues.createInvitation') }}
          </v-btn>
        </div>

        <v-alert
          v-if="error"
          type="error"
          variant="tonal"
          class="mt-4"
          closable
          @click:close="error = null"
        >
          {{ error }}
        </v-alert>
      </v-card-text>
    </v-card>

    <!-- Active Invitations List -->
    <v-card elevation="2" class="mt-4" v-if="invitations.length > 0">
      <v-card-title class="d-flex align-center">
        <v-icon start>mdi-link-variant</v-icon>
        {{ t('leagues.activeInvitations') }}
      </v-card-title>
      <v-divider />
      <v-list>
        <v-list-item
          v-for="invitation in invitations"
          :key="invitation.token"
          @click="openInvitationDetails(invitation)"
          class="cursor-pointer"
        >
          <template v-slot:prepend>
            <v-icon>mdi-account-clock</v-icon>
          </template>

          <v-list-item-title>
            {{ invitation.player_alias }}
          </v-list-item-title>
          <v-list-item-subtitle>
            {{ t('leagues.validUntil') }} {{ formatDate(invitation.expires_at) }}
          </v-list-item-subtitle>

          <template v-slot:append>
            <v-icon>mdi-chevron-right</v-icon>
          </template>
        </v-list-item>
      </v-list>
    </v-card>

    <!-- Expired Invitations List -->
    <v-card elevation="2" class="mt-4" v-if="expiredInvitations.length > 0">
      <v-card-title class="d-flex align-center text-warning">
        <v-icon start color="warning">mdi-clock-alert</v-icon>
        {{ t('leagues.expiredInvitations') }}
      </v-card-title>
      <v-divider />
      <v-list>
        <v-list-item
          v-for="invitation in expiredInvitations"
          :key="invitation.token"
          class="cursor-pointer"
        >
          <template v-slot:prepend>
            <v-icon color="warning">mdi-account-clock-outline</v-icon>
          </template>

          <v-list-item-title class="text-medium-emphasis">
            {{ invitation.player_alias }}
          </v-list-item-title>
          <v-list-item-subtitle>
            {{ t('leagues.expiredAt') }} {{ formatDate(invitation.expires_at) }}
          </v-list-item-subtitle>

          <template v-slot:append>
            <v-btn
              color="primary"
              variant="tonal"
              size="small"
              :loading="extendingToken === invitation.token"
              @click.stop="handleExtendInvitation(invitation)"
            >
              <v-icon start size="small">mdi-refresh</v-icon>
              {{ t('leagues.extendInvitation') }}
            </v-btn>
          </template>
        </v-list-item>
      </v-list>
    </v-card>

    <!-- Loading indicator for invitations list -->
    <v-card elevation="2" class="mt-4" v-if="loadingList">
      <v-card-text class="text-center py-4">
        <v-progress-circular indeterminate color="primary" size="24" />
      </v-card-text>
    </v-card>

    <!-- Create Invitation Dialog -->
    <v-dialog v-model="showCreateDialog" max-width="400">
      <v-card>
        <v-card-title class="bg-primary">
          <v-icon start>mdi-account-plus</v-icon>
          {{ t('leagues.createInvitation') }}
        </v-card-title>
        <v-card-text class="pt-4">
          <v-text-field
            v-model="newPlayerAlias"
            :label="t('leagues.playerAlias')"
            :hint="t('leagues.playerAliasHint')"
            persistent-hint
            variant="outlined"
            autofocus
            @keyup.enter="generateInvitation"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showCreateDialog = false">
            {{ t('common.cancel') }}
          </v-btn>
          <v-btn
            color="primary"
            variant="flat"
            :loading="generating"
            :disabled="!newPlayerAlias.trim()"
            @click="generateInvitation"
          >
            {{ t('leagues.createInvitation') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Invitation Details Dialog -->
    <invitation-details-dialog
      v-model="showDialog"
      :invitation="selectedInvitation"
      @cancel="handleCancelInvitation"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
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
    // Remove from list
    invitations.value = invitations.value.filter(inv => inv.token !== token);
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

<style scoped>
.cursor-pointer {
  cursor: pointer;
}
</style>

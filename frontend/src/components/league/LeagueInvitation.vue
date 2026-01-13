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
            :loading="generating"
            @click="generateInvitation"
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
            <v-icon>mdi-email-outline</v-icon>
          </template>

          <v-list-item-title>
            {{ t('leagues.invitationCreated') }} {{ formatDate(invitation.created_at) }}
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

    <!-- Loading indicator for invitations list -->
    <v-card elevation="2" class="mt-4" v-if="loadingList">
      <v-card-text class="text-center py-4">
        <v-progress-circular indeterminate color="primary" size="24" />
      </v-card-text>
    </v-card>

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
const selectedInvitation = ref<LeagueInvitation | null>(null);
const showDialog = ref(false);
const error = ref<string | null>(null);

const loadInvitations = async () => {
  loadingList.value = true;
  try {
    invitations.value = await leagueStore.listMyInvitations();
  } catch (err) {
    console.error('Error loading invitations:', err);
  } finally {
    loadingList.value = false;
  }
};

const generateInvitation = async () => {
  generating.value = true;
  error.value = null;

  try {
    const result = await leagueStore.createInvitation();
    // Add new invitation to the list
    invitations.value.unshift(result.invitation);
    // Open the dialog to show the new invitation
    selectedInvitation.value = result.invitation;
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

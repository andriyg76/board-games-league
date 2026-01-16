<template>
  <n-modal
    :show="modelValue"
    preset="card"
    :title="t('leagues.invitationDetails')"
    style="max-width: 500px;"
    data-testid="invitation-details-dialog"
    @update:show="$emit('update:modelValue', $event)"
  >
    <template v-if="invitation">
      <!-- Player Alias -->
      <div style="font-size: 0.875rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.playerAlias') }}</div>
      <n-card style="margin-bottom: 16px;" bordered>
        <div style="display: flex; align-items: center; gap: 8px;">
          <n-input
            v-if="editingAlias"
            v-model:value="newAlias"
            :placeholder="t('leagues.playerAlias')"
            size="small"
            style="flex: 1;"
            @keyup.enter="saveAlias"
            @keyup.escape="cancelEditAlias"
          />
          <span v-else style="flex: 1;">{{ invitation?.player_alias }}</span>
          <n-button
            v-if="editingAlias"
            quaternary
            circle
            size="small"
            type="success"
            :loading="savingAlias"
            @click="saveAlias"
          >
            <template #icon>
              <n-icon><CheckIcon /></n-icon>
            </template>
          </n-button>
          <n-button
            v-if="editingAlias"
            quaternary
            circle
            size="small"
            @click="cancelEditAlias"
          >
            <template #icon>
              <n-icon><CloseIcon /></n-icon>
            </template>
          </n-button>
          <n-button
            v-if="!editingAlias"
            quaternary
            circle
            size="small"
            @click="startEditAlias"
          >
            <template #icon>
              <n-icon><PencilIcon /></n-icon>
            </template>
          </n-button>
        </div>
      </n-card>

      <!-- Invitation URL -->
      <div style="font-size: 0.875rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.invitationLink') }}</div>
      <n-card style="margin-bottom: 16px;" bordered>
        <div style="display: flex; align-items: center; gap: 8px;">
          <n-input
            :value="invitationLink"
            readonly
            size="small"
            style="flex: 1;"
            data-testid="invitation-link-input"
          />
          <n-button
            quaternary
            circle
            size="small"
            @click="copyToClipboard"
          >
            <template #icon>
              <n-icon><CopyIcon /></n-icon>
            </template>
          </n-button>
        </div>
      </n-card>

      <n-alert
        v-if="copied"
        type="success"
        size="small"
        style="margin-bottom: 16px;"
      >
        <template #icon>
          <n-icon><CheckCircleIcon /></n-icon>
        </template>
        {{ t('leagues.linkCopied') }}
      </n-alert>

      <!-- QR Code -->
      <div style="font-size: 0.875rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.qrCode') }}</div>
      <div style="display: flex; justify-content: center; margin-bottom: 16px;">
        <div style="padding: 16px; background: white; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.12);">
          <qrcode-vue
            :value="invitationLink"
            :size="200"
            level="M"
          />
        </div>
      </div>

      <!-- Expiry info -->
      <n-card style="background: rgba(240, 160, 32, 0.1);">
        <div style="display: flex; align-items: center; gap: 8px;">
          <n-icon color="#f0a020" size="16"><ClockAlertIcon /></n-icon>
          <div style="font-size: 0.875rem;">
            {{ t('leagues.validUntil') }} {{ formatExpiryDate(invitation?.expires_at) }}
          </div>
        </div>
      </n-card>
    </template>

    <template #action>
      <div style="display: flex; justify-content: space-between; width: 100%;">
        <n-button
          type="error"
          quaternary
          :loading="cancelling"
          @click="handleCancel"
        >
          <template #icon>
            <n-icon><CancelIcon /></n-icon>
          </template>
          {{ t('leagues.cancelInvitation') }}
        </n-button>
        <n-button
          type="primary"
          @click="$emit('update:modelValue', false)"
        >
          {{ t('common.close') }}
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { NModal, NCard, NInput, NButton, NIcon, NAlert } from 'naive-ui';
import { 
  Checkmark as CheckIcon,
  Close as CloseIcon,
  Pencil as PencilIcon,
  Copy as CopyIcon,
  CheckmarkCircle as CheckCircleIcon,
  TimeOutline as ClockAlertIcon,
  CloseCircle as CancelIcon
} from '@vicons/ionicons5';
import { useI18n } from 'vue-i18n';
import QrcodeVue from 'qrcode.vue';
import { useLeagueStore } from '@/store/league';
import type { LeagueInvitation } from '@/api/LeagueApi';

interface Props {
  modelValue: boolean;
  invitation: LeagueInvitation | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  'cancel': [token: string];
  'aliasUpdated': [newAlias: string];
}>();

const { t, locale } = useI18n();
const leagueStore = useLeagueStore();

const copied = ref(false);
const cancelling = ref(false);
const editingAlias = ref(false);
const newAlias = ref('');
const savingAlias = ref(false);

// Reset editing state when invitation changes
watch(() => props.invitation, () => {
  editingAlias.value = false;
  newAlias.value = props.invitation?.player_alias || '';
});

const invitationLink = computed(() => {
  if (!props.invitation) return '';
  return `${window.location.origin}/ui/leagues/join/${props.invitation.token}`;
});

const copyToClipboard = async () => {
  if (!invitationLink.value) return;

  try {
    await navigator.clipboard.writeText(invitationLink.value);
    copied.value = true;
    setTimeout(() => {
      copied.value = false;
    }, 3000);
  } catch (err) {
    console.error('Error copying to clipboard:', err);
  }
};

const handleCancel = async () => {
  if (!props.invitation) return;
  
  cancelling.value = true;
  try {
    emit('cancel', props.invitation.token);
  } finally {
    cancelling.value = false;
  }
};

const startEditAlias = () => {
  newAlias.value = props.invitation?.player_alias || '';
  editingAlias.value = true;
};

const cancelEditAlias = () => {
  editingAlias.value = false;
  newAlias.value = props.invitation?.player_alias || '';
};

const saveAlias = async () => {
  if (!props.invitation?.membership_id || !newAlias.value.trim()) return;

  savingAlias.value = true;
  try {
    await leagueStore.updatePendingMemberAlias(props.invitation.membership_id, newAlias.value.trim());
    emit('aliasUpdated', newAlias.value.trim());
    editingAlias.value = false;
  } catch (err) {
    console.error('Error saving alias:', err);
  } finally {
    savingAlias.value = false;
  }
};

const formatExpiryDate = (dateStr: string | undefined) => {
  if (!dateStr) return '';

  const localeMap: Record<string, string> = { 'uk': 'uk-UA', 'en': 'en-US', 'et': 'et-EE' };
  return new Date(dateStr).toLocaleString(localeMap[locale.value] || 'en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};
</script>

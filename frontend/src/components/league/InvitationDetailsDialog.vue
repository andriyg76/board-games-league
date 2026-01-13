<template>
  <v-dialog
    :model-value="modelValue"
    max-width="500"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <v-card>
      <v-card-title class="d-flex align-center bg-primary">
        <v-icon start>mdi-link-variant</v-icon>
        {{ t('leagues.invitationDetails') }}
      </v-card-title>

      <v-card-text class="py-6">
        <!-- Invitation URL -->
        <div class="text-subtitle-2 mb-2">{{ t('leagues.invitationLink') }}</div>
        <v-card variant="outlined" class="mb-4">
          <v-card-text class="pa-2">
            <div class="d-flex align-center gap-2">
              <v-text-field
                :model-value="invitationLink"
                readonly
                hide-details
                variant="plain"
                density="compact"
                class="flex-grow-1"
              />
              <v-btn
                icon="mdi-content-copy"
                variant="tonal"
                size="small"
                @click="copyToClipboard"
              />
            </div>
          </v-card-text>
        </v-card>

        <v-alert
          v-if="copied"
          type="success"
          variant="tonal"
          density="compact"
          class="mb-4"
        >
          <v-icon start size="small">mdi-check-circle</v-icon>
          {{ t('leagues.linkCopied') }}
        </v-alert>

        <!-- QR Code -->
        <div class="text-subtitle-2 mb-2">{{ t('leagues.qrCode') }}</div>
        <div class="d-flex justify-center mb-4">
          <div class="qr-container pa-4 bg-white rounded-lg">
            <qrcode-vue
              :value="invitationLink"
              :size="200"
              level="M"
            />
          </div>
        </div>

        <!-- Expiry info -->
        <v-card variant="tonal" color="warning">
          <v-card-text class="py-2">
            <div class="d-flex align-center">
              <v-icon start size="small">mdi-clock-alert</v-icon>
              <div class="text-caption">
                {{ t('leagues.validUntil') }} {{ formatExpiryDate(invitation?.expires_at) }}
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-card-text>

      <v-divider />

      <v-card-actions class="pa-4">
        <v-btn
          color="error"
          variant="text"
          :loading="cancelling"
          @click="handleCancel"
        >
          <v-icon start>mdi-cancel</v-icon>
          {{ t('leagues.cancelInvitation') }}
        </v-btn>
        <v-spacer />
        <v-btn
          color="primary"
          variant="flat"
          @click="$emit('update:modelValue', false)"
        >
          {{ t('common.close') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import QrcodeVue from 'qrcode.vue';
import type { LeagueInvitation } from '@/api/LeagueApi';

interface Props {
  modelValue: boolean;
  invitation: LeagueInvitation | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  'cancel': [token: string];
}>();

const { t, locale } = useI18n();

const copied = ref(false);
const cancelling = ref(false);

const invitationLink = computed(() => {
  if (!props.invitation) return '';
  return `${window.location.origin}/invite/${props.invitation.token}`;
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

<style scoped>
.gap-2 {
  gap: 8px;
}

.qr-container {
  display: inline-block;
  border: 1px solid rgba(0, 0, 0, 0.12);
}
</style>


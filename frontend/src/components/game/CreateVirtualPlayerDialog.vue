<template>
  <n-modal
    v-model:show="dialog"
    preset="card"
    :title="$t('game.createVirtualPlayer')"
    style="max-width: 500px;"
    :mask-closable="false"
  >
    <n-form-item :label="$t('game.playerAlias')" :validation-status="errorMessage ? 'error' : undefined" :feedback="errorMessage">
      <n-input
        v-model:value="alias"
        :placeholder="$t('game.playerAlias')"
        :loading="checking"
        data-testid="virtual-player-alias-input"
        autofocus
        @update:value="debouncedCheckAlias"
      />
    </n-form-item>
    
    <n-alert
      v-if="invitationLink"
      type="success"
      size="small"
      style="margin-top: 16px;"
    >
      <div style="display: flex; align-items: center; justify-content: space-between;">
        <span style="font-size: 0.875rem;">{{ $t('game.invitationCreated') }}</span>
        <n-button
          size="small"
          quaternary
          @click="copyLink"
        >
          <template #icon>
            <n-icon><CopyIcon /></n-icon>
          </template>
          {{ $t('common.copy') }}
        </n-button>
      </div>
    </n-alert>

    <template #action>
      <div style="display: flex; justify-content: flex-end; gap: 8px; width: 100%;">
        <n-button @click="close">
          {{ $t('common.cancel') }}
        </n-button>
        <n-button
          type="primary"
          :loading="creating"
          :disabled="!canCreate"
          data-testid="create-virtual-player-button"
          @click="create"
        >
          {{ $t('common.create') }}
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { NModal, NFormItem, NInput, NAlert, NButton, NIcon } from 'naive-ui';
import { Copy as CopyIcon } from '@vicons/ionicons5';
import LeagueApi, { SuggestedPlayer } from '@/api/LeagueApi';

const props = defineProps<{
  modelValue: boolean;
  leagueCode: string;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  'created': [player: SuggestedPlayer];
}>();

const dialog = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
});

const alias = ref('');
const errorMessage = ref('');
const checking = ref(false);
const creating = ref(false);
const invitationLink = ref('');

let checkTimeout: ReturnType<typeof setTimeout> | null = null;

const canCreate = computed(() => 
  alias.value.trim().length >= 2 && 
  !errorMessage.value && 
  !checking.value &&
  !invitationLink.value
);

const debouncedCheckAlias = () => {
  if (checkTimeout) clearTimeout(checkTimeout);
  errorMessage.value = '';
  
  if (alias.value.trim().length < 2) {
    errorMessage.value = 'Alias must be at least 2 characters';
    return;
  }
  
  checking.value = true;
  checkTimeout = setTimeout(async () => {
    // The backend will validate alias uniqueness on create
    // For now, just clear the checking state
    checking.value = false;
  }, 300);
};

const create = async () => {
  if (!canCreate.value) return;
  
  creating.value = true;
  errorMessage.value = '';
  
  try {
    const invitation = await LeagueApi.createInvitation(props.leagueCode, alias.value.trim());
    
    // Build invitation link
    const publicBase = (import.meta.env.VITE_PUBLIC_WEB_BASE_URL || '').trim();
    const baseUrl = publicBase ? publicBase.replace(/\/$/, '') : window.location.origin;
    invitationLink.value = `${baseUrl}/join/${invitation.token}`;
    
    // Copy to clipboard
    await copyLink();
    
    // Create the suggested player object
    const player: SuggestedPlayer = {
      membership_code: invitation.membership_code || invitation.membership_id || '',
      alias: invitation.player_alias,
      is_virtual: true,
    };
    
    emit('created', player);
    
    // Close dialog after short delay to show success message
    setTimeout(() => {
      close();
    }, 1500);
    
  } catch (error: unknown) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to create virtual player';
  } finally {
    creating.value = false;
  }
};

const copyLink = async () => {
  if (!invitationLink.value) return;
  
  try {
    await navigator.clipboard.writeText(invitationLink.value);
    // Could show a toast here
  } catch (error) {
    console.error('Failed to copy link:', error);
  }
};

const close = () => {
  dialog.value = false;
  // Reset state after dialog closes
  setTimeout(() => {
    alias.value = '';
    errorMessage.value = '';
    invitationLink.value = '';
  }, 300);
};

// Reset when dialog opens
watch(dialog, (newValue) => {
  if (newValue) {
    alias.value = '';
    errorMessage.value = '';
    invitationLink.value = '';
  }
});
</script>

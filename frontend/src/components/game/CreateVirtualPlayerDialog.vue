<template>
  <v-dialog v-model="dialog" max-width="500" persistent>
    <v-card>
      <v-card-title>
        {{ $t('game.createVirtualPlayer') }}
      </v-card-title>
      
      <v-card-text>
        <v-text-field
          v-model="alias"
          :label="$t('game.playerAlias')"
          :error-messages="errorMessage"
          :loading="checking"
          variant="outlined"
          autofocus
          @input="debouncedCheckAlias"
        />
        
        <v-alert
          v-if="invitationLink"
          type="success"
          density="compact"
          class="mt-2"
        >
          <div class="d-flex align-center justify-space-between">
            <span class="text-body-2">{{ $t('game.invitationCreated') }}</span>
            <v-btn
              size="small"
              variant="text"
              @click="copyLink"
            >
              <v-icon start>mdi-content-copy</v-icon>
              {{ $t('common.copy') }}
            </v-btn>
          </div>
        </v-alert>
      </v-card-text>
      
      <v-card-actions>
        <v-spacer />
        <v-btn
          variant="text"
          @click="close"
        >
          {{ $t('common.cancel') }}
        </v-btn>
        <v-btn
          color="primary"
          variant="elevated"
          :loading="creating"
          :disabled="!canCreate"
          @click="create"
        >
          {{ $t('common.create') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
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
    invitationLink.value = `${window.location.origin}/join/${invitation.token}`;
    
    // Copy to clipboard
    await copyLink();
    
    // Create the suggested player object
    const player: SuggestedPlayer = {
      membership_id: invitation.membership_id || '',
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

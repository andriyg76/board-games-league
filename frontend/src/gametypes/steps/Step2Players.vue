<template>
  <n-card>
    <template #header>
      {{ $t('game.selectPlayers') }}
    </template>
    <n-spin v-if="loading" size="large" style="display: flex; justify-content: center; padding: 64px;">
      <template #description>
        {{ $t('common.loading') }}
      </template>
    </n-spin>
    <PlayerSelector
      v-else
      ref="playerSelectorRef"
      :suggested-players="suggestedPlayers"
      :min-players="minPlayers"
      :max-players="maxPlayers"
      :has-moderator="hasModerator"
      :league-code="leagueCode"
      @update:selectedPlayers="$emit('update:selectedPlayers', $event)"
      @update:moderatorId="$emit('update:moderatorId', $event)"
    />
    <template #action>
      <div style="display: flex; justify-content: space-between; align-items: center; padding: 16px;">
        <n-button quaternary @click="$emit('back')">
          <template #icon>
            <n-icon><ChevronBackIcon /></n-icon>
          </template>
          {{ $t('game.back') }}
        </n-button>
        <n-button
          type="primary"
          :disabled="!canProceed"
          :loading="saving"
          @click="$emit('next')"
        >
          {{ $t('game.next') }}
          <template #icon>
            <n-icon><ChevronForwardIcon /></n-icon>
          </template>
        </n-button>
      </div>
    </template>
  </n-card>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { NCard, NSpin, NButton, NIcon } from 'naive-ui';
import { ChevronBack as ChevronBackIcon, ChevronForward as ChevronForwardIcon } from '@vicons/ionicons5';
import PlayerSelector from '@/components/game/PlayerSelector.vue';
import type { SuggestedPlayersResponse } from '@/api/LeagueApi';

defineProps<{
  suggestedPlayers: SuggestedPlayersResponse | null;
  minPlayers: number;
  maxPlayers: number;
  hasModerator: boolean;
  leagueCode: string;
  loading: boolean;
  saving: boolean;
  canProceed: boolean;
}>();

import type { SuggestedPlayer } from '@/api/LeagueApi';

defineEmits<{
  'update:selectedPlayers': [players: SuggestedPlayer[]];
  'update:moderatorId': [moderatorId: string | null];
  back: [];
  next: [];
}>();

const playerSelectorRef = ref<InstanceType<typeof PlayerSelector> | null>(null);
</script>

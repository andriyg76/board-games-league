<template>
  <v-card flat>
    <v-card-title>{{ $t('game.selectPlayers') }}</v-card-title>
    <v-card-text>
      <div v-if="loading" class="text-center pa-4">
        <v-progress-circular indeterminate color="primary" />
        <p class="mt-2">{{ $t('common.loading') }}</p>
      </div>
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
    </v-card-text>
    <v-card-actions>
      <v-btn variant="text" @click="$emit('back')">
        <v-icon start>mdi-arrow-left</v-icon>
        {{ $t('game.back') }}
      </v-btn>
      <v-spacer />
      <v-btn
        color="primary"
        :disabled="!canProceed"
        :loading="saving"
        @click="$emit('next')"
      >
        {{ $t('game.next') }}
        <v-icon end>mdi-arrow-right</v-icon>
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
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

<template>
  <v-card flat>
    <v-card-title>{{ $t('game.selectGameType') }}</v-card-title>
    <v-card-text>
      <v-list>
        <v-list-item
          v-for="gameType in gameTypes"
          :key="gameType.code"
          :class="{ 'bg-primary-lighten-4': selectedGameType === gameType.code }"
          @click="$emit('select', gameType)"
        >
          <template #prepend>
            <v-icon>{{ gameType.icon || 'mdi-dice-multiple' }}</v-icon>
          </template>
          <v-list-item-title>
            {{ getLocalizedName(gameType.names) }}
          </v-list-item-title>
          <v-list-item-subtitle>
            {{ gameType.min_players }}-{{ gameType.max_players }} {{ $t('gameTypes.players') }}
          </v-list-item-subtitle>
          <template #append>
            <v-icon v-if="selectedGameType === gameType.code" color="primary">
              mdi-check
            </v-icon>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
    <v-card-actions>
      <v-spacer />
      <v-btn
        color="primary"
        :disabled="!selectedGameType"
        @click="$emit('next')"
      >
        {{ $t('game.next') }}
        <v-icon end>mdi-arrow-right</v-icon>
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { GameType, getLocalizedName } from '@/api/GameApi';

defineProps<{
  gameTypes: GameType[];
  selectedGameType: string;
}>();

defineEmits<{
  select: [gameType: GameType];
  next: [];
}>();
</script>

<style scoped>
.bg-primary-lighten-4 {
  background-color: rgba(var(--v-theme-primary), 0.1);
}
</style>

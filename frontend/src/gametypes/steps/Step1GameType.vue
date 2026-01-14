<template>
  <n-card>
    <template #header>
      {{ $t('game.selectGameType') }}
    </template>
    <n-list>
      <n-list-item
        v-for="gameType in gameTypes"
        :key="gameType.code"
        :class="{ 'selected-game-type': selectedGameType === gameType.code }"
        clickable
        @click="$emit('select', gameType)"
      >
        <template #prefix>
          <n-icon :size="24">
            <component :is="getIconComponent(gameType.icon)" />
          </n-icon>
        </template>
        <div>
          <div style="font-weight: 500;">{{ getLocalizedName(gameType.names) }}</div>
          <div style="font-size: 0.875rem; opacity: 0.7;">
            {{ gameType.min_players }}-{{ gameType.max_players }} {{ $t('gameTypes.players') }}
          </div>
        </div>
        <template #suffix>
          <n-icon v-if="selectedGameType === gameType.code" color="#2080f0" :size="20">
            <CheckIcon />
          </n-icon>
        </template>
      </n-list-item>
    </n-list>
    <template #action>
      <div style="display: flex; justify-content: flex-end; padding: 16px;">
        <n-button
          type="primary"
          :disabled="!selectedGameType"
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
import { NCard, NList, NListItem, NIcon, NButton } from 'naive-ui';
import { Dice as DiceIcon } from '@vicons/ionicons5';
import { Checkmark as CheckIcon, ChevronForward as ChevronForwardIcon } from '@vicons/ionicons5';
import { GameType, getLocalizedName } from '@/api/GameApi';

defineProps<{
  gameTypes: GameType[];
  selectedGameType: string;
}>();

defineEmits<{
  select: [gameType: GameType];
  next: [];
}>();

const getIconComponent = (icon?: string) => {
  // For now, use dice icon as default
  // In a full implementation, you'd map icon strings to components
  return DiceIcon;
};
</script>

<style scoped>
.selected-game-type {
  background-color: rgba(32, 128, 240, 0.1);
}
</style>

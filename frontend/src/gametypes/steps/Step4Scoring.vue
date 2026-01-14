<template>
  <n-card>
    <template #header>
      {{ $t('game.enterScores') }}
    </template>
    <n-table>
      <thead>
        <tr>
          <th>{{ $t('leagues.player') }}</th>
          <th>{{ $t('leagues.points') }}</th>
          <th>{{ $t('game.position') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="player in players" :key="player.membership_id">
          <td>
            <div style="display: flex; align-items: center; gap: 8px;">
              <span>{{ player.alias }}</span>
              <n-tag v-if="player.is_moderator" size="small" type="warning">
                {{ $t('roleTypes.moderator') }}
              </n-tag>
            </div>
          </td>
          <td>
            <n-input-number
              :value="scores[player.membership_id] || 0"
              @update:value="$emit('updateScore', player.membership_id, Number($event || 0))"
              :min="0"
              size="small"
              style="max-width: 100px"
            />
          </td>
          <td>
            <n-select
              :value="positions[player.membership_id] || 1"
              @update:value="$emit('updatePosition', player.membership_id, $event)"
              :options="positionOptions.map(p => ({ label: String(p), value: p }))"
              size="small"
              style="max-width: 80px"
            />
          </td>
        </tr>
      </tbody>
    </n-table>
    <template #action>
      <div style="display: flex; justify-content: space-between; align-items: center; padding: 16px;">
        <n-button quaternary @click="$emit('back')" :disabled="saving">
          <template #icon>
            <n-icon><ChevronBackIcon /></n-icon>
          </template>
          {{ $t('game.back') }}
        </n-button>
        <div style="display: flex; gap: 8px;">
          <n-button
            secondary
            type="primary"
            :loading="saving"
            @click="$emit('save')"
          >
            <template #icon>
              <n-icon><SaveIcon /></n-icon>
            </template>
            {{ $t('common.save') }}
          </n-button>
          <n-button
            type="success"
            :loading="saving"
            @click="$emit('finish')"
          >
            <template #icon>
              <n-icon><CheckIcon /></n-icon>
            </template>
            {{ $t('game.finishGame') }}
          </n-button>
        </div>
      </div>
    </template>
  </n-card>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { NCard, NTable, NTag, NInputNumber, NSelect, NButton, NIcon } from 'naive-ui';
import { ChevronBack as ChevronBackIcon, Save as SaveIcon, Checkmark as CheckIcon } from '@vicons/ionicons5';

export interface ScoringPlayer {
  membership_id: string;
  alias: string;
  is_moderator: boolean;
}

const props = defineProps<{
  players: ScoringPlayer[];
  scores: Record<string, number>;
  positions: Record<string, number>;
  saving: boolean;
}>();

defineEmits<{
  updateScore: [membershipId: string, score: number];
  updatePosition: [membershipId: string, position: number];
  back: [];
  save: [];
  finish: [];
}>();

const positionOptions = computed(() => {
  return props.players.map((_, index) => index + 1);
});
</script>

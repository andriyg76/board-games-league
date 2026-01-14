<template>
  <v-card flat>
    <v-card-title>{{ $t('game.enterScores') }}</v-card-title>
    <v-card-text>
      <v-table>
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
              {{ player.alias }}
              <v-chip v-if="player.is_moderator" size="x-small" color="orange" class="ml-1">
                {{ $t('roleTypes.moderator') }}
              </v-chip>
            </td>
            <td>
              <v-text-field
                :model-value="scores[player.membership_id] || 0"
                @update:model-value="$emit('updateScore', player.membership_id, Number($event))"
                type="number"
                density="compact"
                variant="outlined"
                hide-details
                style="max-width: 100px"
              />
            </td>
            <td>
              <v-select
                :model-value="positions[player.membership_id] || 1"
                @update:model-value="$emit('updatePosition', player.membership_id, $event)"
                :items="positionOptions"
                density="compact"
                variant="outlined"
                hide-details
                style="max-width: 80px"
              />
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card-text>
    <v-card-actions>
      <v-btn variant="text" @click="$emit('back')" :disabled="saving">
        <v-icon start>mdi-arrow-left</v-icon>
        {{ $t('game.back') }}
      </v-btn>
      <v-spacer />
      <v-btn
        variant="outlined"
        color="primary"
        :loading="saving"
        @click="$emit('save')"
        class="mr-2"
      >
        <v-icon start>mdi-content-save</v-icon>
        {{ $t('common.save') }}
      </v-btn>
      <v-btn
        color="success"
        :loading="saving"
        @click="$emit('finish')"
      >
        <v-icon start>mdi-check</v-icon>
        {{ $t('game.finishGame') }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { computed } from 'vue';

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

<template>
  <v-card flat>
    <v-card-title>{{ $t('game.configureRound') }}</v-card-title>
    <v-card-text>
      <v-text-field
        :model-value="gameName"
        @update:model-value="$emit('update:gameName', $event)"
        :label="$t('game.roundName')"
        variant="outlined"
        class="mb-4"
      />

      <v-select
        :model-value="bidRestriction"
        @update:model-value="$emit('update:bidRestriction', $event)"
        :items="bidRestrictions"
        :item-title="(item) => $t(item.title)"
        item-value="value"
        :label="$t('wizard.bidRestriction')"
        variant="outlined"
        class="mb-4"
      >
        <template #prepend>
          <v-icon>mdi-shield-alert</v-icon>
        </template>
      </v-select>

      <h4 class="mb-2">{{ $t('wizard.selectFirstDealer') }}</h4>
      <v-list density="compact" class="mb-4">
        <v-list-item
          v-for="(player, index) in players"
          :key="player.membership_id"
          :class="{ 'bg-primary-lighten-4': firstDealerIndex === index }"
          @click="$emit('update:firstDealerIndex', index)"
        >
          <template #prepend>
            <v-icon v-if="firstDealerIndex === index" color="primary">
              mdi-cards-playing
            </v-icon>
            <span v-else class="text-grey ml-2 mr-3">{{ index + 1 }}</span>
          </template>
          <v-list-item-title>
            {{ player.alias }}
            <v-chip v-if="player.isCurrentUser" size="x-small" color="primary" class="ml-1">
              {{ $t('game.you') }}
            </v-chip>
          </v-list-item-title>
          <template #append v-if="firstDealerIndex === index">
            <v-chip size="small" color="primary">
              {{ $t('wizard.firstDealer') }}
            </v-chip>
          </template>
        </v-list-item>
      </v-list>

      <!-- Wizard game summary -->
      <v-card variant="outlined" class="pa-4">
        <h4 class="text-subtitle-1 mb-2">{{ $t('wizard.gameSummary') }}</h4>
        <v-row dense>
          <v-col cols="6">
            <div class="text-caption text-grey">{{ $t('game.selectedPlayers') }}</div>
            <div class="text-body-1">{{ players.length }}</div>
          </v-col>
          <v-col cols="6">
            <div class="text-caption text-grey">{{ $t('wizard.rounds') }}</div>
            <div class="text-body-1">{{ maxRounds }}</div>
          </v-col>
          <v-col cols="6">
            <div class="text-caption text-grey">{{ $t('wizard.firstDealer') }}</div>
            <div class="text-body-1">{{ players[firstDealerIndex]?.alias || '-' }}</div>
          </v-col>
          <v-col cols="6">
            <div class="text-caption text-grey">{{ $t('wizard.bidRestriction') }}</div>
            <div class="text-body-1 text-truncate">{{ $t(bidRestrictions.find(r => r.value === bidRestriction)?.title || '') }}</div>
          </v-col>
        </v-row>
      </v-card>
    </v-card-text>
    <v-card-actions>
      <v-btn variant="text" @click="$emit('back')" :disabled="saving">
        <v-icon start>mdi-arrow-left</v-icon>
        {{ $t('game.back') }}
      </v-btn>
      <v-spacer />
      <v-btn
        color="success"
        :loading="saving"
        @click="$emit('start')"
      >
        <v-icon start>mdi-play</v-icon>
        {{ $t('game.startGame') }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { BidRestriction } from './types';

export interface WizardPlayer {
  membership_id: string;
  alias: string;
  isCurrentUser?: boolean;
}

const props = defineProps<{
  gameName: string;
  players: WizardPlayer[];
  bidRestriction: BidRestriction;
  firstDealerIndex: number;
  saving: boolean;
}>();

defineEmits<{
  'update:gameName': [name: string];
  'update:bidRestriction': [restriction: BidRestriction];
  'update:firstDealerIndex': [index: number];
  back: [];
  start: [];
}>();

const bidRestrictions = [
  { value: BidRestriction.NO_RESTRICTIONS, title: 'wizard.noRestrictions' },
  { value: BidRestriction.CANNOT_MATCH_CARDS, title: 'wizard.cannotMatchCards' },
  { value: BidRestriction.MUST_MATCH_CARDS, title: 'wizard.mustMatchCards' },
];

const maxRounds = computed(() => {
  if (props.players.length === 0) return 0;
  return Math.floor(60 / props.players.length);
});
</script>

<style scoped>
.bg-primary-lighten-4 {
  background-color: rgba(var(--v-theme-primary), 0.1);
}
</style>

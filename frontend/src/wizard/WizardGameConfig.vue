<template>
  <n-card>
    <template #header>
      {{ $t('game.configureRound') }}
    </template>
    <n-form-item :label="$t('game.roundName')">
      <n-input
        :value="gameName"
        @update:value="$emit('update:gameName', $event)"
      />
    </n-form-item>

    <n-form-item :label="$t('wizard.bidRestriction')">
      <n-select
        :value="bidRestriction"
        @update:value="$emit('update:bidRestriction', $event)"
        :options="bidRestrictions.map(r => ({ label: $t(r.title), value: r.value }))"
      />
    </n-form-item>

    <h4 style="margin-bottom: 8px; font-size: 1rem; font-weight: 500;">{{ $t('wizard.selectFirstDealer') }}</h4>
    <n-list style="margin-bottom: 16px;">
      <n-list-item
        v-for="(player, index) in players"
        :key="player.membership_id"
        :class="{ 'selected-dealer': firstDealerIndex === index }"
        clickable
        @click="$emit('update:firstDealerIndex', index)"
      >
        <template #prefix>
          <n-icon v-if="firstDealerIndex === index" color="#2080f0" :size="24">
            <CardsIcon />
          </n-icon>
          <span v-else style="color: #999; margin-right: 8px;">{{ index + 1 }}</span>
        </template>
        <div>
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="font-weight: 500;">{{ player.alias }}</span>
            <n-tag v-if="player.isCurrentUser" size="small" type="primary">
              {{ $t('game.you') }}
            </n-tag>
          </div>
        </div>
        <template #suffix v-if="firstDealerIndex === index">
          <n-tag size="small" type="primary">
            {{ $t('wizard.firstDealer') }}
          </n-tag>
        </template>
      </n-list-item>
    </n-list>

    <!-- Wizard game summary -->
    <n-card style="padding: 16px;">
      <h4 style="font-size: 1rem; font-weight: 500; margin-bottom: 8px;">{{ $t('wizard.gameSummary') }}</h4>
      <n-grid :cols="24" :x-gap="8">
        <n-gi :span="24" :responsive="{ m: 6 }">
          <div style="font-size: 0.75rem; opacity: 0.7; margin-bottom: 4px;">{{ $t('game.selectedPlayers') }}</div>
          <div style="font-size: 1rem;">{{ players.length }}</div>
        </n-gi>
        <n-gi :span="24" :responsive="{ m: 6 }">
          <div style="font-size: 0.75rem; opacity: 0.7; margin-bottom: 4px;">{{ $t('wizard.rounds') }}</div>
          <div style="font-size: 1rem;">{{ maxRounds }}</div>
        </n-gi>
        <n-gi :span="24" :responsive="{ m: 6 }">
          <div style="font-size: 0.75rem; opacity: 0.7; margin-bottom: 4px;">{{ $t('wizard.firstDealer') }}</div>
          <div style="font-size: 1rem;">{{ players[firstDealerIndex]?.alias || '-' }}</div>
        </n-gi>
        <n-gi :span="24" :responsive="{ m: 6 }">
          <div style="font-size: 0.75rem; opacity: 0.7; margin-bottom: 4px;">{{ $t('wizard.bidRestriction') }}</div>
          <div style="font-size: 1rem; overflow: hidden; text-overflow: ellipsis;">{{ $t(bidRestrictions.find(r => r.value === bidRestriction)?.title || '') }}</div>
        </n-gi>
      </n-grid>
    </n-card>
    <template #action>
      <div style="display: flex; justify-content: space-between; align-items: center; padding: 16px;">
        <n-button quaternary @click="$emit('back')" :disabled="saving">
          <template #icon>
            <n-icon><ChevronBackIcon /></n-icon>
          </template>
          {{ $t('game.back') }}
        </n-button>
        <n-button
          type="success"
          :loading="saving"
          @click="$emit('start')"
        >
          <template #icon>
            <n-icon><PlayIcon /></n-icon>
          </template>
          {{ $t('game.startGame') }}
        </n-button>
      </div>
    </template>
  </n-card>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { NCard, NFormItem, NInput, NSelect, NIcon, NList, NListItem, NTag, NGrid, NGi, NButton } from 'naive-ui';
import { Shield as ShieldIcon, ChevronBack as ChevronBackIcon, Play as PlayIcon, Card as CardsIcon } from '@vicons/ionicons5';
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
.selected-dealer {
  background-color: rgba(32, 128, 240, 0.1);
}
</style>

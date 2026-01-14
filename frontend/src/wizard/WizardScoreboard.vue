<template>
  <n-modal v-model:show="isOpen" preset="card" title="Wizard Scoreboard" style="max-width: 1200px;" :mask-closable="false">
    <n-spin v-if="loading" size="large" style="display: flex; justify-content: center; padding: 64px;">
      <template #description>
        Loading scoreboard...
      </template>
    </n-spin>

    <div v-else-if="scoreboard" class="scoreboard-container">
      <!-- Header Info -->
      <div style="padding: 16px; background-color: #fafafa;">
        <n-grid :cols="24" :x-gap="8">
          <n-gi :span="24" :responsive="{ m: 8 }">
            <div style="font-size: 0.75rem; opacity: 0.7; margin-bottom: 4px;">Game Code</div>
            <div style="font-size: 1rem; font-weight: 500;">{{ scoreboard.game_code }}</div>
          </n-gi>
          <n-gi :span="24" :responsive="{ m: 8 }">
            <div style="font-size: 0.75rem; opacity: 0.7; margin-bottom: 4px;">Current Round</div>
            <div style="font-size: 1rem;">{{ scoreboard.current_round }} / {{ scoreboard.max_rounds }}</div>
          </n-gi>
          <n-gi :span="24" :responsive="{ m: 8 }">
            <div style="font-size: 0.75rem; opacity: 0.7; margin-bottom: 4px;">Players</div>
            <div style="font-size: 1rem;">{{ scoreboard.players.length }}</div>
          </n-gi>
        </n-grid>
      </div>

      <n-divider />

      <!-- Scrollable Table -->
      <div class="table-scroll">
        <table class="wizard-table">
          <thead>
            <tr>
              <th class="sticky-col player-col">Player</th>
              <th class="sticky-col score-col">Total</th>
              <th
                v-for="round in scoreboard.rounds"
                :key="round.round_number"
                class="round-col"
                :class="{ 'current-round': round.round_number === scoreboard.current_round }"
              >
                <div class="round-header">
                  <div class="round-number">R{{ round.round_number }}</div>
                  <div class="round-info">{{ round.cards_count }} cards</div>
                </div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(player, playerIndex) in scoreboard.players"
              :key="player.membership_id"
            >
              <td class="sticky-col player-col">
                <div class="player-name">{{ player.player_name }}</div>
              </td>
              <td class="sticky-col score-col">
                <div class="total-score">{{ player.total_score }}</div>
              </td>
              <td
                v-for="round in scoreboard.rounds"
                :key="`${player.membership_id}-${round.round_number}`"
                class="round-col"
                :class="{
                  'current-round': round.round_number === scoreboard.current_round,
                  'dealer-round': round.dealer_index === playerIndex
                }"
              >
                <div
                  v-if="round.player_results[playerIndex]"
                  class="round-cell"
                  :class="getCellClass(round.player_results[playerIndex])"
                >
                  <div class="score-main">
                    {{ getRoundScore(round.player_results[playerIndex]) }}
                  </div>
                  <div class="tricks-info">
                    {{ getTricksInfo(round.player_results[playerIndex]) }}
                  </div>
                </div>
                <div v-else class="round-cell pending">
                  -
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <n-divider />

      <!-- Legend -->
      <div style="padding: 16px;">
        <div style="font-size: 0.875rem; font-weight: 500; margin-bottom: 8px;">Legend:</div>
        <div style="display: flex; flex-wrap: wrap; gap: 8px;">
          <n-tag size="small" class="cell-match">Match (bid = actual)</n-tag>
          <n-tag size="small" class="cell-miss-small">Miss by 1</n-tag>
          <n-tag size="small" class="cell-miss-large">Miss by 2+</n-tag>
          <n-tag size="small" type="default" style="border: 2px solid #1976d2;">Dealer (border)</n-tag>
          <n-tag size="small" class="current-round-chip">Current Round</n-tag>
        </div>
      </div>
    </div>

    <n-alert v-else type="info" style="margin: 16px;">
      No scoreboard data available
    </n-alert>

    <template #action>
      <div style="display: flex; justify-content: flex-end;">
        <n-button type="primary" @click="close">
          Close
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NModal, NSpin, NGrid, NGi, NDivider, NTag, NAlert, NButton } from 'naive-ui'
import { useWizardStore } from '@/store/wizard'
import type { WizardPlayerResult } from './types'

interface Props {
  modelValue: boolean
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const wizardStore = useWizardStore()
const loading = ref(false)

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const scoreboard = computed(() => wizardStore.scoreboard)

watch(() => props.modelValue, async (newValue) => {
  if (newValue) {
    await loadScoreboard()
  }
})

async function loadScoreboard() {
  loading.value = true
  try {
    await wizardStore.loadScoreboard()
  } catch (error) {
    console.error('Error loading scoreboard:', error)
  } finally {
    loading.value = false
  }
}

function getRoundScore(result: WizardPlayerResult): string {
  if (result.bid < 0 || result.actual < 0) {
    return '-'
  }

  if (result.score === 0) {
    return '0'
  }

  return result.score > 0 ? `+${result.score}` : `${result.score}`
}

function getTricksInfo(result: WizardPlayerResult): string {
  if (result.bid < 0 || result.actual < 0) {
    return ''
  }
  return `${result.actual}/${result.bid}`
}

function getCellClass(result: WizardPlayerResult): string[] {
  const classes: string[] = []

  if (result.bid < 0 || result.actual < 0) {
    classes.push('pending')
    return classes
  }

  // Color based on result
  if (result.bid === result.actual) {
    classes.push('cell-match')
  } else {
    const diff = Math.abs(result.bid - result.actual)
    if (diff === 1) {
      classes.push('cell-miss-small')
    } else {
      classes.push('cell-miss-large')
    }
  }

  return classes
}

function close() {
  isOpen.value = false
}
</script>

<style scoped>
.scoreboard-container {
  position: relative;
}

.table-scroll {
  overflow-x: auto;
  max-height: 600px;
  overflow-y: auto;
}

.wizard-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 14px;
}

.wizard-table th,
.wizard-table td {
  padding: 12px 8px;
  text-align: center;
  border-bottom: 1px solid #e0e0e0;
  border-right: 1px solid #e0e0e0;
}

.wizard-table thead th {
  background: #f5f5f5;
  font-weight: 600;
  position: sticky;
  top: 0;
  z-index: 10;
}

.sticky-col {
  position: sticky;
  left: 0;
  background: white;
  z-index: 5;
}

.wizard-table thead .sticky-col {
  z-index: 15;
  background: #f5f5f5;
}

.player-col {
  min-width: 120px;
  text-align: left;
  font-weight: 500;
}

.score-col {
  min-width: 80px;
  left: 120px !important;
  border-right: 2px solid #bdbdbd;
  font-weight: 600;
  font-size: 16px;
}

.round-col {
  min-width: 80px;
  max-width: 80px;
}

.round-header {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.round-number {
  font-weight: 600;
  font-size: 14px;
}

.round-info {
  font-size: 11px;
  color: #666;
  font-weight: normal;
}

.current-round {
  background: #e3f2fd !important;
}

.current-round-chip {
  background: #e3f2fd;
  color: #1976d2;
}

.dealer-round {
  border: 2px solid #1976d2;
}

.round-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 4px;
  border-radius: 4px;
}

.score-main {
  font-size: 16px;
  font-weight: 600;
}

.tricks-info {
  font-size: 11px;
  color: #666;
}

.player-name {
  font-weight: 500;
}

.total-score {
  font-size: 18px;
  font-weight: 700;
  color: #1976d2;
}

/* Cell colors */
.cell-match {
  background-color: #c8e6c9;
  color: #2e7d32;
}

.cell-miss-small {
  background-color: #fff3e0;
  color: #e65100;
}

.cell-miss-large {
  background-color: #ffcdd2;
  color: #c62828;
}

.pending {
  color: #9e9e9e;
  font-style: italic;
}

.gap-3 {
  gap: 0.75rem;
}

.bg-grey-lighten-5 {
  background-color: #fafafa;
}
</style>

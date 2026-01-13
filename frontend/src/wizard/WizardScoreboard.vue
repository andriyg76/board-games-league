<template>
  <v-dialog v-model="isOpen" max-width="1200" scrollable>
    <v-card>
      <v-card-title class="d-flex align-center">
        <v-icon start>mdi-table</v-icon>
        Wizard Scoreboard
        <v-spacer />
        <v-btn icon="mdi-close" variant="text" @click="close" />
      </v-card-title>

      <v-divider />

      <v-card-text class="pa-0">
        <div v-if="loading" class="text-center pa-8">
          <v-progress-circular indeterminate color="primary" size="64" />
          <p class="mt-4">Loading scoreboard...</p>
        </div>

        <div v-else-if="scoreboard" class="scoreboard-container">
          <!-- Header Info -->
          <div class="pa-4 bg-grey-lighten-5">
            <v-row dense>
              <v-col cols="4">
                <div class="text-caption text-grey">Game Code</div>
                <div class="text-body-1 font-weight-bold">{{ scoreboard.game_code }}</div>
              </v-col>
              <v-col cols="4">
                <div class="text-caption text-grey">Current Round</div>
                <div class="text-body-1">{{ scoreboard.current_round }} / {{ scoreboard.max_rounds }}</div>
              </v-col>
              <v-col cols="4">
                <div class="text-caption text-grey">Players</div>
                <div class="text-body-1">{{ scoreboard.players.length }}</div>
              </v-col>
            </v-row>
          </div>

          <v-divider />

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

          <v-divider />

          <!-- Legend -->
          <div class="pa-4">
            <div class="text-subtitle-2 mb-2">Legend:</div>
            <div class="d-flex flex-wrap gap-3">
              <v-chip size="small" class="cell-match">Match (bid = actual)</v-chip>
              <v-chip size="small" class="cell-miss-small">Miss by 1</v-chip>
              <v-chip size="small" class="cell-miss-large">Miss by 2+</v-chip>
              <v-chip size="small" variant="outlined">Dealer (border)</v-chip>
              <v-chip size="small" class="current-round-chip">Current Round</v-chip>
            </div>
          </div>
        </div>

        <v-alert v-else type="info" class="ma-4">
          No scoreboard data available
        </v-alert>
      </v-card-text>

      <v-divider />

      <v-card-actions>
        <v-spacer />
        <v-btn color="primary" variant="text" @click="close">
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
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

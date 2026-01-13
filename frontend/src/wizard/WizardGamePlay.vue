<template>
  <v-container>
    <v-alert v-if="wizardStore.error" type="error" class="mb-4" closable>
      {{ wizardStore.error }}
    </v-alert>

    <div v-if="wizardStore.loading" class="text-center py-8">
      <v-progress-circular indeterminate color="primary" size="64" />
      <p class="mt-4">Loading game...</p>
    </div>

    <div v-else-if="game">
      <!-- Game Header -->
      <v-card class="mb-4">
        <v-card-title class="d-flex align-center">
          <v-icon start>mdi-wizard-hat</v-icon>
          Wizard Game
          <v-spacer />
          <v-chip color="primary" class="ml-2">
            Round {{ game.current_round }} / {{ game.max_rounds }}
          </v-chip>
        </v-card-title>
        <v-card-subtitle>
          Game Code: {{ game.code }}
        </v-card-subtitle>
      </v-card>

      <!-- Current Round Info -->
      <v-card class="mb-4" v-if="currentRound">
        <v-card-text>
          <div class="d-flex justify-space-between align-center mb-2">
            <div>
              <v-icon>mdi-cards</v-icon>
              <span class="ml-2 text-h6">{{ currentRound.cards_count }} cards</span>
            </div>
            <div>
              <v-icon>mdi-account-star</v-icon>
              <span class="ml-2">
                Dealer: {{ game.players[currentRound.dealer_index]?.player_name }}
              </span>
            </div>
            <div>
              <v-chip :color="roundStatusColor" size="small">
                {{ currentRound.status }}
              </v-chip>
            </div>
          </div>

          <v-divider class="my-3" />

          <!-- Action Buttons -->
          <div class="d-flex gap-2 flex-wrap">
            <v-btn
              v-if="currentRound.status === 'BIDDING'"
              color="primary"
              @click="showBidDialog = true"
            >
              <v-icon start>mdi-hand-coin</v-icon>
              Enter Bids
            </v-btn>

            <v-btn
              v-if="currentRound.status === 'PLAYING' && !wizardStore.areAllResultsSubmitted"
              color="success"
              @click="showResultDialog = true"
            >
              <v-icon start>mdi-trophy</v-icon>
              Enter Results
            </v-btn>

            <v-btn
              v-if="wizardStore.areAllBidsSubmitted && wizardStore.areAllResultsSubmitted"
              color="success"
              variant="elevated"
              @click="completeRound"
              :loading="completing"
            >
              <v-icon start>mdi-check-circle</v-icon>
              Complete Round
            </v-btn>

            <v-btn
              v-if="currentRound.status === 'COMPLETED' && game.current_round < game.max_rounds"
              color="primary"
              @click="moveToNextRound"
            >
              <v-icon start>mdi-arrow-right</v-icon>
              Next Round
            </v-btn>

            <v-btn
              v-if="currentRound.status === 'COMPLETED' && game.current_round === game.max_rounds"
              color="success"
              variant="elevated"
              @click="finalizeGame"
              :loading="finalizing"
            >
              <v-icon start>mdi-flag-checkered</v-icon>
              Finalize Game
            </v-btn>

            <v-spacer />

            <v-btn
              color="grey"
              variant="text"
              @click="showScoreboard"
            >
              <v-icon start>mdi-table</v-icon>
              Scoreboard
            </v-btn>
          </div>
        </v-card-text>
      </v-card>

      <!-- Players List -->
      <v-card>
        <v-card-title>Players</v-card-title>
        <v-list>
          <v-list-item
            v-for="(player, index) in game.players"
            :key="player.membership_id"
          >
            <template v-slot:prepend>
              <v-avatar :color="index === currentRound?.dealer_index ? 'primary' : 'grey'">
                <v-icon>
                  {{ index === currentRound?.dealer_index ? 'mdi-account-star' : 'mdi-account' }}
                </v-icon>
              </v-avatar>
            </template>

            <v-list-item-title>
              {{ player.player_name }}
            </v-list-item-title>

            <v-list-item-subtitle v-if="currentRound">
              <span v-if="currentPlayerResult(index).bid >= 0">
                Bid: {{ currentPlayerResult(index).bid }}
              </span>
              <span v-if="currentPlayerResult(index).actual >= 0" class="ml-2">
                Actual: {{ currentPlayerResult(index).actual }}
              </span>
              <span v-if="currentRound.status === 'COMPLETED'" class="ml-2">
                Score: {{ currentPlayerResult(index).score > 0 ? '+' : '' }}{{ currentPlayerResult(index).score }}
              </span>
            </v-list-item-subtitle>

            <template v-slot:append>
              <div class="text-right">
                <div class="text-h6">{{ player.total_score }}</div>
                <div class="text-caption text-grey">Total</div>
              </div>
            </template>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- Bid Dialog -->
      <WizardBidDialog
        v-model="showBidDialog"
        :roundNumber="game.current_round"
        :cardsCount="currentRound?.cards_count || 1"
        :players="game.players"
        :dealerIndex="currentRound?.dealer_index || 0"
        :bidRestriction="game.config.bid_restriction"
        :existingBids="currentRoundBids"
        @submit="submitBids"
      />

      <!-- Result Dialog -->
      <WizardResultDialog
        v-model="showResultDialog"
        :roundNumber="game.current_round"
        :cardsCount="currentRound?.cards_count || 1"
        :players="game.players"
        :playerBids="currentRoundBids"
        :existingResults="currentRoundResults"
        @submit="submitResults"
      />
    </div>

    <v-card v-else class="text-center pa-8">
      <v-icon size="64" color="grey">mdi-wizard-hat</v-icon>
      <h2 class="mt-4">No game loaded</h2>
      <p class="text-grey mt-2">Game code not found or invalid</p>
      <v-btn color="primary" class="mt-4" to="/">
        Go Home
      </v-btn>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useWizardStore } from '@/store/wizard'
import WizardBidDialog from './WizardBidDialog.vue'
import WizardResultDialog from './WizardResultDialog.vue'

const route = useRoute()
const router = useRouter()
const wizardStore = useWizardStore()

const showBidDialog = ref(false)
const showResultDialog = ref(false)
const completing = ref(false)
const finalizing = ref(false)

const game = computed(() => wizardStore.currentGame)
const currentRound = computed(() => wizardStore.currentRoundData)

const roundStatusColor = computed(() => {
  switch (currentRound.value?.status) {
    case 'BIDDING': return 'warning'
    case 'PLAYING': return 'info'
    case 'COMPLETED': return 'success'
    default: return 'grey'
  }
})

const currentRoundBids = computed(() => {
  if (!currentRound.value) return []
  return currentRound.value.player_results.map(pr => pr.bid)
})

const currentRoundResults = computed(() => {
  if (!currentRound.value) return []
  return currentRound.value.player_results.map(pr => pr.actual)
})

function currentPlayerResult(index: number) {
  if (!currentRound.value) {
    return { bid: -1, actual: -1, score: 0 }
  }
  return currentRound.value.player_results[index] || { bid: -1, actual: -1, score: 0 }
}

async function submitBids(bids: number[]) {
  try {
    await wizardStore.submitBids(bids)
  } catch (error) {
    console.error('Error submitting bids:', error)
  }
}

async function submitResults(results: number[]) {
  try {
    await wizardStore.submitResults(results)
  } catch (error) {
    console.error('Error submitting results:', error)
  }
}

async function completeRound() {
  completing.value = true
  try {
    await wizardStore.completeRound()
  } catch (error) {
    console.error('Error completing round:', error)
  } finally {
    completing.value = false
  }
}

async function moveToNextRound() {
  try {
    await wizardStore.nextRound()
  } catch (error) {
    console.error('Error moving to next round:', error)
  }
}

async function finalizeGame() {
  finalizing.value = true
  try {
    await wizardStore.finalizeGame()
    // Redirect to game rounds list or league page
    router.push('/games')
  } catch (error) {
    console.error('Error finalizing game:', error)
  } finally {
    finalizing.value = false
  }
}

function showScoreboard() {
  // TODO: Implement scoreboard view
  alert('Scoreboard view not implemented yet')
}

onMounted(async () => {
  const code = route.params.code as string
  if (code) {
    try {
      await wizardStore.loadGame(code)
    } catch (error) {
      console.error('Error loading game:', error)
    }
  }
})
</script>

<style scoped>
.gap-2 {
  gap: 0.5rem;
}
</style>

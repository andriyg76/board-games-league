<template>
  <n-grid :cols="24" :x-gap="16">
    <n-gi :span="24">
      <n-alert v-if="wizardStore.error" type="error" style="margin-bottom: 16px;" closable @close="wizardStore.error = null">
        {{ wizardStore.error }}
      </n-alert>

      <n-spin v-if="wizardStore.loading" size="large" style="display: flex; justify-content: center; padding: 64px;">
        <template #description>
          Loading game...
        </template>
      </n-spin>

      <div v-else-if="game">
        <!-- Game Header -->
        <n-card style="margin-bottom: 16px;">
          <template #header>
            <div style="display: flex; align-items: center; justify-content: space-between;">
              <div style="display: flex; align-items: center; gap: 8px;">
                <n-icon :size="24"><WizardIcon /></n-icon>
                <span style="font-weight: 500;">Wizard Game</span>
              </div>
              <div style="display: flex; align-items: center; gap: 8px;">
                <n-tag v-if="wizardStore.isConnected" type="success" size="small">
                  <template #icon>
                    <n-icon><LiveIcon /></n-icon>
                  </template>
                  Live ({{ wizardStore.subscriberCount }})
                </n-tag>
                <n-tag type="primary">
                  Round {{ game.current_round }} / {{ game.max_rounds }}
                </n-tag>
              </div>
            </div>
          </template>
          <div style="font-size: 0.875rem; opacity: 0.7;">
            Game Code: {{ game.code }}
          </div>
        </n-card>

        <!-- Current Round Info -->
        <n-card v-if="currentRound" style="margin-bottom: 16px;">
          <n-grid :cols="24" :x-gap="8" style="margin-bottom: 16px;">
            <n-gi :span="24" :responsive="{ m: 8 }">
              <div style="display: flex; align-items: center; gap: 8px;">
                <n-icon><CardsIcon /></n-icon>
                <span style="font-size: 1.125rem; font-weight: 500;">{{ currentRound.cards_count }} cards</span>
              </div>
            </n-gi>
            <n-gi :span="24" :responsive="{ m: 8 }">
              <div style="display: flex; align-items: center; gap: 8px;">
                <n-icon><StarIcon /></n-icon>
                <span>Dealer: {{ game.players[currentRound.dealer_index]?.player_name }}</span>
              </div>
            </n-gi>
            <n-gi :span="24" :responsive="{ m: 8 }">
              <n-tag :type="getRoundStatusTagType(currentRound.status)" size="small">
                {{ currentRound.status }}
              </n-tag>
            </n-gi>
          </n-grid>

          <n-divider style="margin: 16px 0;" />

          <!-- Action Buttons -->
          <div style="display: flex; flex-wrap: wrap; gap: 8px;">
            <n-button
              v-if="currentRound.status === 'BIDDING'"
              type="primary"
              @click="showBidDialog = true"
            >
              <template #icon>
                <n-icon><HandCoinIcon /></n-icon>
              </template>
              Enter Bids
            </n-button>

            <n-button
              v-if="currentRound.status === 'PLAYING' && !wizardStore.areAllResultsSubmitted"
              type="success"
              @click="showResultDialog = true"
            >
              <template #icon>
                <n-icon><TrophyIcon /></n-icon>
              </template>
              Enter Results
            </n-button>

            <n-button
              v-if="wizardStore.areAllBidsSubmitted && wizardStore.areAllResultsSubmitted"
              type="success"
              @click="completeRound"
              :loading="completing"
            >
              <template #icon>
                <n-icon><CheckCircleIcon /></n-icon>
              </template>
              Complete Round
            </n-button>

            <n-button
              v-if="currentRound.status === 'COMPLETED' && game.current_round < game.max_rounds"
              type="primary"
              @click="moveToNextRound"
            >
              <template #icon>
                <n-icon><ArrowForwardIcon /></n-icon>
              </template>
              Next Round
            </n-button>

            <n-button
              v-if="currentRound.status === 'COMPLETED' && game.current_round === game.max_rounds"
              type="success"
              @click="finalizeGame"
              :loading="finalizing"
            >
              <template #icon>
                <n-icon><FlagIcon /></n-icon>
              </template>
              Finalize Game
            </n-button>

            <div style="flex: 1;"></div>

            <n-button quaternary @click="showScoreboard">
              <template #icon>
                <n-icon><TableIcon /></n-icon>
              </template>
              Scoreboard
            </n-button>
          </div>
        </n-card>

        <!-- Players List -->
        <n-card>
          <template #header>
            Players
          </template>
          <n-list>
            <n-list-item
              v-for="(player, index) in game.players"
              :key="player.membership_id"
            >
              <template #prefix>
                <n-avatar :style="{ backgroundColor: index === currentRound?.dealer_index ? '#2080f0' : '#999' }">
                  <n-icon :color="index === currentRound?.dealer_index ? '#fff' : '#fff'">
                    <component :is="index === currentRound?.dealer_index ? StarIcon : PersonIcon" />
                  </n-icon>
                </n-avatar>
              </template>
              <div style="flex: 1;">
                <div style="font-weight: 500;">{{ player.player_name }}</div>
                <div v-if="currentRound" style="font-size: 0.875rem; opacity: 0.7;">
                  <span v-if="currentPlayerResult(index).bid >= 0">
                    Bid: {{ currentPlayerResult(index).bid }}
                  </span>
                  <span v-if="currentPlayerResult(index).actual >= 0" style="margin-left: 8px;">
                    Actual: {{ currentPlayerResult(index).actual }}
                  </span>
                  <span v-if="currentRound.status === 'COMPLETED'" style="margin-left: 8px;">
                    Score: {{ currentPlayerResult(index).score > 0 ? '+' : '' }}{{ currentPlayerResult(index).score }}
                  </span>
                </div>
              </div>
              <template #suffix>
                <div style="text-align: right;">
                  <div style="font-size: 1.25rem; font-weight: 500;">{{ player.total_score }}</div>
                  <div style="font-size: 0.75rem; opacity: 0.7;">Total</div>
                </div>
              </template>
            </n-list-item>
          </n-list>
        </n-card>

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

        <!-- Scoreboard Dialog -->
        <WizardScoreboard v-model="showScoreboardDialog" />
      </div>

      <n-card v-else style="text-align: center; padding: 64px;">
        <n-icon :size="64" color="#999"><WizardIcon /></n-icon>
        <h2 style="margin-top: 16px; font-size: 1.5rem;">No game loaded</h2>
        <p style="color: #999; margin-top: 8px;">Game code not found or invalid</p>
        <n-button type="primary" style="margin-top: 16px;" @click="$router.push('/')">
          Go Home
        </n-button>
      </n-card>
    </n-gi>
  </n-grid>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { NGrid, NGi, NAlert, NSpin, NCard, NIcon, NTag, NDivider, NButton, NList, NListItem, NAvatar, NBadge } from 'naive-ui'
import { Sparkles as WizardIcon, Card as CardsIcon, Star as StarIcon, Gift as HandCoinIcon, Trophy as TrophyIcon, CheckmarkCircle as CheckCircleIcon, ArrowForward as ArrowForwardIcon, Flag as FlagIcon, Grid as TableIcon, Person as PersonIcon, Radio as LiveIcon } from '@vicons/ionicons5'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useWizardStore } from '@/store/wizard'
import { useLeagueStore } from '@/store/league'
import { useErrorHandler } from '@/composables/useErrorHandler'
import WizardBidDialog from './WizardBidDialog.vue'
import WizardResultDialog from './WizardResultDialog.vue'
import WizardScoreboard from './WizardScoreboard.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const wizardStore = useWizardStore()
const leagueStore = useLeagueStore()
const { handleError, showSuccess } = useErrorHandler()

const showBidDialog = ref(false)
const showResultDialog = ref(false)
const showScoreboardDialog = ref(false)
const completing = ref(false)
const finalizing = ref(false)

const game = computed(() => wizardStore.currentGame)
const currentRound = computed(() => wizardStore.currentRoundData)

const getRoundStatusTagType = (status?: string): 'default' | 'info' | 'success' | 'warning' | 'error' => {
  switch (status) {
    case 'BIDDING': return 'warning'
    case 'PLAYING': return 'info'
    case 'COMPLETED': return 'success'
    default: return 'default'
  }
}

const currentRoundBids = computed(() => {
  if (!currentRound.value || !currentRound.value.player_results) return []
  return currentRound.value.player_results.map(pr => pr.bid)
})

const currentRoundResults = computed(() => {
  if (!currentRound.value || !currentRound.value.player_results) return []
  return currentRound.value.player_results.map(pr => pr.actual)
})

function currentPlayerResult(index: number) {
  if (!currentRound.value || !currentRound.value.player_results) {
    return { bid: -1, actual: -1, score: 0 }
  }
  return currentRound.value.player_results[index] || { bid: -1, actual: -1, score: 0 }
}

async function submitBids(bids: number[]) {
  try {
    await wizardStore.submitBids(bids)
  } catch (error) {
    handleError(error, t('errors.savingData'))
  }
}

async function submitResults(results: number[]) {
  try {
    await wizardStore.submitResults(results)
  } catch (error) {
    handleError(error, t('errors.savingData'))
  }
}

async function completeRound() {
  completing.value = true
  try {
    await wizardStore.completeRound()
  } catch (error) {
    handleError(error, t('errors.savingData'))
  } finally {
    completing.value = false
  }
}

async function moveToNextRound() {
  try {
    await wizardStore.nextRound()
  } catch (error) {
    handleError(error, t('errors.savingData'))
  }
}

async function finalizeGame() {
  finalizing.value = true
  try {
    await wizardStore.finalizeGame()
    // Redirect to game rounds list or league page
    router.push('/ui/game-rounds')
  } catch (error) {
    handleError(error, t('errors.savingData'))
  } finally {
    finalizing.value = false
  }
}

function showScoreboard() {
  showScoreboardDialog.value = true
}

onMounted(async () => {
  const code = route.params.code as string
  const leagueCode = (route.query.league as string) || leagueStore.currentLeagueCode
  if (code && leagueCode) {
    try {
      await wizardStore.loadGame(leagueCode, code)
      // Subscribe to real-time updates after loading the game
      wizardStore.subscribeToEvents()
    } catch (error) {
      handleError(error, t('errors.loadingData'))
    }
  } else if (code) {
    handleError(new Error('League code is required'), t('errors.loadingData'))
  }
})

onUnmounted(() => {
  // Unsubscribe from SSE when leaving the page
  wizardStore.unsubscribeFromEvents()
})
</script>


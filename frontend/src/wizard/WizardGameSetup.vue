<template>
  <v-container>
    <v-card>
      <v-card-title class="text-h5">
        <v-icon start>mdi-wizard-hat</v-icon>
        New Wizard Game
      </v-card-title>

      <v-card-text>
        <v-alert v-if="error" type="error" class="mb-4" closable @click:close="error = null">
          {{ error }}
        </v-alert>

        <v-form @submit.prevent="createGame">
          <!-- League Selection -->
          <v-select
            v-model="selectedLeagueCode"
            :items="leagueStore.leagues"
            item-title="name"
            item-value="code"
            label="League"
            :loading="leagueStore.loading"
            @update:model-value="onLeagueChange"
            required
          >
            <template v-slot:prepend>
              <v-icon>mdi-trophy</v-icon>
            </template>
          </v-select>

          <!-- Game Name -->
          <v-text-field
            v-model="gameName"
            label="Game Name"
            :rules="[v => !!v || 'Game name is required']"
            class="mt-4"
            required
          >
            <template v-slot:prepend>
              <v-icon>mdi-text</v-icon>
            </template>
          </v-text-field>

          <!-- Bid Restriction -->
          <v-select
            v-model="bidRestriction"
            :items="bidRestrictions"
            label="Bid Restriction"
            class="mt-4"
            required
          >
            <template v-slot:prepend>
              <v-icon>mdi-shield-alert</v-icon>
            </template>
          </v-select>

          <v-divider class="my-6" />

          <!-- Player Selection -->
          <div class="mb-4">
            <h3 class="text-h6 mb-2">
              Players ({{ selectedPlayers.length }}/{{ maxPlayers }})
            </h3>
            <v-chip size="small" color="info" class="mb-4">
              Select 3-6 players. {{ 60 / selectedPlayers.length || '∞' }} rounds (60 cards / players)
            </v-chip>
          </div>

          <v-alert
            v-if="selectedPlayers.length < 3"
            type="warning"
            class="mb-4"
          >
            Please select at least 3 players
          </v-alert>

          <v-alert
            v-if="selectedPlayers.length > 6"
            type="error"
            class="mb-4"
          >
            Maximum 6 players allowed
          </v-alert>

          <v-list v-if="members.length > 0">
            <v-list-item
              v-for="member in members"
              :key="member.code"
              @click="togglePlayer(member)"
              :class="{ 'bg-blue-lighten-5': isPlayerSelected(member) }"
            >
              <template v-slot:prepend>
                <v-checkbox
                  :model-value="isPlayerSelected(member)"
                  @click.stop="togglePlayer(member)"
                  hide-details
                  color="primary"
                />
              </template>

              <v-list-item-title>
                {{ member.alias }}
              </v-list-item-title>

              <template v-slot:append v-if="isPlayerSelected(member)">
                <v-chip
                  size="small"
                  :color="getPlayerIndex(member) === firstDealerIndex ? 'primary' : 'grey'"
                  @click.stop="setFirstDealer(member)"
                >
                  {{ getPlayerIndex(member) === firstDealerIndex ? 'First Dealer' : 'Set as Dealer' }}
                </v-chip>
              </template>
            </v-list-item>
          </v-list>

          <v-alert v-else type="info" class="mt-4">
            Please select a league to see available players
          </v-alert>

          <v-divider class="my-6" />

          <!-- Summary -->
          <v-card variant="outlined" class="pa-4 mb-4" v-if="selectedPlayers.length >= 3">
            <h4 class="text-subtitle-1 mb-2">Game Summary</h4>
            <v-row dense>
              <v-col cols="6">
                <div class="text-caption text-grey">Players</div>
                <div class="text-body-1">{{ selectedPlayers.length }}</div>
              </v-col>
              <v-col cols="6">
                <div class="text-caption text-grey">Rounds</div>
                <div class="text-body-1">{{ maxRounds }}</div>
              </v-col>
              <v-col cols="6">
                <div class="text-caption text-grey">First Dealer</div>
                <div class="text-body-1">{{ firstDealerName }}</div>
              </v-col>
              <v-col cols="6">
                <div class="text-caption text-grey">Bid Restriction</div>
                <div class="text-body-1 text-truncate">{{ bidRestrictionName }}</div>
              </v-col>
            </v-row>
          </v-card>

          <!-- Action Buttons -->
          <div class="d-flex gap-2 mt-6">
            <v-btn
              color="grey"
              variant="text"
              @click="cancel"
              :disabled="loading"
            >
              Cancel
            </v-btn>
            <v-spacer />
            <v-btn
              type="submit"
              color="primary"
              variant="elevated"
              :loading="loading"
              :disabled="!isValid"
            >
              <v-icon start>mdi-play</v-icon>
              Create Game
            </v-btn>
          </div>
        </v-form>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useLeagueStore } from '@/store/league'
import { useWizardStore } from '@/store/wizard'
import { BidRestriction } from './types'
import type { LeagueMember } from '@/api/LeagueApi'

const router = useRouter()
const leagueStore = useLeagueStore()
const wizardStore = useWizardStore()

const selectedLeagueCode = ref<string>('')
const gameName = ref<string>('')
const bidRestriction = ref<BidRestriction>(BidRestriction.NO_RESTRICTIONS)
const selectedPlayers = ref<LeagueMember[]>([])
const firstDealerIndex = ref<number>(0)
const loading = ref(false)
const error = ref<string | null>(null)

const maxPlayers = 6

const bidRestrictions = [
  {
    value: BidRestriction.NO_RESTRICTIONS,
    title: 'No Restrictions'
  },
  {
    value: BidRestriction.CANNOT_MATCH_CARDS,
    title: 'Total Bids Cannot Match Cards'
  },
  {
    value: BidRestriction.MUST_MATCH_CARDS,
    title: 'Total Bids Must Match Cards'
  }
]

const members = computed(() => {
  return leagueStore.currentLeagueMembers || []
})

const maxRounds = computed(() => {
  if (selectedPlayers.value.length === 0) return 0
  return Math.floor(60 / selectedPlayers.value.length)
})

const firstDealerName = computed(() => {
  if (selectedPlayers.value.length === 0 || firstDealerIndex.value >= selectedPlayers.value.length) {
    return 'Not set'
  }
  return selectedPlayers.value[firstDealerIndex.value]?.alias || 'Not set'
})

const bidRestrictionName = computed(() => {
  const restriction = bidRestrictions.find(r => r.value === bidRestriction.value)
  return restriction?.title || 'Unknown'
})

const isValid = computed(() => {
  return selectedPlayers.value.length >= 3 &&
         selectedPlayers.value.length <= 6 &&
         gameName.value.trim().length > 0 &&
         selectedLeagueCode.value.length > 0
})

function isPlayerSelected(member: LeagueMember): boolean {
  return selectedPlayers.value.some(p => p.code === member.code)
}

function getPlayerIndex(member: LeagueMember): number {
  return selectedPlayers.value.findIndex(p => p.code === member.code)
}

function togglePlayer(member: LeagueMember) {
  const index = selectedPlayers.value.findIndex(p => p.code === member.code)

  if (index >= 0) {
    // Remove player
    selectedPlayers.value.splice(index, 1)

    // Adjust first dealer index if needed
    if (firstDealerIndex.value >= selectedPlayers.value.length) {
      firstDealerIndex.value = Math.max(0, selectedPlayers.value.length - 1)
    }
  } else {
    // Add player (if under max)
    if (selectedPlayers.value.length < maxPlayers) {
      selectedPlayers.value.push(member)
    } else {
      error.value = `Maximum ${maxPlayers} players allowed`
    }
  }
}

function setFirstDealer(member: LeagueMember) {
  const index = getPlayerIndex(member)
  if (index >= 0) {
    firstDealerIndex.value = index
  }
}

async function onLeagueChange(code: string) {
  if (!code) return

  try {
    await leagueStore.setCurrentLeague(code)
    selectedPlayers.value = []
    firstDealerIndex.value = 0
  } catch (err: any) {
    error.value = err.message || 'Failed to load league members'
  }
}

async function createGame() {
  if (!isValid.value) {
    error.value = 'Please fill in all required fields'
    return
  }

  loading.value = true
  error.value = null

  try {
    const request = {
      league_id: leagueStore.currentLeague?.code || selectedLeagueCode.value,
      game_name: gameName.value,
      bid_restriction: bidRestriction.value,
      game_variant: 'STANDARD',
      first_dealer_index: firstDealerIndex.value,
      player_membership_ids: selectedPlayers.value.map(p => p.code)
    }

    await wizardStore.createGame(request)

    // Navigate to game
    if (wizardStore.currentGame) {
      router.push(`/ui/wizard/${wizardStore.currentGame.code}`)
    }
  } catch (err: any) {
    error.value = err.message || 'Failed to create game'
    console.error('Error creating game:', err)
  } finally {
    loading.value = false
  }
}

function cancel() {
  router.push('/ui/game-rounds')
}

onMounted(async () => {
  // Load leagues if not already loaded
  if (leagueStore.leagues.length === 0) {
    try {
      await leagueStore.loadLeagues()
    } catch (err: any) {
      error.value = err.message || 'Failed to load leagues'
    }
  }

  // Auto-select current league if available
  if (leagueStore.currentLeague) {
    selectedLeagueCode.value = leagueStore.currentLeague.code
  }
})
</script>

<style scoped>
.gap-2 {
  gap: 0.5rem;
}

.bg-blue-lighten-5 {
  background-color: rgba(33, 150, 243, 0.05);
}
</style>

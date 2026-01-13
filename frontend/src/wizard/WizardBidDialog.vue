<template>
  <v-dialog v-model="isOpen" max-width="600" persistent>
    <v-card>
      <v-card-title class="text-h5">
        Enter Bids - Round {{ roundNumber }}
      </v-card-title>

      <v-card-subtitle v-if="bidRestriction !== 'NO_RESTRICTIONS'" class="mt-2">
        <v-chip size="small" color="warning" variant="outlined">
          {{ bidRestrictionText }}
        </v-chip>
      </v-card-subtitle>

      <v-card-text>
        <v-alert v-if="error" type="error" class="mb-4" closable @click:close="error = null">
          {{ error }}
        </v-alert>

        <v-alert v-if="validationError" type="warning" class="mb-4">
          {{ validationError }}
        </v-alert>

        <div class="mb-3">
          <v-chip size="small" class="mr-2">
            <v-icon start>mdi-cards</v-icon>
            {{ cardsCount }} cards
          </v-chip>
          <v-chip size="small">
            <v-icon start>mdi-account-group</v-icon>
            {{ players.length }} players
          </v-chip>
        </div>

        <v-list>
          <v-list-item
            v-for="(player, index) in players"
            :key="player.membership_id"
            class="mb-2"
          >
            <v-row align="center">
              <v-col cols="5">
                <v-list-item-title>
                  {{ player.player_name }}
                  <v-chip
                    v-if="index === dealerIndex"
                    size="x-small"
                    color="primary"
                    class="ml-2"
                  >
                    Dealer
                  </v-chip>
                </v-list-item-title>
              </v-col>
              <v-col cols="7">
                <v-slider
                  v-model="bids[index]"
                  :min="0"
                  :max="cardsCount"
                  :step="1"
                  thumb-label="always"
                  color="primary"
                  hide-details
                  @update:model-value="validateBids"
                >
                  <template v-slot:prepend>
                    <v-btn
                      icon="mdi-minus"
                      size="small"
                      variant="text"
                      @click="decrementBid(index)"
                    />
                  </template>
                  <template v-slot:append>
                    <v-btn
                      icon="mdi-plus"
                      size="small"
                      variant="text"
                      @click="incrementBid(index)"
                    />
                  </template>
                </v-slider>
              </v-col>
            </v-row>
          </v-list-item>
        </v-list>

        <v-divider class="my-4" />

        <div class="d-flex justify-space-between text-body-1">
          <span>Total Bids:</span>
          <span :class="totalBidsColor" class="font-weight-bold">
            {{ totalBids }}
          </span>
        </div>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn
          color="grey"
          variant="text"
          @click="cancel"
          :disabled="loading"
        >
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          variant="elevated"
          @click="submit"
          :loading="loading"
          :disabled="!!validationError"
        >
          Submit Bids
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { WizardPlayer } from './types'
import { BidRestriction } from './types'

interface Props {
  modelValue: boolean
  roundNumber: number
  cardsCount: number
  players: WizardPlayer[]
  dealerIndex: number
  bidRestriction: BidRestriction
  existingBids?: number[]
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', bids: number[]): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const bids = ref<number[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const validationError = ref<string | null>(null)

// Initialize bids
watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    // Initialize with existing bids or zeros
    bids.value = props.existingBids
      ? [...props.existingBids]
      : props.players.map(() => 0)
    validateBids()
  }
}, { immediate: true })

const totalBids = computed(() => {
  return bids.value.reduce((sum, bid) => sum + bid, 0)
})

const totalBidsColor = computed(() => {
  if (props.bidRestriction === BidRestriction.NO_RESTRICTIONS) {
    return 'text-primary'
  }

  const total = totalBids.value
  const cards = props.cardsCount

  if (props.bidRestriction === BidRestriction.CANNOT_MATCH_CARDS) {
    return total === cards ? 'text-error' : 'text-success'
  }

  if (props.bidRestriction === BidRestriction.MUST_MATCH_CARDS) {
    return total === cards ? 'text-success' : 'text-error'
  }

  return 'text-primary'
})

const bidRestrictionText = computed(() => {
  switch (props.bidRestriction) {
    case BidRestriction.CANNOT_MATCH_CARDS:
      return `Total bids cannot equal ${props.cardsCount}`
    case BidRestriction.MUST_MATCH_CARDS:
      return `Total bids must equal ${props.cardsCount}`
    default:
      return ''
  }
})

function validateBids() {
  validationError.value = null

  if (props.bidRestriction === BidRestriction.NO_RESTRICTIONS) {
    return
  }

  const total = totalBids.value
  const cards = props.cardsCount

  if (props.bidRestriction === BidRestriction.CANNOT_MATCH_CARDS) {
    if (total === cards) {
      validationError.value = `Total bids cannot equal ${cards}`
    }
  }

  if (props.bidRestriction === BidRestriction.MUST_MATCH_CARDS) {
    if (total !== cards) {
      validationError.value = `Total bids must equal ${cards} (currently ${total})`
    }
  }
}

function incrementBid(index: number) {
  if (bids.value[index] < props.cardsCount) {
    bids.value[index]++
    validateBids()
  }
}

function decrementBid(index: number) {
  if (bids.value[index] > 0) {
    bids.value[index]--
    validateBids()
  }
}

function cancel() {
  isOpen.value = false
  error.value = null
  validationError.value = null
}

async function submit() {
  if (validationError.value) {
    return
  }

  loading.value = true
  error.value = null

  try {
    emit('submit', [...bids.value])
    isOpen.value = false
  } catch (err: any) {
    error.value = err.message || 'Failed to submit bids'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.v-list-item {
  border: 1px solid rgba(0, 0, 0, 0.12);
  border-radius: 4px;
}
</style>

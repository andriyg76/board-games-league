<template>
  <v-dialog v-model="isOpen" max-width="600" persistent>
    <v-card>
      <v-card-title class="text-h5">
        Enter Results - Round {{ roundNumber }}
      </v-card-title>

      <v-card-subtitle class="mt-2">
        <v-chip size="small" color="info" variant="outlined">
          Total tricks must equal {{ cardsCount }}
        </v-chip>
      </v-card-subtitle>

      <v-card-text>
        <v-alert v-if="error" type="error" class="mb-4" closable @click:close="error = null">
          {{ error }}
        </v-alert>

        <v-alert v-if="validationError" type="warning" class="mb-4">
          {{ validationError }}
        </v-alert>

        <v-alert v-if="allResultsValid" type="success" class="mb-4">
          <v-icon start>mdi-check-circle</v-icon>
          All results valid! Total: {{ totalResults }} = {{ cardsCount }}
        </v-alert>

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
                </v-list-item-title>
                <v-list-item-subtitle v-if="playerBids[index] !== undefined">
                  Bid: {{ playerBids[index] }}
                  <v-chip
                    v-if="results[index] !== -1 && results[index] === playerBids[index]"
                    size="x-small"
                    color="success"
                    class="ml-1"
                  >
                    Match!
                  </v-chip>
                </v-list-item-subtitle>
              </v-col>
              <v-col cols="7">
                <v-slider
                  v-model="results[index]"
                  :min="0"
                  :max="cardsCount"
                  :step="1"
                  thumb-label="always"
                  :color="getSliderColor(index)"
                  hide-details
                  @update:model-value="validateResults"
                >
                  <template v-slot:prepend>
                    <v-btn
                      icon="mdi-minus"
                      size="small"
                      variant="text"
                      @click="decrementResult(index)"
                    />
                  </template>
                  <template v-slot:append>
                    <v-btn
                      icon="mdi-plus"
                      size="small"
                      variant="text"
                      @click="incrementResult(index)"
                    />
                  </template>
                </v-slider>
              </v-col>
            </v-row>
          </v-list-item>
        </v-list>

        <v-divider class="my-4" />

        <div class="d-flex justify-space-between text-body-1">
          <span>Total Tricks:</span>
          <span :class="totalResultsColor" class="font-weight-bold">
            {{ totalResults }} / {{ cardsCount }}
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
          Submit Results
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { WizardPlayer } from './types'

interface Props {
  modelValue: boolean
  roundNumber: number
  cardsCount: number
  players: WizardPlayer[]
  playerBids: number[]
  existingResults?: number[]
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', results: number[]): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const results = ref<number[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const validationError = ref<string | null>(null)

// Initialize results
watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    // Initialize with existing results or zeros
    results.value = props.existingResults
      ? [...props.existingResults]
      : props.players.map(() => 0)
    validateResults()
  }
}, { immediate: true })

const totalResults = computed(() => {
  return results.value.reduce((sum, result) => sum + result, 0)
})

const allResultsValid = computed(() => {
  return !validationError.value && totalResults.value === props.cardsCount
})

const totalResultsColor = computed(() => {
  const total = totalResults.value
  const cards = props.cardsCount

  if (total === cards) {
    return 'text-success'
  } else if (total > cards) {
    return 'text-error'
  } else {
    return 'text-warning'
  }
})

function getSliderColor(index: number): string {
  const result = results.value[index]
  const bid = props.playerBids[index]

  if (result === -1 || bid === undefined) {
    return 'primary'
  }

  return result === bid ? 'success' : 'warning'
}

function validateResults() {
  validationError.value = null

  const total = totalResults.value
  const cards = props.cardsCount

  if (total !== cards) {
    validationError.value = `Total tricks must equal ${cards} (currently ${total})`
  }
}

function incrementResult(index: number) {
  if (results.value[index] < props.cardsCount) {
    results.value[index]++
    validateResults()
  }
}

function decrementResult(index: number) {
  if (results.value[index] > 0) {
    results.value[index]--
    validateResults()
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
    emit('submit', [...results.value])
    isOpen.value = false
  } catch (err: any) {
    error.value = err.message || 'Failed to submit results'
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

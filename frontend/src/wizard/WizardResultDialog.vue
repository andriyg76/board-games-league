<template>
  <n-modal v-model:show="isOpen" preset="card" :title="`Enter Results - Round ${roundNumber}`" style="max-width: 600px;" :mask-closable="false">
    <n-tag type="info" style="margin-bottom: 16px;">
      Total tricks must equal {{ cardsCount }}
    </n-tag>

    <n-alert v-if="error" type="error" style="margin-bottom: 16px;" closable @close="error = null">
      {{ error }}
    </n-alert>

    <n-alert v-if="validationError" type="warning" style="margin-bottom: 16px;">
      {{ validationError }}
    </n-alert>

    <n-alert v-if="allResultsValid" type="success" style="margin-bottom: 16px;">
      <template #icon>
        <n-icon><CheckCircleIcon /></n-icon>
      </template>
      All results valid! Total: {{ totalResults }} = {{ cardsCount }}
    </n-alert>

    <n-list>
      <n-list-item
        v-for="(player, index) in players"
        :key="player.membership_id"
        style="margin-bottom: 16px; padding: 12px; border: 1px solid rgba(0, 0, 0, 0.12); border-radius: 4px;"
      >
        <div style="width: 100%;">
          <div style="margin-bottom: 8px;">
            <div style="font-weight: 500;">{{ player.player_name }}</div>
            <div v-if="playerBids[index] !== undefined" style="font-size: 0.875rem; opacity: 0.7; display: flex; align-items: center; gap: 8px;">
              Bid: {{ playerBids[index] }}
              <n-tag v-if="results[index] !== -1 && results[index] === playerBids[index]" size="small" type="success">
                Match!
              </n-tag>
            </div>
          </div>
          <div style="display: flex; align-items: center; gap: 8px;">
            <n-button size="small" quaternary @click="decrementResult(index)">
              <template #icon>
                <n-icon><RemoveIcon /></n-icon>
              </template>
            </n-button>
            <n-slider
              v-model:value="results[index]"
              :min="0"
              :max="cardsCount"
              :step="1"
              :tooltip="true"
              :style="{ flex: 1, '--n-handle-color': getSliderColor(index) }"
              @update:value="validateResults"
            />
            <n-button size="small" quaternary @click="incrementResult(index)">
              <template #icon>
                <n-icon><AddIcon /></n-icon>
              </template>
            </n-button>
            <span style="min-width: 30px; text-align: center; font-weight: 500;">{{ results[index] }}</span>
          </div>
        </div>
      </n-list-item>
    </n-list>

    <n-divider style="margin: 16px 0;" />

    <div style="display: flex; justify-content: space-between; align-items: center;">
      <span>Total Tricks:</span>
      <span :style="{ color: getTotalResultsColor(), fontWeight: 'bold' }">
        {{ totalResults }} / {{ cardsCount }}
      </span>
    </div>

    <template #action>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <n-button quaternary @click="cancel" :disabled="loading">
          Cancel
        </n-button>
        <n-button
          type="primary"
          @click="submit"
          :loading="loading"
          :disabled="!!validationError"
        >
          Submit Results
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NModal, NTag, NIcon, NAlert, NList, NListItem, NSlider, NButton, NDivider } from 'naive-ui'
import { CheckmarkCircle as CheckCircleIcon, Remove as RemoveIcon, Add as AddIcon } from '@vicons/ionicons5'
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
    // Initialize with existing results or zeros, ensuring array length matches players
    const initialResults: number[] = []
    for (let i = 0; i < props.players.length; i++) {
      // Use existing result if available and valid, otherwise use 0
      if (props.existingResults && i < props.existingResults.length && props.existingResults[i] >= 0) {
        initialResults.push(props.existingResults[i])
      } else {
        initialResults.push(0)
      }
    }
    results.value = initialResults
    validateResults()
  }
}, { immediate: true })

const totalResults = computed(() => {
  return results.value.reduce((sum, result) => sum + result, 0)
})

const allResultsValid = computed(() => {
  return !validationError.value && totalResults.value === props.cardsCount
})

const getTotalResultsColor = () => {
  const total = totalResults.value
  const cards = props.cardsCount

  if (total === cards) {
    return '#18a058'
  } else if (total > cards) {
    return '#d03050'
  } else {
    return '#f0a020'
  }
}

function getSliderColor(index: number): string {
  const result = results.value[index]
  const bid = props.playerBids[index]

  if (result === -1 || bid === undefined) {
    return '#2080f0'
  }

  return result === bid ? '#18a058' : '#f0a020'
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


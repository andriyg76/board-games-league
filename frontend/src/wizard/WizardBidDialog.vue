<template>
  <n-modal v-model:show="isOpen" preset="card" :title="`Enter Bids - Round ${roundNumber}`" style="max-width: 600px;" :mask-closable="false">
    <n-tag v-if="bidRestriction !== 'NO_RESTRICTIONS'" type="warning" style="margin-bottom: 16px;">
      {{ bidRestrictionText }}
    </n-tag>

    <n-alert v-if="error" type="error" style="margin-bottom: 16px;" closable @close="error = null">
      {{ error }}
    </n-alert>

    <n-alert v-if="validationError" type="warning" style="margin-bottom: 16px;">
      {{ validationError }}
    </n-alert>

    <div style="display: flex; gap: 8px; margin-bottom: 16px;">
      <n-tag size="small">
        <template #icon>
          <n-icon><CardsIcon /></n-icon>
        </template>
        {{ cardsCount }} cards
      </n-tag>
      <n-tag size="small">
        <template #icon>
          <n-icon><PeopleIcon /></n-icon>
        </template>
        {{ players.length }} players
      </n-tag>
    </div>

    <n-list>
      <n-list-item
        v-for="(player, index) in players"
        :key="player.membership_id"
        style="margin-bottom: 16px; padding: 12px; border: 1px solid rgba(0, 0, 0, 0.12); border-radius: 4px;"
      >
        <div style="width: 100%;">
          <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
            <span style="font-weight: 500;">{{ player.player_name }}</span>
            <n-tag v-if="index === dealerIndex" size="small" type="primary">
              Dealer
            </n-tag>
          </div>
          <div style="display: flex; align-items: center; gap: 8px;">
            <n-button size="small" quaternary @click="decrementBid(index)">
              <template #icon>
                <n-icon><RemoveIcon /></n-icon>
              </template>
            </n-button>
            <n-slider
              v-model:value="bids[index]"
              :min="0"
              :max="cardsCount"
              :step="1"
              :tooltip="true"
              style="flex: 1;"
              @update:value="validateBids"
            />
            <n-button size="small" quaternary @click="incrementBid(index)">
              <template #icon>
                <n-icon><AddIcon /></n-icon>
              </template>
            </n-button>
            <span style="min-width: 30px; text-align: center; font-weight: 500;">{{ bids[index] }}</span>
          </div>
        </div>
      </n-list-item>
    </n-list>

    <n-divider style="margin: 16px 0;" />

    <div style="display: flex; justify-content: space-between; align-items: center;">
      <span>Total Bids:</span>
      <span :style="{ color: getTotalBidsColor(), fontWeight: 'bold' }">
        {{ totalBids }}
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
          Submit Bids
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NModal, NTag, NIcon, NAlert, NList, NListItem, NSlider, NButton, NDivider } from 'naive-ui'
import { Card as CardsIcon, People as PeopleIcon, Remove as RemoveIcon, Add as AddIcon } from '@vicons/ionicons5'
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
    // Initialize with existing bids or zeros, ensuring array length matches players
    const initialBids: number[] = []
    for (let i = 0; i < props.players.length; i++) {
      // Use existing bid if available and valid, otherwise use 0
      if (props.existingBids && i < props.existingBids.length && props.existingBids[i] >= 0) {
        initialBids.push(props.existingBids[i])
      } else {
        initialBids.push(0)
      }
    }
    bids.value = initialBids
    validateBids()
  }
}, { immediate: true })

const totalBids = computed(() => {
  return bids.value.reduce((sum, bid) => sum + bid, 0)
})

const getTotalBidsColor = () => {
  if (props.bidRestriction === BidRestriction.NO_RESTRICTIONS) {
    return '#2080f0'
  }

  const total = totalBids.value
  const cards = props.cardsCount

  if (props.bidRestriction === BidRestriction.CANNOT_MATCH_CARDS) {
    return total === cards ? '#d03050' : '#18a058'
  }

  if (props.bidRestriction === BidRestriction.MUST_MATCH_CARDS) {
    return total === cards ? '#18a058' : '#d03050'
  }

  return '#2080f0'
}

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


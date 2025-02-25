<template>
  <v-select
      v-model="currentLocale"
      :items="availableLocales"
      dense
      hide-details
      class="language-selector"
      @update:model-value="changeLocale"
  />
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()
const currentLocale = ref(locale.value)
const availableLocales = ['en', 'uk', 'et']

const changeLocale = (value: string) => {
  locale.value = value
  localStorage.setItem('locale', value)
}

// Initialize locale from localStorage
const savedLocale = localStorage.getItem('locale')
if (savedLocale) {
  locale.value = savedLocale
  currentLocale.value = savedLocale
}
</script>
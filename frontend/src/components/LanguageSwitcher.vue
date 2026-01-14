<template>
  <n-dropdown :options="localeOptions" trigger="click" @select="changeLocale">
    <n-button quaternary size="small">
      <img
          :src="`/flags/${currentLocale}.svg`"
          width="24"
          height="24"
          style="object-fit: cover;"
          :alt="currentLocale"
      />
    </n-button>
  </n-dropdown>
</template>

<script lang="ts" setup>
import { ref, computed, h } from 'vue'
import { NButton, NDropdown, NImage } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()
const currentLocale = ref(locale.value)
const availableLocales = ['en', 'uk', 'et']

const localeOptions = computed(() => 
  availableLocales.map(loc => ({
    label: () => h('div', { style: 'display: flex; align-items: center; gap: 8px;' }, [
      h('img', {
        src: `/flags/${loc}.svg`,
        width: 24,
        height: 24,
        style: 'object-fit: cover;'
      }),
      h('span', loc.toUpperCase())
    ]),
    key: loc,
  }))
)

const changeLocale = (value: string) => {
  locale.value = value
  localStorage.setItem('locale', value)
  currentLocale.value = value
}

// Initialize locale from localStorage
const savedLocale = localStorage.getItem('locale')
if (savedLocale) {
  locale.value = savedLocale
  currentLocale.value = savedLocale
}
</script>
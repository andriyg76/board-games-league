<template>
  <v-menu location="bottom">
    <template v-slot:activator="{ props }">
      <v-btn
          icon
          v-bind="props"
          size="small"
      >
        <v-img
            :src="`/flags/${currentLocale}.svg`"
            width="24"
            height="24"
            cover
        ></v-img>
      </v-btn>
    </template>
    <v-list>
      <v-list-item
          v-for="locale in availableLocales"
          :key="locale"
          :value="locale"
          @click="changeLocale(locale)"
      >
        <template v-slot:prepend>
          <v-img
              :src="`/flags/${locale}.svg`"
              width="24"
              height="24"
              cover
              class="mr-2"
          ></v-img>
        </template>
        <v-list-item-title>{{ locale.toUpperCase() }}</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
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
  currentLocale.value = value
}

// Initialize locale from localStorage
const savedLocale = localStorage.getItem('locale')
if (savedLocale) {
  locale.value = savedLocale
  currentLocale.value = savedLocale
}
</script>
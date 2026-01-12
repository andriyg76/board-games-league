<template>
  <v-list-item
    :value="league.code"
    :class="{ 'bg-grey-lighten-3': league.status === 'archived' }"
  >
    <template v-slot:prepend>
      <v-icon :color="statusColor">{{ statusIcon }}</v-icon>
    </template>

    <v-list-item-title>{{ league.name }}</v-list-item-title>
    <v-list-item-subtitle>
      <v-chip
        :color="statusColor"
        size="small"
        variant="flat"
        class="mr-2"
      >
        {{ statusText }}
      </v-chip>
      {{ t('leagues.createdOn') }} {{ formatDate(league.created_at) }}
    </v-list-item-subtitle>

    <template v-slot:append>
      <v-icon>mdi-chevron-right</v-icon>
    </template>
  </v-list-item>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { League } from '@/api/LeagueApi';

interface Props {
  league: League;
}

const { t, locale } = useI18n();
const props = defineProps<Props>();

const statusColor = computed(() => {
  return props.league.status === 'active' ? 'success' : 'grey';
});

const statusIcon = computed(() => {
  return props.league.status === 'active' ? 'mdi-trophy' : 'mdi-archive';
});

const statusText = computed(() => {
  return props.league.status === 'active' ? t('leagues.active') : t('leagues.archived');
});

const formatDate = (dateStr: string) => {
  const localeMap: Record<string, string> = { 'uk': 'uk-UA', 'en': 'en-US', 'et': 'et-EE' };
  return new Date(dateStr).toLocaleDateString(localeMap[locale.value] || 'en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
};
</script>

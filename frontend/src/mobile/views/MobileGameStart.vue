<template>
  <div class="mobile-game-start">
    <h1 class="mobile-game-start__title">{{ t('game.startGame') }}</h1>
    <p class="mobile-game-start__subtitle">Game setup coming soon.</p>
    <n-button @click="goBack">Back</n-button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { NButton } from 'naive-ui';

const { t } = useI18n();
const router = useRouter();

const goBack = () => {
  router.push({ name: 'MobileLeagueHome' });
};
</script>

<style scoped>
.mobile-game-start {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 24px;
}

.mobile-game-start__title {
  margin: 0;
  font-size: 1.5rem;
}

.mobile-game-start__subtitle {
  margin: 0;
  opacity: 0.7;
}
</style>

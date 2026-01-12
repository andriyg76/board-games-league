<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" md="6">
        <v-card elevation="4">
          <v-card-title class="text-h5 text-center bg-primary pa-6">
            <v-icon size="large" class="mr-2">mdi-email-open</v-icon>
            Запрошення до ліги
          </v-card-title>

          <!-- Loading State -->
          <v-card-text v-if="loading" class="text-center py-8">
            <v-progress-circular
              indeterminate
              color="primary"
              size="64"
              class="mb-4"
            />
            <div class="text-h6">Обробка запрошення...</div>
          </v-card-text>

          <!-- Error State -->
          <v-card-text v-else-if="error" class="text-center py-8">
            <v-icon
              size="80"
              color="error"
              class="mb-4"
            >
              mdi-alert-circle
            </v-icon>
            <div class="text-h6 mb-2">Помилка</div>
            <v-alert type="error" variant="tonal">
              {{ error }}
            </v-alert>

            <v-btn
              color="primary"
              class="mt-4"
              @click="goToHome"
            >
              На головну
            </v-btn>
          </v-card-text>

          <!-- Success State -->
          <v-card-text v-else-if="success && league" class="text-center py-8">
            <v-icon
              size="80"
              color="success"
              class="mb-4"
            >
              mdi-check-circle
            </v-icon>
            <div class="text-h6 mb-2">Вітаємо!</div>
            <div class="text-body-1 mb-4">
              Ви успішно приєдналися до ліги <strong>{{ league.name }}</strong>
            </div>

            <v-divider class="my-4" />

            <v-card variant="tonal" color="primary" class="mb-4">
              <v-card-text>
                <div class="text-subtitle-2 mb-2">Що далі?</div>
                <v-list density="compact" bg-color="transparent">
                  <v-list-item>
                    <template v-slot:prepend>
                      <v-icon>mdi-trophy</v-icon>
                    </template>
                    <v-list-item-title>Переглядайте рейтинг ліги</v-list-item-title>
                  </v-list-item>
                  <v-list-item>
                    <template v-slot:prepend>
                      <v-icon>mdi-gamepad-variant</v-icon>
                    </template>
                    <v-list-item-title>Грайте ігри та набирайте очки</v-list-item-title>
                  </v-list-item>
                  <v-list-item>
                    <template v-slot:prepend>
                      <v-icon>mdi-account-group</v-icon>
                    </template>
                    <v-list-item-title>Запрошуйте інших гравців</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-card-text>
            </v-card>

            <div class="d-flex gap-2 justify-center">
              <v-btn
                color="primary"
                variant="flat"
                @click="goToLeague"
              >
                <v-icon start>mdi-arrow-right</v-icon>
                Перейти до ліги
              </v-btn>
              <v-btn
                variant="outlined"
                @click="goToHome"
              >
                На головну
              </v-btn>
            </div>
          </v-card-text>

          <!-- Initial State (no token) -->
          <v-card-text v-else class="text-center py-8">
            <v-icon
              size="80"
              color="warning"
              class="mb-4"
            >
              mdi-help-circle
            </v-icon>
            <div class="text-h6 mb-2">Невірне запрошення</div>
            <div class="text-body-1 mb-4">
              Не знайдено токену запрошення. Переконайтесь, що ви перейшли за правильним посиланням.
            </div>

            <v-btn
              color="primary"
              @click="goToHome"
            >
              На головну
            </v-btn>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useLeagueStore } from '@/store/league';
import type { League } from '@/api/LeagueApi';

const route = useRoute();
const router = useRouter();
const leagueStore = useLeagueStore();

const loading = ref(false);
const success = ref(false);
const error = ref<string | null>(null);
const league = ref<League | null>(null);

const acceptInvitation = async (token: string) => {
  loading.value = true;
  error.value = null;

  try {
    const result = await leagueStore.acceptInvitation(token);
    league.value = result.league;
    success.value = true;
  } catch (err) {
    if (err instanceof Error) {
      if (err.message.includes('404')) {
        error.value = 'Запрошення не знайдено або вже використано';
      } else if (err.message.includes('expired')) {
        error.value = 'Запрошення прострочено (дійсне 7 днів)';
      } else {
        error.value = err.message;
      }
    } else {
      error.value = 'Не вдалося прийняти запрошення';
    }
    console.error('Error accepting invitation:', err);
  } finally {
    loading.value = false;
  }
};

const goToLeague = () => {
  if (league.value) {
    router.push({ name: 'LeagueDetails', params: { code: league.value.code } });
  }
};

const goToHome = () => {
  router.push({ name: 'Home' });
};

onMounted(async () => {
  const token = route.params.token as string;
  if (token) {
    await acceptInvitation(token);
  }
});
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
</style>

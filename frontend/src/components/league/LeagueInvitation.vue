<template>
  <div>
    <v-card variant="tonal" color="primary" class="mb-4">
      <v-card-text>
        <div class="d-flex align-center">
          <v-icon start size="large">mdi-information</v-icon>
          <div>
            <div class="text-subtitle-1">Як запросити гравця</div>
            <div class="text-caption">
              Створіть одноразове запрошення і відправте посилання гравцеві.
              Запрошення дійсне протягом 7 днів.
            </div>
          </div>
        </div>
      </v-card-text>
    </v-card>

    <v-card elevation="2">
      <v-card-text>
        <div class="text-center py-4">
          <v-btn
            color="primary"
            size="large"
            :loading="generating"
            @click="generateInvitation"
          >
            <v-icon start>mdi-link-plus</v-icon>
            Створити запрошення
          </v-btn>
        </div>

        <v-alert
          v-if="error"
          type="error"
          variant="tonal"
          class="mt-4"
          closable
          @click:close="error = null"
        >
          {{ error }}
        </v-alert>

        <v-expand-transition>
          <div v-if="invitationLink" class="mt-4">
            <v-divider class="mb-4" />

            <div class="text-subtitle-2 mb-2">Посилання для запрошення:</div>

            <v-card variant="outlined">
              <v-card-text>
                <div class="d-flex align-center gap-2">
                  <v-text-field
                    :model-value="invitationLink"
                    readonly
                    hide-details
                    variant="plain"
                    density="compact"
                  />
                  <v-btn
                    icon="mdi-content-copy"
                    variant="tonal"
                    @click="copyToClipboard"
                  />
                </div>
              </v-card-text>
            </v-card>

            <v-alert
              v-if="copied"
              type="success"
              variant="tonal"
              class="mt-2"
            >
              <v-icon start>mdi-check-circle</v-icon>
              Посилання скопійовано до буферу обміну!
            </v-alert>

            <div class="mt-4">
              <v-card variant="tonal" color="warning">
                <v-card-text>
                  <div class="d-flex align-center">
                    <v-icon start>mdi-clock-alert</v-icon>
                    <div class="text-caption">
                      Запрошення дійсне до {{ formatExpiryDate(invitation?.expires_at) }}
                    </div>
                  </div>
                </v-card-text>
              </v-card>
            </div>

            <div class="text-center mt-4">
              <v-btn
                variant="text"
                @click="resetInvitation"
              >
                Створити нове запрошення
              </v-btn>
            </div>
          </div>
        </v-expand-transition>
      </v-card-text>
    </v-card>
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useLeagueStore } from '@/store/league';
import type { LeagueInvitation } from '@/api/LeagueApi';

interface Props {
  leagueCode: string;
}

defineProps<Props>();

const leagueStore = useLeagueStore();

const generating = ref(false);
const invitation = ref<LeagueInvitation | null>(null);
const invitationLink = ref('');
const copied = ref(false);
const error = ref<string | null>(null);

const generateInvitation = async () => {
  generating.value = true;
  error.value = null;

  try {
    const result = await leagueStore.createInvitation();
    invitation.value = result.invitation;
    invitationLink.value = result.invitation_link;
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Помилка створення запрошення';
    console.error('Error generating invitation:', err);
  } finally {
    generating.value = false;
  }
};

const copyToClipboard = async () => {
  if (!invitationLink.value) return;

  try {
    await navigator.clipboard.writeText(invitationLink.value);
    copied.value = true;

    // Reset copied state after 3 seconds
    setTimeout(() => {
      copied.value = false;
    }, 3000);
  } catch (err) {
    console.error('Error copying to clipboard:', err);
    error.value = 'Не вдалося скопіювати посилання';
  }
};

const resetInvitation = () => {
  invitation.value = null;
  invitationLink.value = '';
  copied.value = false;
  error.value = null;
};

const formatExpiryDate = (dateStr: string | undefined) => {
  if (!dateStr) return '';

  return new Date(dateStr).toLocaleString('uk-UA', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
</style>

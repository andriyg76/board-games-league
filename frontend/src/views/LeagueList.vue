<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <v-card elevation="2">
          <v-card-title>
            <span class="text-h5">{{ t('leagues.title') }}</span>
            <v-spacer />
            <v-btn
              v-if="canCreateLeague"
              color="primary"
              @click="showCreateDialog = true"
            >
              <v-icon start>mdi-plus</v-icon>
              {{ t('leagues.createLeague') }}
            </v-btn>
          </v-card-title>
          <v-divider />

          <v-card-text v-if="loading">
            <v-progress-linear indeterminate color="primary" />
          </v-card-text>

          <v-card-text v-else-if="error">
            <v-alert type="error" variant="tonal">
              {{ error }}
            </v-alert>
          </v-card-text>

          <v-card-text v-else-if="activeLeagues.length === 0">
            <v-alert type="info" variant="tonal">
              {{ t('leagues.noActiveLeagues') }}
            </v-alert>
          </v-card-text>

          <v-list v-else>
            <league-card
              v-for="league in activeLeagues"
              :key="league.code"
              :league="league"
              @click="selectLeague(league.code)"
            />
          </v-list>

          <v-divider v-if="archivedLeagues.length > 0" />

          <v-expansion-panels v-if="archivedLeagues.length > 0">
            <v-expansion-panel>
              <v-expansion-panel-title>
                {{ t('leagues.archivedLeagues') }} ({{ archivedLeagues.length }})
              </v-expansion-panel-title>
              <v-expansion-panel-text>
                <v-list>
                  <league-card
                    v-for="league in archivedLeagues"
                    :key="league.code"
                    :league="league"
                    @click="selectLeague(league.code)"
                  />
                </v-list>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
        </v-card>
      </v-col>
    </v-row>

    <!-- Create League Dialog -->
    <v-dialog v-model="showCreateDialog" max-width="500">
      <v-card>
        <v-card-title>{{ t('leagues.createLeague') }}</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="newLeagueName"
            :label="t('leagues.leagueName')"
            :rules="[v => !!v || t('leagues.nameRequired')]"
            required
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="showCreateDialog = false">{{ t('leagues.cancel') }}</v-btn>
          <v-btn
            color="primary"
            :disabled="!newLeagueName"
            :loading="creating"
            @click="createLeague"
          >
            {{ t('leagues.create') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useLeagueStore } from '@/store/league';
import LeagueCard from '@/components/league/LeagueCard.vue';

const { t } = useI18n();
const router = useRouter();
const leagueStore = useLeagueStore();

const showCreateDialog = ref(false);
const newLeagueName = ref('');
const creating = ref(false);

// TODO: Implement superadmin check when user store is available
const canCreateLeague = ref(false);

const loading = computed(() => leagueStore.isLoading);
const error = computed(() => leagueStore.errorMessage);
const activeLeagues = computed(() => leagueStore.activeLeagues);
const archivedLeagues = computed(() => leagueStore.archivedLeagues);

const selectLeague = (code: string) => {
  router.push({ name: 'LeagueDetails', params: { code } });
};

const createLeague = async () => {
  if (!newLeagueName.value) return;

  creating.value = true;
  try {
    const league = await leagueStore.createLeague(newLeagueName.value);
    showCreateDialog.value = false;
    newLeagueName.value = '';
    router.push({ name: 'LeagueDetails', params: { code: league.code } });
  } catch (error) {
    console.error('Error creating league:', error);
  } finally {
    creating.value = false;
  }
};

onMounted(async () => {
  try {
    await leagueStore.loadLeagues();
  } catch (error) {
    console.error('Error loading leagues:', error);
  }
});
</script>

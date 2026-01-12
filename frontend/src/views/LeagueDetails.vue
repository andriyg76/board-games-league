<template>
  <v-container>
    <v-row v-if="loading">
      <v-col cols="12" class="text-center">
        <v-progress-circular indeterminate color="primary" size="64" />
      </v-col>
    </v-row>

    <v-row v-else-if="error">
      <v-col cols="12">
        <v-alert type="error" variant="tonal">
          {{ error }}
        </v-alert>
      </v-col>
    </v-row>

    <template v-else-if="currentLeague">
      <v-row>
        <v-col cols="12">
          <v-card elevation="2">
            <v-card-title>
              <v-row align="center">
                <v-col>
                  <div class="d-flex align-center">
                    <v-icon size="large" class="mr-3">mdi-trophy</v-icon>
                    <div>
                      <div class="text-h4">{{ currentLeague.name }}</div>
                      <v-chip
                        :color="currentLeague.status === 'active' ? 'success' : 'grey'"
                        size="small"
                        class="mt-2"
                      >
                        {{ currentLeague.status === 'active' ? 'Активна' : 'Архівна' }}
                      </v-chip>
                    </div>
                  </div>
                </v-col>
                <v-col cols="auto">
                  <v-btn
                    v-if="canManageLeague"
                    icon
                    @click="showManageMenu = !showManageMenu"
                  >
                    <v-icon>mdi-dots-vertical</v-icon>
                  </v-btn>
                  <v-menu v-model="showManageMenu" :close-on-content-click="true">
                    <template v-slot:activator="{ props }">
                      <div v-bind="props"></div>
                    </template>
                    <v-list>
                      <v-list-item
                        v-if="currentLeague.status === 'active'"
                        @click="archiveLeague"
                      >
                        <template v-slot:prepend>
                          <v-icon>mdi-archive</v-icon>
                        </template>
                        <v-list-item-title>Архівувати</v-list-item-title>
                      </v-list-item>
                      <v-list-item
                        v-else
                        @click="unarchiveLeague"
                      >
                        <template v-slot:prepend>
                          <v-icon>mdi-archive-arrow-up</v-icon>
                        </template>
                        <v-list-item-title>Розархівувати</v-list-item-title>
                      </v-list-item>
                    </v-list>
                  </v-menu>
                </v-col>
              </v-row>
            </v-card-title>
            <v-divider />

            <v-tabs v-model="activeTab" bg-color="transparent">
              <v-tab value="standings">
                <v-icon start>mdi-chart-line</v-icon>
                Рейтинг
              </v-tab>
              <v-tab value="members">
                <v-icon start>mdi-account-group</v-icon>
                Учасники ({{ members.length }})
              </v-tab>
              <v-tab value="invitation">
                <v-icon start>mdi-account-plus</v-icon>
                Запрошення
              </v-tab>
            </v-tabs>

            <v-card-text>
              <v-window v-model="activeTab">
                <v-window-item value="standings">
                  <league-standings :standings="standings" />
                </v-window-item>

                <v-window-item value="members">
                  <v-list>
                    <v-list-item
                      v-for="member in members"
                      :key="member.code"
                    >
                      <template v-slot:prepend>
                        <v-avatar :image="member.user_avatar" />
                      </template>

                      <v-list-item-title>{{ member.user_name }}</v-list-item-title>
                      <v-list-item-subtitle>
                        Приєднався {{ formatDate(member.joined_at) }}
                      </v-list-item-subtitle>

                      <template v-slot:append>
                        <v-chip
                          :color="member.status === 'active' ? 'success' : 'error'"
                          size="small"
                        >
                          {{ member.status === 'active' ? 'Активний' : 'Заблокований' }}
                        </v-chip>
                        <v-btn
                          v-if="canManageLeague && member.status === 'active'"
                          icon="mdi-block-helper"
                          size="small"
                          variant="text"
                          @click="banMember(member)"
                        />
                      </template>
                    </v-list-item>
                  </v-list>
                </v-window-item>

                <v-window-item value="invitation">
                  <league-invitation :league-code="currentLeague.code" />
                </v-window-item>
              </v-window>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </template>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useLeagueStore } from '@/store/league';
import LeagueStandings from '@/components/league/LeagueStandings.vue';
import LeagueInvitation from '@/components/league/LeagueInvitation.vue';
import type { LeagueMember } from '@/api/LeagueApi';

const route = useRoute();
const leagueStore = useLeagueStore();

const activeTab = ref('standings');
const showManageMenu = ref(false);

// TODO: Implement superadmin check when user store is available
const canManageLeague = ref(false);

const loading = computed(() => leagueStore.isLoading);
const error = computed(() => leagueStore.errorMessage);
const currentLeague = computed(() => leagueStore.currentLeague);
const members = computed(() => leagueStore.currentLeagueMembers);
const standings = computed(() => leagueStore.currentLeagueStandings);

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('uk-UA', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
};

const archiveLeague = async () => {
  if (!currentLeague.value) return;

  try {
    await leagueStore.archiveLeague(currentLeague.value.code);
  } catch (error) {
    console.error('Error archiving league:', error);
  }
};

const unarchiveLeague = async () => {
  if (!currentLeague.value) return;

  try {
    await leagueStore.unarchiveLeague(currentLeague.value.code);
  } catch (error) {
    console.error('Error unarchiving league:', error);
  }
};

const banMember = async (member: LeagueMember) => {
  if (!currentLeague.value) return;

  const confirmed = confirm(`Ви впевнені, що хочете заблокувати ${member.user_name}?`);
  if (!confirmed) return;

  try {
    await leagueStore.banUser(currentLeague.value.code, member.user_id);
  } catch (error) {
    console.error('Error banning member:', error);
  }
};

onMounted(async () => {
  const leagueCode = route.params.code as string;
  if (leagueCode) {
    try {
      await leagueStore.setCurrentLeague(leagueCode);
    } catch (error) {
      console.error('Error loading league:', error);
    }
  }
});
</script>

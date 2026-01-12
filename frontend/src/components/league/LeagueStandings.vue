<template>
  <div>
    <v-alert v-if="standings.length === 0" type="info" variant="tonal">
      Поки що немає даних для рейтингу. Зіграйте перші ігри в цій лізі!
    </v-alert>

    <v-data-table
      v-else
      :headers="headers"
      :items="standings"
      :items-per-page="10"
      class="elevation-1"
    >
      <template v-slot:item.position="{ index }">
        <div class="d-flex align-center">
          <v-icon
            v-if="index === 0"
            color="gold"
            size="small"
            class="mr-2"
          >
            mdi-medal
          </v-icon>
          <v-icon
            v-else-if="index === 1"
            color="silver"
            size="small"
            class="mr-2"
          >
            mdi-medal
          </v-icon>
          <v-icon
            v-else-if="index === 2"
            color="bronze"
            size="small"
            class="mr-2"
          >
            mdi-medal
          </v-icon>
          <span>{{ index + 1 }}</span>
        </div>
      </template>

      <template v-slot:item.user="{ item }">
        <div class="d-flex align-center py-2">
          <v-avatar
            :image="item.user_avatar"
            size="32"
            class="mr-3"
          />
          <span>{{ item.user_name }}</span>
        </div>
      </template>

      <template v-slot:item.total_points="{ item }">
        <v-chip color="primary" variant="flat">
          {{ item.total_points }}
        </v-chip>
      </template>

      <template v-slot:item.podiums="{ item }">
        <div class="d-flex gap-1">
          <v-chip
            v-if="item.first_place_count > 0"
            color="gold"
            size="small"
            variant="flat"
          >
            <v-icon start size="small">mdi-trophy</v-icon>
            {{ item.first_place_count }}
          </v-chip>
          <v-chip
            v-if="item.second_place_count > 0"
            color="silver"
            size="small"
            variant="flat"
          >
            <v-icon start size="small">mdi-trophy</v-icon>
            {{ item.second_place_count }}
          </v-chip>
          <v-chip
            v-if="item.third_place_count > 0"
            color="bronze"
            size="small"
            variant="flat"
          >
            <v-icon start size="small">mdi-trophy</v-icon>
            {{ item.third_place_count }}
          </v-chip>
        </div>
      </template>

      <template v-slot:item.details="{ item }">
        <v-btn
          icon="mdi-information"
          size="small"
          variant="text"
          @click="showDetails(item)"
        />
      </template>
    </v-data-table>

    <!-- Details Dialog -->
    <v-dialog v-model="detailsDialog" max-width="600">
      <v-card v-if="selectedPlayer">
        <v-card-title>
          <div class="d-flex align-center">
            <v-avatar
              :image="selectedPlayer.user_avatar"
              size="48"
              class="mr-3"
            />
            <div>
              <div>{{ selectedPlayer.user_name }}</div>
              <div class="text-caption text-medium-emphasis">
                Детальна статистика
              </div>
            </div>
          </div>
        </v-card-title>
        <v-divider />
        <v-card-text>
          <v-row dense>
            <v-col cols="6">
              <v-card variant="tonal" color="primary">
                <v-card-text class="text-center">
                  <div class="text-h4">{{ selectedPlayer.total_points }}</div>
                  <div class="text-caption">Всього очок</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6">
              <v-card variant="tonal" color="secondary">
                <v-card-text class="text-center">
                  <div class="text-h4">{{ selectedPlayer.games_played }}</div>
                  <div class="text-caption">Зіграно ігор</div>
                </v-card-text>
              </v-card>
            </v-col>
          </v-row>

          <v-divider class="my-4" />

          <div class="text-subtitle-2 mb-2">Розподіл очок:</div>
          <v-list density="compact">
            <v-list-item>
              <template v-slot:prepend>
                <v-icon color="success">mdi-account-check</v-icon>
              </template>
              <v-list-item-title>Очки за участь</v-list-item-title>
              <template v-slot:append>
                <v-chip size="small">{{ selectedPlayer.participation_points }}</v-chip>
              </template>
            </v-list-item>
            <v-list-item>
              <template v-slot:prepend>
                <v-icon color="warning">mdi-trophy</v-icon>
              </template>
              <v-list-item-title>Очки за позиції</v-list-item-title>
              <template v-slot:append>
                <v-chip size="small">{{ selectedPlayer.position_points }}</v-chip>
              </template>
            </v-list-item>
            <v-list-item>
              <template v-slot:prepend>
                <v-icon color="info">mdi-gavel</v-icon>
              </template>
              <v-list-item-title>Очки за модерацію</v-list-item-title>
              <template v-slot:append>
                <v-chip size="small">{{ selectedPlayer.moderation_points }}</v-chip>
              </template>
            </v-list-item>
          </v-list>

          <v-divider class="my-4" />

          <div class="text-subtitle-2 mb-2">Подіуми:</div>
          <v-row dense>
            <v-col cols="4">
              <v-card variant="tonal" color="gold">
                <v-card-text class="text-center">
                  <v-icon size="large">mdi-trophy</v-icon>
                  <div class="text-h6">{{ selectedPlayer.first_place_count }}</div>
                  <div class="text-caption">1-е місце</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="4">
              <v-card variant="tonal" color="silver">
                <v-card-text class="text-center">
                  <v-icon size="large">mdi-trophy</v-icon>
                  <div class="text-h6">{{ selectedPlayer.second_place_count }}</div>
                  <div class="text-caption">2-е місце</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="4">
              <v-card variant="tonal" color="bronze">
                <v-card-text class="text-center">
                  <v-icon size="large">mdi-trophy</v-icon>
                  <div class="text-h6">{{ selectedPlayer.third_place_count }}</div>
                  <div class="text-caption">3-є місце</div>
                </v-card-text>
              </v-card>
            </v-col>
          </v-row>

          <v-divider class="my-4" />

          <v-list density="compact">
            <v-list-item>
              <v-list-item-title>Ігор як модератор</v-list-item-title>
              <template v-slot:append>
                <v-chip size="small">{{ selectedPlayer.games_moderated }}</v-chip>
              </template>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="detailsDialog = false">Закрити</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import type { LeagueStanding } from '@/api/LeagueApi';

interface Props {
  standings: LeagueStanding[];
}

defineProps<Props>();

const headers = [
  { title: '#', key: 'position', sortable: false, width: 80 },
  { title: 'Гравець', key: 'user', sortable: false },
  { title: 'Очки', key: 'total_points', align: 'center' as const },
  { title: 'Ігор', key: 'games_played', align: 'center' as const },
  { title: 'Подіуми', key: 'podiums', sortable: false, align: 'center' as const },
  { title: '', key: 'details', sortable: false, width: 60 },
];

const detailsDialog = ref(false);
const selectedPlayer = ref<LeagueStanding | null>(null);

const showDetails = (player: LeagueStanding) => {
  selectedPlayer.value = player;
  detailsDialog.value = true;
};
</script>

<style scoped>
.gap-1 {
  gap: 4px;
}
</style>

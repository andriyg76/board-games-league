<template>
  <div>
    <n-grid :cols="24" :x-gap="16">
      <n-gi :span="24">
        <n-card>
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span style="font-size: 1.25rem; font-weight: 500;">{{ t('leagues.title') }}</span>
              <n-button
                v-if="canCreateLeague"
                type="primary"
                @click="showCreateDialog = true"
              >
                <template #icon>
                  <n-icon><AddIcon /></n-icon>
                </template>
                {{ t('leagues.createLeague') }}
              </n-button>
            </div>
          </template>
          <n-divider />

          <n-spin v-if="loading" style="padding: 24px;" />

          <n-alert v-else-if="error" type="error" style="margin: 16px 0;">
            {{ error }}
          </n-alert>

          <n-alert v-else-if="activeLeagues.length === 0" type="info" style="margin: 16px 0;">
            {{ t('leagues.noActiveLeagues') }}
          </n-alert>

          <n-list v-else>
            <league-card
              v-for="league in activeLeagues"
              :key="league.code"
              :league="league"
              @click="selectLeague(league.code)"
            />
          </n-list>

          <n-divider v-if="archivedLeagues.length > 0" />

          <n-collapse v-if="archivedLeagues.length > 0">
            <n-collapse-item :title="`${t('leagues.archivedLeagues')} (${archivedLeagues.length})`" name="archived">
              <n-list>
                <league-card
                  v-for="league in archivedLeagues"
                  :key="league.code"
                  :league="league"
                  @click="selectLeague(league.code)"
                />
              </n-list>
            </n-collapse-item>
          </n-collapse>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- Create League Dialog -->
    <n-modal v-model:show="showCreateDialog" preset="dialog" :title="t('leagues.createLeague')" positive-text="Create" negative-text="Cancel" @positive-click="handleCreate" @negative-click="showCreateDialog = false">
      <n-form-item :label="t('leagues.leagueName')" :required="true">
        <n-input v-model:value="newLeagueName" :placeholder="t('leagues.leagueName')" />
      </n-form-item>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import { NGrid, NGi, NCard, NButton, NIcon, NDivider, NList, NAlert, NSpin, NCollapse, NCollapseItem, NModal, NFormItem, NInput } from 'naive-ui';
import { Add as AddIcon } from '@vicons/ionicons5';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useLeagueStore } from '@/store/league';
import { useUserStore } from '@/store/user';
import LeagueCard from '@/components/league/LeagueCard.vue';

const { t } = useI18n();
const router = useRouter();
const leagueStore = useLeagueStore();
const userStore = useUserStore();

const showCreateDialog = ref(false);
const newLeagueName = ref('');
const creating = ref(false);

const canCreateLeague = computed(() => userStore.isSuperAdmin);

const loading = computed(() => leagueStore.isLoading);
const error = computed(() => leagueStore.errorMessage);
const activeLeagues = computed(() => leagueStore.activeLeagues);
const archivedLeagues = computed(() => leagueStore.archivedLeagues);

const selectLeague = (code: string) => {
  router.push({ name: 'LeagueDetails', params: { code } });
};

const handleCreate = async () => {
  if (!newLeagueName.value) return false;

  creating.value = true;
  try {
    const league = await leagueStore.createLeague(newLeagueName.value);
    showCreateDialog.value = false;
    newLeagueName.value = '';
    router.push({ name: 'LeagueDetails', params: { code: league.code } });
    return true;
  } catch (error) {
    console.error('Error creating league:', error);
    return false;
  } finally {
    creating.value = false;
  }
};

const createLeague = handleCreate;

onMounted(async () => {
  try {
    await leagueStore.loadLeagues();
  } catch (error) {
    console.error('Error loading leagues:', error);
  }
});
</script>

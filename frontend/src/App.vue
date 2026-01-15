<template>
  <n-config-provider>
    <n-message-provider>
      <n-notification-provider>
        <n-dialog-provider>
          <n-layout>
            <n-layout-header bordered style="padding: 12px 16px; display: flex; align-items: center; gap: 8px;">
              <!-- Mobile menu icon -->
              <n-button
                  v-if="isMobile"
                  quaternary
                  @click="drawer = !drawer"
              >
                <template #icon>
                  <n-icon><MenuIcon /></n-icon>
                </template>
              </n-button>

              <!-- Desktop menu -->
              <n-space v-if="!isMobile" :size="8">
                <router-link to="/" style="text-decoration: none;">
                  <n-button quaternary>{{ t('nav.home') }}</n-button>
                </router-link>
                <router-link v-if="loggedIn" to="/ui/admin/game-types" style="text-decoration: none;">
                  <n-button quaternary>{{ t('nav.gameTypes') }}</n-button>
                </router-link>
                <league-menu-item v-if="loggedIn" />
                <gameround-menu-item v-if="loggedIn" />
                <router-link v-if="loggedIn" to="/ui/user" style="text-decoration: none;">
                  <n-button quaternary>{{ t('nav.user') }}</n-button>
                </router-link>
                <router-link v-if="isSuperAdmin" to="/ui/admin/diagnostics" style="text-decoration: none;">
                  <n-button quaternary>{{ t('nav.diagnostics') }}</n-button>
                </router-link>
              </n-space>

              <div style="flex: 1;"></div>
              <n-divider vertical style="margin: 0 8px;" />
              <language-switcher />
              <logout-button/>
            </n-layout-header>

            <!-- Mobile side menu -->
            <n-drawer
                v-model:show="drawer"
                :width="drawerWidth"
                placement="left"
            >
              <n-drawer-content class="mobile-drawer-content" closable>
                <n-list class="mobile-drawer-list">
                  <router-link to="/" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('nav.home') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/admin/game-types" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('nav.gameTypes') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <div v-if="loggedIn" class="mobile-drawer-section">
                    <div class="mobile-drawer-section-title">{{ t('leagues.menu') }}</div>
                    <div v-if="drawerLeagueItems.length === 0" class="mobile-drawer-empty">
                      {{ t('leagues.noActiveLeagues') }}
                    </div>
                    <template v-else>
                      <router-link
                        v-for="league in drawerLeagueItems"
                        :key="league.code"
                        :to="{ name: 'LeagueDetails', params: { code: league.code } }"
                        custom
                        v-slot="{ navigate, href }"
                      >
                        <n-list-item
                          class="mobile-drawer-item"
                          :class="{ 'mobile-drawer-item--active': league.code === currentLeagueCode }"
                        >
                          <a :href="href" class="mobile-drawer-link mobile-league-link" @click="handleDrawerNavigate(navigate)">
                            <span class="mobile-league-check">
                              <n-icon v-if="league.code === currentLeagueCode" size="16" color="#18a058">
                                <CheckmarkIcon />
                              </n-icon>
                            </span>
                            <span class="mobile-league-name">{{ league.name }}</span>
                            <span v-if="league.status === 'archived'" class="mobile-league-status">
                              ({{ t('leagues.archived') }})
                            </span>
                          </a>
                        </n-list-item>
                      </router-link>
                    </template>
                    <router-link to="/ui/leagues" custom v-slot="{ navigate, href, isActive }">
                      <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                        <a :href="href" class="mobile-drawer-link mobile-league-link" @click="handleDrawerNavigate(navigate)">
                          {{ t('nav.leagues') }}
                        </a>
                      </n-list-item>
                    </router-link>
                  </div>
                  <router-link v-if="loggedIn" to="/ui/game-rounds" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('gameRounds.title') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/game-rounds/new" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('gameRounds.start') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/game-rounds/new?gameType=wizard" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('wizard.newGame') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/game-rounds" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('gameRounds.list') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/user" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('nav.user') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="isSuperAdmin" to="/ui/admin/diagnostics" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('nav.diagnostics') }}
                      </a>
                    </n-list-item>
                  </router-link>
                </n-list>
              </n-drawer-content>
            </n-drawer>

            <n-layout-content style="padding: 24px;">
              <router-view/>
            </n-layout-content>
          </n-layout>
        </n-dialog-provider>
      </n-notification-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { 
  NConfigProvider, 
  NMessageProvider, 
  NNotificationProvider, 
  NDialogProvider,
  NLayout, 
  NLayoutHeader, 
  NLayoutContent, 
  NButton, 
  NSpace, 
  NDivider, 
  NDrawer, 
  NList,
  NListItem,
  NDrawerContent,
  NIcon,
} from 'naive-ui';
import { Menu as MenuIcon, Checkmark as CheckmarkIcon } from '@vicons/ionicons5';
import { RouterLink, useRoute } from 'vue-router';
import LogoutButton from "@/components/LogoutButton.vue";
import { useUserStore } from '@/store/user';
import { useLeagueStore } from '@/store/league';
import GameroundMenuItem from "@/components/GameroundMenuItem.vue";
import LeagueMenuItem from "@/components/LeagueMenuItem.vue";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const userStore = useUserStore();
const leagueStore = useLeagueStore();
const route = useRoute();
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'

const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')

const drawer = ref(false);
const loggedIn = computed(() => userStore.$state.loggedIn);
const isSuperAdmin = computed(() => userStore.isSuperAdmin);
const drawerWidth = computed(() => (isMobile.value ? '100%' : 320));
const currentLeagueCode = computed(() => leagueStore.currentLeagueCode);
const drawerLeagueItems = computed(() => {
  const leagues = userStore.isSuperAdmin ? leagueStore.leagues : leagueStore.activeLeagues;
  const currentCode = currentLeagueCode.value;
  return [...leagues].sort((a, b) => {
    if (a.code === currentCode) return -1;
    if (b.code === currentCode) return 1;
    return 0;
  });
});
const handleDrawerNavigate = (navigate: (event?: MouseEvent) => void) => {
  navigate();
  drawer.value = false;
};

const loadLeagues = async () => {
  if (!loggedIn.value || leagueStore.leagues.length > 0) {
    return;
  }
  try {
    await leagueStore.loadLeagues();
  } catch (error) {
    console.error('Error loading leagues:', error);
  }
};

onMounted(() => {
  loadLeagues();
});

watch(loggedIn, (value) => {
  if (value) {
    loadLeagues();
  }
});

watch(
  () => route.fullPath,
  () => {
    drawer.value = false;
  }
);
</script>

<style scoped>
.mobile-drawer-content {
  padding: 8px 0;
}

.mobile-drawer-list {
  font-size: 1.125rem;
}

.mobile-drawer-item {
  padding: 0;
}

.mobile-drawer-item--active .mobile-drawer-link {
  background: rgba(32, 128, 240, 0.12);
  color: #2080f0;
  font-weight: 600;
}

.mobile-drawer-section {
  margin: 4px 0 8px;
}

.mobile-drawer-section-title {
  padding: 6px 20px 2px;
  font-size: 0.7rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  opacity: 0.6;
}

.mobile-drawer-empty {
  padding: 8px 20px 12px;
  font-size: 0.9rem;
  opacity: 0.7;
}

.mobile-drawer-link {
  display: flex;
  align-items: center;
  width: 100%;
  color: inherit;
  text-decoration: none;
  padding: 16px 20px;
  border-radius: 10px;
  transition: background 0.15s ease, color 0.15s ease;
}

.mobile-league-link {
  gap: 8px;
}

.mobile-league-check {
  display: inline-flex;
  width: 16px;
  justify-content: center;
  flex-shrink: 0;
}

.mobile-league-name {
  flex: 1;
  min-width: 0;
}

.mobile-league-status {
  font-size: 0.75rem;
  opacity: 0.6;
}

.mobile-drawer-link:active {
  background: rgba(32, 128, 240, 0.18);
}

.mobile-drawer-link:focus-visible {
  outline: 2px solid rgba(32, 128, 240, 0.4);
  outline-offset: 2px;
}
</style>

<template>
  <div>
    <n-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <div style="font-size: 1.25rem; font-weight: 500;">{{ t('diagnostics.title') }}</div>
          <n-button 
            :loading="activeLoading" 
            @click="refreshDiagnostics"
            type="primary"
            size="small"
          >
            <template #icon>
              <n-icon><RefreshIcon /></n-icon>
            </template>
            {{ t('diagnostics.refresh') }}
          </n-button>
        </div>
      </template>

      <n-tabs v-model:value="activeTab" type="line">
        <n-tab-pane name="request" :tab="t('diagnostics.tabs.request')">
          <div v-if="requestDiagnostics">
            <n-grid :cols="24" :x-gap="16" style="margin-bottom: 16px;">
              <n-gi :span="24" :responsive="{ m: 12 }">
                <n-card style="margin-bottom: 16px;">
                  <template #header>
                    <div style="font-size: 1rem; font-weight: 500;">{{ t('diagnostics.serverInfo') }}</div>
                  </template>
                  <p><strong>{{ t('diagnostics.hostUrl') }}:</strong> {{ requestDiagnostics.server_info.host_url }}</p>
                  <div>
                    <strong>{{ t('diagnostics.trustedOrigins') }}:</strong>
                    <ul v-if="requestDiagnostics.server_info.trusted_origins?.length > 0" style="margin: 8px 0; padding-left: 20px;">
                      <li v-for="origin in requestDiagnostics.server_info.trusted_origins" :key="origin">
                        {{ origin }}
                      </li>
                    </ul>
                    <p v-else style="opacity: 0.7; margin: 8px 0;">{{ t('diagnostics.noneConfigured') }}</p>
                  </div>
                </n-card>
              </n-gi>

              <n-gi :span="24" :responsive="{ m: 12 }">
                <n-card style="margin-bottom: 16px;">
                  <template #header>
                    <div style="font-size: 1rem; font-weight: 500;">{{ t('diagnostics.requestInfo') }}</div>
                  </template>
                  <p><strong>{{ t('diagnostics.ipAddress') }}:</strong> {{ requestDiagnostics.request_info.ip_address }}</p>
                  <p><strong>{{ t('diagnostics.baseUrl') }}:</strong> {{ requestDiagnostics.request_info.base_url }}</p>
                  <p><strong>{{ t('diagnostics.origin') }}:</strong> {{ requestDiagnostics.request_info.origin || t('diagnostics.na') }}</p>
                  <p>
                    <strong>{{ t('diagnostics.isTrusted') }}:</strong>
                    <n-tag :type="requestDiagnostics.request_info.is_trusted ? 'success' : 'error'" size="small" style="margin-left: 8px;">
                      {{ requestDiagnostics.request_info.is_trusted ? t('diagnostics.yes') : t('diagnostics.no') }}
                    </n-tag>
                  </p>
                  <p><strong>{{ t('diagnostics.userAgent') }}:</strong> {{ requestDiagnostics.request_info.user_agent }}</p>
                </n-card>
              </n-gi>
            </n-grid>

            <n-card v-if="requestDiagnostics.request_info.geo_info" style="margin-bottom: 16px;">
              <template #header>
                <div style="font-size: 1rem; font-weight: 500;">{{ t('diagnostics.geoInfo') }}</div>
              </template>
              <n-grid :cols="24" :x-gap="16">
                <n-gi :span="24" :responsive="{ m: 8 }">
                  <p><strong>{{ t('diagnostics.country') }}:</strong> {{ requestDiagnostics.request_info.geo_info.country || t('diagnostics.na') }}</p>
                  <p><strong>{{ t('diagnostics.countryCode') }}:</strong> {{ requestDiagnostics.request_info.geo_info.country_code || t('diagnostics.na') }}</p>
                </n-gi>
                <n-gi :span="24" :responsive="{ m: 8 }">
                  <p><strong>{{ t('diagnostics.region') }}:</strong> {{ requestDiagnostics.request_info.geo_info.region_name || requestDiagnostics.request_info.geo_info.region || t('diagnostics.na') }}</p>
                  <p><strong>{{ t('diagnostics.city') }}:</strong> {{ requestDiagnostics.request_info.geo_info.city || t('diagnostics.na') }}</p>
                </n-gi>
                <n-gi :span="24" :responsive="{ m: 8 }">
                  <p><strong>{{ t('diagnostics.timezone') }}:</strong> {{ requestDiagnostics.request_info.geo_info.timezone || t('diagnostics.na') }}</p>
                  <p><strong>{{ t('diagnostics.isp') }}:</strong> {{ requestDiagnostics.request_info.geo_info.isp || t('diagnostics.na') }}</p>
                </n-gi>
              </n-grid>
            </n-card>

            <n-card v-if="resolutionData.length > 0" style="margin-bottom: 16px;">
              <template #header>
                <div style="font-size: 1rem; font-weight: 500;">{{ t('diagnostics.resolutionInfo') }}</div>
              </template>
              <n-data-table
                :columns="resolutionColumns"
                :data="resolutionData"
                size="small"
              />
            </n-card>
          </div>
          <n-skeleton v-else height="200px" />
        </n-tab-pane>

        <n-tab-pane name="system" :tab="t('diagnostics.tabs.system')">
          <div v-if="systemDiagnostics">
            <n-card v-if="systemDiagnostics.runtime_info" style="margin-bottom: 16px;">
              <template #header>
                <div style="font-size: 1rem; font-weight: 500;">{{ t('diagnostics.runtimeInfo') }}</div>
              </template>
              <n-grid :cols="24" :x-gap="16">
                <n-gi :span="24" :responsive="{ m: 8 }">
                  <p><strong>{{ t('diagnostics.goVersion') }}:</strong> {{ systemDiagnostics.runtime_info.go_version }}</p>
                  <p><strong>{{ t('diagnostics.platform') }}:</strong> {{ systemDiagnostics.runtime_info.goos }}/{{ systemDiagnostics.runtime_info.goarch }}</p>
                  <p><strong>{{ t('diagnostics.numCpu') }}:</strong> {{ systemDiagnostics.runtime_info.num_cpu }}</p>
                </n-gi>
                <n-gi :span="24" :responsive="{ m: 8 }">
                  <p><strong>{{ t('diagnostics.numGoroutine') }}:</strong> {{ systemDiagnostics.runtime_info.num_goroutine }}</p>
                  <p><strong>{{ t('diagnostics.uptime') }}:</strong> {{ systemDiagnostics.runtime_info.uptime }}</p>
                  <p><strong>{{ t('diagnostics.startTime') }}:</strong> {{ formatDateTime(systemDiagnostics.runtime_info.start_time) }}</p>
                </n-gi>
                <n-gi :span="24" :responsive="{ m: 8 }">
                  <p><strong>{{ t('diagnostics.gcCycles') }}:</strong> {{ systemDiagnostics.runtime_info.memory.num_gc }}</p>
                </n-gi>
              </n-grid>
            </n-card>

            <n-card v-if="systemDiagnostics.runtime_info?.memory" style="margin-bottom: 16px;">
              <template #header>
                <div style="font-size: 1rem; font-weight: 500;">{{ t('diagnostics.memoryStats') }}</div>
              </template>
              <n-data-table
                :columns="memoryColumns"
                :data="memoryData"
                size="small"
              />
            </n-card>

            <n-card v-if="systemDiagnostics.environment_vars && systemDiagnostics.environment_vars.length > 0" style="margin-bottom: 16px;">
              <template #header>
                <div style="display: flex; align-items: center; gap: 8px;">
                  <span style="font-size: 1rem; font-weight: 500;">{{ t('diagnostics.environmentVars') }}</span>
                  <n-tag size="small">{{ systemDiagnostics.environment_vars.length }}</n-tag>
                </div>
              </template>
              <n-input
                v-model:value="envFilter"
                :placeholder="t('diagnostics.filterEnvVars')"
                clearable
                size="small"
                style="margin-bottom: 16px;"
              >
                <template #prefix>
                  <n-icon><SearchIcon /></n-icon>
                </template>
              </n-input>
              <n-data-table
                :columns="envColumns"
                :data="filteredEnvVars"
                size="small"
                :scroll-x="800"
                style="max-height: 400px;"
              />
            </n-card>

            <n-card v-if="systemDiagnostics.cache_stats && systemDiagnostics.cache_stats.length > 0" style="margin-bottom: 16px;">
              <template #header>
                <div style="font-size: 1rem; font-weight: 500;">{{ t('diagnostics.cacheStats') }}</div>
              </template>
              <n-data-table
                :columns="cacheStatsColumns"
                :data="systemDiagnostics.cache_stats"
                :pagination="false"
                size="small"
              />
            </n-card>
          </div>
          <n-skeleton v-else height="200px" />
        </n-tab-pane>

        <n-tab-pane name="build" :tab="t('diagnostics.tabs.build')">
          <div v-if="buildDiagnostics">
            <n-grid :cols="24" :x-gap="16" style="margin-bottom: 16px;">
              <n-gi :span="24" :responsive="{ m: 12 }">
                <n-card style="margin-bottom: 16px;">
                  <template #header>
                    <div style="font-size: 1rem; font-weight: 500;">{{ t('diagnostics.backendBuildInfo') }}</div>
                  </template>
                  <p><strong>{{ t('diagnostics.version') }}:</strong> {{ buildDiagnostics.build_info.version }}</p>
                  <p><strong>{{ t('diagnostics.commit') }}:</strong> {{ buildDiagnostics.build_info.commit }}</p>
                  <p><strong>{{ t('diagnostics.branch') }}:</strong> {{ buildDiagnostics.build_info.branch }}</p>
                  <p><strong>{{ t('diagnostics.buildDate') }}:</strong> {{ buildDiagnostics.build_info.date }}</p>
                </n-card>
              </n-gi>

              <n-gi :span="24" :responsive="{ m: 12 }">
                <n-card style="margin-bottom: 16px;">
                  <template #header>
                    <div style="font-size: 1rem; font-weight: 500;">{{ t('diagnostics.frontendBuildInfo') }}</div>
                  </template>
                  <p><strong>{{ t('diagnostics.version') }}:</strong> {{ frontendBuildInfo.version }}</p>
                  <p><strong>{{ t('diagnostics.commit') }}:</strong> {{ frontendBuildInfo.commit }}</p>
                  <p><strong>{{ t('diagnostics.branch') }}:</strong> {{ frontendBuildInfo.branch }}</p>
                  <p><strong>{{ t('diagnostics.buildDate') }}:</strong> {{ frontendBuildInfo.date }}</p>
                </n-card>
              </n-gi>
            </n-grid>
          </div>
          <n-skeleton v-else height="200px" />
        </n-tab-pane>
      </n-tabs>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, watch, h } from 'vue';
import { NGrid, NGi, NCard, NDataTable, NInput, NIcon, NTag, NSkeleton, NButton, NTabs, NTabPane, DataTableColumns } from 'naive-ui';
import { Search as SearchIcon, EyeOff as EyeOffIcon, Refresh as RefreshIcon } from '@vicons/ionicons5';
import DiagnosticsApi, { DiagnosticsRequestResponse, DiagnosticsSystemResponse, DiagnosticsBuildResponse, getFrontendBuildInfo, BuildInfo, EnvVarInfo, CacheStatsInfo } from "@/api/DiagnosticsApi";
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

type DiagnosticsTab = 'request' | 'system' | 'build';

const activeTab = ref<DiagnosticsTab>('request');
const requestDiagnostics = ref<DiagnosticsRequestResponse | null>(null);
const systemDiagnostics = ref<DiagnosticsSystemResponse | null>(null);
const buildDiagnostics = ref<DiagnosticsBuildResponse | null>(null);
const frontendBuildInfo = ref<BuildInfo>({
  version: "unknown",
  commit: "unknown",
  branch: "unknown",
  date: "unknown",
});
const envFilter = ref<string>("");
const loadingTabs = ref<Record<DiagnosticsTab, boolean>>({
  request: false,
  system: false,
  build: false,
});

const activeLoading = computed(() => loadingTabs.value[activeTab.value]);

// Filter environment variables based on search
const filteredEnvVars = computed<EnvVarInfo[]>(() => {
  if (!systemDiagnostics.value?.environment_vars) return [];
  if (!envFilter.value) return systemDiagnostics.value.environment_vars;
  
  const filter = envFilter.value.toLowerCase();
  return systemDiagnostics.value.environment_vars.filter(
    (env) => env.name.toLowerCase().includes(filter) || env.value.toLowerCase().includes(filter)
  );
});

// Resolution info table columns
const resolutionColumns: DataTableColumns = [
  { 
    title: t('diagnostics.headerName'), 
    key: 'name',
    render: (row: any) => h('code', row.name)
  },
  { 
    title: t('diagnostics.headerValue'), 
    key: 'value',
    render: (row: any) => h('code', row.value)
  },
];

const resolutionData = computed(() => {
  if (!requestDiagnostics.value?.request_info.resolution_info) return [];
  return Object.entries(requestDiagnostics.value.request_info.resolution_info).map(([key, value]) => ({
    name: key,
    value: String(value),
    key: key,
  }));
});

// Memory stats table columns
const memoryColumns: DataTableColumns = [
  { title: t('diagnostics.metric'), key: 'metric' },
  { 
    title: t('diagnostics.value'), 
    key: 'value', 
    align: 'right',
    render: (row: any) => h('code', row.value)
  },
];

const memoryData = computed(() => {
  if (!systemDiagnostics.value?.runtime_info?.memory) return [];
  const mem = systemDiagnostics.value.runtime_info.memory;
  return [
    { metric: t('diagnostics.heapAlloc'), value: formatBytes(mem.heap_alloc_bytes), key: 'heapAlloc' },
    { metric: t('diagnostics.heapInuse'), value: formatBytes(mem.heap_inuse_bytes), key: 'heapInuse' },
    { metric: t('diagnostics.heapSys'), value: formatBytes(mem.heap_sys_bytes), key: 'heapSys' },
    { metric: t('diagnostics.alloc'), value: formatBytes(mem.alloc_bytes), key: 'alloc' },
    { metric: t('diagnostics.totalAlloc'), value: formatBytes(mem.total_alloc_bytes), key: 'totalAlloc' },
    { metric: t('diagnostics.sysMem'), value: formatBytes(mem.sys_bytes), key: 'sysMem' },
  ];
});

// Environment variables table columns
const envColumns: DataTableColumns<EnvVarInfo> = [
  { 
    title: t('diagnostics.envName'), 
    key: 'name',
    render: (rowData: EnvVarInfo) => h('code', rowData.name)
  },
  { 
    title: t('diagnostics.envValue'), 
    key: 'value',
    render: (rowData: EnvVarInfo) => h('code', { style: { opacity: rowData.masked ? 0.7 : 1 } }, rowData.value)
  },
  { 
    title: t('diagnostics.envMasked'), 
    key: 'masked',
    align: 'center',
    render: (rowData: EnvVarInfo) => {
      return rowData.masked ? h(NIcon, { color: '#f0a020', size: 16 }, { default: () => h(EyeOffIcon) }) : null;
    }
  },
];

// Cache statistics table columns
const cacheStatsColumns: DataTableColumns<CacheStatsInfo> = [
  {
    title: t('diagnostics.cacheName'),
    key: 'name',
  },
  {
    title: t('diagnostics.cacheSize'),
    key: 'size',
    render(row: CacheStatsInfo) {
      return `${row.current_size} / ${row.max_size}`;
    },
  },
  {
    title: t('diagnostics.cacheUsage'),
    key: 'usage',
    render(row: CacheStatsInfo) {
      return `${row.usage_percent.toFixed(1)}%`;
    },
  },
  {
    title: t('diagnostics.cacheExpired'),
    key: 'expired_count',
  },
  {
    title: t('diagnostics.cacheTTL'),
    key: 'ttl',
  },
];

// Format bytes to human-readable string
function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

// Format ISO date string to localized date/time
function formatDateTime(isoString: string): string {
  try {
    return new Date(isoString).toLocaleString();
  } catch {
    return isoString;
  }
}

async function loadRequestDiagnostics(force = false) {
  if (loadingTabs.value.request || (requestDiagnostics.value && !force)) {
    return;
  }

  try {
    loadingTabs.value.request = true;
    requestDiagnostics.value = await DiagnosticsApi.getRequestDiagnostics();
  } catch (e) {
    console.error("Error loading request diagnostics:", e);
  } finally {
    loadingTabs.value.request = false;
  }
}

async function loadSystemDiagnostics(force = false) {
  if (loadingTabs.value.system || (systemDiagnostics.value && !force)) {
    return;
  }

  try {
    loadingTabs.value.system = true;
    systemDiagnostics.value = await DiagnosticsApi.getSystemDiagnostics();
  } catch (e) {
    console.error("Error loading system diagnostics:", e);
  } finally {
    loadingTabs.value.system = false;
  }
}

async function loadBuildDiagnostics(force = false) {
  if (loadingTabs.value.build || (buildDiagnostics.value && !force)) {
    return;
  }

  loadingTabs.value.build = true;
  const [buildResult, frontendResult] = await Promise.allSettled([
    DiagnosticsApi.getBuildDiagnostics(),
    getFrontendBuildInfo(),
  ]);

  if (buildResult.status === 'fulfilled') {
    buildDiagnostics.value = buildResult.value;
  } else {
    console.error("Error loading build diagnostics:", buildResult.reason);
    buildDiagnostics.value = {
      build_info: {
        version: "unknown",
        commit: "unknown",
        branch: "unknown",
        date: "unknown",
      },
    };
  }

  if (frontendResult.status === 'fulfilled') {
    frontendBuildInfo.value = frontendResult.value;
  } else {
    console.error("Failed to load frontend build info:", frontendResult.reason);
  }

  loadingTabs.value.build = false;
}

async function loadTab(tab: DiagnosticsTab, force = false) {
  switch (tab) {
    case 'request':
      await loadRequestDiagnostics(force);
      break;
    case 'system':
      await loadSystemDiagnostics(force);
      break;
    case 'build':
      await loadBuildDiagnostics(force);
      break;
    default:
      break;
  }
}

async function refreshDiagnostics() {
  await loadTab(activeTab.value, true);
}

watch(activeTab, (tab) => {
  void loadTab(tab);
});

onMounted(() => {
  void loadTab(activeTab.value);
});
</script>

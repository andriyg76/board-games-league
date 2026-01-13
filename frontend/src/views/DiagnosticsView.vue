<template>
  <v-container>
    <v-card v-if="diagnostics">
      <v-card-title>{{ t('diagnostics.title') }}</v-card-title>
      <v-card-text>
        <v-row>
          <v-col cols="12" md="6">
            <v-card class="mb-4">
              <v-card-title>{{ t('diagnostics.backendBuildInfo') }}</v-card-title>
              <v-card-text>
                <p><strong>{{ t('diagnostics.version') }}:</strong> {{ diagnostics.build_info.version }}</p>
                <p><strong>{{ t('diagnostics.commit') }}:</strong> {{ diagnostics.build_info.commit }}</p>
                <p><strong>{{ t('diagnostics.branch') }}:</strong> {{ diagnostics.build_info.branch }}</p>
                <p><strong>{{ t('diagnostics.buildDate') }}:</strong> {{ diagnostics.build_info.date }}</p>
              </v-card-text>
            </v-card>
          </v-col>
          <v-col cols="12" md="6">
            <v-card class="mb-4">
              <v-card-title>{{ t('diagnostics.frontendBuildInfo') }}</v-card-title>
              <v-card-text>
                <p><strong>{{ t('diagnostics.version') }}:</strong> {{ frontendBuildInfo.version }}</p>
                <p><strong>{{ t('diagnostics.commit') }}:</strong> {{ frontendBuildInfo.commit }}</p>
                <p><strong>{{ t('diagnostics.branch') }}:</strong> {{ frontendBuildInfo.branch }}</p>
                <p><strong>{{ t('diagnostics.buildDate') }}:</strong> {{ frontendBuildInfo.date }}</p>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

        <v-row>
          <v-col cols="12" md="6">
            <v-card class="mb-4">
              <v-card-title>{{ t('diagnostics.serverInfo') }}</v-card-title>
              <v-card-text>
                <p><strong>{{ t('diagnostics.hostUrl') }}:</strong> {{ diagnostics.server_info.host_url }}</p>
                <div>
                  <strong>{{ t('diagnostics.trustedOrigins') }}:</strong>
                  <ul v-if="diagnostics.server_info.trusted_origins?.length > 0">
                    <li v-for="origin in diagnostics.server_info.trusted_origins" :key="origin">
                      {{ origin }}
                    </li>
                  </ul>
                  <p v-else class="text-grey">{{ t('diagnostics.noneConfigured') }}</p>
                </div>
              </v-card-text>
            </v-card>
          </v-col>
          <v-col cols="12" md="6">
            <v-card class="mb-4">
              <v-card-title>{{ t('diagnostics.requestInfo') }}</v-card-title>
              <v-card-text>
                <p><strong>{{ t('diagnostics.ipAddress') }}:</strong> {{ diagnostics.request_info.ip_address }}</p>
                <p><strong>{{ t('diagnostics.baseUrl') }}:</strong> {{ diagnostics.request_info.base_url }}</p>
                <p><strong>{{ t('diagnostics.origin') }}:</strong> {{ diagnostics.request_info.origin || t('diagnostics.na') }}</p>
                <p>
                  <strong>{{ t('diagnostics.isTrusted') }}:</strong>
                  <v-chip :color="diagnostics.request_info.is_trusted ? 'success' : 'error'" size="small">
                    {{ diagnostics.request_info.is_trusted ? t('diagnostics.yes') : t('diagnostics.no') }}
                  </v-chip>
                </p>
                <p><strong>{{ t('diagnostics.userAgent') }}:</strong> {{ diagnostics.request_info.user_agent }}</p>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

        <v-card v-if="diagnostics.request_info.geo_info" class="mt-4">
          <v-card-title>{{ t('diagnostics.geoInfo') }}</v-card-title>
          <v-card-text>
            <v-row>
              <v-col cols="12" md="4">
                <p><strong>{{ t('diagnostics.country') }}:</strong> {{ diagnostics.request_info.geo_info.country || t('diagnostics.na') }}</p>
                <p><strong>{{ t('diagnostics.countryCode') }}:</strong> {{ diagnostics.request_info.geo_info.country_code || t('diagnostics.na') }}</p>
              </v-col>
              <v-col cols="12" md="4">
                <p><strong>{{ t('diagnostics.region') }}:</strong> {{ diagnostics.request_info.geo_info.region_name || diagnostics.request_info.geo_info.region || t('diagnostics.na') }}</p>
                <p><strong>{{ t('diagnostics.city') }}:</strong> {{ diagnostics.request_info.geo_info.city || t('diagnostics.na') }}</p>
              </v-col>
              <v-col cols="12" md="4">
                <p><strong>{{ t('diagnostics.timezone') }}:</strong> {{ diagnostics.request_info.geo_info.timezone || t('diagnostics.na') }}</p>
                <p><strong>{{ t('diagnostics.isp') }}:</strong> {{ diagnostics.request_info.geo_info.isp || t('diagnostics.na') }}</p>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>

        <v-card v-if="diagnostics.request_info.resolution_info && Object.keys(diagnostics.request_info.resolution_info).length > 0" class="mt-4">
          <v-card-title>{{ t('diagnostics.resolutionInfo') }}</v-card-title>
          <v-card-text>
            <v-table density="compact">
              <thead>
                <tr>
                  <th class="text-left">{{ t('diagnostics.headerName') }}</th>
                  <th class="text-left">{{ t('diagnostics.headerValue') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(value, key) in diagnostics.request_info.resolution_info" :key="key">
                  <td><code>{{ key }}</code></td>
                  <td><code>{{ value }}</code></td>
                </tr>
              </tbody>
            </v-table>
          </v-card-text>
        </v-card>

        <!-- Go Runtime Information -->
        <v-card v-if="diagnostics.runtime_info" class="mt-4">
          <v-card-title>{{ t('diagnostics.runtimeInfo') }}</v-card-title>
          <v-card-text>
            <v-row>
              <v-col cols="12" md="4">
                <p><strong>{{ t('diagnostics.goVersion') }}:</strong> {{ diagnostics.runtime_info.go_version }}</p>
                <p><strong>{{ t('diagnostics.platform') }}:</strong> {{ diagnostics.runtime_info.goos }}/{{ diagnostics.runtime_info.goarch }}</p>
                <p><strong>{{ t('diagnostics.numCpu') }}:</strong> {{ diagnostics.runtime_info.num_cpu }}</p>
              </v-col>
              <v-col cols="12" md="4">
                <p><strong>{{ t('diagnostics.numGoroutine') }}:</strong> {{ diagnostics.runtime_info.num_goroutine }}</p>
                <p><strong>{{ t('diagnostics.uptime') }}:</strong> {{ diagnostics.runtime_info.uptime }}</p>
                <p><strong>{{ t('diagnostics.startTime') }}:</strong> {{ formatDateTime(diagnostics.runtime_info.start_time) }}</p>
              </v-col>
              <v-col cols="12" md="4">
                <p><strong>{{ t('diagnostics.gcCycles') }}:</strong> {{ diagnostics.runtime_info.memory.num_gc }}</p>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>

        <!-- Memory Statistics -->
        <v-card v-if="diagnostics.runtime_info?.memory" class="mt-4">
          <v-card-title>{{ t('diagnostics.memoryStats') }}</v-card-title>
          <v-card-text>
            <v-table density="compact">
              <thead>
                <tr>
                  <th class="text-left">{{ t('diagnostics.metric') }}</th>
                  <th class="text-right">{{ t('diagnostics.value') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>{{ t('diagnostics.heapAlloc') }}</td>
                  <td class="text-right"><code>{{ formatBytes(diagnostics.runtime_info.memory.heap_alloc_bytes) }}</code></td>
                </tr>
                <tr>
                  <td>{{ t('diagnostics.heapInuse') }}</td>
                  <td class="text-right"><code>{{ formatBytes(diagnostics.runtime_info.memory.heap_inuse_bytes) }}</code></td>
                </tr>
                <tr>
                  <td>{{ t('diagnostics.heapSys') }}</td>
                  <td class="text-right"><code>{{ formatBytes(diagnostics.runtime_info.memory.heap_sys_bytes) }}</code></td>
                </tr>
                <tr>
                  <td>{{ t('diagnostics.alloc') }}</td>
                  <td class="text-right"><code>{{ formatBytes(diagnostics.runtime_info.memory.alloc_bytes) }}</code></td>
                </tr>
                <tr>
                  <td>{{ t('diagnostics.totalAlloc') }}</td>
                  <td class="text-right"><code>{{ formatBytes(diagnostics.runtime_info.memory.total_alloc_bytes) }}</code></td>
                </tr>
                <tr>
                  <td>{{ t('diagnostics.sysMem') }}</td>
                  <td class="text-right"><code>{{ formatBytes(diagnostics.runtime_info.memory.sys_bytes) }}</code></td>
                </tr>
              </tbody>
            </v-table>
          </v-card-text>
        </v-card>

        <!-- Environment Variables -->
        <v-card v-if="diagnostics.environment_vars && diagnostics.environment_vars.length > 0" class="mt-4">
          <v-card-title>
            {{ t('diagnostics.environmentVars') }}
            <v-chip class="ml-2" size="small">{{ diagnostics.environment_vars.length }}</v-chip>
          </v-card-title>
          <v-card-text>
            <v-text-field
              v-model="envFilter"
              :label="t('diagnostics.filterEnvVars')"
              prepend-inner-icon="mdi-magnify"
              clearable
              density="compact"
              class="mb-2"
            ></v-text-field>
            <v-table density="compact" style="max-height: 400px; overflow-y: auto;">
              <thead>
                <tr>
                  <th class="text-left">{{ t('diagnostics.envName') }}</th>
                  <th class="text-left">{{ t('diagnostics.envValue') }}</th>
                  <th class="text-center">{{ t('diagnostics.envMasked') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="envVar in filteredEnvVars" :key="envVar.name">
                  <td><code>{{ envVar.name }}</code></td>
                  <td>
                    <code :class="{ 'text-grey': envVar.masked }">{{ envVar.value }}</code>
                  </td>
                  <td class="text-center">
                    <v-icon v-if="envVar.masked" color="warning" size="small">mdi-eye-off</v-icon>
                  </td>
                </tr>
              </tbody>
            </v-table>
          </v-card-text>
        </v-card>
      </v-card-text>
    </v-card>
    <v-skeleton-loader v-else type="card"></v-skeleton-loader>
  </v-container>
</template>

<script lang="ts" setup>
import DiagnosticsApi, { DiagnosticsResponse, getFrontendBuildInfo, BuildInfo, EnvVarInfo } from "@/api/DiagnosticsApi";
import { ref, computed, onMounted } from "vue";
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const diagnostics = ref<DiagnosticsResponse | null>(null);
const frontendBuildInfo = ref<BuildInfo>({
  version: "unknown",
  commit: "unknown",
  branch: "unknown",
  date: "unknown",
});
const envFilter = ref<string>("");

// Filter environment variables based on search
const filteredEnvVars = computed<EnvVarInfo[]>(() => {
  if (!diagnostics.value?.environment_vars) return [];
  if (!envFilter.value) return diagnostics.value.environment_vars;
  
  const filter = envFilter.value.toLowerCase();
  return diagnostics.value.environment_vars.filter(
    (env) => env.name.toLowerCase().includes(filter) || env.value.toLowerCase().includes(filter)
  );
});

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

onMounted(async () => {
  try {
    diagnostics.value = await DiagnosticsApi.getDiagnostics();
    frontendBuildInfo.value = await getFrontendBuildInfo();
  } catch (e) {
    console.error("Error loading diagnostics:", e);
  }
});
</script>

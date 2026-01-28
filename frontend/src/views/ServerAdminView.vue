<template>
  <div>
    <n-card>
      <template #header>
        <div style="font-size: 1.25rem; font-weight: 500;">Server Administration</div>
      </template>

      <!-- Token Input Section -->
      <n-card style="margin-bottom: 24px;">
        <template #header>
          <div style="font-size: 1rem; font-weight: 500;">Authentication</div>
        </template>
        <n-space vertical>
          <n-input
            v-model:value="token"
            type="password"
            placeholder="Enter admin API token"
            show-password-on="click"
            :disabled="loading"
            @keyup.enter="saveToken"
          />
          <n-space>
            <n-button
              type="primary"
              :loading="loading"
              @click="saveToken"
            >
              Save Token
            </n-button>
            <n-button
              v-if="hasToken"
              @click="clearToken"
            >
              Clear Token
            </n-button>
          </n-space>
          <n-alert v-if="tokenError" type="error" style="margin-top: 8px;">
            {{ tokenError }}
          </n-alert>
        </n-space>
      </n-card>

      <!-- Debug Logging Section -->
      <n-card v-if="hasToken" style="margin-bottom: 24px;">
        <template #header>
          <div style="font-size: 1rem; font-weight: 500;">Debug Logging</div>
        </template>
        <n-space vertical>
          <div>
            <n-space align="center">
              <span>Status:</span>
              <n-tag :type="debugStatus.enabled ? 'success' : 'default'" size="small">
                {{ debugStatus.enabled ? 'Enabled' : 'Disabled' }}
              </n-tag>
              <span v-if="debugStatus.enabled && debugStatus.expires_at">
                (expires: {{ formatDateTime(debugStatus.expires_at) }})
              </span>
            </n-space>
          </div>
          <n-space>
            <n-input-number
              v-model:value="debugDuration"
              :min="1"
              :max="1440"
              :step="5"
              placeholder="Duration (minutes)"
              :disabled="loading || debugStatus.enabled"
              style="width: 200px;"
            />
            <n-button
              type="primary"
              :loading="loading"
              :disabled="!debugDuration || debugStatus.enabled"
              @click="enableDebug"
            >
              Enable Debug Logging
            </n-button>
            <n-button
              :loading="loading"
              :disabled="!debugStatus.enabled"
              @click="disableDebug"
            >
              Disable Debug Logging
            </n-button>
            <n-button
              :loading="loading"
              @click="refreshDebugStatus"
            >
              Refresh Status
            </n-button>
          </n-space>
          <n-alert v-if="debugError" type="error" style="margin-top: 8px;">
            {{ debugError }}
          </n-alert>
        </n-space>
      </n-card>

      <!-- Log Download Section -->
      <n-card v-if="hasToken">
        <template #header>
          <div style="font-size: 1rem; font-weight: 500;">Download Logs</div>
        </template>
        <n-space vertical>
          <div>
            <n-radio-group v-model:value="downloadMode" :disabled="loading">
              <n-space>
                <n-radio value="period">By Period</n-radio>
                <n-radio value="full">Full Logs</n-radio>
              </n-space>
            </n-radio-group>
          </div>

          <!-- Period Selection -->
          <div v-if="downloadMode === 'period'">
            <n-space>
              <span>Duration:</span>
              <n-select
                v-model:value="selectedDuration"
                :options="durationOptions"
                :disabled="loading"
                style="width: 150px;"
              />
            </n-space>
          </div>

          <!-- File Selection -->
          <div>
            <n-checkbox-group v-model:value="selectedFiles" :disabled="loading">
              <n-space>
                <n-checkbox value="server">server.log</n-checkbox>
                <n-checkbox value="debug">debug.log</n-checkbox>
              </n-space>
            </n-checkbox-group>
          </div>

          <n-button
            type="primary"
            :loading="loading"
            :disabled="selectedFiles.length === 0 || (downloadMode === 'period' && !selectedDuration)"
            @click="downloadLogs"
          >
            Download Logs
          </n-button>
          <n-alert v-if="downloadError" type="error" style="margin-top: 8px;">
            {{ downloadError }}
          </n-alert>
        </n-space>
      </n-card>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import {
  NCard,
  NInput,
  NButton,
  NSpace,
  NTag,
  NInputNumber,
  NAlert,
  NRadioGroup,
  NRadio,
  NSelect,
  NCheckboxGroup,
  NCheckbox,
} from 'naive-ui';
import ServerAdminApi, { getStoredToken, setStoredToken, clearStoredToken, DebugLoggingState } from '@/api/ServerAdminApi';

const token = ref<string>('');
const loading = ref(false);
const tokenError = ref<string>('');
const debugStatus = ref<DebugLoggingState>({ enabled: false });
const debugDuration = ref<number | null>(30);
const debugError = ref<string>('');
const downloadMode = ref<'period' | 'full'>('period');
const selectedDuration = ref<number | null>(5);
const selectedFiles = ref<string[]>(['server', 'debug']);
const downloadError = ref<string>('');

const durationOptions = [
  { label: '5 minutes', value: 5 },
  { label: '10 minutes', value: 10 },
  { label: '30 minutes', value: 30 },
  { label: '60 minutes', value: 60 },
];

const hasToken = computed(() => {
  return getStoredToken() !== null;
});

function formatDateTime(isoString: string): string {
  try {
    return new Date(isoString).toLocaleString();
  } catch {
    return isoString;
  }
}

async function saveToken() {
  if (!token.value.trim()) {
    tokenError.value = 'Token cannot be empty';
    return;
  }

  try {
    setStoredToken(token.value.trim());
    tokenError.value = '';
    token.value = '';
    await refreshDebugStatus();
  } catch (error) {
    tokenError.value = error instanceof Error ? error.message : 'Failed to save token';
  }
}

function clearToken() {
  clearStoredToken();
  token.value = '';
  debugStatus.value = { enabled: false };
}

async function enableDebug() {
  if (!debugDuration.value) {
    debugError.value = 'Please specify duration';
    return;
  }

  loading.value = true;
  debugError.value = '';
  try {
    debugStatus.value = await ServerAdminApi.enableDebugLogging(debugDuration.value);
  } catch (error) {
    debugError.value = error instanceof Error ? error.message : 'Failed to enable debug logging';
  } finally {
    loading.value = false;
  }
}

async function disableDebug() {
  loading.value = true;
  debugError.value = '';
  try {
    debugStatus.value = await ServerAdminApi.disableDebugLogging();
  } catch (error) {
    debugError.value = error instanceof Error ? error.message : 'Failed to disable debug logging';
  } finally {
    loading.value = false;
  }
}

async function refreshDebugStatus() {
  if (!hasToken.value) {
    return;
  }

  loading.value = true;
  debugError.value = '';
  try {
    debugStatus.value = await ServerAdminApi.getDebugLoggingStatus();
  } catch (error) {
    debugError.value = error instanceof Error ? error.message : 'Failed to refresh status';
  } finally {
    loading.value = false;
  }
}

async function downloadLogs() {
  if (selectedFiles.value.length === 0) {
    downloadError.value = 'Please select at least one log file';
    return;
  }

  if (downloadMode.value === 'period' && !selectedDuration.value) {
    downloadError.value = 'Please select duration';
    return;
  }

  loading.value = true;
  downloadError.value = '';
  try {
    let blob: Blob;
    let filename: string;

    if (downloadMode.value === 'period') {
      blob = await ServerAdminApi.downloadLogsByPeriod(selectedDuration.value!, selectedFiles.value);
      filename = `logs-${selectedDuration.value}min-${new Date().toISOString().slice(0, 19).replace(/:/g, '')}.txt`;
    } else {
      blob = await ServerAdminApi.downloadFullLogs(selectedFiles.value);
      filename = `logs-full-${new Date().toISOString().slice(0, 19).replace(/:/g, '')}.zip`;
    }

    // Create download link
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  } catch (error) {
    downloadError.value = error instanceof Error ? error.message : 'Failed to download logs';
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  if (hasToken.value) {
    refreshDebugStatus();
  }
});
</script>

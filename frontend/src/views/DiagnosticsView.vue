<template>
  <v-container>
    <v-card v-if="diagnostics">
      <v-card-title>System Diagnostics</v-card-title>
      <v-card-text>
        <v-row>
          <v-col cols="12" md="6">
            <v-card class="mb-4">
              <v-card-title>Backend Build Information</v-card-title>
              <v-card-text>
                <p><strong>Version:</strong> {{ diagnostics.build_info.version }}</p>
                <p><strong>Commit:</strong> {{ diagnostics.build_info.commit }}</p>
                <p><strong>Branch:</strong> {{ diagnostics.build_info.branch }}</p>
                <p><strong>Build Date:</strong> {{ diagnostics.build_info.date }}</p>
              </v-card-text>
            </v-card>
          </v-col>
          <v-col cols="12" md="6">
            <v-card class="mb-4">
              <v-card-title>Frontend Build Information</v-card-title>
              <v-card-text>
                <p><strong>Version:</strong> {{ frontendBuildInfo.version }}</p>
                <p><strong>Commit:</strong> {{ frontendBuildInfo.commit }}</p>
                <p><strong>Branch:</strong> {{ frontendBuildInfo.branch }}</p>
                <p><strong>Build Date:</strong> {{ frontendBuildInfo.date }}</p>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

        <v-row>
          <v-col cols="12" md="6">
            <v-card class="mb-4">
              <v-card-title>Server Information</v-card-title>
              <v-card-text>
                <p><strong>Host URL:</strong> {{ diagnostics.server_info.host_url }}</p>
                <div>
                  <strong>Trusted Origins:</strong>
                  <ul v-if="diagnostics.server_info.trusted_origins.length > 0">
                    <li v-for="origin in diagnostics.server_info.trusted_origins" :key="origin">
                      {{ origin }}
                    </li>
                  </ul>
                  <p v-else class="text-grey">None configured</p>
                </div>
              </v-card-text>
            </v-card>
          </v-col>
          <v-col cols="12" md="6">
            <v-card class="mb-4">
              <v-card-title>Request Information</v-card-title>
              <v-card-text>
                <p><strong>IP Address:</strong> {{ diagnostics.request_info.ip_address }}</p>
                <p><strong>Base URL:</strong> {{ diagnostics.request_info.base_url }}</p>
                <p><strong>Origin:</strong> {{ diagnostics.request_info.origin || 'N/A' }}</p>
                <p>
                  <strong>Is Trusted:</strong>
                  <v-chip :color="diagnostics.request_info.is_trusted ? 'success' : 'error'" size="small">
                    {{ diagnostics.request_info.is_trusted ? 'Yes' : 'No' }}
                  </v-chip>
                </p>
                <p><strong>User Agent:</strong> {{ diagnostics.request_info.user_agent }}</p>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

        <v-card v-if="diagnostics.request_info.geo_info" class="mt-4">
          <v-card-title>Geolocation Information</v-card-title>
          <v-card-text>
            <v-row>
              <v-col cols="12" md="4">
                <p><strong>Country:</strong> {{ diagnostics.request_info.geo_info.country || 'N/A' }}</p>
                <p><strong>Country Code:</strong> {{ diagnostics.request_info.geo_info.country_code || 'N/A' }}</p>
              </v-col>
              <v-col cols="12" md="4">
                <p><strong>Region:</strong> {{ diagnostics.request_info.geo_info.region_name || diagnostics.request_info.geo_info.region || 'N/A' }}</p>
                <p><strong>City:</strong> {{ diagnostics.request_info.geo_info.city || 'N/A' }}</p>
              </v-col>
              <v-col cols="12" md="4">
                <p><strong>Timezone:</strong> {{ diagnostics.request_info.geo_info.timezone || 'N/A' }}</p>
                <p><strong>ISP:</strong> {{ diagnostics.request_info.geo_info.isp || 'N/A' }}</p>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-card-text>
    </v-card>
    <v-skeleton-loader v-else type="card"></v-skeleton-loader>
  </v-container>
</template>

<script lang="ts" setup>
import DiagnosticsApi, { DiagnosticsResponse, getFrontendBuildInfo, BuildInfo } from "@/api/DiagnosticsApi";
import { ref, onMounted } from "vue";

const diagnostics = ref<DiagnosticsResponse | null>(null);
const frontendBuildInfo = ref<BuildInfo>({
  version: "unknown",
  commit: "unknown",
  branch: "unknown",
  date: "unknown",
});

onMounted(async () => {
  try {
    diagnostics.value = await DiagnosticsApi.getDiagnostics();
    frontendBuildInfo.value = await getFrontendBuildInfo();
  } catch (e) {
    console.error("Error loading diagnostics:", e);
  }
});
</script>

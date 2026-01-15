<template>
  <div>
    <h1 style="margin-bottom: 16px;">{{ t('createUser.title') }}</h1>
    <n-alert v-if="message" type="info" style="margin-bottom: 16px;">
      {{ message }}
    </n-alert>
    <div v-if="externalIDs && externalIDs.length > 0">
      <p style="margin-bottom: 16px;">{{ t('createUser.email') }}: {{ externalIDs.join(', ') }}</p>
      <n-button type="primary" @click="createUser">
        {{ t('createUser.create') }}
      </n-button>
    </div>
    <div v-else>
      <p>{{ t('createUser.emailMissing') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { NButton, NAlert } from 'naive-ui';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import UserApi from '../api/UserApi';
import { useErrorHandler } from '@/composables/useErrorHandler';

const { t } = useI18n();
const { handleError, showSuccess } = useErrorHandler();
const externalIDs = ref<string[]>([]);
const message = ref('');
const route = useRoute();

const createUser = async () => {
  try {
    await UserApi.adminCreateUser(externalIDs.value);
    message.value = t('createUser.userCreated');
    showSuccess(t('createUser.userCreated'));
  } catch (error) {
    message.value = t('createUser.createFailed');
    handleError(error, t('createUser.createFailed'));
  }
};

onMounted(() => {
  const ids = (route.query.external_ids || "").toString();
  externalIDs.value = ids ? ids.split(",") : [];
});
</script>

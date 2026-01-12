<template>
  <div>
    <h1>{{ t('createUser.title') }}</h1>
    <p v-if="message">{{ message }}</p>
    <div v-if="externalIDs">
      <p>{{ t('createUser.email') }}: {{ externalIDs }}</p>
      <v-btn @click="createUser">{{ t('createUser.create') }}</v-btn>
    </div>
    <div v-else>
      <p>{{ t('createUser.emailMissing') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import UserApi from '../api/UserApi';

const { t } = useI18n();
const externalIDs = ref([]);
const message = ref('');
const route = useRoute();

const createUser = async () => {
  try {
    await UserApi.adminCreateUser(externalIDs.value);
    message.value = t('createUser.userCreated');
  } catch (_error) {
    message.value = t('createUser.createFailed');
  }
};

onMounted(() => {
  externalIDs.value = (route.query.external_ids || "").split(",");
});
</script>

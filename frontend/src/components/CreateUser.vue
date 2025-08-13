<template>
  <div>
    <h1>Create User</h1>
    <p v-if="message">{{ message }}</p>
    <div v-if="externalIDs">
      <p>Email: {{ externalIDs }}</p>
      <v-btn @click="createUser">Create</v-btn>
    </div>
    <div v-else>
      <p>Email parameter is missing</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import UserApi from '../api/UserApi';

const externalIDs = ref([]);
const message = ref('');
const route = useRoute();

const createUser = async () => {
  try {
    await UserApi.adminCreateUser(externalIDs.value);
    message.value = 'User created successfully';
  } catch (error) {
    message.value = 'Failed to create user';
  }
};

onMounted(() => {
  externalIDs.value = (route.query.external_ids || "").split(",");
});
</script>
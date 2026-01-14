import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import naive from './plugins/naive-ui'
import { i18n } from './i18n'
import {createPinia} from "pinia";

createApp(App)
    .use(naive)
    .use(router)
    .use(i18n)
    .use(createPinia())
    .mount('#app')
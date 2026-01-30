import { createApp } from "vue";
import { createPinia } from 'pinia';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';
import { VueQueryPlugin, QueryClient } from '@tanstack/vue-query';
import "./assets/index.css";
import App from "./App.vue";
import router from "./router";

const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: 0,
            gcTime: 5 * 60 * 1000, // 5 minutes
            refetchOnWindowFocus: true,
        },
    },
});

createApp(App)
    .use(pinia)
    .use(router)
    .use(VueQueryPlugin, { queryClient })
    .mount("#app");

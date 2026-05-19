// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite"

const isDev = process.env.NODE_ENV !== "production"

const domain = process.env.DOMAIN
const ws_domain = process.env.WS_DOMAIN

console.log("Server type:", process.env.SERVER_TYPE)

const nitro = process.env.SERVER_TYPE == "local"
  ? {
    routeRules: {
      "/api/**": { proxy: "http://localhost:4000/api/**" },
      "/pub/**": { proxy: "http://localhost:4000/pub/**" },
    },
  } : {
    routeRules: {
      "/api/**": { proxy: "http://localhost:10000/api/**" },
      "/pub/**": { proxy: "http://localhost:10000/pub/**" },
    },
  }

console.log("nitro", nitro)
console.log(domain)

export default defineNuxtConfig({
  devtools: { enabled: true },

  // Removed: build.transpile vuetify
  // Removed: css vuetify/styles
  // Removed: plugins vuetify.ts
  // Removed: modules vite-plugin-vuetify hook
  // Removed: vite.css.preprocessorOptions scss (was only for vuetify settings.scss)

  css: [
    "~/assets/css/main.css", // Tailwind + shadcn CSS variables live here
  ],

  ssr: false,
  sourcemap: {
    server: false,
    client: false,
  },

  runtimeConfig: {
    public: {
      dev: isDev,
      domain: domain,
      ws_domain: ws_domain,
      baseUrl: process.env.BASE_URL,
    },
  },

  modules: [
    'nuxt-lucide-icons',
  ],
  

  components: [
    { path: '~/app/components/ui', pathPrefix: false },
    '~/app/components',
  ],
  
  vite: {
    plugins: [
      tailwindcss(),
    ],
  },

  nitro,

  lucide: {
    namePrefix: 'Icon',
  },

  app: {
    head: {
      title: "Smart Note",
      titleTemplate: "%s",
      meta: [
        { charset: "utf-8" },
        { name: "viewport", content: "width=device-width, initial-scale=1, viewport-fit=cover" },
      ],
    },
  },
})

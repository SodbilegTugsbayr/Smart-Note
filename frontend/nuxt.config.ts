// https://nuxt.com/docs/api/configuration/nuxt-config
import vuetify, { transformAssetUrls } from "vite-plugin-vuetify"

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
  build: {
    transpile: ["vuetify"],
  },
  css: [
    "vuetify/styles",
  ],
  plugins: ["~/plugins/vuetify.ts"],
  modules: [
    async (_options, nuxt) => {
      nuxt.hooks.hook("vite:extendConfig", (config) => {
        // @ts-expect-error
        config.plugins.push(
          vuetify({
            autoImport: true,
            styles: {
              configFile: "./assets/settings.scss",
            },
          }),
        )
      })
    },
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

  vite: {
    vue: {
      template: {
        transformAssetUrls,
      },
    },
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: '@import "./assets/settings.scss";',
        },
      },
    },
  },

  nitro,

  app: {
    head: {
      title: "My app",
      titleTemplate: "%s %separator %site.name",
      templateParams: {
        site: {
          name: 'My app',
          url: 'https://example.com',
        },
        separator: '-',
      },
      meta: [
        { charset: "utf-8" },
        { name: "viewport", content: "width=device-width, initial-scale=1, viewport-fit=cover" },
      ],
    },
  },
})

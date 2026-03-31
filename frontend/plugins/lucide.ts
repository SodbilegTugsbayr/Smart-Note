import * as icons from 'lucide-vue-next'
import type { Component } from 'vue'

export default defineNuxtPlugin((nuxtApp) => {
  Object.entries(icons).forEach(([name, component]) => {
    if (
      typeof component === 'object' ||
      typeof component === 'function'
    ) {
      if (name === 'createLucideIcon' || name === 'Icon') return

      nuxtApp.vueApp.component(
        `${name}Icon`,
        component as Component
      )
    }
  })
})
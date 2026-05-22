<script setup>
const route = useRoute()
const user = useUser()

defineProps({
  title: { type: String, required: true },
  description: { type: String, default: "" },
})

if (process.client && user.value?.role && user.value.role !== "admin") {
  navigateTo("/")
}

function navClass(path) {
  const active = path === "/admin" ? route.path === "/admin" : route.path.startsWith(path)
  return [
    "inline-flex items-center gap-2 rounded-lg px-3 py-2 text-sm transition-colors",
    active
      ? "bg-indigo-50 text-indigo-700 border border-indigo-100"
      : "text-muted-foreground hover:text-foreground hover:bg-slate-100 border border-transparent",
  ]
}
</script>

<template>
  <NavTopBar />
  <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
    <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
      <div>
        <h1 class="font-heading text-3xl text-foreground">{{ title }}</h1>
        <p v-if="description" class="text-sm text-muted-foreground mt-1">{{ description }}</p>
      </div>

      <nav class="flex items-center gap-2 overflow-x-auto pb-1">
        <NuxtLink to="/admin" :class="navClass('/admin')">
          <LayoutDashboardIcon class="w-4 h-4" />
          Ерөнхий
        </NuxtLink>
        <NuxtLink to="/admin/users" :class="navClass('/admin/users')">
          <UsersIcon class="w-4 h-4" />
          Хэрэглэгчид
        </NuxtLink>
        <NuxtLink to="/admin/courses" :class="navClass('/admin/courses')">
          <BookOpenIcon class="w-4 h-4" />
          Хичээлүүд
        </NuxtLink>
        <NuxtLink to="/admin/notes" :class="navClass('/admin/notes')">
          <FileTextIcon class="w-4 h-4" />
          Тэмдэглэлүүд
        </NuxtLink>
      </nav>
    </div>

    <slot />
  </main>
</template>

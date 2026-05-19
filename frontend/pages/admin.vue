<script setup>
const page = ref(1)
const { data, pending, refresh } = await useFetch(() => `/api/users?page=${page.value}&size=25`)

const users = computed(() => data.value?.items || [])
const total = computed(() => data.value?.total || 0)

async function removeUser(user) {
  await $fetch(`/api/users/${user.id}`, { method: "DELETE" })
  await refresh()
}
</script>

<template>
  <NavTopBar />
  <main class="max-w-5xl mx-auto px-4 sm:px-6 py-8 space-y-6">
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="font-heading text-2xl text-foreground">Админ удирдлага</h1>
        <p class="text-sm text-muted-foreground mt-1">Нийт {{ total }} хэрэглэгч</p>
      </div>
      <button
        class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2"
        @click="refresh"
      >
        <RefreshCwIcon class="w-4 h-4" />
        Шинэчлэх
      </button>
    </div>

    <div class="glass-card rounded-xl overflow-hidden">
      <div v-if="pending" class="p-8 flex items-center justify-center">
        <Loader2Icon class="w-6 h-6 text-indigo-400 animate-spin" />
      </div>

      <table v-else class="w-full text-sm">
        <thead class="bg-white/5 text-muted-foreground">
          <tr>
            <th class="text-left font-medium px-4 py-3">ID</th>
            <th class="text-left font-medium px-4 py-3">Нэр</th>
            <th class="text-left font-medium px-4 py-3">Имэйл</th>
            <th class="text-left font-medium px-4 py-3">Эрх</th>
            <th class="text-right font-medium px-4 py-3">Үйлдэл</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id" class="border-t border-white/5">
            <td class="px-4 py-3 text-muted-foreground">{{ user.id }}</td>
            <td class="px-4 py-3 text-foreground">
              {{ [user.first_name, user.last_name].filter(Boolean).join(" ") || "-" }}
            </td>
            <td class="px-4 py-3 text-muted-foreground">{{ user.email }}</td>
            <td class="px-4 py-3">
              <span class="px-2 py-1 rounded-lg bg-indigo-500/10 text-indigo-300 text-xs">
                {{ user.role }}
              </span>
            </td>
            <td class="px-4 py-3 text-right">
              <button
                class="text-red-400 hover:text-red-300 disabled:opacity-40"
                :disabled="user.role === 'admin'"
                @click="removeUser(user)"
              >
                <Trash2Icon class="w-4 h-4" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </main>
</template>

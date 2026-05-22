<script setup>
const page = ref(1)
const { data, pending, refresh } = await useFetch(() => `/api/users?page=${page.value}&size=25`)

const users = computed(() => data.value?.items || [])
const total = computed(() => data.value?.total || 0)
const totalPages = computed(() => Math.max(Math.ceil(total.value / 25), 1))

async function removeUser(user) {
  await $fetch(`/api/users/${user.id}`, { method: "DELETE" })
  await refresh()
}

function roleLabel(role) {
  if (role === "admin") return "Админ"
  return "Хэрэглэгч"
}
</script>

<template>
  <AdminShell title="Хэрэглэгчид" :description="`Нийт ${total} хэрэглэгч`">
    <div class="flex items-center justify-between gap-4">
      <div class="text-sm text-muted-foreground">Хуудас {{ page }} / {{ totalPages }}</div>
      <button
        class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2"
        @click="refresh"
      >
        <RefreshCwIcon class="w-4 h-4" />
        Шинэчлэх
      </button>
    </div>

    <div class="glass-card rounded-xl overflow-x-auto">
      <div v-if="pending" class="p-8 flex items-center justify-center">
        <Loader2Icon class="w-6 h-6 text-indigo-600 animate-spin" />
      </div>

      <table v-else class="w-full min-w-[720px] text-sm">
        <thead class="bg-slate-100 text-muted-foreground">
          <tr>
            <th class="text-left font-medium px-4 py-3">ID</th>
            <th class="text-left font-medium px-4 py-3">Нэр</th>
            <th class="text-left font-medium px-4 py-3">Имэйл</th>
            <th class="text-left font-medium px-4 py-3">Эрх</th>
            <th class="text-right font-medium px-4 py-3">Үйлдэл</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id" class="border-t border-slate-200">
            <td class="px-4 py-3 text-muted-foreground">{{ user.id }}</td>
            <td class="px-4 py-3 text-foreground">
              {{ [user.firstname, user.lastname].filter(Boolean).join(" ") || "-" }}
            </td>
            <td class="px-4 py-3 text-muted-foreground">{{ user.email }}</td>
            <td class="px-4 py-3">
              <span class="px-2 py-1 rounded-lg bg-indigo-500/10 text-indigo-700 text-xs">
                {{ roleLabel(user.role) }}
              </span>
            </td>
            <td class="px-4 py-3 text-right">
              <button
                class="inline-flex items-center justify-center text-red-600 hover:text-red-700 disabled:opacity-40"
                :disabled="user.role === 'admin'"
                @click="removeUser(user)"
              >
                <Trash2Icon class="w-4 h-4" />
              </button>
            </td>
          </tr>
          <tr v-if="users.length === 0">
            <td colspan="5" class="px-4 py-8 text-center text-muted-foreground">Хэрэглэгч олдсонгүй.</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="flex justify-end gap-2">
      <button
        class="glass-card glass-card-hover px-3 py-2 rounded-lg text-sm text-muted-foreground disabled:opacity-40"
        :disabled="page <= 1"
        @click="page--"
      >
        Өмнөх
      </button>
      <button
        class="glass-card glass-card-hover px-3 py-2 rounded-lg text-sm text-muted-foreground disabled:opacity-40"
        :disabled="page >= totalPages"
        @click="page++"
      >
        Дараах
      </button>
    </div>
  </AdminShell>
</template>

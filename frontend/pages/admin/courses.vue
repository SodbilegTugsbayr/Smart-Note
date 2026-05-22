<script setup>
const page = ref(1)
const keyword = ref("")
const pageSize = 25

const { data, status, refresh } = await useFetch("/api/admin/courses", {
  query: {
    page,
    size: pageSize,
    keyword,
    order_by: "created_at desc",
  },
  default: () => ({ items: [], total: 0, users: {} }),
})

const courses = computed(() => data.value?.items || [])
const users = computed(() => data.value?.users || {})
const total = computed(() => data.value?.total || 0)
const totalPages = computed(() => Math.max(Math.ceil(total.value / pageSize), 1))

watch(keyword, () => {
  page.value = 1
})

function owner(course) {
  return users.value?.[course.user_id] || null
}

function ownerLabel(course) {
  const user = owner(course)
  if (!user) return `Хэрэглэгч #${course.user_id}`
  return user.email || [user.firstname, user.lastname].filter(Boolean).join(" ") || `Хэрэглэгч #${course.user_id}`
}

function statusText(course) {
  if (course.status === "completed") return "Дууссан"
  return "Явагдаж байна"
}

function formatDate(value) {
  if (!value) return "-"
  return new Date(value).toLocaleDateString()
}
</script>

<template>
  <AdminShell title="Хичээлүүд" :description="`Нийт ${total} хичээл`">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="relative w-full sm:max-w-sm">
        <SearchIcon class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <input
          v-model="keyword"
          class="w-full rounded-xl border border-slate-200 bg-white pl-9 pr-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
          placeholder="Хичээл хайх"
        />
      </div>
      <button
        class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2"
        @click="refresh"
      >
        <RefreshCwIcon class="w-4 h-4" />
        Шинэчлэх
      </button>
    </div>

    <div class="glass-card rounded-xl overflow-x-auto">
      <div v-if="status === 'pending'" class="p-8 flex items-center justify-center">
        <Loader2Icon class="w-6 h-6 text-indigo-600 animate-spin" />
      </div>

      <table v-else class="w-full min-w-[900px] text-sm">
        <thead class="bg-slate-100 text-muted-foreground">
          <tr>
            <th class="text-left font-medium px-4 py-3">Хичээл</th>
            <th class="text-left font-medium px-4 py-3">Эзэмшигч</th>
            <th class="text-left font-medium px-4 py-3">Төлөв</th>
            <th class="text-left font-medium px-4 py-3">Тэмдэглэл</th>
            <th class="text-left font-medium px-4 py-3">Үүссэн</th>
            <th class="text-right font-medium px-4 py-3">Нээх</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="course in courses" :key="course.id" class="border-t border-slate-200">
            <td class="px-4 py-3">
              <p class="font-medium text-foreground">{{ course.title }}</p>
              <p class="text-xs text-muted-foreground line-clamp-1">{{ course.description || "-" }}</p>
            </td>
            <td class="px-4 py-3 text-muted-foreground">{{ ownerLabel(course) }}</td>
            <td class="px-4 py-3">
              <span class="px-2 py-1 rounded-lg bg-indigo-500/10 text-indigo-700 text-xs">
                {{ statusText(course) }}
              </span>
            </td>
            <td class="px-4 py-3 text-muted-foreground">{{ course.notes?.length || 0 }}</td>
            <td class="px-4 py-3 text-muted-foreground">{{ formatDate(course.created_at) }}</td>
            <td class="px-4 py-3 text-right">
              <NuxtLink
                :to="`/course/${course.id}`"
                class="inline-flex items-center justify-center text-muted-foreground hover:text-indigo-700"
              >
                <ExternalLinkIcon class="w-4 h-4" />
              </NuxtLink>
            </td>
          </tr>
          <tr v-if="courses.length === 0">
            <td colspan="6" class="px-4 py-8 text-center text-muted-foreground">Хичээл олдсонгүй.</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="flex items-center justify-between gap-2">
      <p class="text-sm text-muted-foreground">Хуудас {{ page }} / {{ totalPages }}</p>
      <div class="flex gap-2">
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
    </div>
  </AdminShell>
</template>

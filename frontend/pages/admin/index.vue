<script setup>
const {
  data: stats,
  status,
  refresh,
} = await useFetch("/api/admin/stats", {
  default: () => ({
    totals: { users: 0, courses: 0, notes: 0, quizzes: 0 },
    users: { admins: 0, regular: 0 },
    courses: { completed: 0, in_progress: 0, public: 0 },
    notes: { completed: 0, processing: 0, failed: 0, draft: 0, with_files: 0 },
    recent_courses: [],
    recent_notes: [],
  }),
})

const totals = computed(() => stats.value?.totals || {})
const userStats = computed(() => stats.value?.users || {})
const courseStats = computed(() => stats.value?.courses || {})
const noteStats = computed(() => stats.value?.notes || {})

function percent(value, total) {
  if (!total) return 0
  return Math.round((Number(value || 0) / Number(total || 0)) * 100)
}

function formatDate(value) {
  if (!value) return "-"
  return new Date(value).toLocaleDateString()
}
</script>

<template>
  <AdminShell title="Админ самбар" description="Платформын статистик болон боловсруулалтын төлөв.">
    <div class="flex justify-end">
      <button
        class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2"
        @click="refresh"
      >
        <RefreshCwIcon class="w-4 h-4" />
        Шинэчлэх
      </button>
    </div>

    <div
      v-if="status === 'pending'"
      class="glass-card rounded-xl p-8 flex items-center justify-center"
    >
      <Loader2Icon class="w-6 h-6 text-indigo-600 animate-spin" />
    </div>

    <template v-else>
      <section class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="glass-card rounded-xl p-5">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm text-muted-foreground">Хэрэглэгчид</p>
            <UsersIcon class="w-4 h-4 text-indigo-600" />
          </div>
          <p class="mt-3 text-3xl font-semibold text-foreground">{{ totals.users || 0 }}</p>
          <p class="mt-1 text-xs text-muted-foreground">{{ userStats.admins || 0 }} админ</p>
        </div>
        <div class="glass-card rounded-xl p-5">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm text-muted-foreground">Хичээлүүд</p>
            <BookOpenIcon class="w-4 h-4 text-teal-600" />
          </div>
          <p class="mt-3 text-3xl font-semibold text-foreground">{{ totals.courses || 0 }}</p>
          <p class="mt-1 text-xs text-muted-foreground">
            {{ courseStats.public || 0 }} нийтэд нээлттэй
          </p>
        </div>
        <div class="glass-card rounded-xl p-5">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm text-muted-foreground">Тэмдэглэлүүд</p>
            <FileTextIcon class="w-4 h-4 text-indigo-600" />
          </div>
          <p class="mt-3 text-3xl font-semibold text-foreground">{{ totals.notes || 0 }}</p>
          <p class="mt-1 text-xs text-muted-foreground">{{ noteStats.with_files || 0 }} файлтай</p>
        </div>
        <div class="glass-card rounded-xl p-5">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm text-muted-foreground">Асуултууд</p>
            <ListChecksIcon class="w-4 h-4 text-teal-600" />
          </div>
          <p class="mt-3 text-3xl font-semibold text-foreground">{{ totals.quizzes || 0 }}</p>
          <p class="mt-1 text-xs text-muted-foreground">Үүсгэсэн асуултууд</p>
        </div>
      </section>

      <section class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div class="glass-card rounded-xl p-5 space-y-4">
          <h2 class="text-sm font-semibold text-foreground">Хэрэглэгчийн эрхүүд</h2>
          <div class="space-y-3">
            <div>
              <div class="flex justify-between text-xs text-muted-foreground">
                <span>Хэрэглэгч</span>
                <span>{{ userStats.regular || 0 }}</span>
              </div>
              <div class="mt-1 h-2 bg-slate-100 rounded-full overflow-hidden">
                <div
                  class="h-full bg-indigo-500"
                  :style="{ width: `${percent(userStats.regular, totals.users)}%` }"
                />
              </div>
            </div>
            <div>
              <div class="flex justify-between text-xs text-muted-foreground">
                <span>Админ</span>
                <span>{{ userStats.admins || 0 }}</span>
              </div>
              <div class="mt-1 h-2 bg-slate-100 rounded-full overflow-hidden">
                <div
                  class="h-full bg-teal-500"
                  :style="{ width: `${percent(userStats.admins, totals.users)}%` }"
                />
              </div>
            </div>
          </div>
        </div>

        <div class="glass-card rounded-xl p-5 space-y-4">
          <h2 class="text-sm font-semibold text-foreground">Хичээлийн төлөв</h2>
          <div class="space-y-3">
            <div>
              <div class="flex justify-between text-xs text-muted-foreground">
                <span>Үргэлжилж байгаа</span>
                <span>{{ courseStats.in_progress || 0 }}</span>
              </div>
              <div class="mt-1 h-2 bg-slate-100 rounded-full overflow-hidden">
                <div
                  class="h-full bg-indigo-500"
                  :style="{ width: `${percent(courseStats.in_progress, totals.courses)}%` }"
                />
              </div>
            </div>
            <div>
              <div class="flex justify-between text-xs text-muted-foreground">
                <span>Дууссан</span>
                <span>{{ courseStats.completed || 0 }}</span>
              </div>
              <div class="mt-1 h-2 bg-slate-100 rounded-full overflow-hidden">
                <div
                  class="h-full bg-teal-500"
                  :style="{ width: `${percent(courseStats.completed, totals.courses)}%` }"
                />
              </div>
            </div>
          </div>
        </div>

        <div class="glass-card rounded-xl p-5 space-y-4">
          <h2 class="text-sm font-semibold text-foreground">Тэмдэглэлийн боловсруулалт</h2>
          <div class="divide-y divide-slate-100 text-sm">
            <div class="flex items-center justify-between py-2">
              <span class="text-muted-foreground">Дууссан</span>
              <span class="font-semibold text-teal-700">{{ noteStats.completed || 0 }}</span>
            </div>
            <div class="flex items-center justify-between py-2">
              <span class="text-muted-foreground">Боловсруулж байна</span>
              <span class="font-semibold text-indigo-700">{{ noteStats.processing || 0 }}</span>
            </div>
            <div class="flex items-center justify-between py-2">
              <span class="text-muted-foreground">Алдаа</span>
              <span class="font-semibold text-red-700">{{ noteStats.failed || 0 }}</span>
            </div>
            <div class="flex items-center justify-between py-2">
              <span class="text-muted-foreground">Ноорог</span>
              <span class="font-semibold text-slate-700">{{ noteStats.draft || 0 }}</span>
            </div>
          </div>
        </div>
      </section>

      <section class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div class="glass-card rounded-xl overflow-hidden">
          <div class="px-4 py-3 border-b border-slate-200">
            <h2 class="text-sm font-semibold text-foreground">Сүүлд үүсгэгдсэн хичээлүүд</h2>
          </div>
          <table class="w-full text-sm">
            <tbody>
              <tr
                v-for="course in stats.recent_courses"
                :key="course.id"
                class="border-t border-slate-100 first:border-t-0"
              >
                <td class="px-4 py-3">
                  <NuxtLink
                    :to="`/course/${course.id}`"
                    class="font-medium text-foreground hover:text-indigo-700"
                  >
                    {{ course.title }}
                  </NuxtLink>
                  <p class="text-xs text-muted-foreground">
                    {{ course.notes?.length || 0 }} тэмдэглэл
                  </p>
                </td>
                <td class="px-4 py-3 text-right text-xs text-muted-foreground">
                  {{ formatDate(course.created_at) }}
                </td>
              </tr>
              <tr v-if="!stats.recent_courses?.length">
                <td class="px-4 py-6 text-sm text-muted-foreground">Одоогоор хичээл алга.</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="glass-card rounded-xl overflow-hidden">
          <div class="px-4 py-3 border-b border-slate-200">
            <h2 class="text-sm font-semibold text-foreground">Сүүлд үүсгэгдсэн тэмдэглэлүүд</h2>
          </div>
          <table class="w-full text-sm">
            <tbody>
              <tr
                v-for="note in stats.recent_notes"
                :key="note.id"
                class="border-t border-slate-100 first:border-t-0"
              >
                <td class="px-4 py-3">
                  <p class="font-medium text-foreground">
                    {{ note.title || "Гарчиггүй тэмдэглэл" }}
                  </p>
                  <p class="text-xs text-muted-foreground">Хичээлийн ID: {{ note.course_id }}</p>
                </td>
                <td class="px-4 py-3 text-right text-xs text-muted-foreground">
                  {{ formatDate(note.created_at) }}
                </td>
              </tr>
              <tr v-if="!stats.recent_notes?.length">
                <td class="px-4 py-6 text-sm text-muted-foreground">Одоогоор тэмдэглэл алга.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </template>
  </AdminShell>
</template>

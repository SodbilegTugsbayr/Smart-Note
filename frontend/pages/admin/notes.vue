<script setup>
const page = ref(1)
const keyword = ref("")
const processStatus = ref("")
const pageSize = 25
const reprocessingId = ref(null)
const actionError = ref("")

const { data, status, refresh } = await useFetch("/api/admin/notes", {
  query: {
    page,
    size: pageSize,
    keyword,
    process_status: processStatus,
    order_by: "created_at desc",
  },
  default: () => ({ items: [], total: 0, courses: {}, users: {} }),
})

const notes = computed(() => data.value?.items || [])
const courses = computed(() => data.value?.courses || {})
const users = computed(() => data.value?.users || {})
const total = computed(() => data.value?.total || 0)
const totalPages = computed(() => Math.max(Math.ceil(total.value / pageSize), 1))

watch([keyword, processStatus], () => {
  page.value = 1
})

function courseFor(note) {
  return courses.value?.[note.course_id] || null
}

function ownerFor(note) {
  const course = courseFor(note)
  if (!course) return null
  return users.value?.[course.user_id] || null
}

function noteStatusText(note) {
  if (note.process_status === "completed") return "Дууссан"
  if (note.process_status === "queued") return "Дараалалд"
  if (note.process_status === "processing") return "Боловсруулж байна"
  if (note.process_status === "failed") return "Алдаа"
  return "Ноорог"
}

function noteStatusClass(note) {
  if (note.process_status === "completed")
    return "bg-teal-500/10 text-teal-700 border border-teal-500/20"
  if (note.process_status === "queued")
    return "bg-slate-100 text-slate-700 border border-slate-200"
  if (note.process_status === "failed") return "bg-red-500/10 text-red-700 border border-red-500/20"
  if (note.process_status === "processing")
    return "bg-indigo-500/10 text-indigo-700 border border-indigo-500/20"
  return "bg-slate-100 text-slate-600 border border-slate-200"
}

function formatDate(value) {
  if (!value) return "-"
  return new Date(value).toLocaleDateString()
}

function replaceNote(updatedNote) {
  data.value = {
    ...(data.value || {}),
    items: notes.value.map((note) => (note.id === updatedNote.id ? updatedNote : note)),
  }
}

async function reprocessNote(note) {
  if (!note?.has_file || reprocessingId.value) return

  reprocessingId.value = note.id
  actionError.value = ""
  try {
    const updatedNote = await $fetch(`/api/admin/notes/${note.id}/reprocess`, { method: "POST" })
    replaceNote(updatedNote)
    setTimeout(() => refresh(), 1500)
  } catch (err) {
    actionError.value =
      err?.data?.message || err?.data || "Тэмдэглэлийг дахин боловсруулахад алдаа гарлаа"
  } finally {
    reprocessingId.value = null
  }
}
</script>

<template>
  <AdminShell title="Тэмдэглэлүүд" :description="`Нийт ${total} тэмдэглэл`">
    <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
        <div class="relative w-full sm:w-80">
          <SearchIcon
            class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
          />
          <input
            v-model="keyword"
            class="w-full rounded-xl border border-slate-200 bg-white pl-9 pr-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
            placeholder="Тэмдэглэл хайх"
          />
        </div>
        <select
          v-model="processStatus"
          class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
        >
          <option value="">Бүх төлөв</option>
          <option value="completed">Дууссан</option>
          <option value="queued">Дараалалд</option>
          <option value="processing">Боловсруулж байна</option>
          <option value="failed">Алдаа</option>
        </select>
      </div>
      <button
        class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2"
        @click="refresh"
      >
        <RefreshCwIcon class="w-4 h-4" />
        Шинэчлэх
      </button>
    </div>

    <p v-if="actionError" class="text-sm text-red-600">{{ actionError }}</p>

    <div class="glass-card rounded-xl overflow-x-auto">
      <div v-if="status === 'pending'" class="p-8 flex items-center justify-center">
        <Loader2Icon class="w-6 h-6 text-indigo-600 animate-spin" />
      </div>

      <table v-else class="w-full min-w-[1040px] text-sm">
        <thead class="bg-slate-100 text-muted-foreground">
          <tr>
            <th class="text-left font-medium px-4 py-3">ID</th>
            <th class="text-left font-medium px-4 py-3">Тэмдэглэл</th>
            <th class="text-left font-medium px-4 py-3">Хичээл</th>
            <th class="text-left font-medium px-4 py-3">Хэрэглэгч</th>
            <th class="text-left font-medium px-4 py-3">Төлөв</th>
            <th class="text-left font-medium px-4 py-3">Эх сурвалж</th>
            <th class="text-left font-medium px-4 py-3">Үүссэн</th>
            <th class="text-right font-medium px-4 py-3">Үйлдэл</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="note in notes" :key="note.id" class="border-t border-slate-200">
            <td class="px-4 py-3">
              <p class="font-medium text-foreground">{{ note.id }}</p>
            </td>
            <td class="px-4 py-3">
              <p class="font-medium text-foreground">{{ note.title || "Гарчиггүй тэмдэглэл" }}</p>
            </td>
            <td class="px-4 py-3">
              <NuxtLink
                v-if="courseFor(note)"
                :to="`/course/${note.course_id}`"
                class="text-foreground hover:text-indigo-700"
              >
                {{ courseFor(note).title }}
              </NuxtLink>
              <span v-else class="text-muted-foreground">Хичээл #{{ note.course_id }}</span>
            </td>
            <td class="px-4 py-3 text-muted-foreground">
              {{ ownerFor(note)?.email || "-" }}
            </td>
            <td class="px-4 py-3">
              <span class="text-xs px-2 py-0.5 rounded-full" :class="noteStatusClass(note)">
                {{ noteStatusText(note) }}
              </span>
            </td>
            <td class="px-4 py-3 text-muted-foreground">
              {{ note.has_file ? "Файл" : "Гараар" }}
            </td>
            <td class="px-4 py-3 text-muted-foreground">{{ formatDate(note.created_at) }}</td>
            <td class="px-4 py-3 text-right">
              <button
                class="inline-flex items-center justify-center gap-2 rounded-lg px-3 py-1.5 text-xs text-muted-foreground hover:text-indigo-700 hover:bg-indigo-50 disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-muted-foreground"
                :disabled="!note.has_file || !!reprocessingId"
                @click="reprocessNote(note)"
              >
                <Loader2Icon v-if="reprocessingId === note.id" class="w-4 h-4 animate-spin" />
                <RotateCwIcon v-else class="w-4 h-4" />
                Дахин боловсруулах
              </button>
            </td>
          </tr>
          <tr v-if="notes.length === 0">
            <td colspan="7" class="px-4 py-8 text-center text-muted-foreground">
              Тэмдэглэл олдсонгүй.
            </td>
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

<script setup>
const props = defineProps({
  course: { type: Object, required: true },
  activeNoteId: { type: [String, Number], default: null },
})
const emit = defineEmits(["update"])

const processing = ref(false)
const saving = ref(false)
const progress = ref(0)

const activeNote = computed(() => {
  const notes = props.course.notes || []
  return notes.find((note) => note.id === props.activeNoteId) || notes[0] || null
})
const content = ref("")

watch(
  activeNote,
  (note) => {
    content.value = note?.summary || ""
  },
  { immediate: true },
)

async function handleSave() {
  if (!activeNote.value) return

  saving.value = true
  try {
    const savedNote = await $fetch(`/api/notes/${activeNote.value.id}`, {
      method: "PATCH",
      body: { summary: content.value },
    })
    content.value = savedNote.summary || ""
    emit("update", {
      ...props.course,
      notes: (props.course.notes || []).map((note) =>
        note.id === savedNote.id ? savedNote : note,
      ),
    })
  } finally {
    saving.value = false
  }
}

async function handleAIProcess() {
  processing.value = true
  progress.value = 0

  const interval = setInterval(() => {
    progress.value = Math.min(progress.value + Math.random() * 15, 90)
  }, 500)

  try {
    const result = await $fetch("/api/ai/process-notes", {
      method: "POST",
      body: { course_id: props.course.id },
    })
    clearInterval(interval)
    progress.value = 100
    const updatedCourse = result?.course || { ...props.course }
    const updatedActiveNote =
      (updatedCourse.notes || []).find((note) => note.id === activeNote.value?.id) ||
      updatedCourse.notes?.[0]

    if (updatedActiveNote) {
      content.value = updatedActiveNote.summary || ""
    }
    emit("update", updatedCourse)
  } catch {
    clearInterval(interval)
  } finally {
    processing.value = false
  }
}

function handleFileAttach() {
  const input = document.createElement("input")
  input.type = "file"
  input.onchange = async (e) => {
    const file = e.target.files[0]
    if (!file) return
    const formData = new FormData()
    formData.append("file", file)
    const { file_url } = await $fetch("/api/upload", { method: "POST", body: formData })
    content.value += `\n[${file.name}](${file_url})`
  }
  input.click()
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-2 flex-wrap">
      <button
        @click="handleAIProcess"
        :disabled="processing"
        class="gradient-indigo text-white px-4 py-2 rounded-xl text-sm font-medium flex items-center gap-2 hover:opacity-90 transition-opacity disabled:opacity-50"
      >
        <Loader2Icon v-if="processing" class="w-4 h-4 animate-spin" />
        <SparklesIcon v-else class="w-4 h-4" />
        AI боловсруулах
      </button>
      <button
        @click="handleFileAttach"
        class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2 transition-colors"
      >
        <PaperclipIcon class="w-4 h-4" />
        Файл хавсаргах
      </button>
      <button
        @click="handleSave"
        :disabled="saving || !activeNote"
        class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2 transition-colors ml-auto"
      >
        <Loader2Icon v-if="saving" class="w-4 h-4 animate-spin" />
        <SaveIcon v-else class="w-4 h-4" />
        Хадгалах
      </button>
    </div>

    <!-- Progress bar -->
    <div v-if="processing" class="glass-card rounded-xl p-4 space-y-3">
      <div class="flex items-center justify-between">
        <span class="text-sm text-muted-foreground">AI боловсруулж байна...</span>
        <span class="text-sm font-medium text-indigo-600">{{ Math.round(progress) }}%</span>
      </div>
      <div class="w-full h-2 bg-slate-100 rounded-full overflow-hidden">
        <div
          class="h-full gradient-indigo rounded-full transition-all duration-500"
          :style="{ width: `${progress}%` }"
        />
      </div>
    </div>

    <!-- Selected note content -->
    <div v-if="activeNote" class="glass-card rounded-xl p-5 space-y-4">
      <div class="flex items-start justify-between gap-2">
        <h3 class="text-sm font-medium text-foreground">
          {{ activeNote.title || "Гарчиггүй тэмдэглэл" }}
        </h3>
        <span
          class="text-xs px-2 py-0.5 rounded-full flex-shrink-0"
          :class="
            activeNote.process_status === 'completed'
              ? 'bg-teal-500/10 text-teal-700 border border-teal-500/20'
              : 'bg-indigo-500/10 text-indigo-700 border border-indigo-500/20'
          "
        >
          {{ activeNote.process_status === "completed" ? "Дууссан" : "Боловсруулж байна" }}
        </span>
      </div>

      <textarea
        v-model="content"
        placeholder="Тэмдэглэл бичих..."
        class="w-full bg-slate-50 border border-slate-200 rounded-xl p-4 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-indigo-500/20 min-h-[240px] resize-y"
      />

      <!-- Key concepts -->
      <div v-if="activeNote.key_concepts?.length" class="space-y-2">
        <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Түлхүүр ойлголтууд
        </p>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <div
            v-for="kc in activeNote.key_concepts"
            :key="kc.concept"
            class="bg-slate-50 border border-slate-200 rounded-lg px-3 py-2"
          >
            <p class="text-xs font-medium text-indigo-600 mb-0.5">{{ kc.concept }}</p>
            <p class="text-xs text-muted-foreground">{{ kc.definition }}</p>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="glass-card rounded-xl p-8 text-center">
      <p class="text-muted-foreground text-sm">Тэмдэглэл сонгоно уу.</p>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  course: { type: Object, required: true },
  activeNoteId: { type: [String, Number], default: null },
})
const emit = defineEmits(["update"])

const uploading = ref(false)
const saving = ref(false)
const progress = ref(0)
const errorMessage = ref("")

const activeNote = computed(() => {
  const notes = props.course.notes || []
  return notes.find((note) => note.id === props.activeNoteId) || notes[0] || null
})
const title = ref("")
const content = ref("")
const canAttachFile = computed(
  () => activeNote.value && !activeNote.value.is_from_book && !activeNote.value.has_file,
)

watch(
  activeNote,
  (note) => {
    title.value = note?.title || ""
    content.value = note?.summary || ""
    errorMessage.value = ""
  },
  { immediate: true },
)

async function handleSave() {
  if (!activeNote.value) return
  const cleanTitle = title.value.trim()
  if (!cleanTitle) {
    errorMessage.value = "Гарчиг оруулна уу"
    return
  }

  saving.value = true
  errorMessage.value = ""
  try {
    const savedNote = await $fetch(`/api/notes/${activeNote.value.id}`, {
      method: "PATCH",
      body: { title: cleanTitle, summary: content.value },
    })
    title.value = savedNote.title || ""
    content.value = savedNote.summary || ""
    emit("update", {
      ...props.course,
      notes: (props.course.notes || []).map((note) =>
        note.id === savedNote.id ? savedNote : note,
      ),
    })
  } catch (err) {
    errorMessage.value = err?.data?.message || "Тэмдэглэл хадгалахад алдаа гарлаа"
  } finally {
    saving.value = false
  }
}

function handleFileAttach() {
  if (!canAttachFile.value || uploading.value) return

  const input = document.createElement("input")
  input.type = "file"
  input.accept = ".pdf,.png,.jpg,.jpeg"
  input.onchange = async (e) => {
    const file = e.target.files?.[0]
    if (!file) return

    const targetNoteId = activeNote.value?.id
    if (!targetNoteId) return

    const formData = new FormData()
    formData.append("file", file)

    uploading.value = true
    progress.value = 0
    errorMessage.value = ""
    const interval = setInterval(() => {
      progress.value = Math.min(progress.value + Math.random() * 12, 90)
    }, 500)

    try {
      const savedNote = await $fetch(`/api/notes/${targetNoteId}/file`, {
        method: "POST",
        body: formData,
      })
      clearInterval(interval)
      progress.value = 100

      if (activeNote.value?.id === savedNote.id) {
        title.value = savedNote.title || ""
        content.value = savedNote.summary || ""
      }
      emit("update", {
        ...props.course,
        notes: (props.course.notes || []).map((note) =>
          note.id === savedNote.id ? savedNote : note,
        ),
      })
    } catch (err) {
      clearInterval(interval)
      errorMessage.value = err?.data?.message || "Файл боловсруулахад алдаа гарлаа"
    } finally {
      uploading.value = false
    }
  }
  input.click()
}

function noteStatusText(note) {
  if (!note) return ""
  if (note.process_status === "completed") return "Дууссан"
  if (note.process_status === "processing") return "Боловсруулж байна"
  if (note.process_status === "failed") return "Алдаа"
  return "Ноорог"
}

function noteStatusClass(note) {
  if (note?.process_status === "completed") {
    return "bg-teal-500/10 text-teal-700 border border-teal-500/20"
  }
  if (note?.process_status === "failed") {
    return "bg-red-500/10 text-red-700 border border-red-500/20"
  }
  if (note?.process_status === "processing") {
    return "bg-indigo-500/10 text-indigo-700 border border-indigo-500/20"
  }
  return "bg-slate-100 text-slate-600 border border-slate-200"
}
</script>

<template>
  <div class="space-y-4">
    <ClientOnly>
      <Teleport to="#course-notes-actions">
        <div class="flex flex-wrap items-center justify-end gap-2">
          <button
            v-if="canAttachFile"
            @click="handleFileAttach"
            :disabled="uploading"
            class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2 transition-colors"
          >
            <Loader2Icon v-if="uploading" class="w-4 h-4 animate-spin" />
            <PaperclipIcon v-else class="w-4 h-4" />
            Файл хавсаргах
          </button>
          <button
            @click="handleSave"
            :disabled="saving || uploading || !activeNote"
            class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2 transition-colors"
          >
            <Loader2Icon v-if="saving" class="w-4 h-4 animate-spin" />
            <SaveIcon v-else class="w-4 h-4" />
            Хадгалах
          </button>
        </div>
      </Teleport>
    </ClientOnly>

    <p v-if="errorMessage" class="text-sm text-red-600">{{ errorMessage }}</p>

    <!-- Progress bar -->
    <div v-if="uploading" class="glass-card rounded-xl p-4 space-y-3">
      <div class="flex items-center justify-between">
        <span class="text-sm text-muted-foreground">
          Файл уншиж, тэмдэглэл үүсгэж байна...
        </span>
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
        <p
          class="min-w-0 flex-1 bg-transparent text-sm font-medium text-foreground placeholder:text-muted-foreground focus:outline-none"
        >
          {{ title }}
        </p>
        <span
          class="text-xs px-2 py-0.5 rounded-full flex-shrink-0"
          :class="noteStatusClass(activeNote)"
        >
          {{ noteStatusText(activeNote) }}
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

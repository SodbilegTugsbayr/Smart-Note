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
const progressByNoteId = ref({})

const activeNote = computed(() => {
  const notes = props.course.notes || []
  return notes.find((note) => note.id === props.activeNoteId) || notes[0] || null
})
const title = ref("")
const content = ref("")
const activeProgress = computed(() =>
  activeNote.value ? progressByNoteId.value[activeNote.value.id] : null,
)
const isActiveNoteProcessing = computed(() => activeNote.value?.process_status === "processing")
const isActiveProgressRunning = computed(
  () => activeProgress.value && !["completed", "failed"].includes(activeProgress.value.stage),
)
const showProcessProgress = computed(
  () => uploading.value || isActiveNoteProcessing.value || isActiveProgressRunning.value,
)
const displayProgress = computed(() => {
  const value = activeProgress.value?.progress ?? (uploading.value ? progress.value : 25)
  return Math.max(0, Math.min(100, Math.round(value)))
})
const displayProgressMessage = computed(() => {
  if (activeProgress.value?.message) return activeProgress.value.message
  if (uploading.value) return "Файл сервер рүү илгээж байна..."
  return "Файлаас текст таньж байна..."
})
const canAttachFile = computed(
  () =>
    activeNote.value &&
    !activeNote.value.is_from_book &&
    !activeNote.value.has_file &&
    activeNote.value.process_status !== "processing",
)

watch(
  activeNote,
  (note, previousNote) => {
    title.value = note?.title || ""
    content.value = note?.summary || ""
    if (note?.id !== previousNote?.id) {
      errorMessage.value = ""
    }
  },
  { immediate: true },
)

useQueue((message) => {
  if (!message || message.Type !== "NOTE_PROCESS_PROGRESS") return

  let payload = message.Text
  if (typeof payload === "string") {
    try {
      payload = JSON.parse(payload)
    } catch {
      return
    }
  }

  if (Number(payload?.course_id) !== Number(props.course.id)) return
  if (!payload?.note_id) return

  setNoteProgress(payload.note_id, {
    stage: payload.stage,
    progress: payload.progress,
    message: payload.message,
  })

  if (payload.note) {
    mergeNote(payload.note)
  }

  if (payload.stage === "failed") {
    errorMessage.value = payload.message || "Файл боловсруулахад алдаа гарлаа"
  }
})

function setNoteProgress(noteId, data) {
  progressByNoteId.value = {
    ...progressByNoteId.value,
    [noteId]: {
      ...(progressByNoteId.value[noteId] || {}),
      ...data,
    },
  }
}

function clearNoteProgress(noteId) {
  const next = { ...progressByNoteId.value }
  delete next[noteId]
  progressByNoteId.value = next
}

function mergeNote(updatedNote) {
  if (!updatedNote?.id) return

  let found = false
  const notes = (props.course.notes || []).map((note) => {
    if (note.id !== updatedNote.id) return note

    found = true
    return { ...note, ...updatedNote }
  })
  if (!found) {
    notes.push(updatedNote)
  }

  emit("update", useCourseWithSyncedProgress(props.course, notes))

  if (activeNote.value?.id === updatedNote.id) {
    title.value = updatedNote.title || ""
    content.value = updatedNote.summary || ""
  }
}

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
    mergeNote(savedNote)
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
    progress.value = 5
    errorMessage.value = ""
    setNoteProgress(targetNoteId, {
      stage: "uploading",
      progress: 5,
      message: "Файл сервер рүү илгээж байна...",
    })

    try {
      const savedNote = await $fetch(`/api/notes/${targetNoteId}/file`, {
        method: "POST",
        body: formData,
      })
      progress.value = 12
      setNoteProgress(savedNote.id, {
        stage: "started",
        progress: 12,
        message: "Файл хадгалагдлаа. AI боловсруулалт эхэллээ...",
      })

      mergeNote(savedNote)
    } catch (err) {
      clearNoteProgress(targetNoteId)
      errorMessage.value = err?.data?.message || "Файл боловсруулахад алдаа гарлаа"
    } finally {
      uploading.value = false
    }
  }
  input.click()
}

function noteStatusText(note) {
  if (!note) return ""
  if (note.process_status === "processing") return "Боловсруулж байна"
  if (note.process_status === "failed") return "Алдаа"
  if (note.status === "completed") return "Дууссан"
  if (note.process_status === "completed") return "Дуусаагүй"
  return "Ноорог"
}

function noteStatusClass(note) {
  if (note?.process_status === "processing") {
    return "bg-indigo-500/10 text-indigo-700 border border-indigo-500/20"
  }
  if (note?.process_status === "failed") {
    return "bg-red-500/10 text-red-700 border border-red-500/20"
  }
  if (note?.status === "completed") {
    return "bg-teal-500/10 text-teal-700 border border-teal-500/20"
  }
  if (note?.process_status === "completed") {
    return "bg-amber-500/10 text-amber-700 border border-amber-500/20"
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
            :disabled="uploading || isActiveNoteProcessing"
            class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2 transition-colors"
          >
            <Loader2Icon v-if="uploading" class="w-4 h-4 animate-spin" />
            <PaperclipIcon v-else class="w-4 h-4" />
            Файл хавсаргах
          </button>
          <button
            @click="handleSave"
            :disabled="saving || uploading || isActiveNoteProcessing || !activeNote"
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

      <div class="relative">
        <div
          v-if="showProcessProgress"
          class="rounded-lg border border-indigo-500/20 bg-white/95 px-3 py-2.5 shadow-sm backdrop-blur"
          role="status"
          aria-live="polite"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="min-w-0 flex items-center gap-2">
              <Loader2Icon class="h-4 w-4 flex-shrink-0 animate-spin text-indigo-600" />
              <span class="truncate text-sm text-slate-700">{{ displayProgressMessage }}</span>
            </div>
            <span class="flex-shrink-0 text-sm font-medium text-indigo-600">
              {{ displayProgress }}%
            </span>
          </div>
          <div class="mt-2 h-2 w-full overflow-hidden rounded-full bg-slate-100">
            <div
              class="h-full rounded-full gradient-indigo transition-all duration-500"
              :style="{ width: `${displayProgress}%` }"
            />
          </div>
        </div>
        <textarea
          v-else
          v-model="content"
          :readonly="showProcessProgress"
          placeholder="Тэмдэглэл бичих..."
          class="block w-full bg-slate-50 border border-slate-200 rounded-xl p-4 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-indigo-500/20 min-h-[240px] resize-y transition-colors"
          :class="showProcessProgress ? 'pb-28 cursor-progress' : ''"
        />
      </div>

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

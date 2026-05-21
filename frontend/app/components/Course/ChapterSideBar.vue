<script setup>
const props = defineProps({
  courseId: { type: [String, Number], required: true },
  notes: { type: Array, default: () => [] },
  activeNoteId: { type: [String, Number], default: null },
})

const emit = defineEmits(["note-select", "note-created"])

const adding = ref(false)
const errorMessage = ref("")

const sortedNotes = computed(() =>
  [...props.notes].sort((a, b) => Number(a.id || 0) - Number(b.id || 0)),
)

async function handleAddNote() {
  adding.value = true
  errorMessage.value = ""

  try {
    const nextNumber = props.notes.length + 1
    const savedNote = await $fetch(`/api/courses/${props.courseId}/notes`, {
      method: "POST",
      body: { title: `Шинэ тэмдэглэл ${nextNumber}` },
    })
    emit("note-created", savedNote)
    emit("note-select", savedNote.id)
  } catch (err) {
    errorMessage.value = err?.data?.message || "Тэмдэглэл нэмэхэд алдаа гарлаа"
  } finally {
    adding.value = false
  }
}
</script>

<template>
  <div class="p-3">
    <div class="flex items-center justify-between gap-2 px-2 mb-3">
      <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">
        Тэмдэглэл
      </p>
      <button
        @click="handleAddNote"
        :disabled="adding"
        class="w-7 h-7 rounded-lg text-muted-foreground hover:text-indigo-700 hover:bg-indigo-500/10 transition-colors flex items-center justify-center disabled:opacity-50"
        title="Тэмдэглэл нэмэх"
      >
        <Loader2Icon v-if="adding" class="w-3.5 h-3.5 animate-spin" />
        <PlusIcon v-else class="w-3.5 h-3.5" />
      </button>
    </div>

    <p v-if="errorMessage" class="text-xs text-red-600 px-2 mb-2">{{ errorMessage }}</p>

    <div v-if="sortedNotes.length === 0" class="text-center py-4">
      <p class="text-xs text-muted-foreground">Тэмдэглэл байхгүй</p>
    </div>

    <div v-else class="space-y-1">
      <button
        v-for="note in sortedNotes"
        :key="note.id"
        @click="emit('note-select', note.id)"
        class="w-full flex items-center gap-2 px-2 py-2 rounded-lg text-sm transition-colors text-left"
        :class="
          activeNoteId === note.id
            ? 'bg-indigo-500/10 text-indigo-700'
            : 'text-foreground/80 hover:bg-slate-100'
        "
      >
        <FileTextIcon class="w-3.5 h-3.5 flex-shrink-0" />
        <span class="truncate flex-1 text-xs">{{ note.title || "Гарчиггүй тэмдэглэл" }}</span>
      </button>
    </div>
  </div>
</template>

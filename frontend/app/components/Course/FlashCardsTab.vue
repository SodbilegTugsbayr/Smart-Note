<script setup>
const props = defineProps({
  course: { type: Object, required: true },
  activeNoteId: { type: [String, Number], default: null },
})
const emit = defineEmits(["update"])

const showAdd = ref(false)
const newTerm = ref("")
const newDef = ref("")
const generating = ref(false)

const activeNote = computed(() => {
  const notes = props.course.notes || []
  return notes.find((note) => note.id === props.activeNoteId) || notes[0] || null
})

const cards = computed(() => {
  return (activeNote.value?.flash_cards || []).map((fc) => ({
    ...fc,
    id: `${activeNote.value.id}-${fc.question}`,
  }))
})

watch(
  () => props.activeNoteId,
  () => {
    showAdd.value = false
    newTerm.value = ""
    newDef.value = ""
  },
)

async function handleAIGenerate() {
  if (!activeNote.value) return
  generating.value = true
  try {
    const updatedCourse = await $fetch("/api/ai/generate-flashcards", {
      method: "POST",
      body: { course_id: props.course.id },
    })
    emit("update", updatedCourse)
  } finally {
    generating.value = false
  }
}

async function handleAdd() {
  if (!activeNote.value || !newTerm.value.trim() || !newDef.value.trim()) return
  const savedNote = await $fetch("/api/flashcards", {
    method: "POST",
    body: {
      course_id: props.course.id,
      note_id: activeNote.value.id,
      term: newTerm.value,
      definition: newDef.value,
    },
  })
  emit("update", {
    ...props.course,
    notes: (props.course.notes || []).map((note) => (note.id === savedNote.id ? savedNote : note)),
  })
  newTerm.value = ""
  newDef.value = ""
  showAdd.value = false
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-2 flex-wrap">
      <button
        @click="handleAIGenerate"
        :disabled="generating || !activeNote"
        class="gradient-indigo text-white px-4 py-2 rounded-xl text-sm font-medium flex items-center gap-2 hover:opacity-90 transition-opacity disabled:opacity-50"
      >
        <Loader2Icon v-if="generating" class="w-4 h-4 animate-spin" />
        <SparklesIcon v-else class="w-4 h-4" />
        AI-аар үүсгэх
      </button>
      <button
        @click="showAdd = !showAdd"
        :disabled="!activeNote"
        class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2 transition-colors"
      >
        <PlusIcon class="w-4 h-4" />
        Нэмэх
      </button>
    </div>

    <div v-if="showAdd" class="glass-card rounded-xl p-4 space-y-3">
      <Input
        v-model="newTerm"
        placeholder="Нэр томьёо"
        class="bg-white border-slate-200 text-foreground"
      />
      <Textarea
        v-model="newDef"
        placeholder="Тодорхойлолт"
        class="bg-white border-slate-200 text-foreground h-20"
      />
      <button
        @click="handleAdd"
        :disabled="!activeNote || !newTerm.trim() || !newDef.trim()"
        class="gradient-indigo text-white px-4 py-2 rounded-xl text-sm font-medium hover:opacity-90 disabled:opacity-50"
      >
        Хадгалах →
      </button>
    </div>

    <div v-if="!activeNote" class="text-center py-12">
      <p class="text-muted-foreground text-sm">Тэмдэглэл сонгоно уу.</p>
    </div>

    <div v-else-if="cards.length === 0 && !generating" class="text-center py-12">
      <p class="text-muted-foreground text-sm">
        Энэ тэмдэглэлд флаш карт байхгүй. AI-аар үүсгэх товчийг дарна уу.
      </p>
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <CourseFlipCard
        v-for="card in cards"
        :key="card.id"
        :term="card.question"
        :definition="card.answer"
      />
    </div>
  </div>
</template>

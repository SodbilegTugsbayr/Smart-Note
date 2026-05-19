<script setup>
const props = defineProps({
  course: { type: Object, required: true },
})
const emit = defineEmits(["update"])

const showAdd = ref(false)
const newTerm = ref("")
const newDef = ref("")
const generating = ref(false)

const cards = computed(() => {
  const all = []
  for (const note of props.course.notes || []) {
    for (const fc of note.flash_cards || []) {
      all.push({ ...fc, id: `${note.id}-${fc.question}` })
    }
  }
  return all
})

async function handleAIGenerate() {
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
  if (!newTerm.value.trim() || !newDef.value.trim()) return
  const savedNote = await $fetch("/api/flashcards", {
    method: "POST",
    body: { course_id: props.course.id, term: newTerm.value, definition: newDef.value },
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
        :disabled="generating"
        class="gradient-indigo text-white px-4 py-2 rounded-xl text-sm font-medium flex items-center gap-2 hover:opacity-90 transition-opacity disabled:opacity-50"
      >
        <Loader2Icon v-if="generating" class="w-4 h-4 animate-spin" />
        <SparklesIcon v-else class="w-4 h-4" />
        AI-аар үүсгэх
      </button>
      <button
        @click="showAdd = !showAdd"
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
        :disabled="!newTerm.trim() || !newDef.trim()"
        class="gradient-indigo text-white px-4 py-2 rounded-xl text-sm font-medium hover:opacity-90 disabled:opacity-50"
      >
        Хадгалах →
      </button>
    </div>

    <div v-if="cards.length === 0 && !generating" class="text-center py-12">
      <p class="text-muted-foreground text-sm">
        Флаш карт байхгүй. AI-аар үүсгэх товчийг дарна уу.
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

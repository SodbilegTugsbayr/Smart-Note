<script setup>
const props = defineProps({
  course: { type: Object, required: true },
  activeNoteId: { type: [String, Number], default: null },
  readonly: { type: Boolean, default: false },
})
const emit = defineEmits(["update"])

const showAdd = ref(false)
const newTerm = ref("")
const newDef = ref("")
const generating = ref(false)
const currentIndex = ref(0)
const cardOrder = ref([])

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
const orderedCards = computed(() => {
  const cardById = new Map(cards.value.map((card) => [card.id, card]))
  const ordered = cardOrder.value.map((id) => cardById.get(id)).filter(Boolean)
  const remaining = cards.value.filter((card) => !cardOrder.value.includes(card.id))

  return [...ordered, ...remaining]
})
const currentCard = computed(() => orderedCards.value[currentIndex.value] || null)
const cardIds = computed(() => cards.value.map((card) => card.id).join("|"))

watch(
  () => props.activeNoteId,
  () => {
    showAdd.value = false
    newTerm.value = ""
    newDef.value = ""
  },
)

watch(
  cardIds,
  () => {
    cardOrder.value = cards.value.map((card) => card.id)
    currentIndex.value = 0
  },
  { immediate: true },
)

async function handleAIGenerate() {
  if (props.readonly || !activeNote.value) return
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
  if (props.readonly || !activeNote.value || !newTerm.value.trim() || !newDef.value.trim()) return
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

function previousCard() {
  if (currentIndex.value <= 0) return
  currentIndex.value--
}

function nextCard() {
  if (currentIndex.value >= orderedCards.value.length - 1) return
  currentIndex.value++
}

function shuffleCards() {
  if (orderedCards.value.length < 2) return

  const shuffled = [...orderedCards.value]
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]]
  }

  if (shuffled.every((card, index) => card.id === orderedCards.value[index]?.id)) {
    shuffled.push(shuffled.shift())
  }

  cardOrder.value = shuffled.map((card) => card.id)
  currentIndex.value = 0
}
</script>

<template>
  <div class="space-y-4">
    <ClientOnly>
      <Teleport to="#course-flashcards-actions">
        <div class="flex flex-wrap items-center justify-end gap-2">
          <button
            v-if="cards.length > 1"
            @click="shuffleCards"
            class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2 transition-colors"
          >
            <ShuffleIcon class="w-4 h-4" />
            Холих
          </button>
          <button
            v-if="!readonly"
            @click="showAdd = !showAdd"
            :disabled="!activeNote"
            class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2 transition-colors"
          >
            <PlusIcon class="w-4 h-4" />
            Нэмэх
          </button>
        </div>
      </Teleport>
    </ClientOnly>

    <div v-if="showAdd && !readonly" class="glass-card rounded-xl p-4 space-y-3">
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
        {{
          readonly
            ? "Энэ тэмдэглэлд флаш карт байхгүй."
            : "Энэ тэмдэглэлд флаш карт байхгүй. AI-аар үүсгэх товчийг дарна уу."
        }}
      </p>
    </div>

    <div v-else class="space-y-4">
      <div class="mx-auto max-w-2xl">
        <CourseFlipCard
          v-if="currentCard"
          :key="currentCard.id"
          :term="currentCard.question"
          :definition="currentCard.answer"
        />
      </div>

      <div class="flex flex-wrap items-center justify-center gap-3">
        <button
          @click="previousCard"
          :disabled="currentIndex === 0"
          class="glass-card glass-card-hover h-10 w-10 rounded-lg text-muted-foreground hover:text-foreground flex items-center justify-center transition-colors disabled:opacity-40"
          title="Өмнөх"
        >
          <ChevronLeftIcon class="w-4 h-4" />
        </button>

        <span class="min-w-20 text-center text-sm text-muted-foreground">
          {{ currentIndex + 1 }} / {{ orderedCards.length }}
        </span>

        <button
          @click="nextCard"
          :disabled="currentIndex >= orderedCards.length - 1"
          class="glass-card glass-card-hover h-10 w-10 rounded-lg text-muted-foreground hover:text-foreground flex items-center justify-center transition-colors disabled:opacity-40"
          title="Дараах"
        >
          <ChevronRightIcon class="w-4 h-4" />
        </button>
      </div>
    </div>
  </div>
</template>

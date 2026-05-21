<script setup>
const props = defineProps({
  course: { type: Object, required: true },
  activeNoteId: { type: [String, Number], default: null },
})
const emit = defineEmits(["update"])

const loading = ref(false)
const currentQ = ref(0)
const selectedAnswer = ref(null)
const showResult = ref(false)
const answers = ref([])
const quizDone = ref(false)
const quizzes = ref([])
const errorMessage = ref("")

const activeNote = computed(() => {
  const notes = props.course.notes || []
  return notes.find((note) => note.id === props.activeNoteId) || notes[0] || null
})
const currentQuiz = computed(() => quizzes.value[currentQ.value] || null)
const score = computed(() => answers.value.filter((a) => a.correct).length)
const pct = computed(() =>
  quizzes.value.length > 0 ? Math.round((score.value / quizzes.value.length) * 100) : 0,
)

watch(
  () => props.activeNoteId,
  (noteId) => {
    fetchQuizzes(noteId)
  },
  { immediate: true },
)

async function fetchQuizzes(noteId) {
  resetQuizState()
  quizzes.value = []
  errorMessage.value = ""

  if (!noteId) return

  loading.value = true
  try {
    const result = await $fetch(`/api/notes/${noteId}/quizzes`)
    if (String(props.activeNoteId) !== String(noteId)) return
    quizzes.value = Array.isArray(result) ? result : result?.items || []
  } catch (err) {
    if (String(props.activeNoteId) === String(noteId)) {
      errorMessage.value = err?.data?.message || "Тест уншихад алдаа гарлаа"
    }
  } finally {
    if (String(props.activeNoteId) === String(noteId)) {
      loading.value = false
    }
  }
}

function handleAnswer(answer) {
  if (showResult.value) return
  const isCorrect = answer === currentQuiz.value?.correct_answer
  selectedAnswer.value = answer
  showResult.value = true
  answers.value.push({ quizId: currentQuiz.value?.id, correct: isCorrect })
}

async function handleNext() {
  if (currentQ.value < quizzes.value.length - 1) {
    currentQ.value++
    selectedAnswer.value = null
    showResult.value = false
  } else {
    quizDone.value = true
    if (pct.value >= 90) {
      await $fetch(`/api/courses/${props.course.id}`, {
        method: "PATCH",
        body: { status: "completed", progress: 100 },
      })
      emit("update", { ...props.course, status: "completed", progress: 100 })
    }
  }
}

function resetQuizState() {
  currentQ.value = 0
  selectedAnswer.value = null
  showResult.value = false
  answers.value = []
  quizDone.value = false
}

function resetQuiz() {
  resetQuizState()
}

function optionClass(option) {
  if (!showResult.value) {
    return "hover:border-indigo-500/30"
  }
  if (option === currentQuiz.value?.correct_answer) {
    return "border-teal-500/40 bg-teal-500/10 text-teal-800"
  }
  if (option === selectedAnswer.value) {
    return "border-red-500/40 bg-red-500/10 text-red-700"
  }
  return "opacity-70"
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-2 flex-wrap">
      <button
        v-if="quizDone && quizzes.length"
        @click="resetQuiz"
        class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2 transition-colors"
      >
        <RotateCcwIcon class="w-4 h-4" /> Дахин эхлүүлэх
      </button>
    </div>

    <div v-if="loading" class="glass-card rounded-xl p-8 text-center">
      <Loader2Icon class="w-8 h-8 text-indigo-600 animate-spin mx-auto mb-3" />
      <p class="text-muted-foreground text-sm">Тест уншиж байна...</p>
    </div>

    <div v-else-if="!activeNote" class="text-center py-12">
      <p class="text-muted-foreground text-sm">Тэмдэглэл сонгоно уу.</p>
    </div>

    <div v-else-if="errorMessage" class="text-center py-12">
      <p class="text-red-600 text-sm">{{ errorMessage }}</p>
    </div>

    <div v-else-if="quizzes.length === 0" class="text-center py-12">
      <p class="text-muted-foreground text-sm">
        Тест байхгүй. Сонгосон тэмдэглэлийг боловсруулна уу.
      </p>
    </div>

    <div v-else-if="quizDone" class="glass-card rounded-xl p-8 text-center">
      <template v-if="pct >= 90">
        <div
          class="w-20 h-20 rounded-full gradient-teal flex items-center justify-center mx-auto mb-4"
        >
          <TrophyIcon class="w-10 h-10 text-white" />
        </div>
        <h3 class="font-heading text-2xl text-foreground mb-2">Баяр хүргэе!</h3>
        <p class="text-teal-700 text-lg font-medium">
          {{ score }}/{{ quizzes.length }} ({{ pct }}%)
        </p>
        <p class="text-muted-foreground text-sm mt-2">Та энэ хичээлийг амжилттай дууслаа!</p>
      </template>
      <template v-else>
        <div
          class="w-20 h-20 rounded-full bg-indigo-500/10 flex items-center justify-center mx-auto mb-4"
        >
          <span class="text-3xl font-heading text-indigo-600">{{ pct }}%</span>
        </div>
        <h3 class="font-heading text-2xl text-foreground mb-2">Тестийн үр дүн</h3>
        <p class="text-muted-foreground text-sm">
          {{ score }}/{{ quizzes.length }} зөв хариулт. 90%-аас дээш авбал дуусгана.
        </p>
      </template>
    </div>

    <div v-else-if="currentQuiz" class="glass-card rounded-xl p-6 space-y-5">
      <div class="flex items-center gap-4">
        <span class="text-xs text-muted-foreground whitespace-nowrap">
          Асуулт {{ currentQ + 1 }}/{{ quizzes.length }}
        </span>
        <div class="h-1.5 flex-1 bg-slate-100 rounded-full overflow-hidden">
          <div
            class="h-full gradient-indigo rounded-full transition-all"
            :style="{ width: `${((currentQ + 1) / quizzes.length) * 100}%` }"
          />
        </div>
      </div>

      <h3 class="text-lg font-medium text-foreground">{{ currentQuiz.question }}</h3>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <button
          v-for="option in currentQuiz.options"
          :key="option"
          @click="handleAnswer(option)"
          :disabled="showResult"
          class="glass-card glass-card-hover px-4 py-3 rounded-xl text-sm text-left transition-all disabled:cursor-default"
          :class="optionClass(option)"
        >
          {{ option }}
        </button>
      </div>

      <div v-if="showResult" class="space-y-3">
        <div class="bg-slate-50 border border-slate-200 rounded-xl px-4 py-3">
          <p class="text-xs text-indigo-600 mb-1 uppercase tracking-wider">Зөв хариулт</p>
          <p class="text-sm text-foreground">{{ currentQuiz.correct_answer }}</p>
        </div>
        <div class="flex items-center gap-2">
          <span
            class="text-sm font-medium"
            :class="selectedAnswer === currentQuiz.correct_answer ? 'text-teal-700' : 'text-red-600'"
          >
            {{ selectedAnswer === currentQuiz.correct_answer ? "Зөв хариуллаа" : "Буруу хариуллаа" }}
          </span>
        </div>
        <button
          @click="handleNext"
          class="gradient-indigo text-white px-6 py-2.5 rounded-xl text-sm font-medium hover:opacity-90 transition-opacity"
        >
          {{ currentQ < quizzes.length - 1 ? "Дараагийнх →" : "Үр дүн харах →" }}
        </button>
      </div>
    </div>
  </div>
</template>

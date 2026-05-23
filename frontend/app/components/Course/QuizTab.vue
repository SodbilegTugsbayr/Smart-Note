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
const submitting = ref(false)
const regenerating = ref(false)
const resultMessage = ref("")
const passed = ref(false)

const activeNote = computed(() => {
  const notes = props.course.notes || []
  return notes.find((note) => note.id === props.activeNoteId) || notes[0] || null
})
const currentQuiz = computed(() => quizzes.value[currentQ.value] || null)
const score = computed(() => answers.value.filter((a) => a.correct).length)
const pct = computed(() =>
  quizzes.value.length > 0 ? Math.round((score.value / quizzes.value.length) * 100) : 0,
)
const hasPassingScore = computed(
  () => quizzes.value.length > 0 && score.value * 100 > 90 * quizzes.value.length,
)

watch(
  () => props.activeNoteId,
  (noteId) => {
    fetchQuizzes(noteId)
  },
  { immediate: true },
)

useQueue((message) => {
  if (!message || message.Type !== "NOTE_QUIZ_REGENERATION") return

  let payload = message.Text
  if (typeof payload === "string") {
    try {
      payload = JSON.parse(payload)
    } catch {
      return
    }
  }

  if (Number(payload?.course_id) !== Number(props.course.id)) return
  if (Number(payload?.note_id) !== Number(activeNote.value?.id)) return

  if (payload.stage === "started") {
    regenerating.value = true
    resultMessage.value = payload.message || "Тест дахин үүсгэж байна"
    return
  }

  if (payload.stage === "completed") {
    regenerating.value = false
    quizzes.value = payload.quizzes || []
    resetQuizState()
    resultMessage.value = payload.message || "Шинэ тест бэлэн боллоо"
    return
  }

  if (payload.stage === "failed") {
    regenerating.value = false
    errorMessage.value = payload.message || "Тест дахин үүсгэхэд алдаа гарлаа"
  }
})

async function fetchQuizzes(noteId) {
  resetQuizState()
  quizzes.value = []
  errorMessage.value = ""
  resultMessage.value = ""
  regenerating.value = false

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
  answers.value.push({
    quiz_id: currentQuiz.value?.id,
    answer,
    correct: isCorrect,
  })
}

async function handleNext() {
  if (submitting.value || regenerating.value) return

  if (currentQ.value < quizzes.value.length - 1) {
    currentQ.value++
    selectedAnswer.value = null
    showResult.value = false
  } else {
    quizDone.value = true
    passed.value = hasPassingScore.value
    await submitQuizResult()
  }
}

function resetQuizState() {
  currentQ.value = 0
  selectedAnswer.value = null
  showResult.value = false
  answers.value = []
  quizDone.value = false
  passed.value = false
}

function resetQuiz() {
  if (submitting.value || regenerating.value) return
  resetQuizState()
  resultMessage.value = ""
}

async function submitQuizResult() {
  if (!activeNote.value) return

  submitting.value = true
  errorMessage.value = ""
  resultMessage.value = ""

  try {
    const result = await $fetch(`/api/notes/${activeNote.value.id}/quizzes/submit`, {
      method: "POST",
      body: {
        answers: answers.value.map((answer) => ({
          quiz_id: answer.quiz_id,
          answer: answer.answer,
        })),
      },
    })

    passed.value = !!result?.passed
    regenerating.value = !!result?.regenerating
    resultMessage.value = result?.message || ""
    mergeQuizSubmission(result)
  } catch (err) {
    errorMessage.value = err?.data?.message || "Тестийн үр дүн хадгалахад алдаа гарлаа"
  } finally {
    submitting.value = false
  }
}

function mergeQuizSubmission(result) {
  if (!result?.course && !result?.note) return

  let nextCourse = props.course
  if (result.course) {
    nextCourse = {
      ...nextCourse,
      ...result.course,
      notes: result.course.notes || nextCourse.notes,
    }
  }

  if (result.note) {
    let found = false
    const notes = (nextCourse.notes || []).map((note) => {
      if (Number(note.id) !== Number(result.note.id)) return note

      found = true
      return { ...note, ...result.note }
    })

    if (!found) {
      notes.push(result.note)
    }

    nextCourse = {
      ...nextCourse,
      notes,
    }
  }

  emit("update", nextCourse)
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
        :disabled="submitting || regenerating"
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

    <div v-else-if="regenerating && !quizDone" class="glass-card rounded-xl p-8 text-center">
      <Loader2Icon class="w-8 h-8 text-indigo-600 animate-spin mx-auto mb-3" />
      <p class="text-muted-foreground text-sm">
        {{ resultMessage || "Тест дахин үүсгэж байна" }}
      </p>
    </div>

    <div v-else-if="quizzes.length === 0" class="text-center py-12">
      <p class="text-muted-foreground text-sm">
        Тест байхгүй. Сонгосон тэмдэглэлийг боловсруулна уу.
      </p>
    </div>

    <div v-else-if="quizDone" class="glass-card rounded-xl p-8 text-center">
      <template v-if="passed">
        <div
          class="w-20 h-20 rounded-full gradient-teal flex items-center justify-center mx-auto mb-4"
        >
          <TrophyIcon class="w-10 h-10 text-white" />
        </div>
        <h3 class="font-heading text-2xl text-foreground mb-2">Баяр хүргэе!</h3>
        <p class="text-teal-700 text-lg font-medium">
          {{ score }}/{{ quizzes.length }} ({{ pct }}%)
        </p>
        <p class="text-muted-foreground text-sm mt-2">
          {{ resultMessage || "Та энэ тэмдэглэлийг амжилттай дууслаа!" }}
        </p>
      </template>
      <template v-else>
        <div
          class="w-20 h-20 rounded-full bg-indigo-500/10 flex items-center justify-center mx-auto mb-4"
        >
          <span class="text-3xl font-heading text-indigo-600">{{ pct }}%</span>
        </div>
        <h3 class="font-heading text-2xl text-foreground mb-2">Тестийн үр дүн</h3>
        <p class="text-muted-foreground text-sm mb-3">
          {{ score }}/{{ quizzes.length }} зөв хариулт. 90%-аас их авбал дуусгана.
        </p>
        <div
          v-if="submitting || regenerating"
          class="flex items-center justify-center gap-2 text-sm text-indigo-700"
        >
          <Loader2Icon class="w-4 h-4 animate-spin" />
          <span>
            {{ submitting ? "Үр дүн хадгалж байна" : resultMessage || "Тест дахин үүсгэж байна" }}
          </span>
        </div>
        <p v-else-if="resultMessage" class="text-sm text-muted-foreground">
          {{ resultMessage }}
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
          :disabled="submitting || regenerating"
          class="gradient-indigo text-white px-6 py-2.5 rounded-xl text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-60 flex items-center gap-2"
        >
          <Loader2Icon v-if="submitting" class="w-4 h-4 animate-spin" />
          <span>{{ currentQ < quizzes.length - 1 ? "Дараагийнх →" : "Үр дүн харах →" }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

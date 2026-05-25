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
const resultAnswers = ref([])
const resultScore = ref(0)
const resultTotal = ref(0)
const resultPercentage = ref(0)
const latestResult = ref(null)
const testStarted = ref(false)

const activeNote = computed(() => {
  const notes = props.course.notes || []
  return notes.find((note) => note.id === props.activeNoteId) || notes[0] || null
})
const currentQuiz = computed(() => quizzes.value[currentQ.value] || null)
const score = computed(() => resultScore.value)
const pct = computed(() => resultPercentage.value)
const resultCount = computed(() => resultTotal.value || quizzes.value.length)

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
  resetQuizState(false)
  quizzes.value = []
  latestResult.value = null
  errorMessage.value = ""
  resultMessage.value = ""
  regenerating.value = false

  if (!noteId) return

  loading.value = true
  try {
    const result = await $fetch(`/api/notes/${noteId}/quizzes`)
    if (String(props.activeNoteId) !== String(noteId)) return
    quizzes.value = Array.isArray(result) ? result : result?.items || []
    latestResult.value = Array.isArray(result) ? null : result?.latest_result || null
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
  selectedAnswer.value = answer
  showResult.value = true
  const nextAnswer = {
    quiz_id: currentQuiz.value?.id,
    answer,
  }
  const existingIndex = answers.value.findIndex((item) => item.quiz_id === nextAnswer.quiz_id)
  if (existingIndex >= 0) {
    answers.value.splice(existingIndex, 1, nextAnswer)
  } else {
    answers.value.push(nextAnswer)
  }
}

async function handleNext() {
  if (submitting.value || regenerating.value) return

  if (currentQ.value < quizzes.value.length - 1) {
    currentQ.value++
    selectedAnswer.value = null
    showResult.value = false
  } else {
    await submitQuizResult()
  }
}

function resetQuizState(started = false) {
  currentQ.value = 0
  selectedAnswer.value = null
  showResult.value = false
  answers.value = []
  quizDone.value = false
  testStarted.value = started
  passed.value = false
  resultAnswers.value = []
  resultScore.value = 0
  resultTotal.value = 0
  resultPercentage.value = 0
}

function resetQuiz() {
  if (submitting.value || regenerating.value) return
  resetQuizState(true)
  resultMessage.value = ""
}

function startQuiz() {
  if (!quizzes.value.length || submitting.value || regenerating.value) return
  resetQuizState(true)
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
    resultAnswers.value = result?.answers || []
    resultScore.value = Number(result?.score || 0)
    resultTotal.value = Number(result?.total || resultAnswers.value.length || quizzes.value.length)
    resultPercentage.value = Number(result?.percentage || 0)
    latestResult.value = result?.result || latestResult.value
    quizDone.value = true
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
  if (option === selectedAnswer.value) {
    return "border-indigo-500/40 bg-indigo-500/10 text-indigo-800"
  }
  return "opacity-70"
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-2 flex-wrap">
      <button
        v-if="(quizDone || testStarted) && quizzes.length"
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

    <div v-else-if="!testStarted && !quizDone" class="glass-card rounded-xl p-8 space-y-6">
      <div class="text-center">
        <div
          class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-indigo-500/10"
        >
          <ClipboardListIcon class="h-8 w-8 text-indigo-600" />
        </div>
        <h3 class="font-heading text-2xl text-foreground mb-2">Тест</h3>
        <p class="text-sm text-muted-foreground">
          {{ quizzes.length }} асуулттай тест. Эхлүүлсний дараа таб солисон ч хариултууд
          хадгалагдана.
        </p>
      </div>

      <div v-if="latestResult" class="rounded-xl border border-slate-200 bg-white/70 px-4 py-3">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Өмнөх үр дүн
            </p>
            <p class="mt-1 text-sm text-foreground">
              {{ latestResult.score }}/{{ latestResult.total }} ({{ latestResult.percentage }}%)
            </p>
          </div>
          <span
            v-if="latestResult.passed"
            class="rounded-full px-2.5 py-1 text-xs font-medium"
            :class="'bg-teal-500/10 text-teal-700'"
          >
            {{ "Давсан" }}
          </span>
        </div>
      </div>

      <button
        @click="startQuiz"
        class="gradient-indigo mx-auto flex items-center gap-2 rounded-xl px-6 py-2.5 text-sm font-medium text-white transition-opacity hover:opacity-90"
      >
        <PlayIcon class="h-4 w-4" />
        Тест эхлүүлэх
      </button>
    </div>

    <div v-else-if="quizDone" class="glass-card rounded-xl p-8 space-y-6">
      <div class="text-center">
        <template v-if="passed">
          <div
            class="w-20 h-20 rounded-full gradient-teal flex items-center justify-center mx-auto mb-4"
          >
            <TrophyIcon class="w-10 h-10 text-white" />
          </div>
          <h3 class="font-heading text-2xl text-foreground mb-2">Баяр хүргэе!</h3>
        </template>
        <template v-else>
          <div
            class="w-20 h-20 rounded-full bg-indigo-500/10 flex items-center justify-center mx-auto mb-4"
          >
            <span class="text-3xl font-heading text-indigo-600">{{ pct }}%</span>
          </div>
          <h3 class="font-heading text-2xl text-foreground mb-2">Тестийн үр дүн</h3>
        </template>

        <p class="text-lg font-medium" :class="passed ? 'text-teal-700' : 'text-indigo-700'">
          {{ score }}/{{ resultCount }} ({{ pct }}%)
        </p>
        <p class="text-muted-foreground text-sm mt-2">
          {{
            resultMessage ||
            (passed ? "Та энэ тэмдэглэлийг амжилттай дууслаа!" : "90%-аас их авбал дуусгана.")
          }}
        </p>

        <div
          v-if="submitting || regenerating"
          class="mt-3 flex items-center justify-center gap-2 text-sm text-indigo-700"
        >
          <Loader2Icon class="w-4 h-4 animate-spin" />
          <span>
            {{ submitting ? "Үр дүн хадгалж байна" : " Шинэ тест үүсгэж байна." }}
          </span>
        </div>
      </div>

      <div v-if="resultAnswers.length" class="space-y-3 text-left">
        <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Хариултын дэлгэрэнгүй
        </p>
        <div
          v-for="(item, idx) in resultAnswers"
          :key="item.quiz_id"
          class="rounded-xl border px-4 py-3"
          :class="
            item.correct ? 'border-teal-500/30 bg-teal-500/5' : 'border-red-500/30 bg-red-500/5'
          "
        >
          <div class="flex items-start justify-between gap-3">
            <p class="text-sm font-medium text-foreground">{{ idx + 1 }}. {{ item.question }}</p>
            <span
              class="flex-shrink-0 rounded-full px-2 py-0.5 text-xs font-medium"
              :class="item.correct ? 'bg-teal-500/10 text-teal-700' : 'bg-red-500/10 text-red-700'"
            >
              {{ item.correct ? "Зөв" : "Буруу" }}
            </span>
          </div>
          <div class="mt-3 grid gap-2 text-sm sm:grid-cols-2">
            <div class="rounded-lg bg-white/70 px-3 py-2">
              <p class="text-xs text-muted-foreground">Таны хариулт</p>
              <p :class="item.correct ? 'text-teal-700' : 'text-red-700'">
                {{ item.selected_answer || "Хариулаагүй" }}
              </p>
            </div>
            <div class="rounded-lg bg-white/70 px-3 py-2">
              <p class="text-xs text-muted-foreground">Зөв хариулт</p>
              <p class="text-teal-700">{{ item.correct_answer }}</p>
            </div>
          </div>
        </div>
      </div>
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
          <p class="text-xs text-indigo-600 mb-1 uppercase tracking-wider">Сонгосон хариулт</p>
          <p class="text-sm text-foreground">{{ selectedAnswer }}</p>
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

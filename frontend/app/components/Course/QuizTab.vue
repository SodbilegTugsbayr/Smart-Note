<script setup>
const props = defineProps({
  course: { type: Object, required: true },
})
const emit = defineEmits(["update"])

const generating = ref(false)
const currentQ = ref(0)
const selected = ref(null)
const showResult = ref(false)
const answers = ref([])
const quizDone = ref(false)

const questions = computed(() => {
  const qs = []
  for (const note of props.course.notes || []) {
    for (const fc of note.flash_cards || []) {
      qs.push({
        question: fc.question,
        options: shuffleWithCorrect(fc.answer),
        correct_index: 0,
      })
    }
  }
  return qs
})

const flashQA = computed(() => {
  const all = []
  for (const note of props.course.notes || []) {
    for (const fc of note.flash_cards || []) {
      all.push(fc)
    }
  }
  return all
})

const currentCard = computed(() => flashQA.value[currentQ.value])
const score = computed(() => answers.value.filter((a) => a.correct).length)
const pct = computed(() =>
  flashQA.value.length > 0 ? Math.round((score.value / flashQA.value.length) * 100) : 0,
)

function handleAnswer(isCorrect) {
  if (showResult.value) return
  selected.value = isCorrect
  showResult.value = true
  answers.value.push({ questionIdx: currentQ.value, correct: isCorrect })
}

async function handleNext() {
  if (currentQ.value < flashQA.value.length - 1) {
    currentQ.value++
    selected.value = null
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

function resetQuiz() {
  currentQ.value = 0
  selected.value = null
  showResult.value = false
  answers.value = []
  quizDone.value = false
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-2 flex-wrap">
      <button
        v-if="quizDone"
        @click="resetQuiz"
        class="glass-card glass-card-hover px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground flex items-center gap-2 transition-colors"
      >
        <RotateCcwIcon class="w-4 h-4" /> Дахин эхлүүлэх
      </button>
    </div>

    <div v-if="flashQA.length === 0" class="text-center py-12">
      <p class="text-muted-foreground text-sm">
        Тест байхгүй. Тэмдэглэлийн хэсгийг боловсруулна уу.
      </p>
    </div>

    <div v-else-if="quizDone" class="glass-card rounded-xl p-8 text-center">
      <template v-if="pct >= 90">
        <div
          class="w-20 h-20 rounded-full gradient-teal flex items-center justify-center mx-auto mb-4"
        >
          <TrophyIcon class="w-10 h-10 text-white" />
        </div>
        <h3 class="font-heading text-2xl text-foreground mb-2">Баяр хүргэе! 🎉</h3>
        <p class="text-teal-400 text-lg font-medium">
          {{ score }}/{{ flashQA.length }} ({{ pct }}%)
        </p>
        <p class="text-muted-foreground text-sm mt-2">Та энэ хичээлийг амжилттай дууслаа!</p>
      </template>
      <template v-else>
        <div
          class="w-20 h-20 rounded-full bg-indigo-500/10 flex items-center justify-center mx-auto mb-4"
        >
          <span class="text-3xl font-heading text-indigo-400">{{ pct }}%</span>
        </div>
        <h3 class="font-heading text-2xl text-foreground mb-2">Тестийн үр дүн</h3>
        <p class="text-muted-foreground text-sm">
          {{ score }}/{{ flashQA.length }} зөв хариулт. 90%-аас дээш авбал дуусгана.
        </p>
      </template>
    </div>

    <!-- Quiz card -->
    <div v-else class="glass-card rounded-xl p-6 space-y-5">
      <!-- Progress -->
      <div class="flex items-center gap-4">
        <span class="text-xs text-muted-foreground whitespace-nowrap"
          >Асуулт {{ currentQ + 1 }}/{{ flashQA.length }}</span
        >
        <div class="h-1.5 flex-1 bg-white/5 rounded-full overflow-hidden">
          <div
            class="h-full gradient-indigo rounded-full transition-all"
            :style="{ width: `${((currentQ + 1) / flashQA.length) * 100}%` }"
          />
        </div>
      </div>

      <h3 class="text-lg font-medium text-foreground">{{ currentCard.question }}</h3>

      <!-- Show answer after reveal -->
      <div v-if="!showResult" class="flex gap-3">
        <button
          @click="handleAnswer(true)"
          class="flex-1 glass-card glass-card-hover px-4 py-3 rounded-xl text-sm text-left transition-all hover:border-teal-500/30"
        >
          ✓ Мэднэ
        </button>
        <button
          @click="handleAnswer(false)"
          class="flex-1 glass-card glass-card-hover px-4 py-3 rounded-xl text-sm text-left transition-all hover:border-red-500/30"
        >
          ✗ Мэдэхгүй
        </button>
      </div>

      <div v-else class="space-y-3">
        <div class="bg-white/5 border border-white/10 rounded-xl px-4 py-3">
          <p class="text-xs text-indigo-400 mb-1 uppercase tracking-wider">Хариулт</p>
          <p class="text-sm text-foreground">{{ currentCard.answer }}</p>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-sm font-medium" :class="selected ? 'text-teal-400' : 'text-red-400'">
            {{ selected ? "✓ Зөв гэж тэмдэглэлээ" : "✗ Буруу гэж тэмдэглэлээ" }}
          </span>
        </div>
        <button
          @click="handleNext"
          class="gradient-indigo text-white px-6 py-2.5 rounded-xl text-sm font-medium hover:opacity-90 transition-opacity"
        >
          {{ currentQ < flashQA.length - 1 ? "Дараагийнх →" : "Үр дүн харах →" }}
        </button>
      </div>
    </div>
  </div>
</template>

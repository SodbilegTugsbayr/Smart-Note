<script setup>
const props = defineProps({
  course: { type: Object, required: true },
})

const messages = ref([])
const question = ref("")
const loading = ref(false)
const error = ref("")

async function sendQuestion() {
  const text = question.value.trim()
  if (!text || loading.value) return

  messages.value.push({ role: "user", text })
  question.value = ""
  error.value = ""
  loading.value = true

  try {
    const result = await $fetch("/api/ai/chat", {
      method: "POST",
      body: { course_id: props.course.id, question: text },
    })
    messages.value.push({
      role: "assistant",
      text: result?.answer || "Хариулт олдсонгүй.",
    })
  } catch (err) {
    error.value = err?.data?.message || "Эхлээд тэмдэглэлийг AI-аар боловсруулна уу."
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <div
      v-if="messages.length === 0"
      class="glass-card rounded-xl p-8 text-center text-sm text-muted-foreground"
    >
      <MessageSquareIcon class="w-8 h-8 mx-auto mb-3 text-indigo-400" />
      <p>Боловсруулсан тэмдэглэл дээр тулгуурлан асуулт асууна.</p>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="(message, index) in messages"
        :key="index"
        class="flex gap-3"
        :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
      >
        <div
          class="max-w-[85%] rounded-xl px-4 py-3 text-sm leading-relaxed"
          :class="
            message.role === 'user'
              ? 'gradient-indigo text-white'
              : 'glass-card text-foreground border border-white/10'
          "
        >
          <div class="flex items-start gap-2">
            <UserIcon v-if="message.role === 'user'" class="w-4 h-4 mt-0.5 flex-shrink-0" />
            <BotIcon v-else class="w-4 h-4 mt-0.5 text-indigo-400 flex-shrink-0" />
            <p class="whitespace-pre-wrap">{{ message.text }}</p>
          </div>
        </div>
      </div>
    </div>

    <p v-if="error" class="text-sm text-red-400">{{ error }}</p>

    <form class="glass-card rounded-xl p-3 flex items-end gap-3" @submit.prevent="sendQuestion">
      <Textarea
        v-model="question"
        placeholder="Асуултаа бичнэ үү..."
        class="min-h-12 max-h-40 bg-transparent border-0 text-foreground resize-none focus-visible:ring-0"
        @keydown.enter.exact.prevent="sendQuestion"
      />
      <button
        type="submit"
        :disabled="!question.trim() || loading"
        class="gradient-indigo text-white h-10 w-10 rounded-lg flex items-center justify-center disabled:opacity-50 flex-shrink-0"
      >
        <Loader2Icon v-if="loading" class="w-4 h-4 animate-spin" />
        <SendIcon v-else class="w-4 h-4" />
      </button>
    </form>
  </div>
</template>

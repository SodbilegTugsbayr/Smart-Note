<script setup>
import { BookOpenIcon, FileTextIcon, Loader2Icon, UploadIcon, XIcon } from "lucide-vue-next"

const props = defineProps({
  open: { type: Boolean, required: true },
})
const emit = defineEmits(["close"])

const router = useRouter()
const step = ref("choose")
const title = ref("")
const loading = ref(false)
const uploadedFile = ref(null)
const extractedChapters = ref([])
const isDragging = ref(false)
const extracting = ref(false)

async function handleBlankCreate() {
  if (!title.value.trim()) return
  loading.value = true
  try {
    const course = await $fetch("/api/courses", {
      method: "POST",
      body: { title: title.value.trim(), status: "in_progress", progress: 0 },
    })
    emit("close")
    router.push(`/course/${course.id}`)
  } finally {
    loading.value = false
  }
}

async function handleFileDrop(files) {
  const file = files?.[0]
  if (!file) return
  extracting.value = true
  try {
    const formData = new FormData()
    formData.append("file", file)
    const { file_url } = await $fetch("/api/upload", { method: "POST", body: formData })
    uploadedFile.value = { name: file.name, url: file_url }

    const result = await $fetch("/api/extract", {
      method: "POST",
      body: { file_url },
    })

    if (result?.title && !title.value) title.value = result.title
    extractedChapters.value = (result?.chapters || []).map((ch, i) => ({
      ...ch,
      id: `ch-${i}`,
      selected: true,
      topics: (ch.topics || []).map((t, j) => ({ ...t, id: `t-${i}-${j}`, selected: true })),
    }))
  } finally {
    extracting.value = false
  }
}

function onDrop(e) {
  e.preventDefault()
  isDragging.value = false
  handleFileDrop(e.dataTransfer.files)
}

function openFilePicker() {
  const input = document.createElement("input")
  input.type = "file"
  input.accept = ".pdf,.png,.jpg,.jpeg"
  input.onchange = (e) => handleFileDrop(e.target.files)
  input.click()
}

function toggleChapter(chIdx) {
  const ch = extractedChapters.value[chIdx]
  const next = !ch.selected
  ch.selected = next
  ch.topics.forEach((t) => (t.selected = next))
}

function toggleTopic(chIdx, tIdx) {
  extractedChapters.value[chIdx].topics[tIdx].selected =
    !extractedChapters.value[chIdx].topics[tIdx].selected
}

async function handleCreateFromFile() {
  if (!title.value.trim() || !uploadedFile.value) return
  loading.value = true
  try {
    const selectedChapters = extractedChapters.value
      .filter((ch) => ch.selected || ch.topics.some((t) => t.selected))
      .map((ch) => ({
        id: ch.id,
        title: ch.title,
        topics: ch.topics.filter((t) => t.selected).map((t) => ({ id: t.id, title: t.title })),
      }))

    const course = await $fetch("/api/courses", {
      method: "POST",
      body: {
        title: title.value.trim(),
        status: "in_progress",
        progress: 0,
        file_url: uploadedFile.value.url,
        chapters: selectedChapters,
      },
    })
    emit("close")
    router.push(`/course/${course.id}`)
  } finally {
    loading.value = false
  }
}

function resetModal() {
  step.value = "choose"
  title.value = ""
  uploadedFile.value = null
  extractedChapters.value = []
  extracting.value = false
}

function handleOpenChange(val) {
  if (!val) {
    emit("close")
    resetModal()
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="handleOpenChange">
    <DialogContent class="glass-card border-white/10 sm:max-w-lg max-h-[85vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle class="font-heading text-xl text-foreground">
          <span v-if="step === 'choose'">Шинэ хичээл үүсгэх</span>
          <span v-else-if="step === 'blank'">Хоосон хичээл</span>
          <span v-else>Файл оруулах</span>
        </DialogTitle>
      </DialogHeader>

      <!-- Step: choose -->
      <div v-if="step === 'choose'" class="grid grid-cols-2 gap-3 mt-4">
        <button
          @click="step = 'blank'"
          class="glass-card glass-card-hover rounded-xl p-6 flex flex-col items-center gap-3 transition-all hover:border-indigo-500/30"
        >
          <div class="w-12 h-12 rounded-xl gradient-indigo flex items-center justify-center">
            <BookOpenIcon class="w-6 h-6 text-white" />
          </div>
          <span class="text-sm font-medium text-foreground">Хоосон хичээл эхлүүлэх</span>
        </button>
        <button
          @click="step = 'upload'"
          class="glass-card glass-card-hover rounded-xl p-6 flex flex-col items-center gap-3 transition-all hover:border-indigo-500/30"
        >
          <div class="w-12 h-12 rounded-xl bg-violet-500/20 flex items-center justify-center">
            <UploadIcon class="w-6 h-6 text-violet-400" />
          </div>
          <span class="text-sm font-medium text-foreground">PDF / Зураг оруулах</span>
        </button>
      </div>

      <!-- Step: blank -->
      <div v-else-if="step === 'blank'" class="space-y-4 mt-4">
        <Input
          v-model="title"
          placeholder="Хичээлийн нэр..."
          class="bg-white/5 border-white/10 text-foreground placeholder:text-muted-foreground"
        />
        <div class="flex gap-2">
          <button
            @click="step = 'choose'"
            class="flex-1 py-2.5 rounded-xl text-sm text-muted-foreground hover:text-foreground glass-card glass-card-hover transition-all"
          >
            Буцах
          </button>
          <button
            @click="handleBlankCreate"
            :disabled="loading || !title.trim()"
            class="flex-1 py-2.5 rounded-xl text-sm font-medium text-white gradient-indigo hover:opacity-90 transition-opacity disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <Loader2Icon v-if="loading" class="w-4 h-4 animate-spin" />
            <template v-else>Үүсгэх <span>→</span></template>
          </button>
        </div>
      </div>

      <!-- Step: upload -->
      <div v-else class="space-y-4 mt-4">
        <Input
          v-model="title"
          placeholder="Хичээлийн нэр..."
          class="bg-white/5 border-white/10 text-foreground placeholder:text-muted-foreground"
        />

        <!-- Dropzone -->
        <div
          v-if="!uploadedFile && !extracting"
          @dragover.prevent="isDragging = true"
          @dragleave="isDragging = false"
          @drop="onDrop"
          @click="openFilePicker"
          class="border-2 border-dashed rounded-xl p-10 flex flex-col items-center gap-3 transition-colors cursor-pointer"
          :class="
            isDragging
              ? 'border-indigo-500/50 bg-indigo-500/5'
              : 'border-white/10 hover:border-white/20'
          "
        >
          <UploadIcon class="w-8 h-8 text-muted-foreground" />
          <p class="text-sm text-muted-foreground text-center">PDF эсвэл зураг чирж оруулна уу</p>
          <p class="text-xs text-muted-foreground/60">.pdf, .png, .jpg</p>
        </div>

        <!-- Extracting -->
        <div v-else-if="extracting" class="flex flex-col items-center gap-3 py-8">
          <Loader2Icon class="w-8 h-8 text-indigo-400 animate-spin" />
          <p class="text-sm text-muted-foreground">Файл задалж байна...</p>
        </div>

        <!-- Uploaded file + chapters -->
        <div v-else class="space-y-3">
          <div class="flex items-center gap-2 glass-card rounded-lg px-3 py-2">
            <FileTextIcon class="w-4 h-4 text-indigo-400" />
            <span class="text-sm text-foreground flex-1 truncate">{{ uploadedFile.name }}</span>
            <button @click="((uploadedFile = null), (extractedChapters = []))">
              <XIcon class="w-4 h-4 text-muted-foreground hover:text-foreground" />
            </button>
          </div>

          <div
            v-if="extractedChapters.length"
            class="glass-card rounded-xl p-4 space-y-2 max-h-52 overflow-y-auto"
          >
            <p class="text-xs text-muted-foreground font-medium uppercase tracking-wider mb-2">
              Бүтэц
            </p>
            <div v-for="(ch, chIdx) in extractedChapters" :key="ch.id">
              <label class="flex items-center gap-2 cursor-pointer py-1">
                <input
                  type="checkbox"
                  :checked="ch.selected"
                  @change="toggleChapter(chIdx)"
                  class="rounded border-white/20 bg-white/5 text-indigo-500 focus:ring-indigo-500/30"
                />
                <BookOpenIcon class="w-3.5 h-3.5 text-indigo-400" />
                <span class="text-sm text-foreground">{{ ch.title }}</span>
              </label>
              <div class="ml-7 space-y-0.5">
                <label
                  v-for="(t, tIdx) in ch.topics"
                  :key="t.id"
                  class="flex items-center gap-2 cursor-pointer py-0.5"
                >
                  <input
                    type="checkbox"
                    :checked="t.selected"
                    @change="toggleTopic(chIdx, tIdx)"
                    class="rounded border-white/20 bg-white/5 text-indigo-500 focus:ring-indigo-500/30"
                  />
                  <span class="text-xs text-muted-foreground">{{ t.title }}</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <div class="flex gap-2">
          <button
            @click="step = 'choose'"
            class="flex-1 py-2.5 rounded-xl text-sm text-muted-foreground hover:text-foreground glass-card glass-card-hover transition-all"
          >
            Буцах
          </button>
          <button
            @click="handleCreateFromFile"
            :disabled="loading || !title.trim() || !uploadedFile"
            class="flex-1 py-2.5 rounded-xl text-sm font-medium text-white gradient-indigo hover:opacity-90 transition-opacity disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <Loader2Icon v-if="loading" class="w-4 h-4 animate-spin" />
            <template v-else>Үүсгэх <span>→</span></template>
          </button>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

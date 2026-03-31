<script setup>
const props = defineProps({
  open: { type: Boolean, required: true },
})
const emit = defineEmits(["close", "created"])

const step = ref("choose")
const loading = ref(false)
const extracting = ref(false)
const isDragging = ref(false)
const isIconOpen = ref(false)
const errorMessage = ref("")

const form = reactive({
  title: "",
  description: "",
  icon: "BookOpen",
  file: null,
  fileName: null,
})

const extractedChapters = ref([])

const iconList = [
  "BookOpen",
  "Notebook",
  "GraduationCap",
  "Brain",
  "Code",
  "Database",
  "Cpu",
  "PenTool",
  "Lightbulb",
  "FlaskConical",
]

// ── Submit (unified) ──────────────────────────────────────────────────────────

async function handleSubmit() {
  if (!form.title.trim()) return
  if (step.value === "upload" && !form.file) return

  loading.value = true
  errorMessage.value = ""

  try {
    const formData = new FormData()
    formData.append("title", form.title.trim())
    formData.append("description", form.description.trim())
    formData.append("icon", form.icon)
    formData.append("hasFile", String(!!form.file))

    if (form.file) {
      formData.append("file", form.file)
      formData.append("sections", JSON.stringify(buildSections()))
    }

    await $fetch("/api/course", { method: "POST", body: formData })
    emit("created")
  } catch {
    errorMessage.value = "Хичээл үүсгэхэд алдаа гарлаа"
  } finally {
    loading.value = false
    resetModal()
  }
}

function buildSections() {
  const sections = []
  for (const ch of extractedChapters.value) {
    const selectedTopics = ch.topics.filter((t) => t.selected)
    if (selectedTopics.length > 0) {
      for (const t of selectedTopics) {
        sections.push({
          section_name: `${ch.title} — ${t.title}`,
          page_range: `${t.startPage}-${t.endPage}`,
        })
      }
    } else if (ch.selected) {
      sections.push({
        section_name: ch.title,
        page_range: `${ch.startPage}-${ch.endPage}`,
      })
    }
  }
  return sections
}

// ── File handling ─────────────────────────────────────────────────────────────

async function handleFileDrop(files) {
  const file = files?.[0]
  if (!file) return

  form.file = file
  form.fileName = file.name

  if (!form.title) {
    form.title = file.name.replace(/\.[^/.]+$/, "").replace(/[-_]/g, " ")
  }

  if (file.type !== "application/pdf") {
    extractedChapters.value = []
    return
  }

  extracting.value = true
  try {
    await loadScript("https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.min.js")
    window.pdfjsLib.GlobalWorkerOptions.workerSrc =
      "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js"
    const pdf = await window.pdfjsLib.getDocument({ data: await file.arrayBuffer() }).promise
    extractedChapters.value = await extractStructure(pdf)
  } catch (e) {
    console.error("PDF parse error:", e)
  } finally {
    extracting.value = false
  }
}

function clearFile() {
  form.file = null
  form.fileName = null
  extractedChapters.value = []
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

// ── PDF parsing ───────────────────────────────────────────────────────────────

function loadScript(src) {
  return new Promise((resolve, reject) => {
    if (document.querySelector(`script[src="${src}"]`)) return resolve()
    const s = Object.assign(document.createElement("script"), {
      src,
      onload: resolve,
      onerror: reject,
    })
    document.head.appendChild(s)
  })
}

async function extractStructure(pdf) {
  const chapterPatterns = [
    /^(chapter|бүлэг|хэсэг)\s*\d*/i,
    /^\d+[\.\s]+[A-ZА-ЯӨҮ]/,
    /^[IVXLC]+[\.\s]+\w/,
  ]
  const sectionPatterns = [/^\d+\.\d+[\.\s]/, /^(section|дэд\s*сэдэв|сэдэв)\s*\d*/i]

  const chapters = []
  let currentChapter = null

  for (let pageNum = 1; pageNum <= pdf.numPages; pageNum++) {
    const page = await pdf.getPage(pageNum)
    const lines = groupIntoLines((await page.getTextContent()).items)

    for (const line of lines) {
      const text = line.trim()
      if (!text || text.length > 80) continue

      if (chapterPatterns.some((p) => p.test(text))) {
        currentChapter = {
          id: `ch-${chapters.length}`,
          title: text,
          startPage: pageNum,
          endPage: pageNum,
          selected: true,
          topics: [],
        }
        chapters.push(currentChapter)
      } else if (sectionPatterns.some((p) => p.test(text)) && currentChapter) {
        currentChapter.topics.push({
          id: `t-${chapters.length - 1}-${currentChapter.topics.length}`,
          title: text,
          startPage: pageNum,
          endPage: pageNum,
          selected: true,
        })
      }
    }

    if (currentChapter) {
      currentChapter.endPage = pageNum
      const topics = currentChapter.topics
      if (topics.length) topics[topics.length - 1].endPage = pageNum
    }
  }

  // Fix page ranges
  for (const ch of chapters) {
    for (let i = 0; i < ch.topics.length - 1; i++)
      ch.topics[i].endPage = ch.topics[i + 1].startPage - 1
  }
  for (let i = 0; i < chapters.length - 1; i++) chapters[i].endPage = chapters[i + 1].startPage - 1

  return chapters
}

function groupIntoLines(items) {
  const lineMap = {}
  for (const item of items) {
    const y = Math.round(item.transform[5])
    ;(lineMap[y] ??= []).push(item.str)
  }
  return Object.values(lineMap).map((parts) => parts.join(" "))
}

// ── Chapter / topic selection ─────────────────────────────────────────────────

function toggleChapter(chIdx) {
  const ch = extractedChapters.value[chIdx]
  ch.selected = !ch.selected
  ch.topics.forEach((t) => (t.selected = ch.selected))
}

function toggleTopic(chIdx, tIdx) {
  const topic = extractedChapters.value[chIdx].topics[tIdx]
  topic.selected = !topic.selected
}

// ── Modal lifecycle ───────────────────────────────────────────────────────────

function resetModal() {
  step.value = "choose"
  Object.assign(form, { title: "", description: "", icon: "BookOpen", file: null, fileName: null })
  extractedChapters.value = []
  extracting.value = false
  errorMessage.value = ""
}

function handleOpenChange(val) {
  if (!val) {
    emit("close")
    resetModal()
  }
}

const isSubmitDisabled = computed(
  () =>
    loading.value ||
    !form.title.trim() ||
    !form.description.trim() ||
    (step.value === "upload" && (!form.file || extracting.value)),
)
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

      <p v-if="errorMessage" class="text-sm text-red-400 mt-1">{{ errorMessage }}</p>

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
            <UploadIcon class="w-6 h-6" />
          </div>
          <span class="text-sm font-medium text-foreground">PDF / Зураг оруулах</span>
        </button>
      </div>

      <!-- Step: blank | upload -->
      <div v-else class="space-y-4 mt-4">
        <!-- Icon + title -->
        <div class="flex items-center gap-2">
          <Popover v-model:open="isIconOpen">
            <PopoverTrigger as-child>
              <Button variant="outline" size="icon" class="bg-white/5 border-white/10">
                <component :is="`${form.icon}Icon`" class="w-5 h-5" />
              </Button>
            </PopoverTrigger>
            <PopoverContent class="w-64">
              <div class="grid grid-cols-5 gap-2">
                <Button
                  v-for="name in iconList"
                  :key="name"
                  variant="ghost"
                  size="icon"
                  @click="((form.icon = name), (isIconOpen = false))"
                >
                  <component :is="`${name}Icon`" class="w-5 h-5" />
                </Button>
              </div>
            </PopoverContent>
          </Popover>
          <Input
            v-model="form.title"
            placeholder="Хичээлийн нэр..."
            class="bg-white/5 border-white/10 text-foreground placeholder:text-muted-foreground"
          />
        </div>

        <Input
          v-model="form.description"
          placeholder="Хичээлийн тайлбар..."
          class="bg-white/5 border-white/10 text-foreground placeholder:text-muted-foreground"
        />

        <!-- Dropzone (upload step only) -->
        <template v-if="step === 'upload'">
          <!-- Empty state -->
          <div
            v-if="!form.file && !extracting"
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

          <!-- Parsing -->
          <div v-else-if="extracting" class="flex flex-col items-center gap-3 py-8">
            <Loader2Icon class="w-8 h-8 text-indigo-400 animate-spin" />
            <p class="text-sm text-muted-foreground">PDF задалж байна...</p>
          </div>

          <!-- File preview + chapter tree -->
          <div v-else class="space-y-3">
            <div class="flex items-center gap-2 glass-card rounded-lg px-3 py-2">
              <FileTextIcon class="w-4 h-4 text-indigo-400" />
              <span class="text-sm text-foreground flex-1 truncate">{{ form.fileName }}</span>
              <button @click="clearFile">
                <XIcon class="w-4 h-4 text-muted-foreground hover:text-foreground" />
              </button>
            </div>

            <div
              v-if="extractedChapters.length === 0"
              class="text-xs text-muted-foreground text-center py-4 glass-card rounded-xl"
            >
              Бүтэц илрүүлсэнгүй — бүх хуудсыг оруулна
            </div>

            <div v-else class="glass-card rounded-xl p-4 space-y-2 max-h-56 overflow-y-auto">
              <p class="text-xs text-muted-foreground font-medium uppercase tracking-wider mb-2">
                Бүтэц сонгох
              </p>
              <div v-for="(ch, chIdx) in extractedChapters" :key="ch.id">
                <label class="flex items-center gap-2 cursor-pointer py-1">
                  <input
                    type="checkbox"
                    :checked="ch.selected"
                    @change="toggleChapter(chIdx)"
                    class="rounded border-white/20 bg-white/5 text-indigo-500 focus:ring-indigo-500/30"
                  />
                  <BookOpenIcon class="w-3.5 h-3.5 text-indigo-400 shrink-0" />
                  <span class="text-sm text-foreground flex-1">{{ ch.title }}</span>
                  <span class="text-xs text-muted-foreground/40 shrink-0 tabular-nums">
                    {{ ch.startPage }}–{{ ch.endPage }}p
                  </span>
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
                    <span class="text-xs text-foreground flex-1">{{ t.title }}</span>
                    <span class="text-xs text-muted-foreground/40 shrink-0 tabular-nums">
                      {{ t.startPage }}–{{ t.endPage }}p
                    </span>
                  </label>
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- Actions -->
        <div class="flex gap-2">
          <button
            @click="step = 'choose'"
            class="flex-1 py-2.5 rounded-xl text-sm text-muted-foreground hover:text-foreground glass-card glass-card-hover transition-all"
          >
            Буцах
          </button>
          <button
            @click="handleSubmit"
            :disabled="isSubmitDisabled"
            class="flex-1 py-2.5 rounded-xl text-sm font-medium text-white gradient-indigo hover:opacity-90 transition-opacity disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <Loader2Icon v-if="loading" class="w-4 h-4 animate-spin" />
            <template v-else>Үүсгэх</template>
          </button>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

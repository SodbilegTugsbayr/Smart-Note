<script setup>
const route = useRoute()
const courseId = route.params.id
const user = useUser()

const sidebarOpen = ref(false)
const activeNoteId = ref(null)
const activeTab = ref("notes")
const publishing = ref(false)
const publishError = ref("")
const bookUploadOpen = ref(false)
const bookUploading = ref(false)
const bookExtracting = ref(false)
const bookDragging = ref(false)
const bookUploadError = ref("")
const MAX_BOOK_UPLOAD_SIZE = 50 * 1024 * 1024
const bookUpload = reactive({
  file: null,
  fileName: "",
})
const bookChapters = ref([])

const { data: course, refresh } = await useFetch(`/api/course/${courseId}`)

const notes = computed(() =>
  [...(course.value?.notes || [])].sort((a, b) => Number(a.id || 0) - Number(b.id || 0)),
)
const canEditCourse = computed(
  () => user.value?.role === "admin" || Number(course.value?.user_id) === Number(user.value?.id),
)
const isReadOnly = computed(() => !canEditCourse.value)

watch(
  notes,
  (nextNotes) => {
    if (!nextNotes.length) {
      activeNoteId.value = null
      return
    }

    const hasActiveNote = nextNotes.some((note) => note.id === activeNoteId.value)
    if (!hasActiveNote) {
      activeNoteId.value = nextNotes[0].id
    }
  },
  { immediate: true },
)

async function togglePublic(checked) {
  if (!course.value || publishing.value || isReadOnly.value) return

  const previousValue = !!course.value.is_public
  const nextValue = !!checked

  publishing.value = true
  publishError.value = ""
  course.value.is_public = nextValue

  try {
    const savedCourse = await $fetch(`/api/course/${courseId}/`, {
      method: "PATCH",
      body: { is_public: nextValue },
    })
    course.value = {
      ...course.value,
      ...savedCourse,
      notes: savedCourse.notes || course.value.notes || [],
    }
  } catch (err) {
    course.value.is_public = previousValue
    publishError.value = err?.data?.message || err?.data || "Нийтлэх төлөв шинэчлэхэд алдаа гарлаа"
  } finally {
    publishing.value = false
  }
}

function copyLink() {
  navigator.clipboard.writeText(window.location.href)
}

function handleUpdate(updated) {
  course.value = updated
}

function handleNoteCreated(note) {
  const nextNotes = [...(course.value?.notes || []), note]
  course.value = useCourseWithSyncedProgress(course.value, nextNotes)
}

function handleNoteUpdated(updatedNote) {
  const nextNotes = (course.value?.notes || []).map((note) =>
    note.id === updatedNote.id ? { ...note, ...updatedNote } : note,
  )
  course.value = useCourseWithSyncedProgress(course.value, nextNotes)
}

function handleNoteDeleted(noteId) {
  const nextNotes = (course.value?.notes || []).filter((note) => note.id !== noteId)
  course.value = useCourseWithSyncedProgress(course.value, nextNotes)
}

function selectNote(id) {
  activeNoteId.value = id
  sidebarOpen.value = false
}

function resetBookUpload() {
  bookUpload.file = null
  bookUpload.fileName = ""
  bookChapters.value = []
  bookExtracting.value = false
  bookDragging.value = false
  bookUploadError.value = ""
}

function closeBookUpload() {
  if (bookUploading.value) return
  bookUploadOpen.value = false
  resetBookUpload()
}

function openBookPicker() {
  const input = document.createElement("input")
  input.type = "file"
  input.accept = ".pdf,.png,.jpg,.jpeg"
  input.onchange = (e) => handleBookFile(e.target.files?.[0])
  input.click()
}

function onBookDrop(e) {
  e.preventDefault()
  bookDragging.value = false
  handleBookFile(e.dataTransfer.files?.[0])
}

async function handleBookFile(file) {
  if (!file) return

  bookUploadError.value = ""
  if (file.size > MAX_BOOK_UPLOAD_SIZE) {
    resetBookUpload()
    bookUploadError.value = "Файлын хэмжээ 50MB-аас бага байх ёстой"
    return
  }

  bookUpload.file = file
  bookUpload.fileName = file.name
  bookChapters.value = []

  if (file.type !== "application/pdf") return

  bookExtracting.value = true
  try {
    await loadPdfJs()
    window.pdfjsLib.GlobalWorkerOptions.workerSrc =
      "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js"
    const pdf = await window.pdfjsLib.getDocument({ data: await file.arrayBuffer() }).promise
    bookChapters.value = await getBookPdfStructure(pdf)
  } catch (err) {
    console.error("PDF parse error:", err)
  } finally {
    bookExtracting.value = false
  }
}

function loadPdfJs() {
  return new Promise((resolve, reject) => {
    const src = "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.min.js"
    if (document.querySelector(`script[src="${src}"]`)) return resolve()
    const script = Object.assign(document.createElement("script"), {
      src,
      onload: resolve,
      onerror: reject,
    })
    document.head.appendChild(script)
  })
}

async function getBookPdfStructure(pdfDoc) {
  const rawOutline = await pdfDoc.getOutline()
  const totalPages = pdfDoc.numPages
  if (!rawOutline) return []

  const flattenedItems = []
  async function traverse(items, depth = 0) {
    for (const item of items) {
      let pageIndex = -1
      try {
        if (item.dest) {
          const dest =
            typeof item.dest === "string" ? await pdfDoc.getDestination(item.dest) : item.dest
          if (dest) pageIndex = await pdfDoc.getPageIndex(dest[0])
        }
      } catch {
        console.warn(`Could not resolve page for: ${item.title}`)
      }

      const node = {
        title: item.title,
        startPage: pageIndex + 1,
        endPage: totalPages,
        depth,
        selected: false,
      }

      if (depth === 0) {
        node.topics = []
        flattenedItems.push(node)
      } else {
        const parentChapter = [...flattenedItems].reverse().find((chapter) => chapter.depth === 0)
        if (parentChapter) parentChapter.topics.push(node)
      }

      if (item.items?.length > 0) {
        await traverse(item.items, depth + 1)
      }
    }
  }

  await traverse(rawOutline)

  const allNodes = flattenedItems.flatMap((chapter) => [chapter, ...chapter.topics])
  for (let i = 0; i < allNodes.length - 1; i++) {
    const next = allNodes[i + 1]
    if (next.startPage > 0) {
      allNodes[i].endPage = Math.max(allNodes[i].startPage, next.startPage - 1)
    }
  }

  return flattenedItems
}

function toggleBookChapter(index) {
  const chapter = bookChapters.value[index]
  chapter.selected = !chapter.selected
  chapter.topics.forEach((topic) => (topic.selected = chapter.selected))
}

function toggleBookTopic(chapterIndex, topicIndex) {
  const topic = bookChapters.value[chapterIndex].topics[topicIndex]
  topic.selected = !topic.selected
}

function buildBookSections() {
  const sections = []
  for (const chapter of bookChapters.value) {
    const selectedTopics = chapter.topics.filter((topic) => topic.selected)
    if (selectedTopics.length > 0) {
      for (const topic of selectedTopics) {
        appendBookSection(sections, `${chapter.title} — ${topic.title}`, topic.startPage, topic.endPage)
      }
    } else if (chapter.selected) {
      appendBookSection(sections, chapter.title, chapter.startPage, chapter.endPage)
    }
  }
  return sections
}

function appendBookSection(sections, title, startPage, endPage) {
  const start = Number(startPage)
  const end = Number(endPage)
  if (!Number.isFinite(start) || !Number.isFinite(end) || start <= 0 || end < start) return
  sections.push({ section_name: title, start_page: start, end_page: end })
}

async function submitBookUpload() {
  if (!bookUpload.file || bookUploading.value || isReadOnly.value) return

  bookUploading.value = true
  bookUploadError.value = ""
  try {
    const formData = new FormData()
    formData.append("file", bookUpload.file)
    formData.append("sections", JSON.stringify(buildBookSections()))

    const savedCourse = await $fetch(`/api/course/${courseId}/book`, {
      method: "POST",
      body: formData,
    })
    course.value = savedCourse
    activeTab.value = "notes"
    closeBookUpload()
  } catch (err) {
    bookUploadError.value = err?.data?.message || err?.data || "Ном оруулахад алдаа гарлаа"
  } finally {
    bookUploading.value = false
  }
}
</script>

<template>
  <!-- <NavTopBar /> -->
  <div v-if="!course" class="flex items-center justify-center h-[60vh]">
    <Loader2Icon class="w-8 h-8 text-indigo-600 animate-spin" />
  </div>

  <div v-else class="min-h-[calc(100vh-64px)] flex flex-col">
    <header class="sticky z-40 glass-card border-b border-slate-200">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="flex flex-wrap items-center justify-between gap-4 py-4">
          <div class="flex min-w-0 items-center gap-6">
            <NuxtLink
              to="/"
              class="shrink-0 text-muted-foreground transition-colors hover:text-foreground"
            >
              <ArrowLeftIcon class="h-5 w-5" />
            </NuxtLink>

            <div class="min-w-0">
              <h1 class="truncate font-heading text-2xl text-foreground sm:text-3xl">
                {{ course.title }}
              </h1>

              <p class="text-xs text-muted-foreground">
                {{
                  course.status === "completed" ? "Дуусгасан" : `${course.progress || 0}% дууссан`
                }}
              </p>
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-3 self-center">
            <button
              v-if="canEditCourse && !course.has_book"
              type="button"
              @click="bookUploadOpen = true"
              class="glass-card glass-card-hover flex items-center gap-2 rounded-lg px-3 py-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              <UploadIcon class="h-4 w-4" />
              Ном оруулах
            </button>

            <div v-if="canEditCourse" class="flex items-center gap-2">
              <GlobeIcon class="h-4 w-4 shrink-0 text-muted-foreground" />

              <span class="text-xs leading-none text-muted-foreground"> Нийтлэх </span>

              <Switch
                :model-value="course.is_public"
                :disabled="publishing"
                @update:model-value="togglePublic"
                class="shrink-0"
              />
            </div>

            <button
              type="button"
              @click="copyLink"
              class="glass-card glass-card-hover flex items-center justify-center rounded-lg p-2 transition-colors"
            >
              <Link2Icon class="h-4 w-4 text-muted-foreground" />
            </button>
          </div>
        </div>

        <p v-if="publishError" class="pb-4 text-right text-xs text-red-600">
          {{ publishError }}
        </p>
      </div>
    </header>

    <div class="flex flex-1 min-h-0">
      <Button
        @click="sidebarOpen = !sidebarOpen"
        class="lg:hidden fixed bottom-4 left-4 z-50 gradient-indigo text-white w-12 h-12 rounded-full flex items-center justify-center shadow-lg"
      >
        <XIcon v-if="sidebarOpen" class="w-5 h-5" />
        <MenuIcon v-else class="w-5 h-5" />
      </Button>

      <!-- Sidebar -->
      <aside
        class="fixed lg:static inset-y-0 left-0 top-32 z-30 w-64 glass-card border-r border-slate-200 overflow-y-auto transition-transform duration-300"
        :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'"
      >
        <CourseChapterSideBar
          :course-id="course.id"
          :notes="notes"
          :active-note-id="activeNoteId"
          :readonly="isReadOnly"
          @note-select="selectNote"
          @note-created="handleNoteCreated"
          @note-updated="handleNoteUpdated"
          @note-deleted="handleNoteDeleted"
        />
      </aside>

      <!-- Main content -->
      <div class="flex-1 overflow-y-auto">
        <div class="max-w-4xl mx-auto px-4 sm:px-6 py-6">
          <!-- Tabs -->
          <Tabs v-model="activeTab" default-value="notes" :unmount-on-hide="false" class="space-y-4">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <TabsList class="bg-slate-100 border border-slate-200 p-1 rounded-xl">
                <TabsTrigger
                  value="notes"
                  class="data-[state=active]:bg-white data-[state=active]:text-indigo-700 rounded-lg text-sm"
                >
                  Тэмдэглэл
                </TabsTrigger>
                <TabsTrigger
                  value="flashcards"
                  class="data-[state=active]:bg-white data-[state=active]:text-indigo-700 rounded-lg text-sm"
                >
                  Флаш карт
                </TabsTrigger>
                <TabsTrigger
                  v-if="!isReadOnly"
                  value="quiz"
                  class="data-[state=active]:bg-white data-[state=active]:text-indigo-700 rounded-lg text-sm"
                >
                  Тест
                </TabsTrigger>
                <TabsTrigger
                  v-if="!isReadOnly"
                  value="chat"
                  class="data-[state=active]:bg-white data-[state=active]:text-indigo-700 rounded-lg text-sm"
                >
                  Chatbot
                </TabsTrigger>
              </TabsList>

              <div
                v-show="activeTab === 'notes'"
                id="course-notes-actions"
                class="flex items-center justify-end gap-2"
              />
              <div
                v-show="activeTab === 'flashcards'"
                id="course-flashcards-actions"
                class="flex items-center justify-end gap-2"
              />
            </div>

            <TabsContent value="notes">
              <CourseNotesTab
                :course="course"
                :active-note-id="activeNoteId"
                :readonly="isReadOnly"
                @update="handleUpdate"
              />
            </TabsContent>
            <TabsContent value="flashcards">
              <CourseFlashCardsTab
                :course="course"
                :active-note-id="activeNoteId"
                :readonly="isReadOnly"
                @update="handleUpdate"
              />
            </TabsContent>
            <TabsContent v-if="!isReadOnly" value="quiz">
              <CourseQuizTab
                :course="course"
                :active-note-id="activeNoteId"
                @update="handleUpdate"
              />
            </TabsContent>
            <TabsContent v-if="!isReadOnly" value="chat">
              <CourseChatTab :course="course" />
            </TabsContent>
          </Tabs>
        </div>
      </div>
    </div>

    <Dialog :open="bookUploadOpen" @update:open="(open) => (open ? (bookUploadOpen = true) : closeBookUpload())">
      <DialogContent class="glass-card border-slate-200 sm:max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle class="font-heading text-xl text-foreground">Ном оруулах</DialogTitle>
          <DialogDescription></DialogDescription>
        </DialogHeader>

        <p v-if="bookUploadError" class="text-sm text-red-600">{{ bookUploadError }}</p>

        <div class="space-y-4">
          <div
            v-if="!bookUpload.file && !bookExtracting"
            @dragover.prevent="bookDragging = true"
            @dragleave="bookDragging = false"
            @drop="onBookDrop"
            @click="openBookPicker"
            class="border-2 border-dashed rounded-xl p-10 flex flex-col items-center gap-3 transition-colors cursor-pointer"
            :class="
              bookDragging
                ? 'border-indigo-500/50 bg-indigo-500/5'
                : 'border-slate-200 hover:border-slate-300'
            "
          >
            <UploadIcon class="w-8 h-8 text-muted-foreground" />
            <p class="text-sm text-muted-foreground text-center">PDF эсвэл зураг чирж оруулна уу</p>
            <p class="text-xs text-muted-foreground/60">.pdf, .png, .jpg</p>
          </div>

          <div v-else-if="bookExtracting" class="flex flex-col items-center gap-3 py-8">
            <Loader2Icon class="w-8 h-8 text-indigo-600 animate-spin" />
            <p class="text-sm text-muted-foreground">PDF задалж байна...</p>
          </div>

          <div v-else class="space-y-3">
            <div class="flex items-center gap-2 glass-card rounded-lg px-3 py-2">
              <FileTextIcon class="w-4 h-4 text-indigo-600" />
              <span class="text-sm text-foreground flex-1 truncate">{{ bookUpload.fileName }}</span>
              <button type="button" :disabled="bookUploading" @click="resetBookUpload">
                <XIcon class="w-4 h-4 text-muted-foreground hover:text-foreground" />
              </button>
            </div>

            <div
              v-if="bookChapters.length === 0"
              class="text-xs text-muted-foreground text-center py-4 glass-card rounded-xl"
            >
              Бүтэц илрүүлсэнгүй — бүх хуудсыг нэг тэмдэглэл болгоно
            </div>

            <div v-else class="glass-card rounded-xl p-4 space-y-2 max-h-56 overflow-y-auto">
              <p class="text-xs text-muted-foreground font-medium uppercase tracking-wider mb-2">
                Бүтэц сонгох
              </p>
              <div v-for="(chapter, chapterIndex) in bookChapters" :key="chapter.title">
                <label class="flex items-center gap-2 cursor-pointer py-1">
                  <input
                    type="checkbox"
                    :checked="chapter.selected"
                    @change="toggleBookChapter(chapterIndex)"
                    class="rounded border-slate-300 bg-white text-indigo-500 focus:ring-indigo-500/30"
                  />
                  <BookOpenIcon class="w-3.5 h-3.5 text-indigo-600 shrink-0" />
                  <span class="text-sm text-foreground flex-1">{{ chapter.title }}</span>
                  <span class="text-xs text-muted-foreground/40 shrink-0 tabular-nums">
                    {{ chapter.startPage }}–{{ chapter.endPage }}p
                  </span>
                </label>
                <div class="ml-7 space-y-0.5">
                  <label
                    v-for="(topic, topicIndex) in chapter.topics"
                    :key="topic.title"
                    class="flex items-center gap-2 cursor-pointer py-0.5"
                  >
                    <input
                      type="checkbox"
                      :checked="topic.selected"
                      @change="toggleBookTopic(chapterIndex, topicIndex)"
                      class="rounded border-slate-300 bg-white text-indigo-500 focus:ring-indigo-500/30"
                    />
                    <span class="text-xs text-foreground flex-1">{{ topic.title }}</span>
                    <span class="text-xs text-muted-foreground/40 shrink-0 tabular-nums">
                      {{ topic.startPage }}–{{ topic.endPage }}p
                    </span>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <DialogFooter class="gap-2">
            <button
              type="button"
              :disabled="bookUploading"
              class="px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground glass-card glass-card-hover transition-all"
              @click="closeBookUpload"
            >
              Болих
            </button>
            <button
              type="button"
              :disabled="!bookUpload.file || bookUploading || bookExtracting"
              class="gradient-indigo text-white px-4 py-2 rounded-xl text-sm font-medium hover:opacity-90 disabled:opacity-50 flex items-center justify-center gap-2"
              @click="submitBookUpload"
            >
              <Loader2Icon v-if="bookUploading" class="w-4 h-4 animate-spin" />
              <template v-else>Оруулах</template>
            </button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>

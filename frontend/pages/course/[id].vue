<script setup>
const route = useRoute()
const courseId = route.params.id

const sidebarOpen = ref(false)
const activeNoteId = ref(null)

const { data: course, refresh } = await useFetch(`/api/course/${courseId}`)

const notes = computed(() =>
  [...(course.value?.notes || [])].sort((a, b) => Number(a.id || 0) - Number(b.id || 0)),
)

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

async function togglePublic() {
  const newVal = !course.value.is_public
  await $fetch(`/api/courses/${courseId}`, {
    method: "PATCH",
    body: { is_public: newVal },
  })
  course.value.is_public = newVal
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
            <div class="flex items-center gap-2">
              <GlobeIcon class="h-4 w-4 shrink-0 text-muted-foreground" />

              <span class="text-xs leading-none text-muted-foreground"> Нийтлэх </span>

              <Switch :checked="course.is_public" @update:checked="togglePublic" class="shrink-0" />
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
          <Tabs default-value="notes" class="space-y-4">
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
                  value="quiz"
                  class="data-[state=active]:bg-white data-[state=active]:text-indigo-700 rounded-lg text-sm"
                >
                  Тест
                </TabsTrigger>
                <TabsTrigger
                  value="chat"
                  class="data-[state=active]:bg-white data-[state=active]:text-indigo-700 rounded-lg text-sm"
                >
                  Chatbot
                </TabsTrigger>
              </TabsList>

              <div id="course-notes-actions" class="flex items-center justify-end gap-2" />
            </div>

            <TabsContent value="notes">
              <CourseNotesTab
                :course="course"
                :active-note-id="activeNoteId"
                @update="handleUpdate"
              />
            </TabsContent>
            <TabsContent value="flashcards">
              <CourseFlashCardsTab
                :course="course"
                :active-note-id="activeNoteId"
                @update="handleUpdate"
              />
            </TabsContent>
            <TabsContent value="quiz">
              <CourseQuizTab
                :course="course"
                :active-note-id="activeNoteId"
                @update="handleUpdate"
              />
            </TabsContent>
            <TabsContent value="chat">
              <CourseChatTab :course="course" />
            </TabsContent>
          </Tabs>
        </div>
      </div>
    </div>
  </div>
</template>

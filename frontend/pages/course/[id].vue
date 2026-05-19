<script setup>
const route = useRoute()
const courseId = route.params.id

const sidebarOpen = ref(false)
const activeTopic = ref(null)

const { data: course, refresh } = await useFetch(`/api/course/${courseId}`)

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
</script>

<template>
  <NavTopBar/>
  <div v-if="!course" class="flex items-center justify-center h-[60vh]">
    <Loader2Icon class="w-8 h-8 text-indigo-400 animate-spin" />
  </div>

  <div v-else class="flex h-[calc(100vh-64px)]">
    <Button
      @click="sidebarOpen = !sidebarOpen"
      class="lg:hidden fixed bottom-4 left-4 z-50 gradient-indigo text-white w-12 h-12 rounded-full flex items-center justify-center shadow-lg"
    >
      <XIcon v-if="sidebarOpen" class="w-5 h-5" />
      <MenuIcon v-else class="w-5 h-5" />
    </button>

    <!-- Sidebar -->
    <aside
      class="fixed lg:static inset-y-0 left-0 top-16 z-40 w-64 glass-card border-r border-white/5 overflow-y-auto transition-transform duration-300"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'"
    >
      <CourseChapterSideBar
        :notes="course.notes || []"
        :active-topic-id="activeTopic"
        @topic-select="
          (id) => {
            activeTopic = id
            sidebarOpen = false
          }
        "
      />
    </aside>

    <!-- Main content -->
    <div class="flex-1 overflow-y-auto">
      <div class="max-w-4xl mx-auto px-4 sm:px-6 py-6">
        <!-- Header -->
        <div class="flex items-start justify-between mb-6 gap-4 flex-wrap">
          <div class="flex items-center gap-3">
            <NuxtLink to="/" class="text-muted-foreground hover:text-foreground transition-colors">
              <ArrowLeftIcon class="w-5 h-5" />
            </NuxtLink>
            <div>
              <h1 class="font-heading text-2xl sm:text-3xl text-foreground">{{ course.title }}</h1>
              <p class="text-xs text-muted-foreground mt-1">
                {{
                  course.status === "completed" ? "Дуусгасан" : `${course.progress || 0}% дууссан`
                }}
              </p>
            </div>
          </div>
          <div class="flex items-center gap-3">
            <div class="flex items-center gap-2">
              <GlobeIcon class="w-4 h-4 text-muted-foreground" />
              <span class="text-xs text-muted-foreground">Нийтлэх</span>
              <Switch :checked="course.is_public" @update:checked="togglePublic" />
            </div>
            <button
              @click="copyLink"
              class="glass-card glass-card-hover p-2 rounded-lg transition-colors"
            >
              <Link2Icon class="w-4 h-4 text-muted-foreground" />
            </button>
          </div>
        </div>

        <!-- Tabs -->
        <Tabs default-value="notes" class="space-y-4">
          <TabsList class="bg-white/5 border border-white/5 p-1 rounded-xl">
            <TabsTrigger
              value="notes"
              class="data-[state=active]:bg-indigo-500/20 data-[state=active]:text-indigo-300 rounded-lg text-sm"
            >
              Тэмдэглэл
            </TabsTrigger>
            <TabsTrigger
              value="flashcards"
              class="data-[state=active]:bg-indigo-500/20 data-[state=active]:text-indigo-300 rounded-lg text-sm"
            >
              Флаш карт
            </TabsTrigger>
            <TabsTrigger
              value="quiz"
              class="data-[state=active]:bg-indigo-500/20 data-[state=active]:text-indigo-300 rounded-lg text-sm"
            >
              Тест
            </TabsTrigger>
            <TabsTrigger
              value="chat"
              class="data-[state=active]:bg-indigo-500/20 data-[state=active]:text-indigo-300 rounded-lg text-sm"
            >
              Chatbot
            </TabsTrigger>
          </TabsList>

          <TabsContent value="notes">
            <CourseNotesTab :course="course" @update="handleUpdate" />
          </TabsContent>
          <TabsContent value="flashcards">
            <CourseFlashCardsTab :course="course" @update="handleUpdate" />
          </TabsContent>
          <TabsContent value="quiz">
            <CourseQuizTab :course="course" @update="handleUpdate" />
          </TabsContent>
          <TabsContent value="chat">
            <CourseChatTab :course="course" />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  </div>
</template>

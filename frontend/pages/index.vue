<script setup>
import { PlusIcon } from "lucide-vue-next"

const showCreate = ref(false)
const editingCourse = ref(null)
const deletingCourse = ref(null)
const savingEdit = ref(false)
const deleting = ref(false)
const courseActionError = ref("")

const editForm = reactive({
  title: "",
  description: "",
  icon: "BookOpen",
})

const {
  data: courses,
  status,
  refresh,
} = await useFetch("/api/course", {
  query: { order_by: "created_at" },
  default: () => ({ items: [], total: 0 }),
})

const courseItems = computed(() => courses.value?.items || [])

const onCreated = async () => {
  showCreate.value = false
  await refresh()
}

function setCourseItems(items, total = courses.value?.total) {
  courses.value = {
    ...(courses.value || { total: 0 }),
    items,
    total: typeof total === "number" ? total : items.length,
  }
}

function openEditDialog(course) {
  editingCourse.value = course
  editForm.title = course.title || ""
  editForm.description = course.description || ""
  editForm.icon = course.icon || "BookOpen"
  courseActionError.value = ""
}

function handleEditOpenChange(open) {
  if (!open && !savingEdit.value) {
    editingCourse.value = null
    courseActionError.value = ""
  }
}

async function saveEdit() {
  if (!editingCourse.value || savingEdit.value) return

  const title = editForm.title.trim()
  if (!title) {
    courseActionError.value = "Хичээлийн нэр оруулна уу"
    return
  }

  savingEdit.value = true
  courseActionError.value = ""
  try {
    const savedCourse = await $fetch(`/api/courses/${editingCourse.value.id}`, {
      method: "PATCH",
      body: {
        title,
        description: editForm.description.trim(),
        icon: editForm.icon,
      },
    })

    setCourseItems(courseItems.value.map((course) => (course.id === savedCourse.id ? savedCourse : course)))
    editingCourse.value = null
  } catch (err) {
    courseActionError.value = err?.data?.message || err?.data || "Хичээл засахад алдаа гарлаа"
  } finally {
    savingEdit.value = false
  }
}

function confirmDelete(course) {
  deletingCourse.value = course
  courseActionError.value = ""
}

function handleDeleteOpenChange(open) {
  if (!open && !deleting.value) {
    deletingCourse.value = null
    courseActionError.value = ""
  }
}

async function deleteSelectedCourse() {
  if (!deletingCourse.value || deleting.value) return

  deleting.value = true
  courseActionError.value = ""
  try {
    await $fetch(`/api/courses/${deletingCourse.value.id}`, { method: "DELETE" })
    const nextItems = courseItems.value.filter((course) => course.id !== deletingCourse.value.id)
    setCourseItems(nextItems, Math.max((courses.value?.total || courseItems.value.length) - 1, 0))
    deletingCourse.value = null
  } catch (err) {
    courseActionError.value = err?.data?.message || err?.data || "Хичээл устгахад алдаа гарлаа"
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <NavTopBar />
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="font-heading text-3xl sm:text-4xl text-foreground">Миний хичээлүүд</h1>
        <!-- <p class="text-sm text-muted-foreground mt-1">AI-ийн тусламжтай суралцах платформ</p> -->
      </div>
      <button
        @click="showCreate = true"
        class="gradient-indigo text-white px-5 py-2.5 rounded-xl font-medium text-sm flex items-center gap-2 hover:opacity-90 transition-opacity shadow-lg shadow-indigo-500/20"
      >
        <PlusIcon class="w-4 h-4" />
        <span class="hidden sm:inline">Шинэ хичээл үүсгэх</span>
        <span class="sm:hidden">Шинэ</span>
      </button>
    </div>

    <!-- Loading skeleton -->
    <div v-if="status === 'pending'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="i in 3" :key="i" class="glass-card rounded-xl p-5 h-48 animate-pulse">
        <div class="h-4 bg-slate-200 rounded w-2/3 mb-4" />
        <div class="h-2 bg-slate-200 rounded w-full mb-6" />
        <div class="h-3 bg-slate-200 rounded w-1/3" />
      </div>
    </div>

    <!-- Empty state -->
    <CourseEmptyState v-else-if="courseItems.length === 0" @create="showCreate = true" />

    <!-- Course grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <CourseCard
        v-for="course in courseItems"
        :key="course.id"
        :course="course"
        @edit="openEditDialog"
        @delete="confirmDelete"
      />
    </div>

    <CourseCreateModal :open="showCreate" @close="showCreate = false" @created="onCreated" />

    <Dialog :open="!!editingCourse" @update:open="handleEditOpenChange">
      <DialogContent class="glass-card border-slate-200 sm:max-w-md">
        <DialogHeader>
          <DialogTitle class="font-heading text-xl text-foreground">Хичээл засах</DialogTitle>
          <DialogDescription></DialogDescription>
        </DialogHeader>

        <p v-if="courseActionError" class="text-sm text-red-600">{{ courseActionError }}</p>

        <div class="space-y-3">
          <Input
            v-model="editForm.title"
            placeholder="Хичээлийн нэр"
            class="bg-white border-slate-200 text-foreground"
          />
          <Input
            v-model="editForm.description"
            placeholder="Хичээлийн тайлбар"
            class="bg-white border-slate-200 text-foreground"
          />
        </div>

        <DialogFooter class="gap-2">
          <button
            type="button"
            class="px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground glass-card glass-card-hover transition-all"
            :disabled="savingEdit"
            @click="handleEditOpenChange(false)"
          >
            Болих
          </button>
          <button
            type="button"
            class="gradient-indigo text-white px-4 py-2 rounded-xl text-sm font-medium hover:opacity-90 disabled:opacity-50 flex items-center justify-center gap-2"
            :disabled="savingEdit || !editForm.title.trim()"
            @click="saveEdit"
          >
            <Loader2Icon v-if="savingEdit" class="w-4 h-4 animate-spin" />
            <template v-else>Хадгалах</template>
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="!!deletingCourse" @update:open="handleDeleteOpenChange">
      <DialogContent class="glass-card border-slate-200 sm:max-w-md">
        <DialogHeader>
          <DialogTitle class="font-heading text-xl text-foreground">Хичээл устгах</DialogTitle>
          <DialogDescription>
            {{ deletingCourse?.title }} хичээлийг устгах уу?
          </DialogDescription>
        </DialogHeader>

        <p v-if="courseActionError" class="text-sm text-red-600">{{ courseActionError }}</p>

        <DialogFooter class="gap-2">
          <button
            type="button"
            class="px-4 py-2 rounded-xl text-sm text-muted-foreground hover:text-foreground glass-card glass-card-hover transition-all"
            :disabled="deleting"
            @click="handleDeleteOpenChange(false)"
          >
            Болих
          </button>
          <button
            type="button"
            class="bg-red-600 text-white px-4 py-2 rounded-xl text-sm font-medium hover:bg-red-700 disabled:opacity-50 flex items-center justify-center gap-2"
            :disabled="deleting"
            @click="deleteSelectedCourse"
          >
            <Loader2Icon v-if="deleting" class="w-4 h-4 animate-spin" />
            <template v-else>Устгах</template>
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

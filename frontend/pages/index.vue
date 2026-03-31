<script setup>
import { PlusIcon } from "lucide-vue-next"

const showCreate = ref(false)

const {
  data: courses,
  status,
  refresh,
} = await useFetch("/api/course", {
  query: { order_by: "created_at" },
  default: () => [],
})

const onCreated = () => {
  showCreate.value = false
  refresh()
}

const confirmDelete = () => {}

const openEditDialog = () => {}
</script>

<template>
  <NavTopBar />
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="font-heading text-3xl sm:text-4xl text-foreground">Миний хичээлүүд</h1>
        <p class="text-sm text-muted-foreground mt-1">AI-ийн тусламжтай суралцах платформ</p>
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
        <div class="h-4 bg-white/5 rounded w-2/3 mb-4" />
        <div class="h-2 bg-white/5 rounded w-full mb-6" />
        <div class="h-3 bg-white/5 rounded w-1/3" />
      </div>
    </div>

    <!-- Empty state -->
    <CourseEmptyState v-else-if="courses.items.length === 0" @create="showCreate = true" />

    <!-- Course grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <CourseCard
        v-for="course in courses.items"
        :key="course.id"
        :course="course"
        @edit="openEditDialog"
        @delete="confirmDelete"
      />
    </div>

    <CourseCreateModal :open="showCreate" @close="showCreate = false" @created="onCreated" />
  </div>
</template>

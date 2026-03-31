<script setup>
import { BookOpenIcon, SparklesIcon } from "lucide-vue-next"

const emit = defineEmits(["create"])
</script>

<template>
  <div class="flex flex-col items-center justify-center py-24 px-4">
    <div class="relative mb-8">
      <div class="w-24 h-24 rounded-2xl glass-card flex items-center justify-center">
        <BookOpenIcon class="w-10 h-10 text-muted-foreground" />
      </div>
      <div
        class="absolute -top-2 -right-2 w-8 h-8 rounded-lg gradient-indigo flex items-center justify-center"
      >
        <SparklesIcon class="w-4 h-4 text-white" />
      </div>
    </div>

    <h3 class="font-heading text-2xl text-foreground mb-2">Хичээл байхгүй байна</h3>
    <p class="text-muted-foreground text-sm text-center max-w-sm mb-8">
      Эхний хичээлээ үүсгэж, AI-ийн тусламжтай суралцаж эхлээрэй
    </p>

    <button
      @click="emit('create')"
      class="gradient-indigo text-white px-6 py-3 rounded-xl font-medium text-sm flex items-center gap-2 hover:opacity-90 transition-opacity"
    >
      Шинэ хичээл үүсгэх
      <span>→</span>
    </button>
  </div>
</template>

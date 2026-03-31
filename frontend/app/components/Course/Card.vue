<script setup>
const props = defineProps({
  course: { type: Object, required: true },
})

const isCompleted = computed(() => props.course.status === "completed")
</script>

<template>
  <div
    class="glass-card glass-card-hover rounded-xl p-5 flex flex-col gap-4 transition-all duration-300 group"
  >
    <!-- Top row -->
    <div class="flex items-start justify-between">
      <div class="flex items-center gap-3">
        <div
          class="w-10 h-10 rounded-lg flex items-center justify-center"
          :class="isCompleted ? 'gradient-teal' : 'gradient-indigo'"
        >
          <BookOpenIcon class="w-5 h-5 text-white" />
        </div>
        <div>
          <h3 class="font-heading text-lg text-foreground leading-tight">{{ course.title }}</h3>
          <p v-if="course.description" class="text-xs text-muted-foreground mt-0.5 line-clamp-1">
            {{ course.description }}
          </p>
        </div>
      </div>
      <span
        class="px-2.5 py-1 rounded-full text-xs font-medium whitespace-nowrap"
        :class="
          isCompleted
            ? 'bg-teal-500/10 text-teal-400 border border-teal-500/20'
            : 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20'
        "
      >
        {{ isCompleted ? "Дуусгасан" : "Үргэлжилж байна" }}
      </span>
    </div>

    <!-- Progress -->
    <div class="space-y-2">
      <div class="flex justify-between text-xs text-muted-foreground">
        <span>Явц</span>
        <span>{{ course.progress || 0 }}%</span>
      </div>
      <Progress :value="course.progress || 0" class="h-1.5 bg-white/5" />
    </div>

    <!-- Footer -->
    <div class="flex items-center justify-between mt-auto pt-2">
      <div class="flex items-center gap-1.5 text-xs text-muted-foreground">
        <CalendarIcon class="w-3.5 h-3.5" />
        {{ course.created_date }}
      </div>
      <NuxtLink
        :to="`/course/${course.id}`"
        class="flex items-center gap-1.5 text-sm font-medium text-indigo-400 hover:text-indigo-300 transition-colors group-hover:gap-2.5"
      >
        Үргэлжлүүлэх
        <ArrowRightIcon class="w-4 h-4 transition-transform group-hover:translate-x-0.5" />
      </NuxtLink>
    </div>
  </div>
</template>

<style lang="css">
.glass-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(20px);
}
.glass-card-hover:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
}
</style>

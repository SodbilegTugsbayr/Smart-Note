<script setup>
const props = defineProps({
  course: { type: Object, required: true },
})

const emit = defineEmits(["edit", "delete"])
const isCompleted = computed(() => props.course.status === "completed")
</script>

<template>
  <NuxtLink :to="`/course/${course.id}`">
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
            <component :is="`${course.icon}Icon`" class="w-5 h-5" />
          </div>
          <div>
            <p class="font-heading text-lg text-foreground leading-tight">
              {{ course.title.length > 10 ? `${course.title.slice(0, 10)}...` : course.title }}
            </p>
            <p v-if="course.description" class="text-xs text-muted-foreground mt-0.5 line-clamp-1">
              {{
                course.description.length > 20
                  ? `${course.description.slice(0, 20)}...`
                  : course.description
              }}
            </p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <span
            class="px-2.5 py-1 rounded-full text-xs font-medium whitespace-nowrap"
            :class="
              isCompleted
                ? 'bg-teal-500/10 text-teal-400 border border-teal-500/20'
                : 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20'
            "
          >
            {{ isCompleted ? "Дууссан" : "Үргэлжилж байгаа" }}
          </span>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <button class="p-1 rounded-md hover:bg-white/10 transition">
                <EllipsisVerticalIcon class="w-5 h-5" />
              </button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end" class="w-44">
              <DropdownMenuItem @click="emit('edit', course)"> Засах </DropdownMenuItem>

              <DropdownMenuSeparator />

              <DropdownMenuItem
                @click="emit('delete', course)"
                class="text-red-400 focus:text-red-400"
              >
                Устгах
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
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
          {{ $fullDate(course.created_at) }}
        </div>
      </div>
    </div>
  </NuxtLink>
</template>

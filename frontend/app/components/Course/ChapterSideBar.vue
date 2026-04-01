<script setup>
const props = defineProps({
  notes: { type: Array, default: () => [] },
  activeTopicId: { type: [String, Number], default: null },
})

const emit = defineEmits(["topic-select"])

const expandedSections = ref(props.notes.map((n) => n.id))

function toggleSection(id) {
  if (expandedSections.value.includes(id)) {
    expandedSections.value = expandedSections.value.filter((i) => i !== id)
  } else {
    expandedSections.value.push(id)
  }
}
</script>

<template>
  <div class="p-3">
    <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider px-2 mb-3">
      Агуулга
    </p>

    <div v-if="notes.length === 0" class="text-center py-4">
      <p class="text-xs text-muted-foreground">Бүлэг байхгүй</p>
    </div>

    <div v-else class="space-y-1">
      <div v-for="note in notes" :key="note.id">
        <button
          @click="toggleSection(note.id)"
          class="w-full flex items-center gap-2 px-2 py-2 rounded-lg text-sm text-foreground/80 hover:bg-white/5 transition-colors"
        >
          <ChevronRightIcon
            class="w-3.5 h-3.5 text-muted-foreground transition-transform flex-shrink-0"
            :class="expandedSections.includes(note.id) ? 'rotate-90' : ''"
          />
          <BookOpenIcon class="w-3.5 h-3.5 text-indigo-400 flex-shrink-0" />
          <span class="truncate text-left flex-1 text-xs">{{ note.title }}</span>
        </button>

        <div v-if="expandedSections.includes(note.id)" class="ml-5 space-y-0.5 mt-0.5">
          <button
            v-for="concept in note.key_concepts || []"
            :key="concept.concept"
            @click="emit('topic-select', note.id)"
            class="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-xs transition-colors text-left"
            :class="
              activeTopicId === note.id
                ? 'bg-indigo-500/10 text-indigo-400'
                : 'text-muted-foreground hover:text-foreground hover:bg-white/5'
            "
          >
            <FileTextIcon class="w-3 h-3 flex-shrink-0" />
            <span class="truncate">{{ concept.concept }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

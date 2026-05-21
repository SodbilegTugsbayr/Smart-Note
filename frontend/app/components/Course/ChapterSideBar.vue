<script setup>
defineProps({
  notes: { type: Array, default: () => [] },
  activeNoteId: { type: [String, Number], default: null },
})

const emit = defineEmits(["note-select"])
</script>

<template>
  <div class="p-3">
    <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider px-2 mb-3">
      Тэмдэглэл
    </p>

    <div v-if="notes.length === 0" class="text-center py-4">
      <p class="text-xs text-muted-foreground">Тэмдэглэл байхгүй</p>
    </div>

    <div v-else class="space-y-1">
      <button
        v-for="note in notes"
        :key="note.id"
        @click="emit('note-select', note.id)"
        class="w-full flex items-center gap-2 px-2 py-2 rounded-lg text-sm transition-colors text-left"
        :class="
          activeNoteId === note.id
            ? 'bg-indigo-500/10 text-indigo-700'
            : 'text-foreground/80 hover:bg-slate-100'
        "
      >
        <FileTextIcon class="w-3.5 h-3.5 flex-shrink-0" />
        <span class="truncate flex-1 text-xs">{{ note.title || "Гарчиггүй тэмдэглэл" }}</span>
      </button>
    </div>
  </div>
</template>

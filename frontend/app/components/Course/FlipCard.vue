<script setup>
defineProps({
  term: { type: String, required: true },
  definition: { type: String, required: true },
})

const flipped = ref(false)
</script>

<template>
  <div class="cursor-pointer h-44" style="perspective: 1000px" @click="flipped = !flipped">
    <div
      class="relative w-full h-full transition-transform duration-500"
      :style="{
        transformStyle: 'preserve-3d',
        transform: flipped ? 'rotateY(180deg)' : 'rotateY(0deg)',
      }"
    >
      <div
        class="absolute inset-0 glass-card glass-card-hover rounded-xl p-5 flex flex-col items-center justify-center"
        style="backface-visibility: hidden"
      >
        <p class="text-xs text-muted-foreground mb-2 uppercase tracking-wider">Асуулт</p>
        <p class="text-base font-medium text-foreground text-center">{{ term }}</p>
        <RotateCcwIcon class="w-3.5 h-3.5 text-muted-foreground/40 absolute bottom-3 right-3" />
      </div>

      <div
        class="absolute inset-0 glass-card rounded-xl p-5 flex flex-col items-center justify-center border-indigo-500/20"
        style="
          backface-visibility: hidden;
          transform: rotateY(180deg);
          background: rgba(99, 102, 241, 0.05);
        "
      >
        <p class="text-xs text-indigo-600 mb-2 uppercase tracking-wider">Хариулт</p>
        <p class="text-sm text-foreground/80 text-center">{{ definition }}</p>
      </div>
    </div>
  </div>
</template>

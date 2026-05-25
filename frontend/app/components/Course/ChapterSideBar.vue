<script setup>
const props = defineProps({
  courseId: { type: [String, Number], required: true },
  notes: { type: Array, default: () => [] },
  activeNoteId: { type: [String, Number], default: null },
  readonly: { type: Boolean, default: false },
})

const emit = defineEmits(["note-select", "note-created", "note-updated", "note-deleted"])

const adding = ref(false)
const savingEdit = ref(false)
const deleting = ref(false)
const errorMessage = ref("")
const editingNote = ref(null)
const deletingNote = ref(null)
const editTitle = ref("")

const sortedNotes = computed(() =>
  [...props.notes].sort((a, b) => Number(a.id || 0) - Number(b.id || 0)),
)

async function handleAddNote() {
  if (props.readonly) return

  adding.value = true
  errorMessage.value = ""

  try {
    const nextNumber = props.notes.length + 1
    const savedNote = await $fetch(`/api/courses/${props.courseId}/notes`, {
      method: "POST",
      body: { title: `Шинэ тэмдэглэл ${nextNumber}` },
    })
    emit("note-created", savedNote)
    emit("note-select", savedNote.id)
  } catch (err) {
    errorMessage.value = err?.data?.message || "Тэмдэглэл нэмэхэд алдаа гарлаа"
  } finally {
    adding.value = false
  }
}

function openEditNote(note) {
  if (props.readonly) return

  editingNote.value = note
  editTitle.value = note.title || ""
  errorMessage.value = ""
}

function handleEditOpenChange(open) {
  if (!open && !savingEdit.value) {
    editingNote.value = null
    editTitle.value = ""
    errorMessage.value = ""
  }
}

async function handleSaveEdit() {
  if (props.readonly || !editingNote.value || savingEdit.value) return

  const title = editTitle.value.trim()
  if (!title) {
    errorMessage.value = "Гарчиг оруулна уу"
    return
  }

  savingEdit.value = true
  errorMessage.value = ""
  try {
    const savedNote = await $fetch(`/api/notes/${editingNote.value.id}`, {
      method: "PATCH",
      body: { title },
    })
    emit("note-updated", savedNote)
    editingNote.value = null
    editTitle.value = ""
  } catch (err) {
    errorMessage.value = err?.data?.message || err?.data || "Тэмдэглэл засахад алдаа гарлаа"
  } finally {
    savingEdit.value = false
  }
}

function openDeleteNote(note) {
  if (props.readonly) return

  deletingNote.value = note
  errorMessage.value = ""
}

function handleDeleteOpenChange(open) {
  if (!open && !deleting.value) {
    deletingNote.value = null
    errorMessage.value = ""
  }
}

async function handleDeleteNote() {
  if (props.readonly || !deletingNote.value || deleting.value) return

  deleting.value = true
  errorMessage.value = ""
  try {
    await $fetch(`/api/notes/${deletingNote.value.id}`, { method: "DELETE" })
    emit("note-deleted", deletingNote.value.id)
    deletingNote.value = null
  } catch (err) {
    errorMessage.value = err?.data?.message || err?.data || "Тэмдэглэл устгахад алдаа гарлаа"
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div class="p-3">
    <div class="flex items-center justify-between gap-2 px-2 mb-3">
      <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">
        Тэмдэглэл
      </p>
      <button
        v-if="!readonly"
        @click="handleAddNote"
        :disabled="adding"
        class="w-7 h-7 rounded-lg text-muted-foreground hover:text-indigo-700 hover:bg-indigo-500/10 transition-colors flex items-center justify-center disabled:opacity-50"
        title="Тэмдэглэл нэмэх"
      >
        <Loader2Icon v-if="adding" class="w-3.5 h-3.5 animate-spin" />
        <PlusIcon v-else class="w-3.5 h-3.5" />
      </button>
    </div>

    <p v-if="errorMessage" class="text-xs text-red-600 px-2 mb-2">{{ errorMessage }}</p>

    <div v-if="sortedNotes.length === 0" class="text-center py-4">
      <p class="text-xs text-muted-foreground">Тэмдэглэл байхгүй</p>
    </div>

    <div v-else class="space-y-1">
      <div
        v-for="note in sortedNotes"
        :key="note.id"
        class="w-full flex items-center gap-1 rounded-lg text-sm transition-colors"
        :class="
          activeNoteId === note.id
            ? 'bg-indigo-500/10 text-indigo-700'
            : 'text-foreground/80 hover:bg-slate-100'
        "
      >
        <button
          @click="emit('note-select', note.id)"
          class="min-w-0 flex-1 flex items-center gap-2 px-2 py-2 text-left"
        >
          <FileTextIcon class="w-3.5 h-3.5 flex-shrink-0" />
          <span class="truncate flex-1 text-xs">{{ note.title || "Гарчиггүй тэмдэглэл" }}</span>
        </button>

        <DropdownMenu v-if="!readonly">
          <DropdownMenuTrigger as-child>
            <button
              class="w-7 h-7 mr-1 rounded-md flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-white/70 transition-colors"
              title="Тэмдэглэлийн үйлдэл"
              @click.stop="() => {}"
            >
              <EllipsisVerticalIcon class="w-3.5 h-3.5" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" class="w-36">
            <DropdownMenuItem @click="openEditNote(note)"> Засах </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              class="text-red-600 focus:text-red-600"
              @click="openDeleteNote(note)"
            >
              Устгах
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>

    <Dialog :open="!!editingNote" @update:open="handleEditOpenChange">
      <DialogContent class="glass-card border-slate-200 sm:max-w-sm">
        <DialogHeader>
          <DialogTitle class="font-heading text-lg text-foreground">Тэмдэглэл засах</DialogTitle>
          <DialogDescription></DialogDescription>
        </DialogHeader>

        <Input
          v-model="editTitle"
          placeholder="Тэмдэглэлийн гарчиг"
          class="bg-white border-slate-200 text-foreground"
        />

        <DialogFooter class="gap-2">
          <button
            type="button"
            :disabled="savingEdit"
            class="px-3 py-2 rounded-lg text-xs text-muted-foreground hover:text-foreground glass-card glass-card-hover transition-all"
            @click="handleEditOpenChange(false)"
          >
            Болих
          </button>
          <button
            type="button"
            :disabled="savingEdit || !editTitle.trim()"
            class="gradient-indigo text-white px-3 py-2 rounded-lg text-xs font-medium hover:opacity-90 disabled:opacity-50 flex items-center justify-center gap-2"
            @click="handleSaveEdit"
          >
            <Loader2Icon v-if="savingEdit" class="w-3.5 h-3.5 animate-spin" />
            <template v-else>Хадгалах</template>
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="!!deletingNote" @update:open="handleDeleteOpenChange">
      <DialogContent class="glass-card border-slate-200 sm:max-w-sm">
        <DialogHeader>
          <DialogTitle class="font-heading text-lg text-foreground">Тэмдэглэл устгах</DialogTitle>
          <DialogDescription>
            {{ deletingNote?.title || "Гарчиггүй тэмдэглэл" }} тэмдэглэлийг устгах уу?
          </DialogDescription>
        </DialogHeader>

        <DialogFooter class="gap-2">
          <button
            type="button"
            :disabled="deleting"
            class="px-3 py-2 rounded-lg text-xs text-muted-foreground hover:text-foreground glass-card glass-card-hover transition-all"
            @click="handleDeleteOpenChange(false)"
          >
            Болих
          </button>
          <button
            type="button"
            :disabled="deleting"
            class="bg-red-600 text-white px-3 py-2 rounded-lg text-xs font-medium hover:bg-red-700 disabled:opacity-50 flex items-center justify-center gap-2"
            @click="handleDeleteNote"
          >
            <Loader2Icon v-if="deleting" class="w-3.5 h-3.5 animate-spin" />
            <template v-else>Устгах</template>
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup>
const user = useUser()

const lessons = ref([
  {
    title: "Веб хөгжүүлэлтийн үндэс",
    status: "In progress",
    progress: 65,
    lastViewed: "2024-08-25",
    description: "HTML, CSS, JavaScript суурь ойлголтууд.",
  },
  {
    title: "Алгоритм ба өгөгдлийн бүтэц",
    status: "Completed",
    progress: 100,
    lastViewed: "2024-08-20",
    description: "Эрэмбэлэлт, хайлт, граф алгоритмууд.",
  },
  {
    title: "Өгөгдлийн сангийн загварчлал",
    status: "In progress",
    progress: 42,
    lastViewed: "2024-08-22",
    description: "SQL асуулга, нормалчлал, индекс.",
  },
  {
    title: "Кибер аюулгүй байдлын үндэс",
    status: "Not started",
    progress: 0,
    lastViewed: "—",
    description: "Эрсдэл, халдлагын төрөл, OWASP Top 10.",
  },
])

const summary = computed(() => {
  const total = lessons.value.length
  const completed = lessons.value.filter((l) => l.status === "Completed").length
  const inProgress = lessons.value.filter((l) => l.status === "In progress").length
  const averageProgress =
    lessons.value.reduce((acc, cur) => acc + cur.progress, 0) / Math.max(total, 1)

  return { total, completed, inProgress, averageProgress: Math.round(averageProgress) }
})

function statusColor(status) {
  if (status === "Completed") return "success"
  if (status === "In progress") return "warning"
  return "default"
}

async function logout() {
  try {
    await $fetch("/pub/logout", { method: "GET", credentials: "include" })
  } catch (e) {
    // ignore error and proceed with logout UX
  }
  const u = useUser()
  u.value = null
  await navigateTo("/landing")
}
</script>

<template>
  <v-container class="py-8">
    <v-row class="mb-6" align="center" justify="space-between">
      <v-col cols="12" md="8">
        <div class="text-h4 font-weight-bold mb-2">Таны хичээлүүд</div>
        <div class="text-body-1 text-medium-emphasis">
          Сайн уу, {{ user?.value?.name || "сургагч" }}! Хичээлүүдийн явцаа эндээс хянаарай.
        </div>
      </v-col>
      <v-col cols="12" md="4" class="d-flex align-end flex-column ga-3">
        <v-card class="pa-4" elevation="4">
          <div class="text-caption text-medium-emphasis mb-1">Дундаж явц</div>
          <div class="text-h5 font-weight-bold mb-2">{{ summary.averageProgress }}%</div>
          <v-progress-linear
            :model-value="summary.averageProgress"
            color="primary"
            height="10"
            rounded
          />
        </v-card>
        <v-btn color="error" variant="text" @click="logout"> Гарах </v-btn>
      </v-col>
    </v-row>

    <v-row class="mb-4" dense>
      <v-col cols="12" sm="4">
        <v-card class="pa-4" variant="tonal">
          <div class="text-caption text-medium-emphasis">Нийт хичээл</div>
          <div class="text-h5 font-weight-bold">{{ summary.total }}</div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="4">
        <v-card class="pa-4" variant="tonal" color="primary">
          <div class="text-caption">Явцад</div>
          <div class="text-h5 font-weight-bold">{{ summary.inProgress }}</div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="4">
        <v-card class="pa-4" variant="tonal" color="success">
          <div class="text-caption">Дууссан</div>
          <div class="text-h5 font-weight-bold">{{ summary.completed }}</div>
        </v-card>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12" md="6" lg="4" v-for="(lesson, idx) in lessons" :key="idx">
        <v-card class="pa-4 h-100" elevation="6">
          <div class="d-flex align-center justify-space-between mb-2">
            <div class="text-subtitle-1 font-weight-bold">{{ lesson.title }}</div>
            <v-chip :color="statusColor(lesson.status)" label size="small" variant="flat">
              {{ lesson.status }}
            </v-chip>
          </div>
          <div class="text-body-2 text-medium-emphasis mb-4">
            {{ lesson.description }}
          </div>

          <div class="mb-2 d-flex justify-space-between align-center">
            <span class="text-caption text-medium-emphasis">Уншсан хувь</span>
            <span class="text-caption font-weight-medium">{{ lesson.progress }}%</span>
          </div>
          <v-progress-linear
            :model-value="lesson.progress"
            height="8"
            color="primary"
            rounded
            class="mb-3"
          />

          <div class="text-caption text-medium-emphasis">
            Сүүлд үзсэн: <strong>{{ lesson.lastViewed }}</strong>
          </div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

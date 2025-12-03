<script setup lang="ts">
const router = useRouter()

function goToLogin() {
  router.push("/login")
}

function goToSignup() {
  router.push("/signup")
}
</script>

<template>
  <v-app-bar flat color="transparent" density="comfortable">
    <v-toolbar-title class="text-h6 font-weight-bold">Smart Note</v-toolbar-title>
    <v-spacer />
    <v-btn variant="text" @click="goToLogin">Нэвтрэх</v-btn>
    <v-btn color="primary" class="ml-2" @click="goToSignup">Бүртгүүлэх</v-btn>
  </v-app-bar>

  <v-container class="fill-height" fluid>
    <v-row class="py-12 px-4 px-md-12" align="center" justify="center">
      <v-col cols="12" md="6">
        <div class="text-h3 font-weight-bold mb-4">
          Хичээлээ цэгцтэй удирдаж, ахиц дэвшлээ хянана.
        </div>
        <div class="text-body-1 text-medium-emphasis mb-6">
          Smart Note нь таны хичээлүүдийг нэг дор эмхлэн үзүүлж, явцыг хэмжихэд тусална. Өнөөдрөөс
          эхлээд зорилгоо биелүүлээрэй.
        </div>
        <div class="d-flex flex-wrap ga-3">
          <v-btn color="primary" size="large" @click="goToSignup">Бүртгүүлэх</v-btn>
          <v-btn variant="outlined" size="large" color="primary" @click="goToLogin">
            Нэвтрэх
          </v-btn>
        </div>
      </v-col>

      <v-col cols="12" md="5">
        <v-card elevation="12" class="pa-6">
          <v-card-title class="text-h6 font-weight-bold"> Яагаад Smart Note? </v-card-title>
          <v-card-text>
            <v-list density="comfortable">
              <v-list-item
                v-for="(item, i) in [
                  'Хичээлүүдийн явц, уншсан хувь, сүүлд үзсэн огноог нэг дор.',
                  'Ойлгомжтой интерфэйс, хурдан нэвтрэлт.',
                  'Таны ахиц дэвшлийг хэмжих визуал график.',
                ]"
                :key="i"
                prepend-icon="mdi-check-circle"
              >
                <v-list-item-title>{{ item }}</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

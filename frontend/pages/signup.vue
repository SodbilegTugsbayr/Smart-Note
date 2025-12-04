<script setup>
const router = useRouter()
const loading = ref(false)
const error = ref("")

const form = reactive({
  name: "",
  email: "",
  password: "",
})

async function submit() {
  loading.value = true
  error.value = ""
  try {
    await $fetch("/pub/auth/signup", {
      method: "POST",
      body: form,
      credentials: "include",
    })
    router.push("/")
  } catch (err) {
    error.value = err?.data || err?.message || "Бүртгүүлэхэд алдаа гарлаа"
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="5" lg="4">
        <v-card class="pa-6" elevation="8">
          <v-card-title class="text-h5 text-center">Шинээр бүртгэл үүсгэх</v-card-title>
          <v-card-subtitle class="text-center mb-4">
            Мэйл ээ ашиглан бүртгүүлнэ үү.
          </v-card-subtitle>

          <v-alert v-if="error" type="error" variant="tonal" class="mb-4" border="start">
            {{ error }}
          </v-alert>

          <v-form @submit.prevent="submit">
            <v-text-field
              v-model="form.name"
              label="Овог нэр"
              required
              variant="outlined"
              class="mb-3"
            />
            <v-text-field
              v-model="form.email"
              label="Мэйл"
              type="email"
              required
              variant="outlined"
              class="mb-3"
            />
            <v-text-field
              v-model="form.password"
              label="Нууц үг"
              type="password"
              required
              variant="outlined"
              class="mb-4"
              hint="Доод тал нь 8 тэмдэгт"
              persistent-hint
            />

            <v-btn color="primary" block size="large" :loading="loading" type="submit">
              Бүртгүүлэх
            </v-btn>
          </v-form>

          <div class="d-flex justify-center mt-4">
            <span class="me-1">Бүртгэлтэй юу?</span>
            <NuxtLink to="/login">Нэвтрэх</NuxtLink>
          </div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

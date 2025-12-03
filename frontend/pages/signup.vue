<script setup lang="ts">
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
  } catch (err: any) {
    error.value = err?.data || err?.message || "Sign up failed"
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
          <v-card-title class="text-h5 text-center">Create account</v-card-title>
          <v-card-subtitle class="text-center mb-4">
            Join with your email and start using Smart Note.
          </v-card-subtitle>

          <v-alert
            v-if="error"
            type="error"
            variant="tonal"
            class="mb-4"
            border="start"
          >
            {{ error }}
          </v-alert>

          <v-form @submit.prevent="submit">
            <v-text-field
              v-model="form.name"
              label="Full name"
              required
              variant="outlined"
              class="mb-3"
            />
            <v-text-field
              v-model="form.email"
              label="Email"
              type="email"
              required
              variant="outlined"
              class="mb-3"
            />
            <v-text-field
              v-model="form.password"
              label="Password"
              type="password"
              required
              variant="outlined"
              class="mb-4"
              hint="At least 8 characters"
              persistent-hint
            />

            <v-btn
              color="primary"
              block
              size="large"
              :loading="loading"
              type="submit"
            >
              Sign up
            </v-btn>
          </v-form>

          <div class="d-flex justify-center mt-4">
            <span class="me-1">Already have an account?</span>
            <NuxtLink to="/login">Log in</NuxtLink>
          </div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

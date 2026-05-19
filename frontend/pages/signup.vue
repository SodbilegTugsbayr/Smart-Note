<script setup>
const router = useRouter()
const loading = ref(false)
const error = ref("")

const form = reactive({
  firstname: "",
  lastname: "",
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
  <div class="login-root">
    <div class="bg-orb bg-orb-1" />
    <div class="bg-orb bg-orb-2" />
    <div class="bg-grid" />

    <div class="login-center">
      <NuxtLink to="/" class="brand-link">
        <span class="brand-icon">✦</span>
        Smart Note
      </NuxtLink>

      <Card class="login-card">
        <CardHeader class="card-header">
          <CardTitle class="card-title">Шинэ бүртгэл</CardTitle>
          <CardDescription class="card-subtitle"> Мэйл ээ ашиглан бүртгүүлнэ үү. </CardDescription>
        </CardHeader>

        <CardContent>
          <Alert v-if="error" variant="destructive" class="alert-error">
            <AlertDescription class="flex items-center gap-2">
              <span>⚠</span>
              <span>{{ error }}</span>
            </AlertDescription>
          </Alert>

          <form @submit.prevent="submit" class="login-form">
            <div class="field">
              <Label for="firstname" class="field-label">Овог</Label>
              <Input
                id="name"
                v-model="form.firstname"
                type="text"
                required
                placeholder="Бат"
                class="field-input"
                :class="{ 'field-input--error': error }"
                autocomplete="name"
              />
            </div>
            <div class="field">
              <Label for="lastname" class="field-label">Нэр</Label>
              <Input
                id="name"
                v-model="form.lastname"
                type="text"
                required
                placeholder="Болд"
                class="field-input"
                :class="{ 'field-input--error': error }"
                autocomplete="name"
              />
            </div>
            <div class="field">
              <Label for="email" class="field-label">Мэйл</Label>
              <Input
                id="email"
                v-model="form.email"
                required
                placeholder="you@example.com"
                class="field-input"
                :class="{ 'field-input--error': error }"
                autocomplete="email"
              />
            </div>

            <div class="field">
              <Label for="password" class="field-label">Нууц үг</Label>
              <Input
                id="password"
                v-model="form.password"
                type="password"
                required
                placeholder="••••••••"
                class="field-input"
                :class="{ 'field-input--error': error }"
                autocomplete="new-password"
              />
              <p class="field-hint">Доод тал нь 8 тэмдэгт</p>
            </div>

            <Button type="submit" :disabled="loading" class="btn-submit">
              <span v-if="!loading">
                Бүртгүүлэх
                <span class="btn-arrow">→</span>
              </span>
              <span v-else class="btn-loading">
                <span class="spinner" />
                Түр хүлээнэ үү...
              </span>
            </Button>
          </form>
        </CardContent>

        <CardFooter class="card-footer">
          <p>
            Бүртгэлтэй юу?
            <NuxtLink to="/login" class="footer-link">Нэвтрэх</NuxtLink>
          </p>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<style scoped>
.login-root {
  min-height: 100vh;
  background: #f8fafc;
  color: #0f172a;
  font-family: "DM Sans", sans-serif;
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bg-orb {
  position: fixed;
  border-radius: 50%;
  filter: blur(120px);
  pointer-events: none;
  z-index: 0;
}
.bg-orb-1 {
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.12) 0%, transparent 70%);
  top: -150px;
  right: -100px;
}
.bg-orb-2 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, rgba(20, 184, 166, 0.1) 0%, transparent 70%);
  bottom: -100px;
  left: -80px;
}
.bg-grid {
  position: fixed;
  inset: 0;
  background-image:
    linear-gradient(rgba(15, 23, 42, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.04) 1px, transparent 1px);
  background-size: 48px 48px;
  z-index: 0;
  pointer-events: none;
}

.login-center {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 28px;
  animation: fadeUp 0.5s ease both;
}

.brand-link {
  font-family: "Instrument Serif", serif;
  font-size: 1.3rem;
  letter-spacing: -0.02em;
  color: #0f172a;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: opacity 0.2s;
}
.brand-link:hover {
  opacity: 0.75;
}
.brand-icon {
  color: #818cf8;
  font-size: 0.9rem;
}

.login-card {
  width: 100%;
  background: rgba(255, 255, 255, 0.92) !important;
  border: 1px solid rgba(15, 23, 42, 0.08) !important;
  border-radius: 20px !important;
  backdrop-filter: blur(20px);
  box-shadow:
    0 32px 64px rgba(15, 23, 42, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.8) !important;
}

.card-header {
  text-align: center;
  padding-bottom: 0 !important;
}
.card-title {
  font-family: "Instrument Serif", serif !important;
  font-size: 1.9rem !important;
  letter-spacing: -0.03em !important;
  color: #0f172a !important;
  font-weight: 400 !important;
}
.card-subtitle {
  font-size: 0.875rem !important;
  color: #64748b !important;
  line-height: 1.5 !important;
}

.alert-error {
  background: rgba(239, 68, 68, 0.08) !important;
  border: 1px solid rgba(239, 68, 68, 0.25) !important;
  color: #b91c1c !important;
  font-size: 0.85rem !important;
  border-radius: 10px !important;
  margin-bottom: 20px;
  animation: fadeUp 0.3s ease both;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 7px;
}
.field-label {
  font-size: 0.82rem !important;
  font-weight: 500 !important;
  color: #475569 !important;
  letter-spacing: 0.01em !important;
}
.field-hint {
  font-size: 0.78rem;
  color: #64748b;
  margin: 0;
}

.field-input {
  background: #ffffff !important;
  border: 1px solid rgba(15, 23, 42, 0.12) !important;
  border-radius: 10px !important;
  color: #0f172a !important;
  font-family: "DM Sans", sans-serif !important;
  font-size: 0.925rem !important;
  transition:
    border-color 0.2s,
    background 0.2s,
    box-shadow 0.2s !important;
}
.field-input::placeholder {
  color: #94a3b8 !important;
}
.field-input:focus-visible {
  border-color: rgba(99, 102, 241, 0.6) !important;
  background: #ffffff !important;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12) !important;
  outline: none !important;
}
.field-input--error {
  border-color: rgba(239, 68, 68, 0.4) !important;
}
.field-input--error:focus-visible {
  border-color: rgba(239, 68, 68, 0.6) !important;
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1) !important;
}

.btn-submit {
  width: 100%;
  margin-top: 6px;
  background: linear-gradient(135deg, #6366f1 0%, #818cf8 100%) !important;
  border: none !important;
  color: #fff !important;
  font-family: "DM Sans", sans-serif !important;
  font-size: 0.95rem !important;
  font-weight: 600 !important;
  padding: 13px 20px !important;
  height: auto !important;
  border-radius: 10px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  gap: 6px !important;
  transition:
    opacity 0.2s,
    transform 0.15s,
    box-shadow 0.2s !important;
  box-shadow: 0 4px 20px rgba(99, 102, 241, 0.3) !important;
}
.btn-submit:hover:not(:disabled) {
  opacity: 0.92 !important;
  transform: translateY(-1px) !important;
  box-shadow: 0 6px 24px rgba(99, 102, 241, 0.4) !important;
}
.btn-submit:disabled {
  opacity: 0.6 !important;
}
.btn-arrow {
  transition: transform 0.2s;
}
.btn-submit:hover:not(:disabled) .btn-arrow {
  transform: translateX(3px);
}
.btn-loading {
  display: flex;
  align-items: center;
  gap: 10px;
}
.spinner {
  width: 15px;
  height: 15px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.card-footer {
  justify-content: center !important;
  padding-top: 0 !important;
}
.card-footer p {
  font-size: 0.85rem;
  color: #64748b;
  text-align: center;
}
.footer-link {
  color: #818cf8;
  text-decoration: none;
  font-weight: 500;
  margin-left: 4px;
  transition: color 0.2s;
}
.footer-link:hover {
  color: #a5b4fc;
}

@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>

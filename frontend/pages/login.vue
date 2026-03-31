<script setup>
const router = useRouter()
const loading = ref(false)
const error = ref("")

const form = reactive({
  email: "",
  password: "",
})

async function submit() {
  loading.value = true
  error.value = ""
  try {
    await $fetch("/pub/auth/login", {
      method: "POST",
      body: form,
      credentials: "include",
    })
    router.push("/")
  } catch (err) {
    error.value = err?.data || err?.message || "Нэврэхэд алдаа гарлаа"
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-root">
    <!-- Ambient background -->
    <div class="bg-orb bg-orb-1" />
    <div class="bg-orb bg-orb-2" />
    <div class="bg-grid" />

    <div class="login-center">
      <!-- Brand mark -->
      <NuxtLink to="/" class="brand-link">
        <span class="brand-icon">✦</span>
        Smart Note
      </NuxtLink>

      <Card class="login-card">
        <CardHeader class="card-header">
          <CardTitle class="card-title">Өдрийн мэнд</CardTitle>
          <CardDescription class="card-subtitle">
            Бүртгэлтэй мэйл болон нууц үгээ ашиглан нэвтэрнэ үү.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <!-- Error alert -->
          <Alert v-if="error" variant="destructive" class="alert-error">
            <AlertDescription class="flex items-center gap-2">
              <span>⚠</span>
              <span>{{ error }}</span>
            </AlertDescription>
          </Alert>

          <!-- Form -->
          <form @submit.prevent="submit" class="login-form">
            <!-- Email -->
            <div class="field">
              <Label for="email" class="field-label">Мэйл</Label>
              <Input
                id="email"
                v-model="form.email"
                type="email"
                required
                placeholder="you@example.com"
                class="field-input"
                :class="{ 'field-input--error': error }"
                autocomplete="email"
              />
            </div>

            <!-- Password -->
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
                autocomplete="current-password"
              />
            </div>

            <!-- Submit -->
            <Button type="submit" :disabled="loading" class="btn-submit">
              <span v-if="!loading">
                Нэвтрэх
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
            Бүртгүүлэх үү?
            <NuxtLink to="/signup" class="footer-link">Шинэ бүртгэл үүсгэх</NuxtLink>
          </p>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<style scoped>
/* ── Root ── */
.login-root {
  min-height: 100vh;
  background: #08090c;
  color: #e8e9f0;
  font-family: "DM Sans", sans-serif;
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ── Ambient ── */
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
  background: radial-gradient(circle, rgba(99, 102, 241, 0.18) 0%, transparent 70%);
  top: -150px;
  right: -100px;
}
.bg-orb-2 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, rgba(20, 184, 166, 0.12) 0%, transparent 70%);
  bottom: -100px;
  left: -80px;
}
.bg-grid {
  position: fixed;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.025) 1px, transparent 1px);
  background-size: 48px 48px;
  z-index: 0;
  pointer-events: none;
}

/* ── Center wrapper ── */
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

/* ── Brand link ── */
.brand-link {
  font-family: "Instrument Serif", serif;
  font-size: 1.3rem;
  letter-spacing: -0.02em;
  color: #f0f0f5;
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

/* ── Card (override shadcn) ── */
.login-card {
  width: 100%;
  background: rgba(255, 255, 255, 0.03) !important;
  border: 1px solid rgba(255, 255, 255, 0.08) !important;
  border-radius: 20px !important;
  backdrop-filter: blur(20px);
  box-shadow:
    0 32px 64px rgba(0, 0, 0, 0.4),
    inset 0 1px 0 rgba(255, 255, 255, 0.06) !important;
}

/* ── Card header ── */
.card-header {
  text-align: center;
  padding-bottom: 0 !important;
}
.card-title {
  font-family: "Instrument Serif", serif !important;
  font-size: 1.9rem !important;
  letter-spacing: -0.03em !important;
  color: #f0f0f5 !important;
  font-weight: 400 !important;
}
.card-subtitle {
  font-size: 0.875rem !important;
  color: #4b5563 !important;
  line-height: 1.5 !important;
}

/* ── Error alert (override shadcn destructive) ── */
.alert-error {
  background: rgba(239, 68, 68, 0.08) !important;
  border: 1px solid rgba(239, 68, 68, 0.25) !important;
  color: #fca5a5 !important;
  font-size: 0.85rem !important;
  border-radius: 10px !important;
  margin-bottom: 20px;
  animation: fadeUp 0.3s ease both;
}

/* ── Form ── */
.login-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

/* ── Field ── */
.field {
  display: flex;
  flex-direction: column;
  gap: 7px;
}
.field-label {
  font-size: 0.82rem !important;
  font-weight: 500 !important;
  color: #9ca3af !important;
  letter-spacing: 0.01em !important;
}

/* ── Input (override shadcn) ── */
.field-input {
  background: rgba(255, 255, 255, 0.04) !important;
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  border-radius: 10px !important;
  color: #f0f0f5 !important;
  font-family: "DM Sans", sans-serif !important;
  font-size: 0.925rem !important;
  transition:
    border-color 0.2s,
    background 0.2s,
    box-shadow 0.2s !important;
}
.field-input::placeholder {
  color: #374151 !important;
}
.field-input:focus-visible {
  border-color: rgba(99, 102, 241, 0.6) !important;
  background: rgba(99, 102, 241, 0.05) !important;
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

/* ── Submit button (override shadcn) ── */
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

/* ── Footer ── */
.card-footer {
  justify-content: center !important;
  padding-top: 0 !important;
}
.card-footer p {
  font-size: 0.85rem;
  color: #4b5563;
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

/* ── Animation ── */
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

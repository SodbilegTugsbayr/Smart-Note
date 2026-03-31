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
    <!-- Ambient background matching landing page -->
    <div class="bg-orb bg-orb-1" />
    <div class="bg-orb bg-orb-2" />
    <div class="bg-grid" />

    <div class="login-center">
      <!-- Brand mark -->
      <NuxtLink to="/" class="brand-link">
        <span class="brand-icon">✦</span>
        Smart Note
      </NuxtLink>

      <div class="login-card">
        <!-- Header -->
        <div class="card-header">
          <h1 class="card-title">Өдрийн мэнд</h1>
          <p class="card-subtitle">Бүртгэлтэй мэйл болон нууц үгээ ашиглан нэвтэрнэ үү.</p>
        </div>

        <!-- Error alert -->
        <div v-if="error" class="alert-error" role="alert">
          <span class="alert-icon">⚠</span>
          <span>{{ error }}</span>
        </div>

        <!-- Form -->
        <form @submit.prevent="submit" class="login-form">
          <!-- Email -->
          <div class="field">
            <label class="field-label" for="email">Мэйл</label>
            <input
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
            <div class="field-label-row">
              <label class="field-label" for="password">Нууц үг</label>
            </div>
            <input
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
          <button type="submit" class="btn-submit" :disabled="loading">
            <span v-if="!loading">
              Нэвтрэх
              <span class="btn-arrow">→</span>
            </span>
            <span v-else class="btn-loading">
              <span class="spinner" />
              Түр хүлээнэ үү...
            </span>
          </button>
        </form>

        <!-- Footer link -->
        <p class="card-footer">
          Бүртгүүлэх үү?
          <NuxtLink to="/signup" class="footer-link">Шинэ бүртгэл үүсгэх</NuxtLink>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
@import url("https://fonts.googleapis.com/css2?family=Instrument+Serif:ital@0;1&family=DM+Sans:wght@300;400;500;600&display=swap");

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

/* ── Ambient (same as landing) ── */
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

/* ── Card ── */
.login-card {
  width: 100%;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 20px;
  padding: 36px 32px;
  backdrop-filter: blur(20px);
  box-shadow:
    0 32px 64px rgba(0, 0, 0, 0.4),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

/* ── Card header ── */
.card-header {
  margin-bottom: 28px;
  text-align: center;
}
.card-title {
  font-family: "Instrument Serif", serif;
  font-size: 1.9rem;
  letter-spacing: -0.03em;
  color: #f0f0f5;
  margin: 0 0 8px;
  font-weight: 400;
}
.card-subtitle {
  font-size: 0.875rem;
  color: #4b5563;
  line-height: 1.5;
  margin: 0;
}

/* ── Error alert ── */
.alert-error {
  display: flex;
  align-items: center;
  gap: 10px;
  background: rgba(239, 68, 68, 0.08);
  border: 1px solid rgba(239, 68, 68, 0.25);
  color: #fca5a5;
  font-size: 0.85rem;
  padding: 11px 14px;
  border-radius: 10px;
  margin-bottom: 20px;
  animation: fadeUp 0.3s ease both;
}
.alert-icon {
  font-size: 1rem;
  flex-shrink: 0;
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
.field-label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.field-label {
  font-size: 0.82rem;
  font-weight: 500;
  color: #9ca3af;
  letter-spacing: 0.01em;
}
.field-input {
  width: 100%;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  color: #f0f0f5;
  font-family: "DM Sans", sans-serif;
  font-size: 0.925rem;
  padding: 11px 14px;
  outline: none;
  transition:
    border-color 0.2s,
    background 0.2s,
    box-shadow 0.2s;
  box-sizing: border-box;
}
.field-input::placeholder {
  color: #374151;
}
.field-input:focus {
  border-color: rgba(99, 102, 241, 0.6);
  background: rgba(99, 102, 241, 0.05);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
}
.field-input--error {
  border-color: rgba(239, 68, 68, 0.4);
}
.field-input--error:focus {
  border-color: rgba(239, 68, 68, 0.6);
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
}

/* ── Submit button ── */
.btn-submit {
  width: 100%;
  margin-top: 6px;
  background: linear-gradient(135deg, #6366f1 0%, #818cf8 100%);
  border: none;
  color: #fff;
  font-family: "DM Sans", sans-serif;
  font-size: 0.95rem;
  font-weight: 600;
  padding: 13px 20px;
  border-radius: 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition:
    opacity 0.2s,
    transform 0.15s,
    box-shadow 0.2s;
  box-shadow: 0 4px 20px rgba(99, 102, 241, 0.3);
}
.btn-submit:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-1px);
  box-shadow: 0 6px 24px rgba(99, 102, 241, 0.4);
}
.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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
  margin: 24px 0 0;
  text-align: center;
  font-size: 0.85rem;
  color: #4b5563;
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

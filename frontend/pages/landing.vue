<script setup lang="ts">
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Progress } from "@/components/ui/progress"

const router = useRouter()

function goToLogin() {
  router.push("/login")
}

function goToSignup() {
  router.push("/signup")
}

const features = [
  {
    icon: "📊",
    text: "Хичээлүүдийн явц, уншсан хувь, сүүлд үзсэн огноог нэг дор.",
  },
  {
    icon: "⚡",
    text: "Ойлгомжтой интерфэйс, хурдан нэвтрэлт.",
  },
  {
    icon: "📈",
    text: "Таны ахиц дэвшлийг хэмжих визуал график.",
  },
]
</script>

<template>
  <div class="landing-root">
    <!-- Ambient background -->
    <div class="bg-orb bg-orb-1" />
    <div class="bg-orb bg-orb-2" />
    <div class="bg-grid" />

    <!-- Navbar -->
    <header class="navbar">
      <div class="navbar-inner">
        <span class="brand">
          <span class="brand-icon">✦</span>
          Smart Note
        </span>
        <nav class="nav-actions">
          <Button variant="ghost" class="btn-ghost" @click="goToLogin"> Нэвтрэх </Button>
          <Button class="btn-primary" @click="goToSignup"> Бүртгүүлэх </Button>
        </nav>
      </div>
    </header>

    <!-- Hero -->
    <main class="hero">
      <div class="hero-left">
        <Badge class="badge">
          <span class="badge-dot" />
          Cурагчдад зориулсан
        </Badge>

        <h1 class="hero-title">
          Хичээлээ цэгцтэй<br />
          удирдаж, ахиц<br />
          <span class="title-accent">дэвшлээ хянана.</span>
        </h1>

        <p class="hero-sub">
          Smart Note нь таны хичээлүүдийг нэг дор эмхлэн үзүүлж, явцыг хэмжихэд тусална. Өнөөдрөөс
          эхлээд зорилгоо биелүүлээрэй.
        </p>

        <div class="hero-cta">
          <Button class="btn-primary btn-lg" @click="goToSignup">
            Бүртгүүлэх
            <span class="btn-arrow">→</span>
          </Button>
          <Button variant="outline" class="btn-outline btn-lg" @click="goToLogin"> Нэвтрэх </Button>
        </div>

        <div class="social-proof">
          <div class="avatars">
            <div v-for="n in 4" :key="n" class="avatar-bubble" :style="`--i:${n}`" />
          </div>
          <span class="social-text">500+ сурагч аль хэдийн нэгдсэн</span>
        </div>
      </div>

      <!-- Hero Right -->
      <div class="hero-right">
        <Card class="feature-card">
          <CardHeader class="card-header">
            <div class="card-dot card-dot-red" />
            <div class="card-dot card-dot-yellow" />
            <div class="card-dot card-dot-green" />
            <span class="card-title-label">Яагаад Smart Note?</span>
          </CardHeader>

          <CardContent class="p-0">
            <ul class="feature-list">
              <li
                v-for="(item, i) in features"
                :key="i"
                class="feature-item"
                :style="`animation-delay: ${0.1 + i * 0.12}s`"
              >
                <span class="feature-icon">{{ item.icon }}</span>
                <span class="feature-text">{{ item.text }}</span>
              </li>
            </ul>

            <div class="card-stats">
              <div class="stat">
                <span class="stat-value">94%</span>
                <span class="stat-label">Хэрэглэгчийн сэтгэл ханамж</span>
              </div>
              <Separator orientation="vertical" class="stat-divider" />
              <div class="stat">
                <span class="stat-value">3x</span>
                <span class="stat-label">Суралцах хурд нэмэгдэнэ</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Floating progress pill -->
        <div class="float-pill float-pill-1">
          <span class="pill-icon">🎯</span>
          <div>
            <div class="pill-label">Өнөөдрийн зорилго</div>
            <Progress :model-value="72" class="pill-bar" />
          </div>
          <span class="pill-pct">72%</span>
        </div>

        <div class="float-pill float-pill-2">
          <span class="pill-icon">🔥</span>
          <div class="pill-streak">
            <div class="pill-label">Тасралтгүй</div>
            <div class="pill-value">14 өдөр</div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
@import url("https://fonts.googleapis.com/css2?family=Instrument+Serif:ital@0;1&family=DM+Sans:wght@300;400;500;600&display=swap");

/* ── Root ── */
.landing-root {
  min-height: 100vh;
  background: #08090c;
  color: #e8e9f0;
  font-family: "DM Sans", sans-serif;
  position: relative;
  overflow: hidden;
}

/* ── Ambient background ── */
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

/* ── Navbar ── */
.navbar {
  position: relative;
  z-index: 10;
  padding: 20px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  backdrop-filter: blur(12px);
}
.navbar-inner {
  max-width: 1160px;
  margin: 0 auto;
  padding: 0 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.brand {
  font-family: "Instrument Serif", serif;
  font-size: 1.4rem;
  letter-spacing: -0.02em;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #f0f0f5;
}
.brand-icon {
  color: #818cf8;
  font-size: 1rem;
}
.nav-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* ── Button overrides ── */
.btn-ghost {
  background: transparent !important;
  border: none !important;
  color: #9ca3af !important;
  font-family: "DM Sans", sans-serif;
  font-size: 0.875rem;
  font-weight: 500;
  padding: 8px 16px;
  border-radius: 8px;
  box-shadow: none !important;
}
.btn-ghost:hover {
  color: #e8e9f0 !important;
  background: transparent !important;
}
.btn-primary {
  background: linear-gradient(135deg, #6366f1 0%, #818cf8 100%) !important;
  border: none !important;
  color: #fff !important;
  font-family: "DM Sans", sans-serif;
  font-size: 0.875rem;
  font-weight: 600;
  padding: 9px 20px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition:
    opacity 0.2s,
    transform 0.15s;
  box-shadow: none !important;
}
.btn-primary:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}
.btn-outline {
  background: transparent !important;
  border: 1px solid rgba(255, 255, 255, 0.15) !important;
  color: #c4c4d4 !important;
  font-family: "DM Sans", sans-serif;
  font-weight: 500;
  border-radius: 10px;
  transition:
    border-color 0.2s,
    color 0.2s;
  box-shadow: none !important;
}
.btn-outline:hover {
  border-color: rgba(99, 102, 241, 0.5) !important;
  color: #e8e9f0 !important;
  background: transparent !important;
}
.btn-lg {
  font-size: 1rem !important;
  padding: 13px 28px !important;
  height: auto !important;
}
.btn-arrow {
  transition: transform 0.2s;
}
.btn-primary:hover .btn-arrow {
  transform: translateX(3px);
}

/* ── Hero ── */
.hero {
  position: relative;
  z-index: 1;
  max-width: 1160px;
  margin: 0 auto;
  padding: 80px 32px 60px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 64px;
  align-items: center;
}
@media (max-width: 860px) {
  .hero {
    grid-template-columns: 1fr;
    padding: 48px 20px 40px;
    gap: 40px;
  }
}

/* ── Badge override ── */
.badge {
  display: inline-flex !important;
  align-items: center;
  gap: 8px;
  background: rgba(99, 102, 241, 0.12) !important;
  border: 1px solid rgba(99, 102, 241, 0.25) !important;
  color: #a5b4fc !important;
  font-size: 0.78rem;
  font-weight: 500;
  padding: 5px 12px;
  border-radius: 20px;
  margin-bottom: 24px;
  animation: fadeUp 0.6s ease both;
}
.badge-dot {
  width: 6px;
  height: 6px;
  background: #818cf8;
  border-radius: 50%;
  flex-shrink: 0;
  animation: pulse 2s ease infinite;
}
@keyframes pulse {
  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.5;
    transform: scale(0.8);
  }
}

.hero-title {
  font-family: "Instrument Serif", serif;
  font-size: clamp(2.2rem, 4vw, 3.4rem);
  line-height: 1.1;
  letter-spacing: -0.03em;
  color: #f0f0f5;
  margin: 0 0 20px;
  animation: fadeUp 0.6s ease 0.1s both;
}
.title-accent {
  background: linear-gradient(90deg, #818cf8 0%, #34d399 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-style: italic;
}
.hero-sub {
  font-size: 1rem;
  line-height: 1.7;
  color: #6b7280;
  max-width: 440px;
  margin: 0 0 32px;
  animation: fadeUp 0.6s ease 0.2s both;
}
.hero-cta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 36px;
  animation: fadeUp 0.6s ease 0.3s both;
}
.social-proof {
  display: flex;
  align-items: center;
  gap: 12px;
  animation: fadeUp 0.6s ease 0.4s both;
}
.avatars {
  display: flex;
}
.avatar-bubble {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: 2px solid #08090c;
  background: linear-gradient(
    135deg,
    hsl(calc(var(--i) * 60), 60%, 55%),
    hsl(calc(var(--i) * 60 + 40), 70%, 45%)
  );
  margin-left: -8px;
}
.avatar-bubble:first-child {
  margin-left: 0;
}
.social-text {
  font-size: 0.82rem;
  color: #4b5563;
}

/* ── Hero Right ── */
.hero-right {
  position: relative;
  animation: fadeUp 0.7s ease 0.2s both;
}

/* ── Card overrides ── */
.feature-card {
  background: rgba(255, 255, 255, 0.03) !important;
  border: 1px solid rgba(255, 255, 255, 0.08) !important;
  border-radius: 20px !important;
  padding: 28px !important;
  backdrop-filter: blur(20px);
  box-shadow:
    0 32px 64px rgba(0, 0, 0, 0.4),
    inset 0 1px 0 rgba(255, 255, 255, 0.06) !important;
}
.card-header {
  flex-direction: row !important;
  align-items: center !important;
  gap: 6px !important;
  margin-bottom: 24px !important;
  padding: 0 !important;
}
/* suppress the default space-y from CardHeader */
.card-header > :deep(*) {
  margin-top: 0 !important;
}

.card-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}
.card-dot-red {
  background: #ef4444;
}
.card-dot-yellow {
  background: #f59e0b;
}
.card-dot-green {
  background: #22c55e;
}
.card-title-label {
  margin-left: 10px;
  font-size: 0.82rem;
  color: #4b5563;
  font-weight: 500;
}

.feature-list {
  list-style: none;
  margin: 0 0 24px;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.feature-item {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  animation: fadeUp 0.5s ease both;
  transition:
    border-color 0.2s,
    background 0.2s;
}
.feature-item:hover {
  background: rgba(99, 102, 241, 0.06);
  border-color: rgba(99, 102, 241, 0.2);
}
.feature-icon {
  font-size: 1.2rem;
  flex-shrink: 0;
  margin-top: 1px;
}
.feature-text {
  font-size: 0.875rem;
  line-height: 1.55;
  color: #9ca3af;
}

.card-stats {
  display: flex;
  align-items: center;
  padding: 16px 0 0;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}
.stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.stat-value {
  font-family: "Instrument Serif", serif;
  font-size: 1.8rem;
  color: #f0f0f5;
  line-height: 1;
}
.stat-label {
  font-size: 0.72rem;
  color: #4b5563;
  text-align: center;
  max-width: 90px;
  line-height: 1.3;
}

/* ── Separator override ── */
.stat-divider {
  width: 1px !important;
  height: 36px !important;
  background-color: rgba(255, 255, 255, 0.07) !important;
  flex-shrink: 0;
  align-self: center;
}

/* ── Floating pills ── */
.float-pill {
  position: absolute;
  background: rgba(18, 20, 28, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 14px;
  padding: 10px 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  backdrop-filter: blur(16px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  font-size: 0.8rem;
  white-space: nowrap;
}
.float-pill-1 {
  bottom: -20px;
  left: -30px;
  animation: float 4s ease-in-out infinite;
}
.float-pill-2 {
  top: -16px;
  right: -20px;
  animation: float 4s ease-in-out 2s infinite;
}
.pill-icon {
  font-size: 1.1rem;
}
.pill-label {
  font-size: 0.7rem;
  color: #4b5563;
  margin-bottom: 4px;
}

/* ── Progress override ── */
.pill-bar {
  width: 80px !important;
  height: 4px !important;
  border-radius: 2px !important;
  background: rgba(255, 255, 255, 0.08) !important;
}
.pill-bar :deep([role="progressbar"] > div),
.pill-bar :deep(.bg-primary) {
  background: linear-gradient(90deg, #6366f1, #34d399) !important;
  border-radius: 2px !important;
}

.pill-pct {
  font-size: 0.82rem;
  font-weight: 600;
  color: #818cf8;
}
.pill-streak {
  display: flex;
  flex-direction: column;
}
.pill-value {
  font-size: 0.9rem;
  font-weight: 600;
  color: #f0f0f5;
}

/* ── Animations ── */
@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(18px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
@keyframes float {
  0%,
  100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-8px);
  }
}
</style>

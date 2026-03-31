<script setup>
import {
  ChevronDownIcon,
  LogOutIcon,
  SettingsIcon,
  ShieldIcon,
  SparklesIcon,
  UserIcon,
} from "lucide-vue-next"

const router = useRouter()
const user = useUser()

const initials = computed(() => {
  return `${user.value.firstname.toUpperCase().slice(0, 1)}${user.value.lastname.toUpperCase().slice(0, 1)}`
})

async function logout() {
  await $fetch("/api/logout", { method: "POST" })
  router.push("/landing")
}
</script>

<template>
  <nav class="sticky top-0 z-50 glass-card border-b border-white/5">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between h-16">
        <!-- Logo -->
        <NuxtLink to="/" class="flex items-center gap-2.5 group">
          <div class="w-9 h-9 rounded-lg gradient-indigo flex items-center justify-center">
            <SparklesIcon class="w-5 h-5 text-white" />
          </div>
          <span class="font-heading text-xl text-foreground tracking-wide">Smart Note</span>
        </NuxtLink>

        <!-- Right side -->
        <div class="flex items-center gap-3">
          <!-- Admin link -->
          <NuxtLink
            v-if="user?.role === 'admin'"
            to="/admin"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm text-muted-foreground hover:text-foreground hover:bg-white/5 transition-colors"
          >
            <ShieldIcon class="w-4 h-4" />
            <span class="hidden sm:inline">Админ</span>
          </NuxtLink>

          <!-- User dropdown -->
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <button
                class="flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-white/5 transition-colors"
              >
                <Avatar class="w-8 h-8">
                  <AvatarFallback class="gradient-indigo text-white text-xs font-medium">
                    {{ initials }}
                  </AvatarFallback>
                </Avatar>
                <ChevronDownIcon class="w-3.5 h-3.5 text-muted-foreground" />
              </button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end" class="w-48 glass-card border-white/10">
              <template v-if="user">
                <div class="px-3 py-2">
                  <p class="text-sm font-medium text-foreground">
                    {{ user.firstname }} {{ user.lastname }}
                  </p>
                  <p class="text-xs text-muted-foreground">{{ user.email }}</p>
                </div>
                <DropdownMenuSeparator class="bg-white/5" />
              </template>

              <DropdownMenuItem class="text-muted-foreground hover:text-foreground cursor-pointer">
                <UserIcon class="w-4 h-4 mr-2" /> Профайл
              </DropdownMenuItem>
              <DropdownMenuItem class="text-muted-foreground hover:text-foreground cursor-pointer">
                <SettingsIcon class="w-4 h-4 mr-2" /> Тохиргоо
              </DropdownMenuItem>

              <DropdownMenuSeparator class="bg-white/5" />

              <DropdownMenuItem
                class="text-red-400 hover:text-red-300 cursor-pointer"
                @click="logout"
              >
                <LogOutIcon class="w-4 h-4 mr-2" /> Гарах
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </div>
  </nav>
</template>

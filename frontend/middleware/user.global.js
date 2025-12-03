export default defineNuxtRouteMiddleware(async (to) => {
  // Authentication check
  const publicPaths = ["/login", "/signup", "/landing"]
  const user = useUser()

  if (user.value === null) {
    await useFetchMe()
  }

  const isPublic = publicPaths.includes(to.path)

  if (!user.value && !isPublic) {
    return navigateTo("/landing")
  }

  if (user.value && isPublic) {
    return navigateTo("/")
  }
})

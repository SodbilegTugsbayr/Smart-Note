export default defineNuxtRouteMiddleware(async (to) => {
  console.log("HEre")
  const publicPaths = ["/login", "/signup", "/landing"]
  await useFetchMe()
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

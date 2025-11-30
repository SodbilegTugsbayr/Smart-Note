export default defineNuxtRouteMiddleware(async (to) => {
  // Authentication check
  let user = useUser()
  if (user.value === null) {
    await useFetchMe()
  }
})

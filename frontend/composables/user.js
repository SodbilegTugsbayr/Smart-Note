export const useUser = () => {
  return useState("user", () => null)
}

export const useFetchMe = async () => {
  const user = useUser()
  const { data: c, error } = await useFetch("/api/me")
  if (error.value) {
    user.value = null
  }
  user.value = c.value
}

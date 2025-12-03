export const useUser = () => {
  return useState("user", () => null)
}

export const useFetchMe = async () => {
  const user = useUser()
  const { data, error } = await useFetch("/api/me", {
    credentials: "include",
  })

  if (error.value) {
    user.value = null
    return null
  }

  user.value = data.value || null
  return user.value
}

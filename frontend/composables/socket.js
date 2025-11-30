let socket = null

function connect_ws() {
  const config = useRuntimeConfig()

  if (config.public.dev) {
    socket = new WebSocket("ws://localhost:4000/api/ws")
  } else {
    socket = new WebSocket(`wss://${config.public.domain}/api/ws`)
  }

  socket.onmessage = (event) => {
    const queue = useState("socket_queue", () => 0)
    let data = JSON.parse(event.data)
    if (data.Type === "PING") {
      socket.send(JSON.stringify({ Type: "PONG", Text: ".." }))
      return
    }

    queue.value = data
  }

  socket.onopen = () => {}
  socket.onclose = () => {
    connect_ws()
  }

  socket.onerror = (err) => {
    socket.close()
  }
}

export const useQueue = (callback) => {
  onMounted(() => {
    if (socket == null) {
      connect_ws()
    }
  })
  const queue = useState("socket_queue", () => 0)
  watch(queue, callback)
}

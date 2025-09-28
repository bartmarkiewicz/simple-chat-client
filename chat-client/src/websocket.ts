import { ref, onMounted, onBeforeUnmount } from "vue";

export type ChatMessage = {
  role: "USER" | "SYSTEM";
  text: string;
  sender: string;
};

export function useWebSocketConnection(url = "ws://localhost:12345/ws") {
  const websocket = ref<WebSocket | null>(null);
  const connectionReady = ref(false);
  const connectionError = ref(false);
  const messages = ref<ChatMessage[]>([]);

  function connect() {
    connectionReady.value = false;
    connectionError.value = false;

    const ws = new WebSocket(url);
    websocket.value = ws;

    ws.onopen = () => {
      connectionReady.value = true;
    };

    ws.onmessage = (evt: MessageEvent) => {
      console.log(evt.data.content);
      if (evt.data) {
        const chatMessage = JSON.parse(evt.data) as ChatMessage;
        messages.value.push(chatMessage);
      } else {
        console.error("Invalid message data:", evt);
      }
    };

    ws.onerror = () => {
      connectionError.value = true;
    };

    ws.onclose = () => {
      connectionReady.value = false;
    };
  }

  function send(data: ChatMessage) {
    const ws = websocket.value;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    const payload = JSON.stringify(data);
    console.log(`Sending payload ${JSON.stringify(payload)}`)
    ws.send(payload);
  }

  onMounted(connect);

  onBeforeUnmount(() => {
    websocket.value?.close(1000, "Unmount");
    websocket.value = null;
  });

  return {
    websocket,
    connectionReady,
    connectionError,
    messages,

    connect,
    send,
  };
}

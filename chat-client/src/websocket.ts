import { ref, onMounted, onBeforeUnmount } from "vue";

export type ChatMessage = {
  role: "USER" | "SYSTEM";
  text: string;
  sender: string;
};

type ServerMessage = {
  sender?: string;
  content: {
    text: string;
    role: "USER" | "SYSTEM";
  };
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
      try {
        const raw = JSON.parse(String(evt.data)) as ServerMessage;
        const chatMessage: ChatMessage = {
          role: raw.content.role,
          text: raw.content.text,
          sender: raw.sender?.trim() ?? "System",
        };
        messages.value.push(chatMessage);
      } catch (e) {
        console.error("Invalid message data:", evt.data, e);
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
    ws.send(data.text);
  }

  onMounted(connect);

  onBeforeUnmount(() => {
    websocket.value?.close(1000, "Unmount");
  });

  return {
    websocket,
    connectionReady,
    connectionError,
    messages,

    send,
  };
}

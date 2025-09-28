<template>
  <section class="chat">
    <ul id="messages">
      <li v-for="(message, idx) in messages" :key="idx">
        <strong> Sender: {{ message.sender }}</strong>
        {{message.text}}
        <div :class="{ system: message.role === 'SYSTEM' }">{{ message.text }}</div>
      </li>
    </ul>

    <form @submit.prevent="sendChatMessage" class="composer" v-if="connectionReady">
      <input
        v-model="chatBox"
        autocomplete="off"
        placeholder="Type a message..."
      />
      <button type="submit">Send</button>
    </form>
    <div v-else-if="!connectionReady && !connectionError">
      Connecting...
    </div>
    <div v-else>
      Connection Error Please Refresh Page
    </div>
  </section>
</template>

<script setup lang="ts">
import {onMounted, ref} from "vue";
import {type ChatMessage, useWebSocketConnection} from "@/websocket.ts";

const { connectionReady, connectionError, messages, send, connect }
  = useWebSocketConnection("ws://localhost:8080/ws");

const chatBox = ref("");

function sendChatMessage() {
  const text = chatBox.value.trim();
  if (!text) return;
  const chatMessage: ChatMessage = {
    text,
    role: "USER",
  }
  send(chatMessage);
  chatBox.value = "";
}

onMounted(() => {
  connect()
});

</script>

<style scoped>
.chat {
  max-width: 640px;
  margin: 2rem auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

#messages {
  list-style: none;
  padding: 0;
  margin: 0;
  border: 1px solid #e3e3e3;
  border-radius: 8px;
  min-height: 200px;
  max-height: 50vh;
  overflow: auto;
}

#messages li {
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
}

#messages li:last-child {
  border-bottom: none;
}

.system {
  font-weight: 700; /* System messages in bold */
}

.composer {
  display: flex;
  gap: 8px;
}

.composer input {
  flex: 1;
  padding: 8px 10px;
  border: 1px solid #d0d0d0;
  border-radius: 6px;
}

.composer button {
  padding: 8px 14px;
  border: none;
  background: #3b82f6;
  color: white;
  border-radius: 6px;
  cursor: pointer;
}

.composer button:hover {
  background: #2563eb;
}
</style>

<template>
  <section class="chat">
    <ul>
      <li v-for="(message, idx) in messages" :key="idx">
        <strong> Sender: {{ message.sender }}</strong>
        <div :class="{ system: message.role === 'SYSTEM' }">{{ message.text }}</div>
      </li>
    </ul>

    <form @submit.prevent="sendChatMessage" class="form-input" v-if="connectionReady">
      <input v-model="chatBox" autocomplete="off" placeholder="Type a message..." />
      <button type="submit">Send</button>
    </form>
    <div v-else-if="!connectionReady && !connectionError">Connecting...</div>
    <div v-else>Connection Error Please Refresh Page</div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { type ChatMessage, useWebSocketConnection } from '@/websocket.ts'

const { connectionReady, connectionError, messages, send } =
  useWebSocketConnection('ws://localhost:8080/ws')

const chatBox = ref('')

function sendChatMessage() {
  const text = chatBox.value.trim()
  if (!text) return
  const chatMessage: ChatMessage = {
    text,
    role: 'USER',
    sender: 'You',
  }
  send(chatMessage)
  chatBox.value = ''
}
</script>

<style scoped>
.chat {
  max-width: 640px;
  margin: 2rem auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

li {
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
}

li:last-child {
  border-bottom: none;
}

.system {
  font-weight: 700;
}

.form-input {
  display: flex;
  gap: 8px;
}

.form-input input {
  flex: 1;
  padding: 8px 10px;
  border: 1px solid #d0d0d0;
  border-radius: 6px;
}

.form-input button {
  padding: 8px 14px;
  border: none;
  background: #3b82f6;
  color: white;
  border-radius: 6px;
  cursor: pointer;
}

.form-input button:hover {
  background: #2563eb;
}
</style>

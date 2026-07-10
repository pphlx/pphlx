<script setup>
import { ref } from 'vue';
defineProps({
  title: String,
  initialValue: Number
});

const count = ref(0);
</script>

<template>
  <div class="vue-card">
    <h3>Vue Component: {{ title }}</h3>
    <p>Value from props: {{ initialValue }}</p>
    <button @click="count++" class="vue-btn">Clicks: {{ count }}</button>
  </div>
</template>

<style scoped>
.vue-card {
  padding: 20px;
  border: 1px solid #41B883;
  border-radius: 8px;
  background: #1a1a24;
  color: #fff;
  margin: 15px 0;
}
.vue-btn {
  background: #41B883;
  color: #fff;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
  transition: background 0.2s ease;
}
.vue-btn:hover {
  background: #35495e;
}
</style>

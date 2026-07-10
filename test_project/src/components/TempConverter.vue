<script setup>
import { ref, computed } from 'vue';

const props = defineProps({
  label: String,
  initialValue: Number
});

const celsius = ref(props.initialValue || 25);
const fahrenheit = computed({
  get: () => (celsius.value * 9/5 + 32).toFixed(1),
  set: (val) => {
    celsius.value = Math.round((val - 32) * 5/9);
  }
});
</script>

<template>
  <div class="vue-card">
    <h3>Vue 3: {{ label || "Temp Converter" }}</h3>
    <div class="inputs">
      <div class="input-group">
        <label>Celsius: </label>
        <input type="number" v-model="celsius" class="conv-input" />
      </div>
      <div class="input-group">
        <label>Fahrenheit: </label>
        <input type="number" v-model="fahrenheit" class="conv-input" />
      </div>
    </div>
    <div class="result">
      <p v-if="celsius < 10">❄️ Cold weather! ({{ celsius }}°C)</p>
      <p v-else-if="celsius > 30">🔥 Hot weather! ({{ celsius }}°C)</p>
      <p v-else>🌤️ Comfortable weather! ({{ celsius }}°C)</p>
    </div>
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
h3 {
  color: #41B883;
  margin-top: 0;
}
.inputs {
  display: flex;
  gap: 15px;
  margin-bottom: 15px;
}
.input-group {
  display: flex;
  flex-direction: column;
}
.conv-input {
  background: #2d2d3d;
  border: 1px solid #444;
  color: #fff;
  padding: 8px;
  border-radius: 4px;
  width: 100px;
  margin-top: 5px;
}
.result {
  font-weight: bold;
}
</style>

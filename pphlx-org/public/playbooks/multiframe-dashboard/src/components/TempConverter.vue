<script setup>
import { ref, computed } from 'vue';

const props = defineProps({
  label: String,
  initialValue: [Number, String]
});

const celsius = ref(Number(props.initialValue) || 25);
const fahrenheit = computed({
  get: () => (celsius.value * 9/5 + 32).toFixed(1),
  set: (val) => {
    celsius.value = Math.round((val - 32) * 5/9);
  }
});
</script>

<template>
  <div class="p-6 border border-[#41B883] rounded-lg bg-gray-900 shadow-md">
    <h3 class="text-xl font-bold text-[#41B883] mb-3">4. Vue 3: {{ label || "Temp Converter" }}</h3>
    <div class="flex gap-4 mb-4">
      <div class="flex flex-col">
        <label class="text-[10px] text-gray-400 block mb-1">Celsius</label>
        <input type="number" v-model="celsius" class="w-[90px] bg-gray-800 border border-gray-700 text-white p-2 rounded text-xs focus:outline-none focus:border-[#41B883]" />
      </div>
      <div class="flex flex-col">
        <label class="text-[10px] text-gray-400 block mb-1">Fahrenheit</label>
        <input type="number" v-model="fahrenheit" class="w-[90px] bg-gray-800 border border-gray-700 text-white p-2 rounded text-xs focus:outline-none focus:border-[#41B883]" />
      </div>
    </div>
    <div class="text-xs font-bold text-gray-300">
      <p v-if="celsius < 10">❄️ Cold weather! ({{ celsius }}°C)</p>
      <p v-else-if="celsius > 30">🔥 Hot weather! ({{ celsius }}°C)</p>
      <p v-else>🌤️ Comfortable weather! ({{ celsius }}°C)</p>
    </div>
  </div>
</template>

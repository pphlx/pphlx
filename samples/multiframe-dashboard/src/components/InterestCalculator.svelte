<script>
  export let label = "Svelte Interest Calculator";
  
  export let principal = 1000;
  export let rate = 5;
  export let years = 5;

  $: total = (principal * Math.pow((1 + (rate / 100)), years)).toFixed(2);
  $: interest = (total - principal).toFixed(2);
</script>

<div class="p-6 border border-[#ff3e00] rounded-lg bg-gray-900 shadow-md">
  <h3 class="text-xl font-bold text-[#ff3e00] mb-3">5. Svelte 4: {label}</h3>
  
  <div class="flex flex-col mb-3">
    <label class="text-[10px] text-gray-400 mb-1">Principal Investment: ${principal}</label>
    <input type="range" min="100" max="10000" step="100" bind:value={principal} class="w-full accent-[#ff3e00]" />
  </div>

  <div class="flex flex-col mb-3">
    <label class="text-[10px] text-gray-400 mb-1">Annual Interest Rate: {rate}%</label>
    <input type="range" min="1" max="20" step="0.5" bind:value={rate} class="w-full accent-[#ff3e00]" />
  </div>

  <div class="flex flex-col mb-3">
    <label class="text-[10px] text-gray-400 mb-1">Duration: {years} Years</label>
    <input type="range" min="1" max="30" step="1" bind:value={years} class="w-full accent-[#ff3e00]" />
  </div>

  <div class="mt-4 pt-3 border-t border-gray-800 text-xs text-gray-300">
    <p class="mb-1">Interest Earned: <strong class="text-[#ff3e00]">${interest}</strong></p>
    <p>Total Balance: <strong class="text-white">${total}</strong></p>
  </div>
</div>

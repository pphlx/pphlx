<script>
  export let stock = 0;
  let count = 0;
  function handleAdd() {
    if (count < stock) {
      count++;
    }
  }
</script>
<div class="flex flex-col items-center gap-1.5">
  <button on:click={handleAdd} class="bg-gradient-to-r from-[#4bf3c8] to-[#00d2a0] hover:from-[#00d2a0] hover:to-[#4bf3c8] text-[#080a0d] text-[11px] font-extrabold py-2 px-4 rounded-xl transition-all shadow-[0_4px_20px_rgba(75,243,200,0.35)] flex items-center gap-1.5 select-none active:scale-95 hover:scale-105 border-none">
    <svg class="w-3.5 h-3.5 stroke-[#080a0d]" fill="none" viewBox="0 0 24 24" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"/></svg>
    Add {count > 0 ? `(${count})` : ''}
  </button>
  {#if count >= stock}
    <span class="text-[#ef4444] text-[9px] font-bold mt-1 shadow-[0_0_8px_rgba(239,68,68,0.2)] bg-[#ef4444]/10 border border-[#ef4444]/20 px-2 py-0.5 rounded-md">Max stock reached</span>
  {/if}
</div>

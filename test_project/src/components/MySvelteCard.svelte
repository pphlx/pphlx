<script>
  export let label = "Svelte Component";
  export let steps = 1;
  
  let count = 0;
  function increment() {
    count += steps;
  }
</script>

<div class="svelte-card">
  <h3>{label}</h3>
  <p>Step configuration: {steps}</p>
  <button on:click={increment} class="svelte-btn">
    Count: {count}
  </button>
</div>

<style>
  .svelte-card {
    padding: 20px;
    border: 1px solid #ff3e00;
    border-radius: 8px;
    background: #1a1a24;
    color: #fff;
    margin: 15px 0;
  }
  .svelte-btn {
    background: #ff3e00;
    color: #fff;
    border: none;
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
    font-weight: bold;
    transition: background 0.2s ease;
  }
  .svelte-btn:hover {
    background: #cc3200;
  }
</style>

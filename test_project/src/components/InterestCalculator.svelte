<script>
  export let label = "Svelte Interest Calculator";
  
  export let principal = 1000;
  export let rate = 5;
  export let years = 5;

  $: total = (principal * Math.pow((1 + (rate / 100)), years)).toFixed(2);
  $: interest = (total - principal).toFixed(2);
</script>

<div class="svelte-card">
  <h3>Svelte 4: {label}</h3>
  
  <div class="slider-group">
    <label>Principal Investment: ${principal}</label>
    <input type="range" min="100" max="10000" step="100" bind:value={principal} />
  </div>

  <div class="slider-group">
    <label>Annual Interest Rate: {rate}%</label>
    <input type="range" min="1" max="20" step="0.5" bind:value={rate} />
  </div>

  <div class="slider-group">
    <label>Duration: {years} Years</label>
    <input type="range" min="1" max="30" step="1" bind:value={years} />
  </div>

  <div class="results">
    <p>Interest Earned: <strong>${interest}</strong></p>
    <p>Total Balance: <strong>${total}</strong></p>
  </div>
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
  h3 {
    color: #ff3e00;
    margin-top: 0;
  }
  .slider-group {
    display: flex;
    flex-direction: column;
    margin-bottom: 12px;
  }
  .slider-group label {
    margin-bottom: 5px;
    font-size: 0.9em;
  }
  .results {
    margin-top: 15px;
    border-top: 1px solid #444;
    padding-top: 10px;
  }
  .results p {
    margin: 5px 0;
  }
</style>

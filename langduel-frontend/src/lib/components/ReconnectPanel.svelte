<script>
  export let reconnectNote = '';
  export let onReconnect = () => {};
  export let onBack = () => {};
</script>

<div class="panel">
  <h3>Connection lost</h3>
  <div class="small">We lost connection. You can try again.</div>
  <div class="controls two">
    <button on:click={onReconnect}>Reconnect</button>
    <button on:click={onBack}>Back to start</button>
  </div>
  {#if reconnectNote}
    <div class="small">{reconnectNote}</div>
  {/if}
</div>

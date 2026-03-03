<script>
  export let title = 'Chat';
  let text = '';
  let items = [];

  function send() {
    if (!text.trim()) return;
    items = [...items, { id: Date.now(), text }];
    text = '';
  }
</script>

<div class="chat">
  <div class="chat-title">{title}</div>
  <div class="chat-list">
    {#if items.length === 0}
      <div class="chat-empty">No messages yet</div>
    {/if}
    {#each items as item}
      <div class="chat-item">{item.text}</div>
    {/each}
  </div>
  <div class="chat-input">
    <input placeholder="Type message" bind:value={text} />
    <button type="button" on:click={send}>Send</button>
  </div>
  <div class="chat-note">Local only (not synced)</div>
</div>

<script>
  import { _ } from 'svelte-i18n';

  export let reconnectNote = '';
  export let onReconnect = () => {};
  export let onBack = () => {};
</script>

<div class="panel">
  <h3>{$_('reconnect.title')}</h3>
  <div class="small">{$_('battle.playerLeft')}</div>
  <div class="controls two">
    <button on:click={onReconnect}>{$_('reconnect.reconnecting')}</button>
    <button on:click={onBack}>{$_('profile.backToHome')}</button>
  </div>
  {#if reconnectNote}
    <div class="small">{reconnectNote}</div>
  {/if}
</div>

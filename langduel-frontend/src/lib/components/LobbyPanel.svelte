<script>
  import { _, locale, isLoading } from 'svelte-i18n';
  
  export let lobbyText = '';
  export let isCreator = false;
  export let currentRoom = '';
  export let lobbyCopyNote = '';
  export let buildRoomLink = () => '';
  export let onCopy = () => {};

  $: displayText = (() => {
    if (!lobbyText) return '';
    if (lobbyText === 'lobby.waiting') return $_('lobby.waiting');
    if (lobbyText === 'lobby.opponentJoined') return $_('lobby.opponentJoined');
    return lobbyText;
  })();
</script>

<div class="panel">
  <h3>Lobby</h3>
  <div class="small">{displayText || $_('lobby.waiting')}</div>
  {#if isCreator}
    <div class="controls two">
      <div class="link">Room link: <span>{currentRoom ? buildRoomLink(currentRoom) : '-'}</span></div>
      <button on:click={onCopy}>Copy Link</button>
    </div>
    {#if lobbyCopyNote}
      <div class="small">{lobbyCopyNote}</div>
    {/if}
  {/if}
</div>

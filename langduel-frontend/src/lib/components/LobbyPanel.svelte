<script>
  import { _ } from 'svelte-i18n';
  
  export let lobbyText = '';
  export let isCreator = false;
  export let currentRoom = '';
  export let lobbyCopyNote = '';
  export let buildRoomLink = () => '';
  export let onCopy = () => {};

  function getDisplayText(text) {
    if (!text) return '';
    if (text.startsWith('lobby.')) {
      if (text === 'lobby.waiting') return $_('lobby.waiting');
      if (text === 'lobby.opponentJoined') return $_('lobby.opponentJoined');
    }
    return text;
  }
</script>

<div class="panel">
  <h3>Lobby</h3>
  <div class="small">{getDisplayText(lobbyText) || $_('lobby.waiting')}</div>
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

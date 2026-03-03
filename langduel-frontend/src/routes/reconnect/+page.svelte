<script>
  import { onMount } from 'svelte';
  import { duel } from '$lib/stores/duel.js';
  import ReconnectPanel from '$lib/components/ReconnectPanel.svelte';

  onMount(() => {
    duel.init();
    const params = new URLSearchParams(window.location.search);
    const room = params.get('room');
    if (room) {
      duel.setField('currentRoom', room);
    }
  });
</script>

<div class="wrap">
  <header>
    <div>
      <div class="title">LangDuel</div>
      <div class="title-badge">RECONNECT</div>
    </div>
    <div class="status">{$duel.status}</div>
  </header>

  <ReconnectPanel
    reconnectNote={$duel.reconnectNote}
    onReconnect={() => {
      duel.setField('reconnectNote', 'Reconnecting...');
      duel.reconnect();
    }}
    onBack={() => {
      duel.setField('currentRoom', '');
      duel.setField('currentUser', '');
      window.location.assign('/');
    }}
  />
</div>

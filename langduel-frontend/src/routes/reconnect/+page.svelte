<script>
  import { onMount } from 'svelte';
  import { _ } from 'svelte-i18n';
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
      <div class="title-badge">{$_('reconnect.title')}</div>
    </div>
    <div class="status">{$_('reconnect.disconnected')}</div>
  </header>

  <ReconnectPanel
    reconnectNote={$duel.reconnectNote}
    onReconnect={() => {
      duel.setField('reconnectNote', $_('reconnect.reconnecting'));
      duel.reconnect();
    }}
    onBack={() => {
      duel.setField('currentRoom', '');
      duel.setField('currentUser', '');
      window.location.assign('/');
    }}
  />
</div>

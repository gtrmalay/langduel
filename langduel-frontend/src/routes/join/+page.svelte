<script>
  import { onMount } from 'svelte';

  onMount(() => {
    const params = new URLSearchParams(window.location.search);
    const room = params.get('room');
    if (room) {
      window.location.assign(`/play?room=${encodeURIComponent(room)}`);
      return;
    }
    window.location.assign('/play');
  });
</script>

<div class="wrap">
  <div class="panel">
    <h3>Redirecting</h3>
    <div class="small">Moving to /play...</div>
  </div>
</div>

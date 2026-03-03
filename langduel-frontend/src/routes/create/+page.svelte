<script>
  import { onMount } from 'svelte';

  onMount(() => {
    window.location.assign('/play');
  });
</script>

<div class="wrap">
  <div class="panel">
    <h3>Redirecting</h3>
    <div class="small">Moving to /play...</div>
  </div>
</div>

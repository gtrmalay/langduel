<script>
  import HPBar from '$lib/components/HPBar.svelte';

  export let name = 'Player';
  export let hp = 100;
  export let status = '';
</script>

<div class="player-card">
  <div class="player-name">{name}</div>
  <HPBar value={hp} />
  <div class="player-meta">
    <span>HP: {hp}</span>
    {#if status}
      <span>{status}</span>
    {/if}
  </div>
</div>

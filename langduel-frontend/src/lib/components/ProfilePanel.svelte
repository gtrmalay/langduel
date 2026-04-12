<script>
  export let open = false;
  export let profileUser = '-';
  export let profileDuels = '-';
  export let profileWins = '-';
  export let profileAcc = '-';
  export let profileStreak = '-';
  export let profileDuelsCount = '0';
  export let profileDuelsList = [];
  export let onClose = () => {};
  export let onLogout = () => {};
</script>

{#if open}
  <button
    type="button"
    class="profile-backdrop"
    aria-label="Close profile"
    on:click={onClose}
  ></button>
  <div class="profile-panel">
    <div class="profile-title">
      <h3 style="margin:0">Profile</h3>
      <button on:click={onClose}>Close</button>
    </div>
    <div class="profile">
      <div class="profile-row"><span>User</span><span>{profileUser}</span></div>
      <div class="profile-row"><span>Duels played</span><span>{profileDuels}</span></div>
      <div class="profile-row"><span>Duels won</span><span>{profileWins}</span></div>
      <div class="profile-row"><span>Accuracy</span><span>{profileAcc}</span></div>
      <div class="profile-row"><span>Best win streak</span><span>{profileStreak}</span></div>
    </div>
    <div class="profile">
      <div class="profile-row"><span>Recent duels</span><span>{profileDuelsCount}</span></div>
      <div class="duel-list">
        {#each profileDuelsList as d}
          <div class="duel-card">
            <div class="duel-row">
              <span>{d.opponentName ? $_('lobby.vs', { values: { player: d.opponentName } }) : $_('lobby.waiting')}</span>
              <span class={`badge ${d.badgeClass}`}>{d.resultLabel}</span>
            </div>
            <div class="duel-row">
              <span>Room: {d.room}</span>
              <span>{d.created}</span>
            </div>
          </div>
        {/each}
      </div>
    </div>
    <div class="profile-actions">
      <button on:click={onLogout}>Logout</button>
    </div>
  </div>
{/if}

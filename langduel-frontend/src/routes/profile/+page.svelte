<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { duel } from '$lib/stores/duel.js';

  onMount(() => {
    duel.init();
  });

  function handleLogout() {
    duel.logout();
    goto('/');
  }
</script>

<div class="wrap">
  <div class="hero">
    <div class="avatar">
      {$duel.profileUser && $duel.profileUser !== '-' ? $duel.profileUser.charAt(0).toUpperCase() : '?'}
    </div>
    <h1 class="username">
      {$duel.profileUser && $duel.profileUser !== '-' ? $duel.profileUser : 'Guest'}
    </h1>
    <div class="badge">
      {$duel.authedUsername ? 'AUTHENTICATED' : 'GUEST'}
    </div>
  </div>

  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-value">{$duel.profileDuels || '0'}</div>
      <div class="stat-label">GAMES</div>
    </div>
    <div class="stat-card">
      <div class="stat-value">{$duel.profileWins || '0'}</div>
      <div class="stat-label">WINS</div>
    </div>
    <div class="stat-card">
      <div class="stat-value">{$duel.profileAcc && $duel.profileAcc !== '-' ? $duel.profileAcc + '%' : '-'}</div>
      <div class="stat-label">ACCURACY</div>
    </div>
    <div class="stat-card">
      <div class="stat-value">{$duel.profileStreak || '0'}</div>
      <div class="stat-label">STREAK</div>
    </div>
  </div>

  <div class="section">
    <h2 class="section-title">Recent Games</h2>
    {#if $duel.profileDuelsList && $duel.profileDuelsList.length > 0}
      <div class="duel-list">
        {#each $duel.profileDuelsList as d}
          <div class="duel-card">
            <div class="duel-opponent">{d.opponent}</div>
            <div class="duel-info">
              <span class="duel-room">{d.room}</span>
              <span class="duel-date">{d.created}</span>
            </div>
            <div class={`duel-result ${d.badgeClass}`}>{d.resultLabel}</div>
          </div>
        {/each}
      </div>
    {:else}
      <div class="empty-state">
        No games yet. <a href="/play">Play a game!</a>
      </div>
    {/if}
  </div>

  <div class="actions">
    {#if $duel.authedUsername}
      <button class="logout-btn" on:click={handleLogout}>
        LOGOUT
      </button>
    {:else}
      <button class="login-btn" on:click={() => goto('/auth?next=/profile')}>
        LOGIN TO SAVE PROGRESS
      </button>
    {/if}
  </div>
</div>

<style>
  .wrap {
    max-width: 520px;
    margin: 0 auto;
    padding: 40px 20px;
    display: flex;
    flex-direction: column;
    gap: 28px;
  }

  .hero {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .avatar {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.3), rgba(37, 244, 183, 0.1));
    border: 2px solid var(--accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: "Press Start 2P", cursive;
    font-size: 28px;
    color: var(--accent);
  }

  .username {
    font-family: "Space Grotesk", sans-serif;
    font-size: 24px;
    font-weight: 700;
    color: var(--text);
    margin: 0;
  }

  .badge {
    font-size: 10px;
    padding: 4px 12px;
    border-radius: 6px;
    background: rgba(37, 244, 183, 0.15);
    border: 1px solid rgba(37, 244, 183, 0.4);
    color: var(--accent);
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .stat-card {
    background: var(--card);
    border-radius: 14px;
    padding: 20px;
    text-align: center;
    border: 1px solid var(--outline);
  }

  .stat-value {
    font-family: "Space Grotesk", sans-serif;
    font-size: 28px;
    font-weight: 700;
    color: var(--accent);
  }

  .stat-label {
    font-size: 10px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-top: 4px;
  }

  .section {
    background: var(--card);
    border-radius: 16px;
    padding: 24px;
    border: 1px solid var(--outline);
  }

  .section-title {
    font-size: 12px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 1px;
    margin: 0 0 16px 0;
  }

  .duel-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .duel-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px;
    background: rgba(11, 15, 31, 0.6);
    border-radius: 10px;
    border: 1px solid var(--outline);
  }

  .duel-opponent {
    flex: 1;
    font-size: 14px;
    font-weight: 600;
    color: var(--text);
  }

  .duel-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    text-align: right;
  }

  .duel-room {
    font-size: 11px;
    color: var(--accent);
  }

  .duel-date {
    font-size: 10px;
    color: var(--muted);
  }

  .duel-result {
    font-size: 10px;
    padding: 4px 10px;
    border-radius: 6px;
    text-transform: uppercase;
    letter-spacing: 1px;
    font-weight: 600;
  }

  .duel-result.win {
    background: rgba(37, 244, 183, 0.15);
    border: 1px solid rgba(37, 244, 183, 0.4);
    color: var(--accent);
  }

  .duel-result.loss {
    background: rgba(255, 92, 122, 0.15);
    border: 1px solid rgba(255, 92, 122, 0.4);
    color: var(--danger);
  }

  .duel-result.pending {
    background: rgba(246, 193, 68, 0.15);
    border: 1px solid rgba(246, 193, 68, 0.4);
    color: var(--accent-2);
  }

  .empty-state {
    text-align: center;
    padding: 24px;
    color: var(--muted);
    font-size: 13px;
  }

  .empty-state a {
    color: var(--accent);
    text-decoration: none;
  }

  .empty-state a:hover {
    text-decoration: underline;
  }

  .actions {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .login-btn, .logout-btn {
    width: 100%;
    padding: 16px 24px;
    border-radius: 12px;
    font-family: "Space Grotesk", sans-serif;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .login-btn {
    border: 1px solid var(--accent-2);
    background: rgba(246, 193, 68, 0.1);
    color: var(--accent-2);
  }

  .login-btn:hover {
    background: rgba(246, 193, 68, 0.2);
    box-shadow: 0 0 20px rgba(246, 193, 68, 0.3);
  }

  .logout-btn {
    border: 1px solid var(--outline);
    background: transparent;
    color: var(--muted);
  }

  .logout-btn:hover {
    border-color: var(--danger);
    color: var(--danger);
    background: rgba(255, 92, 122, 0.1);
  }
</style>

<script>
  import { onMount } from 'svelte';
  import { duel } from '$lib/stores/duel.js';
  import { _ } from 'svelte-i18n';

  let loading = true;
  let currentUserID = '';

  $: leaderboard = $duel.leaderboard;

  onMount(async () => {
    await duel.init();
    currentUserID = $duel.authedUserID;
    await duel.fetchLeaderboard();
    loading = false;
  });

  function getAvatarEmoji(avatarId) {
    return duel.getAvatarEmoji(avatarId);
  }

  function getRankColor(rank) {
    const colors = {
      'master': '#B9F2FF',
      'expert': '#FFD700',
      'apprentice': '#C0C0C0',
      'newbie': '#CD7F32',
      'struggler': '#888888'
    };
    return colors[rank] || colors['newbie'];
  }

  function getTranslatedRankName(rank) {
    const rankNames = {
      'newbie': $_('rank.newbie'),
      'apprentice': $_('rank.apprentice'),
      'expert': $_('rank.expert'),
      'master': $_('rank.master'),
      'struggler': $_('rank.struggler')
    };
    return rankNames[rank] || rankNames['newbie'];
  }
</script>

<div class="wrap">
  <div class="hero">
    <h1 class="title">🏆 {$_('leaderboard.title')}</h1>
  </div>

  {#if loading}
    <div class="loading">{$_('leaderboard.loading')}</div>
  {:else if leaderboard.length === 0}
    <div class="empty">
      <p>{$_('leaderboard.empty')}</p>
      <a href="/play" class="play-link">{$_('leaderboard.playNow')}</a>
    </div>
  {:else}
    <div class="list">
      {#each leaderboard as entry, i}
        <div 
          class="entry" 
          class:highlight={entry.user_id === currentUserID}
        >
          <div class="rank" style="color: {getRankColor(entry.rank_tier)}">
            {#if entry.rank === 1}
              🥇
            {:else if entry.rank === 2}
              🥈
            {:else if entry.rank === 3}
              🥉
            {:else}
              #{entry.rank}
            {/if}
          </div>
          
          <div class="avatar">
            {getAvatarEmoji(entry.avatar)}
          </div>
          
          <div class="info">
            <div class="username">{entry.username}</div>
            <div class="rank-name">{getTranslatedRankName(entry.rank_tier)}</div>
          </div>
          
          <div class="elo">
            <span class="elo-value">{entry.elo}</span>
            <span class="elo-label">ELO</span>
          </div>
          
          <div class="games">
            <span class="games-value">{entry.games_played}</span>
            <span class="games-label">{$_('leaderboard.games')}</span>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .wrap {
    max-width: 600px;
    margin: 0 auto;
    padding: 40px 20px;
  }

  .hero {
    text-align: center;
    margin-bottom: 32px;
  }

  .title {
    font-family: "Press Start 2P", cursive;
    font-size: 20px;
    color: var(--text);
    margin: 0;
  }

  .loading {
    text-align: center;
    padding: 40px;
    color: var(--muted);
  }

  .empty {
    text-align: center;
    padding: 40px;
    background: var(--card);
    border-radius: 16px;
    border: 1px solid var(--outline);
  }

  .empty p {
    color: var(--muted);
    margin-bottom: 16px;
  }

  .play-link {
    display: inline-block;
    padding: 12px 24px;
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.2), rgba(37, 244, 183, 0.05));
    border: 1px solid var(--accent);
    border-radius: 10px;
    color: var(--accent);
    text-decoration: none;
    font-weight: 600;
    transition: all 0.2s;
  }

  .play-link:hover {
    box-shadow: 0 0 15px rgba(37, 244, 183, 0.35);
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .entry {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px 18px;
    background: var(--card);
    border-radius: 12px;
    border: 1px solid var(--outline);
    transition: all 0.2s;
  }

  .entry:hover {
    border-color: var(--accent);
    transform: translateX(4px);
  }

  .entry.highlight {
    border-color: var(--accent);
    background: rgba(37, 244, 183, 0.08);
  }

  .rank {
    width: 40px;
    font-size: 14px;
    font-weight: 700;
    text-align: center;
  }

  .avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: rgba(11, 15, 31, 0.8);
    border: 2px solid var(--outline);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
  }

  .info {
    flex: 1;
    min-width: 0;
  }

  .username {
    font-weight: 600;
    font-size: 14px;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .rank-name {
    font-size: 11px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .elo {
    text-align: right;
    min-width: 70px;
  }

  .elo-value {
    display: block;
    font-family: "Space Grotesk", sans-serif;
    font-size: 18px;
    font-weight: 700;
    color: var(--accent);
  }

  .elo-label {
    font-size: 9px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .games {
    text-align: right;
    min-width: 50px;
  }

  .games-value {
    display: block;
    font-size: 14px;
    font-weight: 600;
    color: var(--text);
  }

  .games-label {
    font-size: 9px;
    color: var(--muted);
    text-transform: uppercase;
  }

  @media (max-width: 480px) {
    .entry {
      padding: 12px 14px;
      gap: 10px;
    }
    
    .rank {
      width: 30px;
      font-size: 12px;
    }
    
    .avatar {
      width: 36px;
      height: 36px;
      font-size: 18px;
    }
    
    .username {
      font-size: 13px;
    }
    
    .elo-value {
      font-size: 16px;
    }
    
    .games {
      display: none;
    }
  }
</style>

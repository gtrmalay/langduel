<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { duel } from '$lib/stores/duel.js';
  import { _ } from 'svelte-i18n';

  let showLeaveConfirm = false;

  $: topicLabels = {
    'default': $_('topics.default'),
    'animals': $_('topics.animals'),
    'travel': $_('topics.travel'),
    'food': $_('topics.food'),
    'movies': $_('topics.movies'),
    'sports': $_('topics.sports')
  };

  $: difficultyLabels = {
    'beginner': $_('difficulty.beginner'),
    'intermediate': $_('difficulty.intermediate'),
    'advanced': $_('difficulty.advanced')
  };

  $: currentUser = $duel.currentUser || '';
  $: playerA = $duel.playerA || '';
  $: playerB = $duel.playerB || '';
  $: isPlayerA = currentUser && playerA === currentUser;
  $: myAvatarEmoji = duel.getAvatarEmoji($duel.userAvatar || 'default');
  $: opponentAvatarEmoji = duel.getAvatarEmoji($duel.opponentAvatar || 'default');
  $: userAvatarEmoji = isPlayerA ? myAvatarEmoji : opponentAvatarEmoji;
  $: oppAvatarEmoji = isPlayerA ? opponentAvatarEmoji : myAvatarEmoji;

  onMount(() => {
    duel.init();
    const params = new URLSearchParams(window.location.search);
    const room = params.get('room');
    if (room) {
      duel.setField('currentRoom', room);
    }
  });

  function handleLeave() {
    if (showLeaveConfirm) {
      duel.leaveMatch();
    } else {
      showLeaveConfirm = true;
      setTimeout(() => showLeaveConfirm = false, 3000);
    }
  }
</script>

<div class="wrap">
  <div class="hero">
    <h1 class="title">{$_('play.create').toUpperCase()}</h1>
  </div>

  <div class="room-info">
    <div class="room-code">
      <span class="label">{$_('play.room').toUpperCase()}</span>
      <span class="code">{$duel.currentRoom || '-'}</span>
      {#if $duel.currentRoom}
        <button class="copy-btn" on:click={() => duel.copyLink($duel.currentRoom, 'lobbyCopyNote')}>
          {$_('play.copy')}
        </button>
      {/if}
    </div>

    {#if $duel.currentTopic || $duel.createDifficulty}
      <div class="game-info">
        {#if $duel.currentTopic || $duel.createTopic}
          <span class="info-badge">
            📚 {topicLabels[$duel.currentTopic] || topicLabels[$duel.createTopic] || $_('topics.default')}
          </span>
        {/if}
        {#if $duel.createDifficulty}
          <span class="info-badge">
            🎯 {difficultyLabels[$duel.createDifficulty] || $_('difficulty.intermediate')}
          </span>
        {/if}
      </div>
    {/if}
  </div>

  {#if $duel.lobbyCopyNote}
    <div class="note">{$duel.lobbyCopyNote}</div>
  {/if}

  <div class="players">
    <div class="player-card you">
      <div class="player-icon">{myAvatarEmoji}</div>
      <div class="player-name">{$duel.currentUser || $_('profile.guest')}</div>
      <div class="player-status ready">{$_('lobby.ready').toUpperCase()}</div>
    </div>

    <div class="vs">VS</div>

    <div class="player-card opponent">
      <div class="player-icon">{opponentAvatarEmoji}</div>
      <div class="player-name">
        {$duel.playerB && $duel.playerB !== 'Player B' ? $duel.playerB : $_('lobby.waiting')}
      </div>
      <div class="player-status">
        {$duel.playerB && $duel.playerB !== 'Player B' ? $_('lobby.ready').toUpperCase() : $_('lobby.waitingOpponent').toUpperCase()}
      </div>
    </div>
  </div>

  <div class="status-text">
    {#if $duel.lobbyText === 'lobby.opponentJoined'}
      {$_('lobby.opponentJoined')}
    {:else if $duel.lobbyText === 'lobby.waiting'}
      {$_('lobby.waiting')}
    {:else}
      {$duel.lobbyText || $_('lobby.waiting')}
    {/if}
  </div>

  <div class="leave-section">
    <button class="leave-btn" class:confirm={showLeaveConfirm} on:click={handleLeave}>
      {showLeaveConfirm ? $_('lobby.confirmLeave').toUpperCase() : $_('lobby.leave').toUpperCase()}
    </button>
  </div>
</div>

<style>
  .wrap {
    max-width: 480px;
    margin: 0 auto;
    padding: 40px 20px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 24px;
  }

  .hero {
    text-align: center;
  }

  .title {
    font-family: "Press Start 2P", cursive;
    font-size: 20px;
    color: var(--text);
    margin: 0;
    letter-spacing: 3px;
  }

  .room-info {
    display: flex;
    flex-direction: column;
    gap: 12px;
    width: 100%;
  }

  .room-code {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px 20px;
    background: var(--card);
    border-radius: 12px;
    border: 1px solid var(--outline);
    justify-content: center;
  }

  .label {
    font-size: 10px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .code {
    font-family: "Space Grotesk", sans-serif;
    font-size: 16px;
    font-weight: 600;
    color: var(--accent);
    letter-spacing: 1px;
  }

  .copy-btn {
    padding: 6px 12px;
    border-radius: 6px;
    border: 1px solid var(--outline);
    background: transparent;
    color: var(--muted);
    font-size: 10px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .copy-btn:hover {
    border-color: var(--accent);
    color: var(--accent);
  }

  .game-info {
    display: flex;
    justify-content: center;
    gap: 10px;
  }

  .info-badge {
    padding: 8px 14px;
    border-radius: 8px;
    background: rgba(37, 244, 183, 0.1);
    border: 1px solid rgba(37, 244, 183, 0.3);
    font-size: 12px;
    color: var(--accent);
  }

  .note {
    font-size: 12px;
    color: var(--accent);
  }

  .players {
    display: flex;
    align-items: center;
    gap: 16px;
    width: 100%;
    justify-content: center;
  }

  .player-card {
    flex: 1;
    max-width: 200px;
    padding: 24px 16px;
    background: var(--card);
    border-radius: 16px;
    border: 2px solid var(--outline);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .player-card.you {
    border-color: rgba(37, 244, 183, 0.4);
  }

  .player-card.opponent {
    border-color: rgba(246, 193, 68, 0.3);
  }

  .player-icon {
    font-size: 32px;
    width: 56px;
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(11, 15, 31, 0.8);
    border-radius: 50%;
  }

  .player-name {
    font-size: 13px;
    font-weight: 600;
    color: var(--text);
    text-align: center;
    word-break: break-word;
    overflow-wrap: break-word;
  }

  .player-status {
    font-size: 10px;
    padding: 4px 10px;
    border-radius: 6px;
    background: rgba(11, 15, 31, 0.8);
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 1px;
    white-space: nowrap;
  }

  .player-status.ready {
    background: rgba(37, 244, 183, 0.2);
    color: var(--accent);
    border: 1px solid rgba(37, 244, 183, 0.4);
  }

  .vs {
    font-family: "Press Start 2P", cursive;
    font-size: 14px;
    color: var(--accent-2);
    background: var(--card);
    padding: 12px 8px;
    border-radius: 10px;
    border: 1px solid rgba(246, 193, 68, 0.5);
  }

  .status-text {
    font-size: 14px;
    color: var(--muted);
    text-align: center;
  }

  .leave-section {
    margin-top: 16px;
  }

  .leave-btn {
    padding: 12px 24px;
    border: 1px solid var(--outline);
    border-radius: 10px;
    background: transparent;
    color: var(--muted);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .leave-btn:hover {
    border-color: var(--danger);
    color: var(--danger);
  }

  .leave-btn.confirm {
    background: rgba(255, 92, 122, 0.15);
    border-color: var(--danger);
    color: var(--danger);
  }
</style>

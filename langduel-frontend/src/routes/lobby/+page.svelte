<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { duel } from '$lib/stores/duel.js';

  let showLeaveConfirm = false;

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
    <h1 class="title">LOBBY</h1>
  </div>

  <div class="room-code">
    <span class="label">ROOM</span>
    <span class="code">{$duel.currentRoom || '-'}</span>
    {#if $duel.currentRoom}
      <button class="copy-btn" on:click={() => duel.copyLink($duel.currentRoom, 'lobbyCopyNote')}>
        COPY
      </button>
    {/if}
  </div>

  {#if $duel.lobbyCopyNote}
    <div class="note">{$duel.lobbyCopyNote}</div>
  {/if}

  <div class="players">
    <div class="player-card you">
      <div class="player-icon">👤</div>
      <div class="player-name">{$duel.currentUser || 'You'}</div>
      <div class="player-status ready">READY</div>
    </div>

    <div class="vs">VS</div>

    <div class="player-card opponent">
      <div class="player-icon">?</div>
      <div class="player-name">
        {$duel.playerB && $duel.playerB !== 'Player B' ? $duel.playerB : 'Waiting...'}
      </div>
      <div class="player-status">
        {$duel.playerB && $duel.playerB !== 'Player B' ? 'READY' : 'WAITING'}
      </div>
    </div>
  </div>

  <div class="status-text">
    {$duel.lobbyText}
  </div>

  <div class="leave-section">
    <button class="leave-btn" class:confirm={showLeaveConfirm} on:click={handleLeave}>
      {showLeaveConfirm ? 'CONFIRM LEAVE' : 'LEAVE LOBBY'}
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

  .room-code {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px 20px;
    background: var(--card);
    border-radius: 12px;
    border: 1px solid var(--outline);
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
    max-width: 160px;
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
    word-break: break-all;
  }

  .player-status {
    font-size: 10px;
    padding: 4px 10px;
    border-radius: 6px;
    background: rgba(11, 15, 31, 0.8);
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 1px;
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

<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { duel } from '$lib/stores/duel.js';

  let activeTab = 'create';

  onMount(() => {
    duel.init();
    const params = new URLSearchParams(window.location.search);
    const room = params.get('room');
    if (room) {
      duel.setField('joinRoom', room);
      activeTab = 'join';
    }
    if (!$duel.authChoiceMade) {
      const next = room ? `/play?room=${encodeURIComponent(room)}` : '/play';
      goto(`/auth?next=${encodeURIComponent(next)}`);
    }
  });

  function handleTab(tab) {
    activeTab = tab;
    duel.setFlowMode(tab);
  }
</script>

<div class="wrap">
  <div class="hero">
    <h1 class="title">LANGDUEL</h1>
    <p class="subtitle">TRANSLATE BATTLE</p>
  </div>

  <div class="tabs">
    <button 
      class="tab" 
      class:active={activeTab === 'create'} 
      on:click={() => handleTab('create')}
    >
      CREATE
    </button>
    <button 
      class="tab" 
      class:active={activeTab === 'join'} 
      on:click={() => handleTab('join')}
    >
      JOIN
    </button>
  </div>

  <div class="card">
    {#if activeTab === 'create'}
      <div class="form-group">
        <label for="room-id">Room ID</label>
        <div class="room-input">
          <input 
            id="room-id"
            type="text" 
            placeholder="room-xxxxx"
            value={$duel.createRoom}
            on:input={(e) => duel.setField('createRoom', e.target.value)}
          />
          <button class="gen-btn" on:click={() => duel.ensureRoomId()}>GEN</button>
        </div>
      </div>
      
      <div class="form-group">
        <label for="nickname-create">Your Nickname</label>
        <input 
          id="nickname-create"
          type="text" 
          placeholder="Guest-xxxx"
          value={$duel.createUser}
          on:input={(e) => duel.setField('createUser', e.target.value)}
        />
      </div>

      <div class="invite-section">
        <span class="invite-label">Invite Link:</span>
        <span class="invite-link">
          {$duel.createRoom ? duel.buildRoomLink($duel.createRoom) : 'Generate room first'}
        </span>
        {#if $duel.createRoom}
          <button class="copy-btn" on:click={() => duel.copyLink($duel.createRoom, 'createCopyNote')}>
            COPY
          </button>
        {/if}
      </div>
      
      {#if $duel.createCopyNote}
        <div class="note">{$duel.createCopyNote}</div>
      {/if}

      <button class="action-btn" on:click={() => duel.createAndConnect()}>
        CREATE ROOM
      </button>

    {:else}
      <div class="form-group">
        <label for="room-code">Room Code</label>
        <input 
          id="room-code"
          type="text" 
          placeholder="Enter room code"
          value={$duel.joinRoom}
          on:input={(e) => duel.setField('joinRoom', e.target.value)}
        />
      </div>

      <div class="form-group">
        <label for="nickname-join">Your Nickname</label>
        <input 
          id="nickname-join"
          type="text" 
          placeholder="Guest-xxxx"
          value={$duel.joinUser}
          on:input={(e) => duel.setField('joinUser', e.target.value)}
        />
      </div>

      <button class="action-btn" on:click={() => duel.joinAndConnect()}>
        JOIN ROOM
      </button>
    {/if}

    {#if $duel.startError}
      <div class="error">{$duel.startError}</div>
    {/if}
  </div>

  <button class="back-btn" on:click={() => goto('/')}>
    ← BACK
  </button>
</div>

<style>
  .wrap {
    max-width: 440px;
    margin: 0 auto;
    padding: 40px 20px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .hero {
    text-align: center;
    margin-bottom: 8px;
  }

  .title {
    font-family: "Press Start 2P", cursive;
    font-size: 28px;
    color: var(--text);
    margin: 0;
    letter-spacing: 3px;
  }

  .subtitle {
    font-family: "Press Start 2P", cursive;
    font-size: 10px;
    color: var(--accent);
    margin: 12px 0 0 0;
    letter-spacing: 2px;
  }

  .tabs {
    display: flex;
    gap: 8px;
    background: var(--card);
    padding: 6px;
    border-radius: 12px;
  }

  .tab {
    flex: 1;
    padding: 14px 20px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    font-family: "Space Grotesk", sans-serif;
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 1px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .tab.active {
    background: var(--accent);
    color: var(--bg);
  }

  .tab:hover:not(.active) {
    background: rgba(37, 244, 183, 0.1);
  }

  .card {
    background: var(--card);
    border-radius: 16px;
    padding: 28px;
    display: flex;
    flex-direction: column;
    gap: 20px;
    border: 1px solid rgba(37, 244, 183, 0.15);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  label {
    font-size: 11px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  input {
    width: 100%;
    padding: 14px 16px;
    border-radius: 10px;
    border: 1px solid var(--outline);
    background: rgba(11, 15, 31, 0.9);
    color: var(--text);
    font-size: 15px;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(37, 244, 183, 0.15);
  }

  .room-input {
    display: flex;
    gap: 8px;
  }

  .room-input input {
    flex: 1;
  }

  .gen-btn {
    padding: 0 16px;
    border-radius: 10px;
    border: 1px solid var(--accent);
    background: rgba(37, 244, 183, 0.15);
    color: var(--accent);
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .gen-btn:hover {
    background: rgba(37, 244, 183, 0.25);
  }

  .invite-section {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    padding: 12px;
    background: rgba(11, 15, 31, 0.6);
    border-radius: 8px;
  }

  .invite-label {
    font-size: 11px;
    color: var(--muted);
    text-transform: uppercase;
  }

  .invite-link {
    flex: 1;
    font-size: 12px;
    color: var(--accent);
    word-break: break-all;
    min-width: 0;
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
    text-align: center;
  }

  .action-btn {
    width: 100%;
    padding: 18px 24px;
    border-radius: 12px;
    border: 2px solid var(--accent);
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.25), rgba(37, 244, 183, 0.1));
    color: var(--accent);
    font-family: "Press Start 2P", cursive;
    font-size: 12px;
    letter-spacing: 1px;
    cursor: pointer;
    transition: all 0.2s;
    margin-top: 8px;
  }

  .action-btn:hover {
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.35), rgba(37, 244, 183, 0.15));
    box-shadow: 0 0 20px rgba(37, 244, 183, 0.3);
  }

  .action-btn:active {
    transform: translateY(2px);
  }

  .error {
    padding: 12px;
    border-radius: 8px;
    background: rgba(255, 92, 122, 0.15);
    border: 1px solid rgba(255, 92, 122, 0.4);
    color: var(--danger);
    font-size: 12px;
    text-align: center;
  }

  .back-btn {
    padding: 14px 24px;
    border: none;
    border-radius: 10px;
    background: transparent;
    color: var(--muted);
    font-size: 13px;
    cursor: pointer;
    transition: all 0.2s;
    align-self: center;
  }

  .back-btn:hover {
    color: var(--text);
    background: rgba(255, 255, 255, 0.05);
  }
</style>

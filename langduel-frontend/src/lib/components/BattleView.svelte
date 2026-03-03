<script>
  export let playerA = 'Player A';
  export let playerB = 'Player B';
  export let hp = {};
  export let promptText = '';
  export let timerText = '';
  export let roundInfo = '';
  export let correctCount = 0;
  export let wrongCount = 0;
  export let totalDamage = 0;
  export let avgSpeedValue = '-';
  export let answer = '';
  export let hitA = false;
  export let hitB = false;
  export let gameOverOpen = false;
  export let gameOverText = '';
  export let gameOverHP = '';
  export let onSend = () => {};
  export let onLeave = () => {};
  export let onPlayAgain = () => {};

  let showLeaveConfirm = false;

  function handleLeave() {
    if (showLeaveConfirm) {
      onLeave();
    } else {
      showLeaveConfirm = true;
      setTimeout(() => showLeaveConfirm = false, 3000);
    }
  }
</script>

<div class="battle-container">
  <div class="top-bar">
    <div class={`player-hp ${hitA ? 'hit' : ''}`}>
      <span class="hp-name">{playerA}</span>
      <div class="hp-bar">
        <div class="hp-fill" style={`width: ${hp[playerA] != null ? Math.max(0, Math.min(100, hp[playerA])) : 100}%`}></div>
      </div>
      <span class="hp-value">{hp[playerA] ?? 100}</span>
    </div>

    <div class="timer-section">
      <div class="round-label">{roundInfo || 'ROUND'}</div>
      <div class="timer">{timerText || '0:00'}</div>
    </div>

    <div class={`player-hp opponent ${hitB ? 'hit' : ''}`}>
      <span class="hp-name">{playerB}</span>
      <div class="hp-bar">
        <div class="hp-fill" style={`width: ${hp[playerB] != null ? Math.max(0, Math.min(100, hp[playerB])) : 100}%`}></div>
      </div>
      <span class="hp-value">{hp[playerB] ?? 100}</span>
    </div>
  </div>

  <div class="prompt-section">
    <div class="prompt-card">
      <div class="prompt-text">{promptText || 'Waiting...'}</div>
      <div class="prompt-hint">translate this</div>
    </div>
  </div>

  <div class="input-section">
    <input 
      class="answer-input" 
      placeholder="type answer..." 
      bind:value={answer} 
      autocomplete="off"
      on:keydown={(e) => e.key === 'Enter' && answer.trim() && onSend()}
    />
    <button class="submit-btn" on:click={onSend} disabled={!answer.trim()}>
      SUBMIT
    </button>
  </div>

  <div class="quit-section">
    <button class="quit-btn" class:confirm={showLeaveConfirm} on:click={handleLeave}>
      {showLeaveConfirm ? 'CONFIRM QUIT' : 'quit'}
    </button>
  </div>

  {#if gameOverOpen}
    <div class="game-over-overlay">
      <div class="game-over-modal">
        <div class="game-over-title">GAME OVER</div>
        <div class="game-over-result">{gameOverText}</div>
        <div class="game-over-hp">{gameOverHP}</div>
        <div class="game-over-stats">
          <div class="stat">
            <span class="stat-label">Correct</span>
            <span class="stat-value">{correctCount}</span>
          </div>
          <div class="stat">
            <span class="stat-label">Wrong</span>
            <span class="stat-value">{wrongCount}</span>
          </div>
          <div class="stat">
            <span class="stat-label">Damage</span>
            <span class="stat-value">{totalDamage}</span>
          </div>
        </div>
        <button class="play-again-btn" on:click={onPlayAgain}>
          PLAY AGAIN
        </button>
        <button class="home-btn" on:click={onLeave}>
          HOME
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .battle-container {
    width: 100%;
    max-width: 600px;
    display: flex;
    flex-direction: column;
    gap: 32px;
  }

  .top-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
  }

  .player-hp {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 16px;
    background: var(--card);
    border-radius: 12px;
    border: 1px solid rgba(37, 244, 183, 0.2);
    transition: all 0.3s;
  }

  .player-hp.opponent {
    text-align: right;
  }

  .player-hp.hit {
    animation: hitFlash 0.35s ease;
    border-color: var(--danger);
  }

  @keyframes hitFlash {
    0% { box-shadow: 0 0 0 rgba(255, 92, 122, 0); }
    50% { box-shadow: 0 0 25px rgba(255, 92, 122, 0.8); }
    100% { box-shadow: 0 0 0 rgba(255, 92, 122, 0); }
  }

  .hp-name {
    font-size: 11px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .hp-bar {
    height: 12px;
    background: rgba(11, 15, 31, 0.9);
    border-radius: 6px;
    overflow: hidden;
    border: 1px solid rgba(37, 244, 183, 0.2);
  }

  .hp-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--accent), #19d9a4);
    transition: width 0.3s ease;
  }

  .player-hp.opponent .hp-fill {
    margin-left: auto;
  }

  .hp-value {
    font-family: "Space Grotesk", sans-serif;
    font-size: 18px;
    font-weight: 700;
    color: var(--accent);
  }

  .player-hp.opponent .hp-value {
    text-align: right;
  }

  .timer-section {
    text-align: center;
    padding: 0 20px;
  }

  .round-label {
    font-size: 10px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 4px;
  }

  .timer {
    font-family: "Press Start 2P", cursive;
    font-size: 28px;
    color: var(--accent-2);
    text-shadow: 0 0 20px rgba(246, 193, 68, 0.5);
  }

  .prompt-section {
    display: flex;
    justify-content: center;
  }

  .prompt-card {
    width: 100%;
    padding: 40px 32px;
    background: var(--card);
    border-radius: 20px;
    border: 2px solid rgba(37, 244, 183, 0.3);
    text-align: center;
    box-shadow: 0 0 40px rgba(37, 244, 183, 0.15);
  }

  .prompt-text {
    font-family: "Space Grotesk", sans-serif;
    font-size: 36px;
    font-weight: 700;
    color: var(--text);
    letter-spacing: 2px;
    text-transform: uppercase;
  }

  .prompt-hint {
    font-size: 12px;
    color: var(--muted);
    margin-top: 12px;
    text-transform: uppercase;
    letter-spacing: 2px;
  }

  .input-section {
    display: flex;
    gap: 12px;
  }

  .answer-input {
    flex: 1;
    padding: 18px 24px;
    border-radius: 14px;
    border: 2px solid var(--outline);
    background: var(--card);
    color: var(--text);
    font-size: 18px;
    text-transform: lowercase;
  }

  .answer-input:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 4px rgba(37, 244, 183, 0.15);
  }

  .answer-input::placeholder {
    color: var(--muted);
    font-size: 14px;
  }

  .submit-btn {
    padding: 18px 32px;
    border-radius: 14px;
    border: 2px solid var(--accent);
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.25), rgba(37, 244, 183, 0.1));
    color: var(--accent);
    font-family: "Press Start 2P", cursive;
    font-size: 11px;
    letter-spacing: 1px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .submit-btn:hover:not(:disabled) {
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.35), rgba(37, 244, 183, 0.15));
    box-shadow: 0 0 20px rgba(37, 244, 183, 0.4);
  }

  .submit-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .quit-section {
    display: flex;
    justify-content: center;
  }

  .quit-btn {
    padding: 8px 16px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    font-size: 11px;
    cursor: pointer;
    transition: all 0.2s;
    text-transform: lowercase;
  }

  .quit-btn:hover {
    color: var(--danger);
    background: rgba(255, 92, 122, 0.1);
  }

  .quit-btn.confirm {
    background: rgba(255, 92, 122, 0.2);
    color: var(--danger);
  }

  .game-over-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.85);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    animation: fadeIn 0.3s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .game-over-modal {
    background: var(--card);
    border-radius: 24px;
    padding: 40px;
    text-align: center;
    border: 2px solid var(--accent-2);
    box-shadow: 0 0 60px rgba(246, 193, 68, 0.3);
    max-width: 400px;
    width: 90%;
    animation: scaleIn 0.3s ease;
  }

  @keyframes scaleIn {
    from { transform: scale(0.9); opacity: 0; }
    to { transform: scale(1); opacity: 1; }
  }

  .game-over-title {
    font-family: "Press Start 2P", cursive;
    font-size: 24px;
    color: var(--accent-2);
    margin-bottom: 20px;
  }

  .game-over-result {
    font-size: 18px;
    font-weight: 600;
    color: var(--text);
    margin-bottom: 8px;
  }

  .game-over-hp {
    font-size: 13px;
    color: var(--muted);
    margin-bottom: 24px;
  }

  .game-over-stats {
    display: flex;
    justify-content: center;
    gap: 24px;
    margin-bottom: 28px;
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .stat-label {
    font-size: 10px;
    color: var(--muted);
    text-transform: uppercase;
  }

  .stat-value {
    font-size: 20px;
    font-weight: 700;
    color: var(--accent);
  }

  .play-again-btn {
    width: 100%;
    padding: 18px 24px;
    border-radius: 14px;
    border: 2px solid var(--accent);
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.25), rgba(37, 244, 183, 0.1));
    color: var(--accent);
    font-family: "Press Start 2P", cursive;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s;
    margin-bottom: 12px;
  }

  .play-again-btn:hover {
    box-shadow: 0 0 25px rgba(37, 244, 183, 0.4);
  }

  .home-btn {
    padding: 12px 24px;
    border: none;
    border-radius: 10px;
    background: transparent;
    color: var(--muted);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .home-btn:hover {
    color: var(--text);
  }
</style>

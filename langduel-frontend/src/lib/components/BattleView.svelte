<script>
  import { _ } from 'svelte-i18n';
  import { duel } from '$lib/stores/duel.js';
  
  export let playerA = 'Player A';
  export let playerB = 'Player B';
  export let playerAEmoji = '?';
  export let playerBEmoji = '?';
  export let hp = {};
  export let promptText = '';
  export let timerText = '';
  export let roundInfo = '';
  export let correctCount = 0;
  export let wrongCount = 0;
  export let totalDamage = 0;
  export let playerADamage = 0;
  export let playerBDamage = 0;
  export let avgSpeedValue = '-';
  export let answer = '';
  export let hitA = false;

  $: translatedPromptText = promptText && promptText.startsWith('lobby.') ? $_(promptText) : promptText;
  $: translatedRoundInfo = roundInfo && roundInfo.startsWith('battle.') ? $_(roundInfo) : roundInfo;
  export let hitB = false;
  export let lastDamage = 0;
  export let lastDamageTo = '';
  export let attackA = false;
  export let attackB = false;
  export let inputCorrect = false;
  export let inputWrong = false;
  export let gameOverOpen = false;
  export let gameOverText = '';
  export let gameOverHP = '';
  export let gameOverReason = '';
  export let isGameWinner = null;
  export let duelId = '';
  export let connectionStatus = 'disconnected';
  export let ping = -1;
  export let onSend = () => {};
  export let onLeave = () => {};
  export let onPlayAgain = () => {};

  let showLeaveConfirm = false;
  let answerInput;
  let lastPrompt = '';
  let showAnalysis = false;
  let analysisData = null;
  let loadingAnalysis = false;

  let screenShake = false;
  let particles = [];
  let showRoundAnnounce = false;
  let announceRound = 1;
  let halftimeCountdown = '';
  let timerDanger = false;
  let timerCritical = false;
  let localInputCorrect = false;
  let localInputWrong = false;
  let avatarAttackA = false;
  let avatarAttackB = false;
  let avatarHitA = false;
  let avatarHitB = false;
  let confetti = [];
  let showPromptTyping = false;
  let typedPrompt = '';
  let projectileA = false;
  let projectileB = false;
  let particleCounter = 0;
  let confettiCounter = 0;

  const characters = {
    knight: { emoji: '⚔️', color: '#3498db', projectile: 'blade' },
    wizard: { emoji: '🧙', color: '#9b59b6', projectile: 'orb' },
    archer: { emoji: '🏹', color: '#2ecc71', projectile: 'arrow' },
    dragon: { emoji: '🐉', color: '#e74c3c', projectile: 'fire' },
    skull: { emoji: '💀', color: '#95a5a6', projectile: 'skull' },
    fire: { emoji: '🔥', color: '#f39c12', projectile: 'flame' },
    ice: { emoji: '❄️', color: '#00cec9', projectile: 'shard' },
    lightning: { emoji: '⚡', color: '#fdcb6e', projectile: 'bolt' },
    sword: { emoji: '🗡️', color: '#d63031', projectile: 'blade' },
    shield: { emoji: '🛡️', color: '#3498db', projectile: 'shield' },
    potion: { emoji: '🧪', color: '#00b894', projectile: 'orb' },
    crown: { emoji: '👑', color: '#fdcb6e', projectile: 'crown' },
    star: { emoji: '⭐', color: '#ffeaa7', projectile: 'star' },
    moon: { emoji: '🌙', color: '#a29bfe', projectile: 'beam' },
    default: { emoji: '👤', color: '#74b9ff', projectile: 'circle' }
  };

  function getCharacter(avatar) {
    return characters[avatar] || characters.default;
  }

  $: if (promptText && promptText !== lastPrompt && !gameOverOpen) {
    lastPrompt = promptText;
    triggerRoundAnnounce();
    // Auto-focus input after round change
    if (answerInput) setTimeout(() => answerInput.focus(), 100);
  }

  $: if (hitA && !avatarHitA) triggerPlayerHit('A');
  $: if (hitB && !avatarHitB) triggerPlayerHit('B');

  $: if (attackA) {
    avatarAttackA = true;
    projectileA = true;
    setTimeout(() => { avatarAttackA = false; }, 400);
    setTimeout(() => { projectileA = false; }, 300);
  }

  $: if (attackB) {
    avatarAttackB = true;
    projectileB = true;
    setTimeout(() => { avatarAttackB = false; }, 400);
    setTimeout(() => { projectileB = false; }, 300);
  }

  $: if (inputCorrect) {
    localInputCorrect = true;
    setTimeout(() => localInputCorrect = false, 400);
  }

  $: if (inputWrong) {
    localInputWrong = true;
    setTimeout(() => localInputWrong = false, 400);
  }

  $: {
    const seconds = parseInt(timerText) || 0;
    timerDanger = seconds <= 5 && seconds > 3;
    timerCritical = seconds <= 3 && seconds > 0;
  }

  async function openAnalysis() {
    if (!duelId) return;
    loadingAnalysis = true;
    showAnalysis = true;
    analysisData = null;
    try {
      const result = await duel.fetchDuelAnalysis(duelId);
      analysisData = result;
    } catch (e) {
      console.error('Analysis error:', e);
    }
    loadingAnalysis = false;
  }

  function handleSend() {
    onSend();
    if (answerInput) answerInput.focus();
  }

  function triggerRoundAnnounce() {
    // Check if it's halftime
    if (translatedRoundInfo && (translatedRoundInfo.includes('HALFTIME') || translatedRoundInfo.includes('Half 2'))) {
      announceRound = 'HALF 2';
      showRoundAnnounce = true;
      
      // Start halftime countdown (5 seconds)
      let halftimeLeft = 5;
      halftimeCountdown = '5';
      
      const halftimeInterval = setInterval(() => {
        halftimeLeft--;
        halftimeCountdown = halftimeLeft.toString();
        if (halftimeLeft <= 0) {
          clearInterval(halftimeInterval);
          halftimeCountdown = '';
        }
      }, 1000);
      
      setTimeout(() => {
        showRoundAnnounce = false;
        halftimeCountdown = '';
      }, 5000);
      return;
    }
    
    // Regular round announcement - only for rounds 1 and 11 (start of each half)
    const match = roundInfo.match(/Round (\d+)/);
    if (!match) return;
    
    const roundNum = parseInt(match[1]);
    
    // Only show announcement for first round of each half
    if (roundNum !== 1 && roundNum !== 11) {
      return;
    }
    
    announceRound = roundNum;
    showRoundAnnounce = true;
    
    showPromptTyping = true;
    typedPrompt = '';
    let i = 0;
    const typeInterval = setInterval(() => {
      if (i < promptText.length) {
        typedPrompt += promptText[i];
        i++;
      } else {
        clearInterval(typeInterval);
        setTimeout(() => showPromptTyping = false, 200);
      }
    }, 40);

    setTimeout(() => {
      showRoundAnnounce = false;
      answer = '';
      if (answerInput) setTimeout(() => answerInput.focus(), 100);
    }, 1500);
  }

  function triggerPlayerHit(player) {
    if (player === 'A') {
      avatarHitA = true;
      screenShake = true;
      spawnParticles('A');
      setTimeout(() => { avatarHitA = false; screenShake = false; }, 350);
    } else {
      avatarHitB = true;
      screenShake = true;
      spawnParticles('B');
      setTimeout(() => { avatarHitB = false; screenShake = false; }, 350);
    }
  }

  function spawnParticles(player) {
    const char = player === 'A' ? getCharacter(playerAEmoji) : getCharacter(playerBEmoji);
    const newParticles = [];
    const baseId = particleCounter++;
    for (let i = 0; i < 10; i++) {
      const angle = (Math.PI * 2 / 10) * i + Math.random() * 0.5;
      const distance = 50 + Math.random() * 40;
      const size = 6 + Math.random() * 10;
      newParticles.push({
        id: baseId * 100 + i,
        x: Math.cos(angle) * distance,
        y: Math.sin(angle) * distance,
        size,
        delay: i * 25,
        color: Math.random() > 0.5 ? '#25f4b7' : char.color
      });
    }
    particles = [...particles, ...newParticles];
    setTimeout(() => { particles = particles.filter(p => !newParticles.includes(p)); }, 700);
  }

  function launchConfetti() {
    const colors = ['#25f4b7', '#f6c144', '#ff5c7a', '#4ecdc4', '#ffe66d', '#ff6b6b', '#c44dff'];
    const baseId = confettiCounter++;
    confetti = Array(60).fill().map((_, i) => ({
      id: baseId * 1000 + i,
      x: Math.random() * 100,
      color: colors[Math.floor(Math.random() * colors.length)],
      size: 8 + Math.random() * 8,
      delay: Math.random() * 500,
      duration: 2000 + Math.random() * 1500,
      rotation: Math.random() * 360,
      drift: (Math.random() - 0.5) * 100
    }));
    setTimeout(() => confetti = [], 4000);
  }

  $: if (gameOverOpen && gameOverText && gameOverText.includes('WIN')) {
    launchConfetti();
  }

  function handleLeave() {
    if (showLeaveConfirm) {
      onLeave();
    } else {
      showLeaveConfirm = true;
      setTimeout(() => showLeaveConfirm = false, 3000);
    }
  }

  $: charA = getCharacter(playerAEmoji);
  $: charB = getCharacter(playerBEmoji);
  $: hpA = hp[playerA] ?? 100;
  $: hpB = hp[playerB] ?? 100;
  $: pingColor = ping < 0 ? '#ff5c7a' : ping < 100 ? '#25f4b7' : ping < 300 ? '#f6c144' : '#ff5c7a';
  $: pingIcon = ping < 0 ? '🔴' : ping < 100 ? '🟢' : ping < 300 ? '🟡' : '🔴';
  $: pingText = ping < 0 ? '-' : ping + 'ms';
  
  // Compute current half from round number
  $: currentHalf = (() => {
    const match = roundInfo.match(/Round (\d+)/);
    if (!match) return 1;
    const roundNum = parseInt(match[1]);
    return roundNum <= 10 ? 1 : 2;
  })();
  
  // Check if we're in halftime (input disabled)
  $: isHalftime = translatedRoundInfo && (translatedRoundInfo.includes('HALFTIME') || translatedRoundInfo.includes('Half 2'));
  $: inputDisabled = isHalftime || gameOverOpen;
  
  // Combined round display
  $: roundDisplay = translatedRoundInfo ? translatedRoundInfo + (currentHalf ? ` (Half ${currentHalf}/2)` : '') : $_('battle.round').toUpperCase();
</script>

<div class="battle-container" class:shake={screenShake}>
  {#if connectionStatus === 'disconnected'}
    <div class="connection-overlay">
      <div class="connection-lost">
        <span class="conn-icon">📡</span>
        <span class="conn-text">Connection lost. Reconnecting...</span>
      </div>
    </div>
  {/if}
  
  <div class="ping-indicator" style="background: {pingColor}20; border-color: {pingColor}">
    <span class="ping-icon">{pingIcon}</span>
    <span class="ping-value" style="color: {pingColor}">{pingText}</span>
  </div>
  {#if showRoundAnnounce}
    <div class="round-announce-overlay">
      <div class="round-announce">
        <span class="round-sword">⚔️</span>
        <span class="round-text">{typeof announceRound === 'string' ? announceRound : 'ROUND ' + announceRound}</span>
        <span class="round-sword">⚔️</span>
        {#if halftimeCountdown}
          <div class="halftime-countdown">{halftimeCountdown}</div>
        {/if}
      </div>
    </div>
  {/if}

  {#if confetti.length > 0}
    <div class="confetti-container">
      {#each confetti as c (c.id)}
        <div class="confetti-piece" style="left: {c.x}%; background: {c.color}; width: {c.size}px; height: {c.size}px; animation-delay: {c.delay}ms; animation-duration: {c.duration}ms; --drift: {c.drift}px; --rotation: {c.rotation}deg;"></div>
      {/each}
    </div>
  {/if}

  {#each particles as p (p.id)}
    <div class="particle" style="--x: {p.x}px; --y: {p.y}px; --size: {p.size}px; --color: {p.color}; --delay: {p.delay}ms;"></div>
  {/each}

  <div class="battle-arena" class:critical={timerCritical}>
    <div class="player-side player-a" class:attacking={avatarAttackA} class:hit={avatarHitA}>
      <div class="player-header">
        <span class="player-avatar-small">{charA.emoji}</span>
        <span class="player-name">{playerA}</span>
      </div>
      <div class="character-wrapper">
        <div class="character-glow" style="background: {charA.color}"></div>
        <div class="character" class:attack={avatarAttackA} class:hit={avatarHitA}>
          {charA.emoji}
        </div>
        {#if avatarHitA}
          <div class="damage-number">-{playerA === lastDamageTo ? lastDamage : 0}</div>
        {/if}
        {#if projectileA}
          <div class="projectile projectile-a projectile-{charA.projectile}" style="background: {charA.color}; box-shadow: 0 0 20px {charA.color};"></div>
        {/if}
      </div>
      <div class="player-stats">
        <div class="hp-bar-container">
          <div class="hp-bar">
            <div class="hp-fill" class:low={hpA < 30} style="width: {Math.max(0, Math.min(100, hpA))}%"></div>
          </div>
          <span class="hp-value" style="color: {hpA < 30 ? 'var(--danger)' : 'var(--accent)'}">{hpA}</span>
        </div>
      </div>
    </div>

    <div class="vs-divider">
      <div class="battle-line"></div>
      <div class="vs-badge">VS</div>
      <div class="battle-line"></div>
    </div>

    <div class="player-side player-b" class:hit={avatarHitB}>
      <div class="player-header">
        <span class="player-avatar-small">{charB.emoji}</span>
        <span class="player-name">{playerB}</span>
      </div>
      <div class="character-wrapper character-wrapper-b">
        <div class="character-glow" style="background: {charB.color}"></div>
        <div class="character character-b" class:attack={avatarAttackB} class:hit={avatarHitB}>
          {charB.emoji}
        </div>
        {#if avatarHitB}
          <div class="damage-number damage-number-b">-{playerB === lastDamageTo ? lastDamage : 0}</div>
        {/if}
        {#if projectileB}
          <div class="projectile projectile-b projectile-{charB.projectile}" style="background: {charB.color}; box-shadow: 0 0 20px {charB.color};"></div>
        {/if}
      </div>
      <div class="player-stats">
        <div class="hp-bar-container">
          <div class="hp-bar">
            <div class="hp-fill hp-fill-b" class:low={hpB < 30} style="width: {Math.max(0, Math.min(100, hpB))}%"></div>
          </div>
          <span class="hp-value" style="color: {hpB < 30 ? 'var(--danger)' : 'var(--accent)'}">{hpB}</span>
        </div>
      </div>
    </div>
  </div>

  <div class="timer-section" class:danger={timerDanger} class:critical={timerCritical}>
    <div class="round-label">{roundDisplay}</div>
    <div class="timer">{timerText || '0:00'}</div>
  </div>

  <div class="prompt-section" class:hidden={showRoundAnnounce}>
    <div class="prompt-card" class:typing={showPromptTyping}>
      {#if showPromptTyping}
        <div class="prompt-text">{typedPrompt}<span class="cursor">|</span></div>
      {:else}
        <div class="prompt-text">{translatedPromptText || $_('lobby.waiting')}</div>
      {/if}
      <div class="prompt-hint">{$_('battle.translate')}</div>
    </div>
  </div>

  <div class="input-section" class:hidden={showRoundAnnounce}>
    <input 
      class="answer-input" 
      class:correct-flash={localInputCorrect}
      class:wrong-shake={localInputWrong}
      placeholder={$_('battle.typeAnswer')} 
      bind:value={answer} 
      bind:this={answerInput}
      autocomplete="off"
      maxlength="200"
      disabled={inputDisabled}
      on:keydown={(e) => e.key === 'Enter' && !inputDisabled && answer.trim() && handleSend()}
    />
    <button class="submit-btn" on:click={handleSend} disabled={!answer.trim() || inputDisabled}>
      {$_('battle.submit').toUpperCase()}
    </button>
  </div>

  <div class="quit-section">
    <button class="quit-btn" class:confirm={showLeaveConfirm} on:click={handleLeave}>
      {showLeaveConfirm ? $_('battle.confirmQuit').toUpperCase() : $_('battle.quit')}
    </button>
  </div>

  {#if gameOverOpen}
    <div class="game-over-overlay">
      <div class="game-over-modal" class:winner={isGameWinner} class:loser={isGameWinner === false}>
        <div class="game-over-title" class:win={isGameWinner} class:lose={isGameWinner === false}>
          {isGameWinner ? $_('battle.winner') : isGameWinner === false ? $_('battle.defeat') : $_('battle.gameOver')}
        </div>
        <div class="game-over-result">{gameOverText}</div>
        {#if gameOverReason}
          <div class="game-over-reason">{gameOverReason}</div>
        {/if}
        <div class="game-over-hp">{gameOverHP}</div>
        <div class="game-over-stats">
          <div class="stat" style="animation-delay: 0ms">
            <span class="stat-label">{$_('battle.correct')}</span>
            <span class="stat-value">{correctCount}</span>
          </div>
          <div class="stat" style="animation-delay: 150ms">
            <span class="stat-label">{$_('battle.wrong')}</span>
            <span class="stat-value">{wrongCount}</span>
          </div>
          <div class="stat" style="animation-delay: 300ms">
            <span class="stat-label">{$_('battle.damage')}</span>
            <span class="stat-value">
              {#if playerADamage !== playerBDamage}
                {playerADamage} / {playerBDamage}
              {:else}
                {playerADamage || totalDamage}
              {/if}
            </span>
          </div>
        </div>
        <button class="play-again-btn" on:click={onPlayAgain}>{$_('gameOver.playAgain').toUpperCase()}</button>
        <button class="analysis-btn" on:click={openAnalysis}>📊 {$_('gameOver.viewAnalysis').toUpperCase() || 'ANALYSIS'}</button>
        <button class="home-btn" on:click={onLeave}>{$_('gameOver.home').toUpperCase()}</button>
      </div>
    </div>
  {/if}

  {#if showAnalysis}
    <div class="analysis-overlay" on:click={() => showAnalysis = false}>
      <div class="analysis-modal" on:click|stopPropagation>
        <h3 class="analysis-title">📊 {$_('gameOver.duelAnalysis') || 'DUEL ANALYSIS'}</h3>
        {#if loadingAnalysis}
          <div class="loading">{$_('leaderboard.loading')}</div>
        {:else if analysisData && analysisData.rounds && analysisData.rounds.length > 0}
          <div class="analysis-content">
            {#if analysisData.participants && analysisData.participants.length > 0}
              <div class="participants-summary">
                {#each analysisData.participants as p}
                  <div class="participant-row">
                    <span class="participant-name">{p.username}</span>
                    <span class="participant-stats">
                      <span class="correct">{p.correct || 0}✓</span>
                      <span class="wrong">{p.wrong || 0}✗</span>
                    </span>
                  </div>
                {/each}
              </div>
            {/if}
            <div class="rounds-list">
              {#each analysisData.rounds || [] as round, i}
                <div class="round-card">
                  <div class="round-header">
                    <span class="round-number">#{round.round_number || i + 1}</span>
                    <span class="round-phrase">{round.phrase || $_('battle.unknownPhrase')}</span>
                  </div>
                  {#if round.correct_answer}
                    <div class="correct-answer">
                      <span class="correct-label">{$_('analysis.correctAnswer')}:</span>
                      <span class="correct-value">{round.correct_answer}</span>
                    </div>
                  {/if}
                  <div class="answers-list">
                    {#each (() => {
                      const seen = new Map();
                      for (const ans of (round.answers || [])) {
                        if (!seen.has(ans.user_id)) {
                          seen.set(ans.user_id, ans);
                        }
                      }
                      return Array.from(seen.values());
                    })() as ans}
                      <div class="answer-item" class:correct={ans.is_correct} class:wrong={!ans.is_correct}>
                        <span class="answer-user">{ans.username}:</span>
                        <span class="answer-text">"{ans.answer || ''}"</span>
                        <span class="answer-status">{ans.is_correct ? '✓' : '✗'}</span>
                      </div>
                    {/each}
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {:else}
            <div class="no-analysis">
              <p>📊 {$_('analysis.noAnalysis')}</p>
              <p class="no-analysis-hint">{$_('analysis.analysisWillBeAvailable')}</p>
            </div>
        {/if}
        <button class="close-analysis-btn" on:click={() => showAnalysis = false}>{$_('gameOver.close').toUpperCase() || 'CLOSE'}</button>
      </div>
    </div>
  {/if}
</div>

<style>
  .battle-container {
    width: 100%;
    max-width: 700px;
    display: flex;
    flex-direction: column;
    gap: 24px;
    position: relative;
    padding: 20px;
  }

  .battle-container.shake {
    animation: screenShake 0.2s ease-out;
  }

  @keyframes screenShake {
    0%, 100% { transform: translate(0, 0); }
    20% { transform: translate(-4px, 2px); }
    40% { transform: translate(4px, -2px); }
    60% { transform: translate(-2px, 2px); }
    80% { transform: translate(2px, -2px); }
  }

  .round-announce-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.85);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 500;
    animation: fadeIn 0.2s ease;
  }

  .round-announce {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 24px;
    animation: roundSlam 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
    position: relative;
  }

  .round-announce > span {
    display: flex;
    align-items: center;
    gap: 24px;
  }

  .round-sword {
    font-size: 56px;
    animation: swordSlash 0.4s ease-out;
  }

  .round-sword:last-child {
    animation: swordSlashReverse 0.4s ease-out;
  }

  .round-text {
    font-family: "Press Start 2P", cursive;
    font-size: 40px;
    color: var(--accent-2);
    text-shadow: 0 0 30px rgba(246, 193, 68, 0.8);
    letter-spacing: 4px;
  }

  .halftime-countdown {
    position: absolute;
    bottom: -60px;
    font-family: "Press Start 2P", cursive;
    font-size: 32px;
    color: var(--danger);
    text-shadow: 0 0 20px rgba(255, 92, 122, 0.8);
    animation: pulse 1s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.2); }
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes roundSlam {
    0% { transform: scale(0) rotate(-5deg); opacity: 0; }
    60% { transform: scale(1.15) rotate(2deg); }
    100% { transform: scale(1) rotate(0deg); opacity: 1; }
  }

  @keyframes swordSlash {
    0% { transform: translateX(-40px) rotate(-30deg); opacity: 0; }
    100% { transform: translateX(0) rotate(0deg); opacity: 1; }
  }

  @keyframes swordSlashReverse {
    0% { transform: translateX(40px) rotate(30deg); opacity: 0; }
    100% { transform: translateX(0) rotate(0deg); opacity: 1; }
  }

  .confetti-container {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    overflow: hidden;
    z-index: 1000;
  }

  .confetti-piece {
    position: absolute;
    top: -20px;
    border-radius: 2px;
    animation: confettiFall linear forwards;
    transform: rotate(var(--rotation));
  }

  @keyframes confettiFall {
    0% { transform: translateY(0) rotate(0deg); opacity: 1; }
    100% { transform: translateY(100vh) rotate(720deg) translateX(var(--drift)); opacity: 0; }
  }

  .particle {
    position: absolute;
    width: var(--size);
    height: var(--size);
    background: var(--color);
    border-radius: 50%;
    pointer-events: none;
    top: 50%;
    left: 50%;
    animation: particleExplode 0.6s ease-out forwards;
    animation-delay: var(--delay);
    box-shadow: 0 0 12px var(--color);
  }

  @keyframes particleExplode {
    0% { transform: translate(-50%, -50%) translate(0, 0) scale(1); opacity: 1; }
    100% { transform: translate(-50%, -50%) translate(var(--x), var(--y)) scale(0); opacity: 0; }
  }

  .battle-arena {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 60px;
    padding: 40px 60px;
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.05), rgba(246, 193, 68, 0.05));
    border-radius: 24px;
    border: 2px solid rgba(37, 244, 183, 0.2);
    position: relative;
    transition: border-color 0.3s, box-shadow 0.3s;
  }

  .battle-arena.critical {
    border-color: var(--danger);
    animation: arenaPulse 0.5s ease-in-out infinite;
  }

  @keyframes arenaPulse {
    0%, 100% { box-shadow: inset 0 0 30px rgba(255, 92, 122, 0.3), 0 0 20px rgba(255, 92, 122, 0.2); }
    50% { box-shadow: inset 0 0 50px rgba(255, 92, 122, 0.5), 0 0 40px rgba(255, 92, 122, 0.3); }
  }

  .player-side {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    transition: transform 0.2s ease;
  }

  .player-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 16px;
    background: rgba(0, 0, 0, 0.3);
    border-radius: 20px;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .player-b .player-header {
    flex-direction: row-reverse;
  }

  .player-avatar-small {
    font-size: 28px;
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 50%;
  }

  .player-name {
    font-family: "Press Start 2P", cursive;
    font-size: 10px;
    color: var(--text);
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-shadow: 0 0 10px var(--accent);
  }

  .player-side.attacking {
    animation: attackMove 0.3s ease-out;
  }

  .player-side.hit {
    animation: hitMove 0.3s ease-out;
  }

  .player-b {
  }

  .character-wrapper-b {
    transform: scaleX(-1);
  }

  .player-side.attacking {
    animation: attackMove 0.3s ease-out;
  }

  .player-side.hit {
    animation: hitMove 0.3s ease-out;
  }

  .player-b.attacking {
    animation: attackMoveB 0.3s ease-out;
  }

  .player-b.hit {
    animation: hitMoveB 0.3s ease-out;
  }

  @keyframes attackMove {
    0% { transform: translateX(0); }
    30% { transform: translateX(60px); }
    100% { transform: translateX(0); }
  }

  @keyframes hitMove {
    0% { transform: translateX(0); filter: brightness(1); }
    30% { transform: translateX(-20px); filter: brightness(1.5); }
    100% { transform: translateX(0); filter: brightness(1); }
  }

  @keyframes attackMoveB {
    0% { transform: translateX(0); }
    30% { transform: translateX(-60px); }
    100% { transform: translateX(0); }
  }

  @keyframes hitMoveB {
    0% { transform: translateX(0); filter: brightness(1); }
    30% { transform: translateX(20px); filter: brightness(1.5); }
    100% { transform: translateX(0); filter: brightness(1); }
  }

  @keyframes attackMoveB {
    0% { transform: scaleX(-1) translateX(0); }
    30% { transform: scaleX(-1) translateX(60px); }
    100% { transform: scaleX(-1) translateX(0); }
  }

  @keyframes hitMove {
    0% { transform: translateX(0); filter: brightness(1); }
    30% { transform: translateX(-20px); filter: brightness(1.5); }
    100% { transform: translateX(0); filter: brightness(1); }
  }

  @keyframes hitMoveB {
    0% { transform: scaleX(-1) translateX(0); filter: brightness(1); }
    30% { transform: scaleX(-1) translateX(-20px); filter: brightness(1.5); }
    100% { transform: scaleX(-1) translateX(0); filter: brightness(1); }
  }

  .character-wrapper {
    position: relative;
    width: 150px;
    height: 150px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .character-glow {
    position: absolute;
    width: 100%;
    height: 100%;
    border-radius: 50%;
    filter: blur(40px);
    opacity: 0.5;
    animation: glowPulse 2s ease-in-out infinite;
  }

  @keyframes glowPulse {
    0%, 100% { transform: scale(1); opacity: 0.4; }
    50% { transform: scale(1.15); opacity: 0.6; }
  }

  .character {
    font-size: 90px;
    animation: characterIdle 3s ease-in-out infinite;
    filter: drop-shadow(0 6px 12px rgba(0,0,0,0.4));
  }

  .character.attack {
    animation: characterAttack 0.3s ease-out;
  }

  .character.hit {
    animation: characterHit 0.3s ease-out;
    filter: brightness(1.5) saturate(1.5);
  }

  @keyframes characterIdle {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-8px); }
  }

  @keyframes characterAttack {
    0% { transform: translateX(0) scale(1); }
    30% { transform: translateX(30px) scale(1.2); }
    100% { transform: translateX(0) scale(1); }
  }

  @keyframes characterHit {
    0% { transform: translateX(0) scale(1); }
    30% { transform: translateX(-15px) scale(0.9); filter: brightness(2); }
    100% { transform: translateX(0) scale(1); }
  }

  .damage-number {
    position: absolute;
    top: -10px;
    right: 10px;
    font-family: "Press Start 2P", cursive;
    font-size: 24px;
    color: var(--danger);
    text-shadow: 0 0 15px var(--danger), 2px 2px 0 #000;
    animation: damageFloat 0.8s ease-out forwards;
  }

  .damage-number-b {
    right: auto;
    left: 10px;
  }

  @keyframes damageFloat {
    0% { transform: translateY(0) scale(0.5); opacity: 1; }
    100% { transform: translateY(-50px) scale(1.2); opacity: 0; }
  }

  .projectile {
    position: absolute;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    animation: projectileFlyA 0.25s ease-out forwards;
  }

  /* Shield - square */
  .projectile-shield {
    border-radius: 4px;
    transform: rotate(45deg);
  }

  /* Orb - circle with glow */
  .projectile-orb {
    border-radius: 50%;
    box-shadow: 0 0 30px currentColor, 0 0 60px currentColor;
  }

  /* Arrow - elongated */
  .projectile-arrow {
    width: 30px;
    height: 12px;
    border-radius: 6px 2px 2px 6px;
  }

  /* Fire - irregular */
  .projectile-fire {
    border-radius: 50% 30% 50% 30%;
    filter: blur(1px);
  }

  /* Skull - with eyes */
  .projectile-skull {
    border-radius: 50% 50% 40% 40%;
  }

  /* Flame - flickering */
  .projectile-flame {
    width: 20px;
    height: 28px;
    border-radius: 50% 50% 50% 50% / 60% 60% 40% 40%;
    filter: blur(0.5px);
  }

  /* Shard - sharp */
  .projectile-shard {
    width: 16px;
    height: 28px;
    clip-path: polygon(50% 0%, 100% 100%, 0% 100%);
    border-radius: 0;
  }

  /* Bolt - zigzag */
  .projectile-bolt {
    width: 24px;
    height: 8px;
    clip-path: polygon(0% 50%, 40% 0%, 60% 50%, 100% 0%, 60% 50%, 40% 100%);
  }

  /* Blade - long */
  .projectile-blade {
    width: 32px;
    height: 10px;
    border-radius: 2px;
  }

  /* Crown - with points */
  .projectile-crown {
    width: 28px;
    height: 20px;
    clip-path: polygon(0% 100%, 0% 40%, 20% 60%, 35% 20%, 50% 50%, 65% 20%, 80% 60%, 100% 40%, 100% 100%);
  }

  /* Star - rotating */
  .projectile-star {
    clip-path: polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%);
  }

  /* Beam - curved */
  .projectile-beam {
    width: 40px;
    height: 6px;
    border-radius: 3px;
    opacity: 0.8;
  }

  .projectile-a {
    right: -40px;
  }

  .projectile-b {
    left: -40px;
  }

  @keyframes projectileFlyA {
    0% { transform: translateX(0) scale(1); opacity: 1; }
    100% { transform: translateX(80px) scale(0.3); opacity: 0; }
  }

  @keyframes projectileFlyB {
    0% { transform: translateX(0) scale(1); opacity: 1; }
    100% { transform: translateX(-80px) scale(0.3); opacity: 0; }
  }

  .projectile-b {
    left: -40px;
    animation: projectileFlyB 0.25s ease-out forwards;
  }

  .player-stats {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .hp-bar-container {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .player-b .hp-bar-container {
    flex-direction: row-reverse;
  }

  .hp-bar {
    flex: 1;
    height: 20px;
    background: rgba(11, 15, 31, 0.95);
    border-radius: 4px;
    overflow: hidden;
    border: 2px solid var(--accent);
    box-shadow: 0 0 10px rgba(37, 244, 183, 0.3), inset 0 2px 4px rgba(0,0,0,0.5);
    position: relative;
  }

  .hp-fill {
    height: 100%;
    transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    background: linear-gradient(180deg, var(--accent) 0%, #1dd4a4 100%);
    box-shadow: 0 0 15px var(--accent), inset 0 -2px 4px rgba(0,0,0,0.3);
    position: relative;
  }

  .hp-fill::after {
    content: '';
    position: absolute;
    top: 2px;
    left: 2px;
    right: 2px;
    height: 6px;
    background: linear-gradient(180deg, rgba(255,255,255,0.5), transparent);
    border-radius: 2px;
  }

  .hp-fill.low {
    background: linear-gradient(180deg, var(--danger) 0%, #cc4a62 100%);
    box-shadow: 0 0 15px var(--danger);
    animation: hpCritical 0.5s ease-in-out infinite;
  }

  @keyframes hpCritical {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.7; }
  }

  .hp-fill-b {
    margin-left: auto;
  }

  .hp-value {
    font-family: "Press Start 2P", cursive;
    font-size: 20px;
    min-width: 60px;
    text-align: center;
    text-shadow: 0 0 15px currentColor;
  }

  .vs-divider {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 0 20px;
  }

  .battle-line {
    width: 3px;
    height: 50px;
    background: linear-gradient(180deg, transparent, var(--accent-2), transparent);
  }

  .vs-badge {
    font-family: "Press Start 2P", cursive;
    font-size: 20px;
    color: var(--accent-2);
    text-shadow: 0 0 20px rgba(246, 193, 68, 0.8);
    animation: vsPulse 1.5s ease-in-out infinite;
  }

  @keyframes vsPulse {
    0%, 100% { transform: scale(1); opacity: 0.8; }
    50% { transform: scale(1.15); opacity: 1; }
  }

  .timer-section {
    text-align: center;
    padding: 10px 20px;
  }

  .timer-section.danger .timer {
    color: #ff9500;
    animation: timerPulse 0.5s ease-in-out infinite;
  }

  .timer-section.critical .timer {
    color: var(--danger);
    animation: timerShake 0.15s linear infinite, timerCritical 0.3s ease-in-out infinite;
  }

  .round-label {
    font-family: "Press Start 2P", cursive;
    font-size: 8px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 2px;
    margin-bottom: 8px;
  }

  .timer {
    font-family: "Press Start 2P", cursive;
    font-size: 36px;
    color: var(--accent-2);
    text-shadow: 0 0 20px rgba(246, 193, 68, 0.5);
    transition: color 0.3s;
  }

  @keyframes timerPulse {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.08); }
  }

  @keyframes timerShake {
    0%, 100% { transform: translateX(0); }
    25% { transform: translateX(-2px); }
    75% { transform: translateX(2px); }
  }

  @keyframes timerCritical {
    0%, 100% { text-shadow: 0 0 20px rgba(255, 92, 122, 0.5); }
    50% { text-shadow: 0 0 40px rgba(255, 92, 122, 0.9); }
  }

  .prompt-section {
    transition: opacity 0.2s ease;
  }

  .prompt-section.hidden {
    opacity: 0;
    pointer-events: none;
  }

  .prompt-card {
    width: 100%;
    padding: 32px 24px;
    background: var(--card);
    border-radius: 20px;
    border: 2px solid rgba(37, 244, 183, 0.3);
    text-align: center;
    box-shadow: 0 0 40px rgba(37, 244, 183, 0.1);
  }

  .prompt-card.typing {
    border-color: var(--accent);
    animation: promptReveal 0.3s ease-out;
  }

  @keyframes promptReveal {
    0% { transform: scale(0.98); opacity: 0.5; }
    100% { transform: scale(1); opacity: 1; }
  }

  .prompt-text {
    font-family: "Space Grotesk", sans-serif;
    font-size: 32px;
    font-weight: 700;
    color: var(--text);
    letter-spacing: 2px;
    text-transform: uppercase;
  }

  .cursor {
    animation: cursorBlink 0.5s step-end infinite;
  }

  @keyframes cursorBlink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0; }
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
    transition: opacity 0.2s ease;
  }

  .input-section.hidden {
    opacity: 0;
    pointer-events: none;
  }

  .answer-input {
    flex: 1;
    padding: 16px 20px;
    border-radius: 14px;
    border: 2px solid var(--outline);
    background: var(--card);
    color: var(--text);
    font-size: 18px;
    text-transform: lowercase;
    transition: all 0.2s;
  }

  .answer-input:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 4px rgba(37, 244, 183, 0.15), 0 0 25px rgba(37, 244, 183, 0.15);
  }

  .answer-input.correct-flash {
    animation: correctFlash 0.4s ease-out;
  }

  .answer-input.wrong-shake {
    animation: wrongShake 0.4s ease-out;
  }

  @keyframes correctFlash {
    0% { box-shadow: 0 0 0 0 rgba(37, 244, 183, 0); }
    30% { box-shadow: 0 0 0 6px rgba(37, 244, 183, 0.4), 0 0 30px rgba(37, 244, 183, 0.3); border-color: var(--accent); }
    100% { box-shadow: 0 0 0 0 rgba(37, 244, 183, 0); }
  }

  @keyframes wrongShake {
    0%, 100% { transform: translateX(0); }
    15%, 45%, 75% { transform: translateX(-6px); border-color: var(--danger); }
    30%, 60%, 90% { transform: translateX(6px); border-color: var(--danger); }
  }

  .answer-input::placeholder {
    color: var(--muted);
    font-size: 14px;
  }

  .submit-btn {
    padding: 16px 28px;
    border-radius: 14px;
    border: 2px solid var(--accent);
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.2), rgba(37, 244, 183, 0.1));
    color: var(--accent);
    font-family: "Press Start 2P", cursive;
    font-size: 10px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .submit-btn:hover:not(:disabled) {
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.3), rgba(37, 244, 183, 0.15));
    box-shadow: 0 0 20px rgba(37, 244, 183, 0.3);
    transform: translateY(-2px);
  }

  .submit-btn:disabled {
    opacity: 0.4;
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
    background: rgba(0, 0, 0, 0.9);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    animation: fadeIn 0.3s ease;
  }

  .game-over-modal {
    background: var(--card);
    border-radius: 24px;
    padding: 40px;
    text-align: center;
    border: 2px solid var(--accent-2);
    box-shadow: 0 0 60px rgba(246, 193, 68, 0.3);
    max-width: 420px;
    width: 90%;
    animation: scaleIn 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
    transition: all 0.3s ease;
  }

  .game-over-modal.winner {
    border-color: var(--accent);
    box-shadow: 0 0 80px rgba(37, 244, 183, 0.5);
  }

  .game-over-modal.loser {
    border-color: var(--danger);
    box-shadow: 0 0 80px rgba(255, 92, 122, 0.4);
  }

  @keyframes scaleIn {
    from { transform: scale(0.8) translateY(20px); opacity: 0; }
    to { transform: scale(1) translateY(0); opacity: 1; }
  }

  .game-over-title {
    font-family: "Press Start 2P", cursive;
    font-size: 28px;
    color: var(--accent-2);
    margin-bottom: 20px;
    animation: titleGlow 2s ease-in-out infinite;
    text-transform: uppercase;
    letter-spacing: 2px;
  }

  .game-over-title.win {
    color: var(--accent);
    animation: winGlow 1s ease-in-out infinite;
    font-size: 36px;
  }

  .game-over-title.lose {
    color: var(--danger);
    animation: loseGlow 1.5s ease-in-out infinite;
    font-size: 36px;
  }

  @keyframes titleGlow {
    0%, 100% { text-shadow: 0 0 20px rgba(246, 193, 68, 0.5); }
    50% { text-shadow: 0 0 40px rgba(246, 193, 68, 0.8); }
  }

  @keyframes winGlow {
    0%, 100% { text-shadow: 0 0 30px rgba(37, 244, 183, 0.6); }
    50% { text-shadow: 0 0 60px rgba(37, 244, 183, 1); }
  }

  @keyframes loseGlow {
    0%, 100% { text-shadow: 0 0 20px rgba(255, 92, 122, 0.4); }
    50% { text-shadow: 0 0 40px rgba(255, 92, 122, 0.7); }
  }

  .game-over-result {
    font-size: 20px;
    font-weight: 700;
    color: var(--text);
    margin-bottom: 8px;
  }

  .game-over-hp {
    font-size: 13px;
    color: var(--muted);
    margin-bottom: 24px;
  }

  .game-over-reason {
    font-size: 12px;
    color: var(--accent-2);
    margin-bottom: 16px;
    font-weight: 600;
  }

  .game-over-stats {
    display: flex;
    justify-content: center;
    gap: 32px;
    margin-bottom: 28px;
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 6px;
    animation: statReveal 0.5s ease-out backwards;
  }

  @keyframes statReveal {
    from { transform: translateY(20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  }

  .stat-label {
    font-size: 10px;
    color: var(--muted);
    text-transform: uppercase;
  }

  .stat-value {
    font-family: "Space Grotesk", sans-serif;
    font-size: 28px;
    font-weight: 800;
    color: var(--accent);
    text-shadow: 0 0 15px rgba(37, 244, 183, 0.4);
  }

  .play-again-btn {
    width: 100%;
    padding: 16px 24px;
    border-radius: 14px;
    border: 2px solid var(--accent);
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.2), rgba(37, 244, 183, 0.1));
    color: var(--accent);
    font-family: "Press Start 2P", cursive;
    font-size: 11px;
    cursor: pointer;
    transition: all 0.2s;
    margin-bottom: 12px;
  }

  .play-again-btn:hover {
    box-shadow: 0 0 25px rgba(37, 244, 183, 0.4);
    transform: translateY(-2px);
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

  .analysis-btn {
    width: 100%;
    padding: 12px 24px;
    border: 2px solid var(--accent);
    border-radius: 10px;
    background: rgba(37, 244, 183, 0.1);
    color: var(--accent);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s;
    margin-bottom: 12px;
  }

  .analysis-btn:hover {
    background: rgba(37, 244, 183, 0.2);
  }

  .analysis-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.92);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 20px;
    animation: fadeIn 0.3s ease;
  }

  .analysis-modal {
    background: var(--card);
    border: 2px solid var(--accent);
    border-radius: 20px;
    padding: 24px;
    max-width: 600px;
    max-height: 85vh;
    width: 100%;
    overflow-y: auto;
    animation: scaleIn 0.3s ease;
  }

  .analysis-title {
    font-size: 20px;
    font-weight: 700;
    color: var(--accent);
    text-align: center;
    margin-bottom: 20px;
  }

  .analysis-content {
    max-height: 60vh;
    overflow-y: auto;
  }

  .loading, .no-analysis {
    text-align: center;
    padding: 40px;
    color: var(--muted);
  }

  .no-analysis-hint {
    font-size: 12px;
    margin-top: 8px;
    opacity: 0.7;
  }

  .participants-summary {
    display: flex;
    justify-content: center;
    gap: 24px;
    margin-bottom: 20px;
    padding: 16px;
    background: rgba(11, 16, 32, 0.6);
    border-radius: 12px;
  }

  .participant-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .participant-name {
    font-weight: 600;
    color: var(--text);
  }

  .participant-stats {
    display: flex;
    gap: 8px;
    font-size: 14px;
  }

  .participant-stats .correct { color: var(--accent); }
  .participant-stats .wrong { color: var(--danger); }

  .rounds-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .round-card {
    background: rgba(11, 16, 32, 0.6);
    border-radius: 12px;
    padding: 16px;
    border: 1px solid var(--outline);
    transition: all 0.2s;
  }

  .round-card:hover {
    border-color: var(--accent);
    transform: translateX(4px);
  }

  .round-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 8px;
  }

  .correct-answer {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    margin-bottom: 12px;
    background: rgba(37, 244, 183, 0.1);
    border: 1px solid var(--accent);
    border-radius: 8px;
  }

  .correct-label {
    font-size: 12px;
    color: var(--accent);
    font-weight: 600;
  }

  .correct-value {
    font-size: 14px;
    color: var(--accent);
    font-weight: 700;
  }

  .round-number {
    background: var(--accent-2);
    color: #0b1020;
    padding: 4px 10px;
    border-radius: 6px;
    font-size: 12px;
    font-weight: 700;
  }

  .round-phrase {
    font-size: 16px;
    font-weight: 700;
    text-transform: uppercase;
    color: var(--text);
  }

  .answers-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .answer-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    border-radius: 8px;
    background: rgba(11, 16, 32, 0.8);
  }

  .answer-item.correct { border-left: 4px solid var(--accent); }
  .answer-item.wrong { border-left: 4px solid var(--danger); }

  .answer-user {
    font-weight: 600;
    color: var(--muted);
    min-width: 80px;
  }

  .answer-text {
    flex: 1;
    color: var(--text);
  }

  .answer-status {
    font-weight: 700;
    font-size: 16px;
  }

  .answer-item.correct .answer-status { color: var(--accent); }
  .answer-item.wrong .answer-status { color: var(--danger); }

  .close-analysis-btn {
    width: 100%;
    padding: 14px;
    margin-top: 20px;
    border: none;
    border-radius: 10px;
    background: var(--outline);
    color: var(--text);
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .close-analysis-btn:hover {
    background: #3a4560;
  }

  .ping-indicator {
    position: absolute;
    top: 12px;
    right: 12px;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: 20px;
    border: 1px solid;
    font-family: 'Space Grotesk', sans-serif;
    font-size: 12px;
    font-weight: 600;
    z-index: 100;
  }

  .ping-icon {
    font-size: 10px;
  }

  .connection-overlay {
    position: fixed;
    inset: 0;
    background: rgba(11, 16, 32, 0.95);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .connection-lost {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    animation: pulse 1.5s ease-in-out infinite;
  }

  .conn-icon {
    font-size: 48px;
  }

  .conn-text {
    font-family: 'Space Grotesk', sans-serif;
    font-size: 18px;
    color: var(--text);
    font-weight: 600;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }
</style>

<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { duel } from '$lib/stores/duel.js';
  import { _ } from 'svelte-i18n';

  const avatars = [
    { id: 'default', emoji: '?', price: 0 },
    { id: 'knight', emoji: '⚔️', price: 50 },
    { id: 'wizard', emoji: '🧙', price: 75 },
    { id: 'archer', emoji: '🏹', price: 75 },
    { id: 'dragon', emoji: '🐉', price: 100 },
    { id: 'skull', emoji: '💀', price: 50 },
    { id: 'fire', emoji: '🔥', price: 60 },
    { id: 'ice', emoji: '❄️', price: 60 },
    { id: 'lightning', emoji: '⚡', price: 80 },
    { id: 'sword', emoji: '🗡️', price: 50 },
    { id: 'shield', emoji: '🛡️', price: 50 },
    { id: 'potion', emoji: '🧪', price: 60 },
    { id: 'crown', emoji: '👑', price: 150 },
    { id: 'star', emoji: '⭐', price: 100 },
    { id: 'moon', emoji: '🌙', price: 80 },
  ];

  const defaultAchievements = [
    { id: 'first_win', nameKey: 'profile.firstVictory', descKey: 'profile.firstVictoryDesc', icon: '🏆', xpReward: 10, coinsReward: 5 },
    { id: 'warrior', nameKey: 'profile.warrior', descKey: 'profile.warriorDesc', icon: '⚔️', xpReward: 25, coinsReward: 15 },
    { id: 'veteran', nameKey: 'profile.veteran', descKey: 'profile.veteranDesc', icon: '🛡️', xpReward: 50, coinsReward: 30 },
    { id: 'champion', nameKey: 'profile.champion', descKey: 'profile.championDesc', icon: '👑', xpReward: 100, coinsReward: 75 },
    { id: 'streak_5', nameKey: 'profile.onFire', descKey: 'profile.onFireDesc', icon: '🔥', xpReward: 20, coinsReward: 10 },
    { id: 'streak_10', nameKey: 'profile.unstoppable', descKey: 'profile.unstoppableDesc', icon: '💥', xpReward: 50, coinsReward: 35 },
    { id: 'games_10', nameKey: 'profile.beginner', descKey: 'profile.beginnerDesc', icon: '🎮', xpReward: 10, coinsReward: 5 },
    { id: 'games_50', nameKey: 'profile.regular', descKey: 'profile.regularDesc', icon: '🎯', xpReward: 25, coinsReward: 20 },
  ];

  const achievementRequirements = {
    'first_win': { need: 1, field: 'wins' },
    'warrior': { need: 10, field: 'wins' },
    'veteran': { need: 50, field: 'wins' },
    'champion': { need: 100, field: 'wins' },
    'streak_5': { need: 5, field: 'streak' },
    'streak_10': { need: 10, field: 'streak' },
    'games_10': { need: 10, field: 'games' },
    'games_50': { need: 50, field: 'games' },
  };

  let editingUsername = false;
  let newUsername = '';
  let usernameError = '';
  let avatarError = '';
  let avatarSuccess = '';
  let savingUsername = false;
  let savingAvatar = false;
  let showAllAchievements = false;
  let showAllGames = false;
  let showAvatarModal = false;
  let showRankModal = false;
  let selectedAchievement = null;

  const ranks = [
    { nameKey: 'rank.newbie', icon: '🥉', eloMin: 0, eloMax: 999 },
    { nameKey: 'rank.apprentice', icon: '🥈', eloMin: 1000, eloMax: 1999 },
    { nameKey: 'rank.expert', icon: '🥇', eloMin: 2000, eloMax: 2999 },
    { nameKey: 'rank.master', icon: '💎', eloMin: 3000, eloMax: 99999 },
    { nameKey: 'rank.struggler', icon: '😔', eloMin: 0, eloMax: 0 },
  ];

  function getRankIcon(rankName) {
    if (rankName.includes('Master')) return '💎';
    if (rankName.includes('Expert')) return '🥇';
    if (rankName.includes('Apprentice')) return '🥈';
    if (rankName.includes('Struggler')) return '😔';
    return '🥉';
  }

  function getTranslatedRankName(rank) {
    const names = {
      'newbie': $_('rank.newbie'),
      'apprentice': $_('rank.apprentice'),
      'expert': $_('rank.expert'),
      'master': $_('rank.master'),
      'struggler': $_('rank.struggler')
    };
    return names[rank] || names['newbie'];
  }

  $: currentAvatar = avatars.find(a => a.id === ($duel.userAvatar || 'default')) || avatars[0];
  $: wins = parseInt($duel.profileWins) || 0;
  $: games = parseInt($duel.profileDuels) || 0;
  $: streak = parseInt($duel.profileStreak) || 0;
  $: userCoins = $duel.profileCoins || 0;
  $: unlockedAvatars = $duel.unlockedAvatars || ['default'];
  
  function isAvatarUnlocked(avatarId) {
    return unlockedAvatars.includes(avatarId);
  }
  
  function checkLocallyUnlocked(achievementId, achWins, achGames, achStreak) {
    const req = achievementRequirements[achievementId];
    if (!req) return false;
    if (req.field === 'wins') return achWins >= req.need;
    if (req.field === 'games') return achGames >= req.need;
    if (req.field === 'streak') return achStreak >= req.need;
    return false;
  }
  
  $: unlockedFromApi = $duel.achievements?.filter(a => a.unlocked).length || 0;
  $: unlockedCount = $duel.achievements && $duel.achievements.length > 0 
    ? unlockedFromApi 
    : allAchievements.filter(a => checkLocallyUnlocked(a.id, wins, games, streak)).length;
  
  $: allAchievements = (() => {
    const apiUnlocked = new Set($duel.achievements?.filter(a => a.unlocked).map(a => a.id) || []);
    const apiById = {};
    ($duel.achievements || []).forEach(a => { apiById[a.id] = a; });
    return defaultAchievements.map(a => {
      const apiData = apiById[a.id] || {};
      const locallyUnlocked = checkLocallyUnlocked(a.id, wins, games, streak);
      return {
        ...a,
        unlocked: apiUnlocked.has(a.id) || locallyUnlocked,
        xpReward: apiData.xp_reward ?? a.xpReward,
        coinsReward: apiData.coins_reward ?? a.coinsReward
      };
    });
  })();

  function getProgress(achievementId) {
    if (achievementId.unlocked) return 100;
    const req = achievementRequirements[achievementId.id];
    if (!req) return 0;
    let current = 0;
    if (req.field === 'wins') current = wins;
    else if (req.field === 'games') current = games;
    else if (req.field === 'streak') current = streak;
    return Math.min(100, Math.round((current / req.need) * 100));
  }

  function getCurrent(achievementId) {
    if (achievementId.unlocked) return '';
    const req = achievementRequirements[achievementId.id];
    if (!req) return '';
    if (req.field === 'wins') return `${wins}/${req.need}`;
    if (req.field === 'games') return `${games}/${req.need}`;
    if (req.field === 'streak') return `${streak}/${req.need}`;
    return '';
  }

  function getLevelProgress(xp) {
    // Level formula: floor(sqrt(xp / 100)), starts at level 1
    // Progress to next level: (xp / xpForNextLevel) * 100
    const currentLevel = Math.floor(Math.sqrt(xp / 100)) + 1;
    const xpForCurrent = Math.pow(currentLevel - 1, 2) * 100;
    const xpForNext = Math.pow(currentLevel, 2) * 100;
    const xpInLevel = xp - xpForCurrent;
    const xpNeeded = xpForNext - xpForCurrent;
    return Math.min(100, Math.round((xpInLevel / xpNeeded) * 100));
  }

  onMount(async () => {
    duel.init();
    if ($duel.authMode === 'auth') {
      await duel.fetchUserRating();
      await duel.fetchAchievements();
    }
  });

  function handleLogout() {
    if (confirm($_('confirm.logout'))) {
      duel.logout();
      goto('/');
    }
  }

  function startEditUsername() {
    newUsername = $duel.authedUsername || '';
    editingUsername = true;
    usernameError = '';
  }

  function cancelEditUsername() {
    editingUsername = false;
    newUsername = '';
    usernameError = '';
  }

  async function saveUsername() {
    if (!newUsername.trim()) {
      usernameError = $_('profile.usernameEmpty');
      return;
    }
    if (newUsername.length < 3 || newUsername.length > 30) {
      usernameError = $_('profile.usernameLength');
      return;
    }
    savingUsername = true;
    usernameError = '';
    const result = await duel.updateUsername(newUsername.trim());
    savingUsername = false;
    if (result.error) {
      usernameError = result.error;
    } else {
      editingUsername = false;
      newUsername = '';
    }
  }

  async function selectAvatar(avatarId) {
    avatarError = '';
    avatarSuccess = '';
    
    const avatarData = avatars.find(a => a.id === avatarId);
    if (!avatarData) return;
    
    if (!isAvatarUnlocked(avatarId)) {
      if (userCoins < avatarData.price) {
        avatarError = 'Недостаточно монет!';
        return;
      }
      
      savingAvatar = true;
      const result = await duel.buyAvatar(avatarId);
      savingAvatar = false;
      
      if (result.error) {
        avatarError = result.error;
        return;
      }
      
      avatarSuccess = 'Аватарка куплена!';
      await duel.fetchUserRating();
    }
    
    savingAvatar = true;
    const result = await duel.updateAvatar(avatarId);
    savingAvatar = false;
    if (result.error) {
      avatarError = result.error;
    }
  }
  
</script>

<div class="wrap">
  <!-- Header Card -->
  <div class="header-card">
    <div class="avatar-section">
      <div class="avatar-large">
        {currentAvatar.emoji}
      </div>
      {#if $duel.authMode === 'auth'}
        <button 
          class="avatar-edit-btn" 
          on:click={() => showAvatarModal = true}
          disabled={savingAvatar}
        >
          ✏️
        </button>
      {/if}
    </div>
    
    <div class="user-info">
      {#if editingUsername}
        <div class="username-edit">
          <input 
            type="text" 
            bind:value={newUsername} 
            placeholder={$_('profile.enterNewUsername')}
            maxlength="30"
            class="username-input"
          />
          <div class="username-actions">
            <button class="save-btn" on:click={saveUsername} disabled={savingUsername}>
              {savingUsername ? $_('profile.saving') : $_('profile.save')}
            </button>
            <button class="cancel-btn" on:click={cancelEditUsername}>
              {$_('profile.cancel')}
            </button>
          </div>
          {#if usernameError}
            <div class="error-text">{usernameError}</div>
          {/if}
        </div>
      {:else}
        <h1 class="username">{$duel.profileUser || $_('profile.guestName')}</h1>
        {#if $duel.authMode === 'auth'}
          <button class="edit-btn" on:click={startEditUsername}>✏️</button>
        {/if}
      {/if}
      
      {#if $duel.authMode === 'auth'}
        <button class="rank-badge" on:click={() => showRankModal = true}>
          <span class="rank-icon">{getRankIcon($duel.profileRankName)}</span>
          <span class="rank-text">{getTranslatedRankName($duel.profileRank)}</span>
        </button>
      {/if}
    </div>
  </div>

  <!-- Main Content Grid -->
  <div class="content-grid">
    <!-- Left Column: Stats -->
    <div class="left-column">
        <div class="module-card">
          <h3 class="module-title">📊 {$_('profile.stats')}</h3>
        <div class="stats-list">
          <div class="stat-row">
            <span class="stat-name">{$_('profile.elo')}</span>
            <span class="stat-value highlight">{$duel.profileElo || 1000}</span>
          </div>
          <div class="stat-row">
            <span class="stat-name">{$_('profile.games')}</span>
            <span class="stat-value">{games}</span>
          </div>
          <div class="stat-row">
            <span class="stat-name">{$_('profile.wins')}</span>
            <span class="stat-value win">{wins}</span>
          </div>
          <div class="stat-row">
            <span class="stat-name">{$_('profile.accuracy')}</span>
            <span class="stat-value">{$duel.profileAcc || '-'}%</span>
          </div>
          <div class="stat-row">
            <span class="stat-name">{$_('profile.streak')}</span>
            <span class="stat-value streak">🔥 {streak}</span>
          </div>
          {#if $duel.authMode === 'auth'}
            <div class="stat-row">
              <span class="stat-name">{$_('profile.coins')}</span>
              <span class="stat-value coins">🪙 {$duel.profileCoins || 0}</span>
            </div>
            <div class="level-section">
              <div class="level-header">
                <span class="stat-name">{$_('profile.level')} {$duel.profileLevel || 1}</span>
                <span class="stat-name">{$duel.profileXP || 0} XP</span>
              </div>
              <div class="level-progress">
                <div class="level-bar" style="width: {getLevelProgress($duel.profileXP || 0)}%"></div>
              </div>
            </div>
          {/if}
        </div>
      </div>

      {#if $duel.authMode !== 'auth'}
        <div class="module-card guest-prompt">
          <p>{$_('profile.loginToSave')}</p>
          <button class="login-btn" on:click={() => goto('/auth')}>
            {$_('profile.loginRegister')}
          </button>
        </div>
      {:else}
        <div class="module-card">
          <h3 class="module-title">⚙️ {$_('profile.settings')}</h3>
          <button class="logout-btn" on:click={handleLogout}>
            {$_('profile.logout')}
          </button>
        </div>
      {/if}
    </div>

    <!-- Right Column: Achievements + Recent Games -->
    <div class="right-column">
      {#if $duel.authMode === 'auth'}
        <div class="module-card achievements-card">
          <div class="module-header">
            <h3 class="module-title">🏆 {$_('profile.achievements')}</h3>
            <span class="achievement-count">{unlockedCount}/{allAchievements.length}</span>
          </div>
          
          <div class="achievements-list">
            {#each allAchievements as achievement}
              <button 
                class="achievement-row" 
                class:unlocked={achievement.unlocked}
                class:locked={!achievement.unlocked}
                on:click={() => selectedAchievement = achievement}
              >
                <div class="achievement-icon">{achievement.unlocked ? achievement.icon : '🔒'}</div>
                <div class="achievement-info">
                  <div class="achievement-name">{$_(achievement.nameKey)}</div>
                  <div class="achievement-desc">
                    {achievement.unlocked ? $_(achievement.descKey) : getCurrent(achievement)}
                  </div>
                  {#if !achievement.unlocked}
                    <div class="progress-bar">
                      <div class="progress-fill" style="width: {getProgress(achievement)}%"></div>
                    </div>
                  {/if}
                </div>
                {#if achievement.unlocked}
                  <div class="achievement-check">✓</div>
                {/if}
              </button>
            {/each}
          </div>
        </div>
      {/if}

      <div class="module-card">
        <button class="module-header-btn" on:click={() => showAllGames = !showAllGames}>
          <h3 class="module-title">🎮 {$_('profile.recentGames')}</h3>
          {#if $duel.profileDuelsList && $duel.profileDuelsList.length > 0}
            <span class="expand-icon">{showAllGames ? '▼' : '▶'}</span>
          {/if}
        </button>
        {#if $duel.profileDuelsList && $duel.profileDuelsList.length > 0}
          <div class="games-list" class:expanded={showAllGames}>
            {#each (showAllGames ? $duel.profileDuelsList : $duel.profileDuelsList.slice(0, 5)) as d}
              <div class="game-row">
                <div class="game-result" class:win={d.badgeClass === 'win'} class:loss={d.badgeClass === 'loss'}>
                  {d.resultLabel}
                </div>
                <div class="game-info">
                  <span class="game-opponent">{d.opponent}</span>
                  <span class="game-date">{d.created}</span>
                </div>
              </div>
            {/each}
          </div>
          {#if $duel.profileDuelsList.length > 5}
            <button class="expand-btn" on:click={() => showAllGames = !showAllGames}>
              {showAllGames ? $_('profile.showLess') : $_('profile.showAll', { values: { count: $duel.profileDuelsList.length } })}
            </button>
          {/if}
        {:else}
          <div class="empty-games">
            <p>{$_('profile.noGamesYet')}</p>
            <a href="/play" class="play-link">{$_('profile.playNow')}</a>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<!-- Avatar Modal -->
{#if showAvatarModal}
  <div class="modal-overlay" on:click={() => showAvatarModal = false}>
    <div class="modal-content avatar-shop" on:click|stopPropagation>
      <div class="shop-header">
        <h3 class="modal-title">{$_('shop.title')}</h3>
        <div class="shop-coins">
          <span class="coin-icon">🪙</span>
          <span class="coin-amount">{userCoins}</span>
        </div>
      </div>
      
      {#if avatarError}
        <div class="shop-error">{avatarError}</div>
      {/if}
      {#if avatarSuccess}
        <div class="shop-success">{avatarSuccess}</div>
      {/if}
      
      <div class="avatar-grid">
        {#each avatars as avatar}
          {@const unlocked = isAvatarUnlocked(avatar.id)}
          {@const selected = $duel.userAvatar === avatar.id}
          {@const canAfford = userCoins >= avatar.price}
          <button 
            class="avatar-card" 
            class:selected={selected}
            class:locked={!unlocked}
            on:click={() => selectAvatar(avatar.id)}
            disabled={savingAvatar}
          >
            <span class="avatar-emoji">{avatar.emoji}</span>
            {#if selected}
              <span class="avatar-status selected">Выбран</span>
            {:else if unlocked}
              <span class="avatar-status owned">Куплено</span>
            {:else}
              <span class="avatar-price" class:cant-afford={!canAfford}>🪙 {avatar.price}</span>
            {/if}
          </button>
        {/each}
      </div>
      
      <div class="shop-actions">
        <button class="shop-close-btn" on:click={() => { showAvatarModal = false; avatarError = ''; avatarSuccess = ''; }}>
          Закрыть
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Achievement Detail Modal -->
{#if selectedAchievement}
  <div class="modal-overlay" on:click={() => selectedAchievement = null}>
    <div class="modal-content achievement-detail" on:click|stopPropagation>
      <div class="achievement-detail-icon">{selectedAchievement.icon}</div>
      <h3 class="modal-title">{$_(selectedAchievement.nameKey)}</h3>
      <p class="achievement-detail-desc">{$_(selectedAchievement.descKey)}</p>
      <div class="achievement-rewards">
        <span class="reward-item xp">+{selectedAchievement.xpReward} XP</span>
        <span class="reward-item coins">+{selectedAchievement.coinsReward} 🪙</span>
      </div>
      {#if selectedAchievement.unlocked}
        <div class="achievement-status unlocked">✓ {$_('profile.unlocked')}</div>
      {:else}
        <div class="achievement-progress-text">{getCurrent(selectedAchievement)}</div>
      {/if}
      <div class="modal-actions">
        <button class="close-btn" on:click={() => selectedAchievement = null}>
          {$_('profile.close')}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Rank Info Modal -->
{#if showRankModal}
  <div class="modal-overlay" on:click={() => showRankModal = false}>
    <div class="modal-content rank-modal" on:click|stopPropagation>
      <h3 class="modal-title">🏆 {$_('profile.ranks')}</h3>
      <div class="ranks-list">
        {#each ranks as rank}
          <div class="rank-row" class:current={$duel.profileRankName.includes($_(rank.nameKey))}>
            <span class="rank-row-icon">{rank.icon}</span>
            <div class="rank-row-info">
              <span class="rank-row-name">{$_(rank.nameKey)}</span>
              <span class="rank-row-elo">{rank.eloMin} - {rank.eloMax === 99999 ? '∞' : rank.eloMax} ELO</span>
            </div>
          </div>
        {/each}
      </div>
      <div class="modal-actions">
        <button class="close-btn" on:click={() => showRankModal = false}>
          {$_('profile.close')}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .wrap {
    max-width: 1100px;
    margin: 0 auto;
    padding: 24px;
  }

  /* Header Card */
  .header-card {
    display: flex;
    align-items: center;
    gap: 24px;
    background: var(--card);
    border-radius: 20px;
    padding: 24px 32px;
    border: 1px solid var(--outline);
    margin-bottom: 24px;
  }

  .avatar-section {
    position: relative;
  }

  .avatar-large {
    width: 100px;
    height: 100px;
    border-radius: 50%;
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.2), rgba(246, 193, 68, 0.1));
    border: 3px solid var(--accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 48px;
  }

  .avatar-edit-btn {
    position: absolute;
    bottom: 0;
    right: 0;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    border: 2px solid var(--accent);
    background: var(--card);
    cursor: pointer;
    font-size: 14px;
    transition: all 0.2s;
  }

  .avatar-edit-btn:hover {
    background: var(--accent);
  }

  .user-info {
    flex: 1;
    position: relative;
  }

  .username {
    font-size: 28px;
    font-weight: 700;
    margin: 0 0 8px 0;
    color: var(--text);
  }

  .edit-btn {
    position: absolute;
    top: 0;
    right: 0;
    background: none;
    border: 1px solid var(--outline);
    border-radius: 8px;
    padding: 6px 10px;
    cursor: pointer;
    font-size: 14px;
    transition: all 0.2s;
  }

  .edit-btn:hover {
    border-color: var(--accent);
  }

  .rank-badge {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.15), rgba(246, 193, 68, 0.1));
    border: 1px solid var(--accent);
    border-radius: 20px;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .rank-badge:hover {
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.25), rgba(246, 193, 68, 0.2));
    transform: scale(1.02);
  }

  .rank-icon {
    font-size: 18px;
  }

  .rank-text {
    font-weight: 600;
    color: var(--accent);
  }

  .rank-modal {
    max-width: 350px;
  }

  .ranks-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 20px;
  }

  .rank-row {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: rgba(11, 16, 32, 0.6);
    border-radius: 10px;
    border: 1px solid var(--outline);
  }

  .rank-row.current {
    border-color: var(--accent);
    background: rgba(37, 244, 183, 0.1);
  }

  .rank-row-icon {
    font-size: 28px;
  }

  .rank-row-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .rank-row-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text);
  }

  .rank-row-elo {
    font-size: 12px;
    color: var(--muted);
  }

  .username-edit {
    max-width: 300px;
  }

  .username-input {
    width: 100%;
    padding: 12px 16px;
    border-radius: 10px;
    border: 1px solid var(--outline);
    background: rgba(11, 16, 32, 0.8);
    color: var(--text);
    font-size: 16px;
    margin-bottom: 8px;
  }

  .username-actions {
    display: flex;
    gap: 8px;
  }

  .save-btn, .cancel-btn {
    padding: 8px 16px;
    border-radius: 8px;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .save-btn {
    background: var(--accent);
    border: none;
    color: var(--bg);
  }

  .cancel-btn {
    background: transparent;
    border: 1px solid var(--outline);
    color: var(--muted);
  }

  /* Content Grid */
  .content-grid {
    display: grid;
    grid-template-columns: 320px 1fr;
    gap: 24px;
  }

  .left-column, .right-column {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  /* Module Card */
  .module-card {
    background: var(--card);
    border-radius: 16px;
    padding: 20px;
    border: 1px solid var(--outline);
  }

  .module-title {
    font-size: 14px;
    font-weight: 700;
    margin: 0 0 16px 0;
    color: var(--text);
    letter-spacing: 1px;
  }

  .module-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .achievement-count {
    font-size: 12px;
    color: var(--accent);
    font-weight: 600;
  }

  /* Stats */
  .stats-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .stat-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 0;
    border-bottom: 1px solid rgba(43, 52, 74, 0.5);
  }

  .stat-row:last-child {
    border-bottom: none;
  }

  .stat-name {
    font-size: 13px;
    color: var(--muted);
  }

  .stat-value {
    font-size: 16px;
    font-weight: 700;
    color: var(--text);
  }

  .stat-value.highlight {
    color: var(--accent);
    font-size: 20px;
  }

  .stat-value.win {
    color: var(--accent);
  }

  .stat-value.streak {
    color: var(--accent-2);
  }

  .stat-value.coins {
    color: #f6c144;
  }

  .level-section {
    margin-top: 8px;
  }

  .level-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 6px;
  }

  .level-progress {
    height: 8px;
    background: rgba(43, 52, 74, 0.5);
    border-radius: 4px;
    overflow: hidden;
  }

  .level-bar {
    height: 100%;
    background: linear-gradient(90deg, var(--accent), var(--accent-2));
    border-radius: 4px;
    transition: width 0.3s ease;
  }

  /* Achievements */
  .achievements-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-height: 400px;
    overflow-y: auto;
  }

  .achievement-row {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: rgba(11, 16, 32, 0.6);
    border-radius: 10px;
    border: 1px solid var(--outline);
    transition: all 0.2s;
  }

  .achievement-row.unlocked {
    border-color: var(--accent);
    background: rgba(37, 244, 183, 0.08);
  }

  .achievement-icon {
    font-size: 28px;
    width: 40px;
    text-align: center;
  }

  .achievement-info {
    flex: 1;
  }

  .achievement-name {
    font-size: 13px;
    font-weight: 600;
    color: var(--text);
    margin-bottom: 2px;
  }

  .achievement-desc {
    font-size: 11px;
    color: var(--muted);
  }

  .progress-bar {
    height: 4px;
    background: rgba(43, 52, 74, 0.5);
    border-radius: 2px;
    margin-top: 6px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--accent), var(--accent-2));
    border-radius: 2px;
    transition: width 0.3s;
  }

  .achievement-check {
    color: var(--accent);
    font-weight: 700;
    font-size: 16px;
  }

  /* Games */
  .module-header-btn {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
    margin-bottom: 16px;
  }

  .expand-icon {
    font-size: 12px;
    color: var(--muted);
  }

  .games-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-height: 300px;
    overflow: hidden;
    transition: max-height 0.3s ease;
  }

  .games-list.expanded {
    max-height: none;
  }

  .expand-btn {
    width: 100%;
    padding: 12px;
    margin-top: 12px;
    background: rgba(11, 16, 32, 0.6);
    border: 1px solid var(--outline);
    border-radius: 10px;
    color: var(--accent);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .expand-btn:hover {
    background: rgba(37, 244, 183, 0.1);
    border-color: var(--accent);
  }

  .game-row {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px;
    background: rgba(11, 16, 32, 0.6);
    border-radius: 8px;
  }

  .game-result {
    padding: 4px 10px;
    border-radius: 6px;
    font-size: 10px;
    font-weight: 700;
    min-width: 50px;
    text-align: center;
  }

  .game-result.win {
    background: rgba(37, 244, 183, 0.2);
    color: var(--accent);
  }

  .game-result.loss {
    background: rgba(255, 92, 122, 0.2);
    color: var(--danger);
  }

  .game-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .game-opponent {
    font-size: 13px;
    color: var(--text);
  }

  .game-date {
    font-size: 10px;
    color: var(--muted);
  }

  .empty-games {
    text-align: center;
    padding: 20px;
    color: var(--muted);
  }

  .play-link {
    display: inline-block;
    margin-top: 12px;
    padding: 10px 24px;
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.2), rgba(37, 244, 183, 0.05));
    border: 1px solid var(--accent);
    border-radius: 10px;
    color: var(--accent);
    text-decoration: none;
    font-weight: 600;
    transition: all 0.2s;
  }

  .more-link {
    text-align: center;
    padding: 10px;
    color: var(--muted);
    font-size: 12px;
  }

  /* Buttons */
  .logout-btn, .login-btn {
    width: 100%;
    padding: 12px;
    border-radius: 10px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .logout-btn {
    background: rgba(255, 92, 122, 0.1);
    border: 1px solid var(--danger);
    color: var(--danger);
  }

  .logout-btn:hover {
    background: rgba(255, 92, 122, 0.2);
  }

  .login-btn {
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.2), rgba(37, 244, 183, 0.05));
    border: 1px solid var(--accent);
    color: var(--accent);
  }

  .login-btn:hover {
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.3), rgba(37, 244, 183, 0.1));
  }

  /* Guest Prompt */
  .guest-prompt {
    text-align: center;
    color: var(--muted);
  }

  .guest-prompt p {
    margin: 0 0 16px 0;
  }

  /* Modal Overlay */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(6px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal-content {
    background: var(--card);
    border: 2px solid var(--accent);
    border-radius: 20px;
    padding: 24px;
    max-width: 420px;
    width: 90%;
    box-shadow: 0 0 50px rgba(37, 244, 183, 0.25);
  }

  .avatar-shop {
    padding: 20px;
  }

  .modal-title {
    font-size: 18px;
    font-weight: 700;
    color: var(--text);
  }

  /* Shop Header */
  .shop-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .shop-coins {
    display: flex;
    align-items: center;
    gap: 6px;
    background: rgba(246, 193, 68, 0.15);
    padding: 6px 14px;
    border-radius: 16px;
    border: 1px solid rgba(246, 193, 68, 0.3);
  }

  .coin-icon {
    font-size: 18px;
  }

  .coin-amount {
    color: #f6c144;
    font-weight: 700;
    font-size: 15px;
  }

  .shop-error {
    color: var(--danger);
    font-size: 13px;
    margin-bottom: 12px;
    text-align: center;
    padding: 8px;
    background: rgba(255, 92, 122, 0.1);
    border-radius: 8px;
  }

  .shop-success {
    color: var(--accent);
    font-size: 13px;
    margin-bottom: 12px;
    text-align: center;
    padding: 8px;
    background: rgba(37, 244, 183, 0.1);
    border-radius: 8px;
  }

  /* Avatar Grid */
  .avatar-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 12px;
    margin-bottom: 16px;
  }

  .avatar-card {
    aspect-ratio: 1;
    border-radius: 14px;
    border: 2px solid var(--outline);
    background: rgba(11, 16, 32, 0.9);
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: 8px;
  }

  .avatar-card:hover {
    border-color: var(--accent);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(37, 244, 183, 0.15);
  }

  .avatar-card.selected {
    border-color: var(--accent);
    background: rgba(37, 244, 183, 0.15);
    box-shadow: 0 0 15px rgba(37, 244, 183, 0.2);
  }

  .avatar-card.locked {
    opacity: 0.75;
  }

  .avatar-card.locked:hover {
    opacity: 1;
  }

  .avatar-card:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .avatar-emoji {
    font-size: 36px;
    line-height: 1;
  }

  .avatar-status {
    font-size: 10px;
    font-weight: 600;
    padding: 2px 6px;
    border-radius: 4px;
  }

  .avatar-status.selected {
    background: rgba(37, 244, 183, 0.25);
    color: var(--accent);
  }

  .avatar-status.owned {
    background: rgba(37, 244, 183, 0.15);
    color: var(--accent);
  }

  .avatar-price {
    font-size: 11px;
    font-weight: 600;
    color: #f6c144;
  }

  .avatar-price.cant-afford {
    color: var(--danger);
  }

  /* Shop Actions */
  .shop-actions {
    display: flex;
    justify-content: center;
    margin-top: 8px;
  }

  .shop-close-btn {
    padding: 10px 28px;
    border-radius: 10px;
    border: 1px solid var(--outline);
    background: transparent;
    color: var(--muted);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .shop-close-btn:hover {
    border-color: var(--text);
    color: var(--text);
  }

  /* Achievement Detail Modal */
  .achievement-detail {
    text-align: center;
  }

  .achievement-detail-icon {
    font-size: 64px;
    margin-bottom: 16px;
  }

  .achievement-detail-desc {
    color: var(--muted);
    margin: 16px 0;
    font-size: 14px;
  }

  .achievement-progress-text {
    color: var(--accent-2);
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 16px;
  }

  .achievement-reward {
    color: var(--accent);
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 16px;
  }

  .achievement-rewards {
    display: flex;
    justify-content: center;
    gap: 16px;
    margin-bottom: 16px;
  }

  .reward-item {
    padding: 6px 12px;
    border-radius: 20px;
    font-size: 13px;
    font-weight: 600;
  }

  .reward-item.xp {
    background: rgba(37, 244, 183, 0.15);
    color: var(--accent);
  }

  .reward-item.coins {
    background: rgba(246, 193, 68, 0.15);
    color: #f6c144;
  }

  .achievement-status.unlocked {
    color: var(--accent);
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 16px;
  }

  /* Achievement Row Button */
  .achievement-row {
    width: 100%;
    text-align: left;
    background: none;
    border: 1px solid var(--outline);
    cursor: pointer;
  }

  .achievement-row:hover {
    border-color: var(--accent);
  }

  /* Responsive */
  @media (max-width: 768px) {
    .content-grid {
      grid-template-columns: 1fr;
    }
    
    .header-card {
      flex-direction: column;
      text-align: center;
    }
    
    .username {
      font-size: 24px;
    }
  }
</style>

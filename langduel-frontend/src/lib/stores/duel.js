import { writable, get } from 'svelte/store';
import { goto } from '$app/navigation';

const initialState = {
  status: 'Disconnected',
  flowMode: 'create',
  authMode: 'guest',
  authChoiceMade: false,
  authTab: 'login',
  wsUrl: '',
  connectionStatus: 'disconnected', // connected, connecting, disconnected
  ping: -1,
  createUser: '',
  joinUser: '',
  createRoom: '',
  joinRoom: '',
  createLang: 'en-ru',
  createTopic: 'default',
  createDifficulty: 'intermediate',
  createRounds: 5,
  generatingPhrases: false,
  phrasesGenerated: false,
  generationError: '',
  jwtToken: '',
  authedUsername: '',
  authedUserID: '',
  userAvatar: 'default',
  opponentAvatar: 'default',
  startError: '',
  gameEndedNormally: false,
  authError: '',
  isCreator: false,
  currentRoom: '',
  currentUser: '',
  currentLang: 'en',
  currentTopic: 'default',
  currentDifficulty: 'intermediate',
  lobbyText: 'lobby.waiting',
  lobbyCopyNote: '',
  createCopyNote: '',
  reconnectNote: '',
  playerA: 'Player A',
  playerB: 'Player B',
  hp: {},
  elo: {},
  promptText: 'Waiting for round...',
  timerText: '-',
  roundInfo: '-',
  totalPhrases: 0,
  correctCount: 0,
  wrongCount: 0,
  totalDamage: 0,
  lastDamage: 0,
  lastDamageTo: '',
  playerADamage: 0,
  playerBDamage: 0,
  totalSpeed: 0,
  speedCount: 0,
  gameOverOpen: false,
  gameOverText: 'Game over',
  gameOverHP: '-',
  gameOverReason: '',
  isGameWinner: null,
  eloChange: {},
  currentDuelId: '',
  profileUser: '-',
  profileDuels: '-',
  profileWins: '-',
  profileAcc: '-',
  profileStreak: '-',
  profileDuelsCount: '0',
  profileDuelsList: [],
  profileElo: 1000,
  profileRank: 'newbie',
  profileRankName: '🥉 Newbie',
  profileCoins: 0,
  profileXP: 0,
  profileLevel: 1,
  leaderboard: [],
  achievements: [],
  hitA: false,
  hitB: false,
  attackA: false,
  attackB: false,
  selfDamageA: false,
  selfDamageB: false,
  inputCorrect: false,
  inputWrong: false,
  coins: 0,
  unlockedAvatars: ['default'],
};

const STORAGE_TOKEN_KEY = 'langduel_token';
const STORAGE_USER_KEY = 'langduel_user';
const STORAGE_GUEST_KEY = 'langduel_guest';
const STORAGE_LAST_KEY = 'langduel_last_session';
const STORAGE_AUTH_CHOICE_KEY = 'langduel_auth_choice';
const STORAGE_AVATAR_KEY = 'langduel_avatar';
const roundDurationMs = 10000;

const state = writable({ ...initialState });

let ws = null;
let roundStartAt = null;
let countdownTimer = null;
let initialized = false;

function setState(patch) {
  state.update((s) => ({ ...s, ...patch }));
}

function httpBaseFromWs(url) {
  try {
    const u = new URL(url);
    const proto = u.protocol === 'wss:' ? 'https:' : 'http:';
    return proto + '//' + u.host;
  } catch {
    if (url.startsWith('ws://')) return 'http://' + url.slice(5).split('/')[0];
    if (url.startsWith('wss://')) return 'https://' + url.slice(6).split('/')[0];
    return url;
  }
}

function buildRoomLink(roomId) {
  if (typeof window === 'undefined') return '';
  const base = window.location.origin;
  return `${base}/play?room=${encodeURIComponent(roomId)}`;
}

function showNote(key, text) {
  setState({ [key]: text });
  setTimeout(() => setState({ [key]: '' }), 1500);
}

async function copyLink(roomId, noteKey) {
  const link = buildRoomLink(roomId);
  try {
    await navigator.clipboard.writeText(link);
    showNote(noteKey, 'Link copied');
  } catch {
    showNote(noteKey, 'Copy failed');
  }
}

function resetStats() {
  setState({
    correctCount: 0,
    wrongCount: 0,
    totalDamage: 0,
    lastDamage: 0,
    lastDamageTo: '',
    playerADamage: 0,
    playerBDamage: 0,
    totalSpeed: 0,
    speedCount: 0
  });
}

function updateStats(correct, damage, speed, targetId) {
  state.update((s) => {
    const next = { ...s };
    if (correct) next.correctCount += 1;
    else next.wrongCount += 1;
    next.totalDamage += damage;
    next.lastDamage = damage;
    next.lastDamageTo = targetId || '';
    
    // Track damage per player
    if (targetId === next.playerA) {
      next.playerADamage = (next.playerADamage || 0) + damage;
    } else if (targetId === next.playerB) {
      next.playerBDamage = (next.playerBDamage || 0) + damage;
    }
    
    if (speed > 0) {
      next.totalSpeed += speed;
      next.speedCount += 1;
    }
    return next;
  });
}

function avgSpeed(s) {
  if (!s.speedCount) return '-';
  return Math.round(s.totalSpeed / s.speedCount);
}

const avatarEmojis = {
  'default': '?',
  'knight': '🛡️',
  'wizard': '🧙',
  'archer': '🏹',
  'dragon': '🐉',
  'skull': '💀',
  'fire': '🔥',
  'ice': '❄️',
  'lightning': '⚡',
  'sword': '⚔️',
  'shield': '🛡️',
  'potion': '🧪',
  'crown': '👑',
  'star': '⭐',
  'moon': '🌙',
};

const avatarPrices = {
  'default': 0,
  'knight': 50,
  'wizard': 75,
  'archer': 75,
  'dragon': 100,
  'skull': 50,
  'fire': 60,
  'ice': 60,
  'lightning': 80,
  'sword': 50,
  'shield': 50,
  'potion': 60,
  'crown': 150,
  'star': 100,
  'moon': 80,
};

const avatarProjectiles = {
  'default': { shape: 'circle', color: '#25f4b7' },
  'knight': { shape: 'shield', color: '#3498db' },
  'wizard': { shape: 'orb', color: '#9b59b6' },
  'archer': { shape: 'arrow', color: '#2ecc71' },
  'dragon': { shape: 'fire', color: '#e74c3c' },
  'skull': { shape: 'skull', color: '#95a5a6' },
  'fire': { shape: 'flame', color: '#f39c12' },
  'ice': { shape: 'shard', color: '#00cec9' },
  'lightning': { shape: 'bolt', color: '#fdcb6e' },
  'sword': { shape: 'blade', color: '#d63031' },
  'shield': { shape: 'shield', color: '#3498db' },
  'potion': { shape: 'orb', color: '#00b894' },
  'crown': { shape: 'crown', color: '#fdcb6e' },
  'star': { shape: 'star', color: '#ffeaa7' },
  'moon': { shape: 'beam', color: '#a29bfe' },
};

function getAvatarEmoji(avatarId) {
  return avatarEmojis[avatarId] || avatarEmojis['default'];
}

function getAvatarPrice(avatarId) {
  return avatarPrices[avatarId] || 0;
}

function getAvatarProjectile(avatarId) {
  return avatarProjectiles[avatarId] || avatarProjectiles['default'];
}

function applyHP(nextHP) {
  setState({ hp: nextHP || {} });
}

function ensurePlayers(list) {
  if (!list || list.length === 0) return;
  setState({
    playerA: list[0] || 'Player A',
    playerB: list[1] || 'Player B'
  });
}

function showRoundInfo(data) {
  if (data && data.round) {
    const total = data.total_phrases || data.totalPhrases || 0;
    if (total > 0) {
      setState({ roundInfo: `Round ${data.round} / ${total}`, totalPhrases: total });
    } else {
      setState({ roundInfo: `Round ${data.round}` });
    }
  }
}

function startCountdown() {
  stopCountdown();
  const end = Date.now() + roundDurationMs;
  countdownTimer = setInterval(() => {
    const left = Math.max(0, end - Date.now());
    setState({ timerText: `${Math.ceil(left / 1000)}s` });
    if (left <= 0) {
      stopCountdown();
    }
  }, 250);
}

function stopCountdown() {
  if (countdownTimer) {
    clearInterval(countdownTimer);
    countdownTimer = null;
  }
}

function hitEffect(playerId) {
  const s = get(state);
  const currentUser = s.currentUser;
  
  // Attacker's animation (the one who dealt damage)
  if (playerId === s.playerA) {
    // Player B hit, so player A attacked
    setState({ attackA: true });
    setTimeout(() => setState({ attackA: false }), 250);
  }
  if (playerId === s.playerB) {
    // Player A hit, so player B attacked
    setState({ attackB: true });
    setTimeout(() => setState({ attackB: false }), 250);
  }
  
  // Defender's hit animation
  if (playerId === s.playerA) {
    setState({ hitA: true });
    setTimeout(() => setState({ hitA: false }), 350);
  }
  if (playerId === s.playerB) {
    setState({ hitB: true });
    setTimeout(() => setState({ hitB: false }), 350);
  }
}

function saveLastSession() {
  const s = get(state);
  const payload = {
    room: s.currentRoom,
    user: s.currentUser,
    creator: s.isCreator,
    flow: s.flowMode,
    auth: s.authMode,
    topic: s.currentTopic || s.createTopic,
    difficulty: s.currentDifficulty || s.createDifficulty
  };
  localStorage.setItem(STORAGE_LAST_KEY, JSON.stringify(payload));
}

function clearLastSession() {
  localStorage.removeItem(STORAGE_LAST_KEY);
}

function ensureRoomId() {
  const s = get(state);
  if (!s.createRoom.trim()) {
    setState({ createRoom: 'room-' + Math.random().toString(36).slice(2, 8) });
  }
}

let pingInterval = null;
let pingStartTime = 0;

function startPingMeasurement() {
  if (pingInterval) clearInterval(pingInterval);
  pingInterval = setInterval(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      pingStartTime = Date.now();
      ws.send(JSON.stringify({ type: 'ping', ts: pingStartTime }));
    }
  }, 3000);
}

function stopPingMeasurement() {
  if (pingInterval) {
    clearInterval(pingInterval);
    pingInterval = null;
  }
}

function handlePingResponse(ts) {
  const latency = Date.now() - ts;
  setState({ ping: latency });
}

function connectAndJoin() {
  const s = get(state);
  if (!s.currentRoom || !s.currentUser) return;

  if (ws) {
    try {
      ws.close();
    } catch {}
  }

  const token = s.jwtToken.trim();
  const url = token ? `${s.wsUrl}?token=${encodeURIComponent(token)}` : s.wsUrl;
  ws = new WebSocket(url);

  ws.onopen = () => {
    setState({ 
      status: 'Connected', 
      lobbyText: 'lobby.waiting',
      connectionStatus: 'connected',
      ping: -1
    });
    startPingMeasurement();
    const msg = {
      type: 'join',
      room_id: s.currentRoom,
      user_id: s.currentUser,
      lang: s.currentLang,
      topic: s.currentTopic,
      difficulty: s.currentDifficulty || s.createDifficulty,
      avatar: s.authMode === 'auth' ? (s.userAvatar || 'default') : 'guest'
    };
    ws.send(JSON.stringify(msg));
    resetStats();
    goto(`/lobby?room=${encodeURIComponent(s.currentRoom)}`);
  };

  ws.onclose = () => {
    setState({ status: 'Disconnected', connectionStatus: 'disconnected', ping: -1 });
    stopPingMeasurement();
    stopCountdown();
    
    // If game ended normally (game_over received), go to home instead of reconnect
    if (get(state).gameEndedNormally) {
      setState({ 
        gameEndedNormally: false,
        gameOverOpen: false,
        currentRoom: '',
        hp: {},
        promptText: 'Waiting for round...',
        roundInfo: '-'
      });
      goto('/');
      return;
    }
    
    const current = get(state).currentRoom;
    if (current) {
      goto(`/reconnect?room=${encodeURIComponent(current)}`);
    } else {
      goto('/');
    }
  };

  ws.onerror = () => {
    setState({ connectionStatus: 'disconnected', ping: -1 });
  };

  ws.onmessage = (ev) => {
    let data = null;
    try {
      data = JSON.parse(ev.data);
    } catch {
      return;
    }

    if (data.type === 'error') {
      const raw = data.error || 'Unknown error';
      const nice =
        raw.includes('room is full') ? 'Room is full' :
        raw.includes('room not found') ? 'Room not found' :
        raw.includes('user already in room') ? 'User already in room' :
        raw;
      setState({ startError: nice });
      return;
    }

    if (data.type === 'pong') {
      handlePingResponse(data.ts);
      return;
    }

    if (data.type === 'room_state') {
      ensurePlayers(data.players);
      applyHP(data.hp);
      if (data.prompt) setState({ promptText: data.prompt });
      if (data.duel_id) setState({ currentDuelId: data.duel_id });
      showRoundInfo(data);
    }

    if (data.type === 'player_joined') {
      ensurePlayers(data.players);
      applyHP(data.hp);
      const opponentName = data.players.find(p => p !== get(state).currentUser);
      if (opponentName && data.avatars) {
        const opponentAvatar = data.avatars[opponentName];
        if (opponentAvatar && opponentAvatar !== 'guest') {
          setState({ opponentAvatar: opponentAvatar });
        } else {
          setState({ opponentAvatar: 'default' });
        }
      } else if (opponentName) {
        const isGuest = opponentName.toLowerCase().includes('guest');
        if (isGuest) {
          setState({ opponentAvatar: 'default' });
        } else {
          const avatars = ['knight', 'wizard', 'archer', 'dragon', 'skull', 'fire', 'ice', 'lightning', 'sword', 'potion', 'crown', 'star', 'moon'];
          const hash = opponentName.split('').reduce((a, c) => ((a << 5) - a + c.charCodeAt(0)) | 0, 0);
          const avatarIndex = Math.abs(hash) % avatars.length;
          setState({ opponentAvatar: avatars[avatarIndex] });
        }
      }
      setState({ lobbyText: 'lobby.opponentJoined' });
    }

    if (data.type === 'player_left') {
      ensurePlayers(data.players);
      applyHP(data.hp);
      setState({ promptText: 'lobby.waiting', roundInfo: 'battle.playerLeft' });
      stopCountdown();
      goto(`/lobby?room=${encodeURIComponent(get(state).currentRoom)}`);
    }

    if (data.type === 'round_start') {
      setState({ promptText: data.prompt || 'Round started' });
      roundStartAt = Date.now();
      applyHP(data.hp);
      showRoundInfo(data);
      startCountdown();
      goto(`/battle?room=${encodeURIComponent(get(state).currentRoom)}`);
    }

    if (data.type === 'halftime') {
      setState({ promptText: data.prompt || '⏸ HALF TIME ⏸', roundInfo: 'Half 2 / 2' });
      applyHP(data.hp);
      stopCountdown();
      roundStartAt = null;
      // After 5 seconds, start next round
      setTimeout(() => {
        if (get(state).currentRoom && get(state).connectionStatus === 'connected') {
          ws.send(JSON.stringify({ type: 'next_round', room_id: get(state).currentRoom }));
        }
      }, 5000);
    }

    if (data.type === 'round_end') {
      setState({ promptText: 'Время вышло. Следующий раунд...', roundInfo: 'Причина: ' + (data.reason || 'timeout') });
      stopCountdown();
      roundStartAt = null;
    }

    if (data.type === 'update') {
      applyHP(data.hp);
      
      // Determine who received damage (opponent or self)
      let damageTarget = '';
      let damageAmount = 0;
      
      if (data.correct) {
        // Correct answer: damage to opponent
        damageTarget = data.defender_id || '';
        damageAmount = data.damage || 0;
      } else if (data.self_damage && data.self_damage > 0) {
        // Wrong answer: self-damage
        damageTarget = data.attacker_id || '';
        damageAmount = data.self_damage;
      }
      
      updateStats(!!data.correct, damageAmount, data.speed || 0, damageTarget);
      
      const s = get(state);
      const isMyAnswer = data.attacker_id === s.currentUser;
      
      if (isMyAnswer) {
        if (data.correct) {
          setState({ inputCorrect: true, roundInfo: `✓ Правильно! dmg: ${data.damage}` });
          setTimeout(() => setState({ inputCorrect: false }), 400);
        } else {
          setState({ inputWrong: true, roundInfo: `✗ Попробуй ещё! -${data.self_damage || 0} HP` });
          setTimeout(() => setState({ inputWrong: false }), 600);
        }
      } else {
        if (data.correct) {
          setState({ roundInfo: `Соперник ответил правильно! -${data.damage} HP` });
        } else {
          setState({ roundInfo: `Соперник ошибся! -${data.self_damage || 0} HP` });
        }
      }
      
      // Show damage on defender
      if (data.defender_id) hitEffect(data.defender_id);
      
      // Show self-damage on attacker if wrong answer
      if (data.self_damage && data.self_damage > 0 && data.attacker_id) {
        hitEffect(data.attacker_id);
      }
    }

    if (data.type === 'game_over') {
      setState({ 
        gameEndedNormally: true,
        promptText: 'Winner: ' + data.winner_id, 
        roundInfo: 'Game over' 
      });
      if (data.duel_id) {
        setState({ currentDuelId: data.duel_id });
      }
      applyHP(data.hp);
      if (data.elo) {
        setState({ elo: data.elo });
      }
      if (data.elo_change) {
        setState({ eloChange: data.elo_change });
      }
      
      if (data.correct_count && data.wrong_count) {
        const userID = get(state).currentUser;
        const userCorrect = data.correct_count[userID] || 0;
        const userWrong = data.wrong_count[userID] || 0;
        
        // Use the damage tracked during the game (more accurate)
        const s = get(state);
        setState({ 
          correctCount: userCorrect, 
          wrongCount: userWrong,
          playerADamage: s.playerADamage || 0,
          playerBDamage: s.playerBDamage || 0
        });
      }
      
      stopCountdown();
      const last = data.hp || {};
      const ids = Object.keys(last);
      const a = ids[0] ? (ids[0] + ': ' + last[ids[0]]) : '-';
      const b = ids[1] ? (ids[1] + ': ' + last[ids[1]]) : '-';
      
      const winner = data.winner_id;
      const isWinner = winner && winner === get(state).currentUser;
      const winMessages = [
        'VICTORY! 🏆',
        'YOU WIN! ⚔️',
        'LEGENDARY! 🌟',
        'UNSTOPPABLE! 🔥',
        'CHAMPION! 👑',
        'MASTERFUL! 🎯'
      ];
      const loseMessages = [
        'DEFEAT... 💔',
        'GAME OVER 💀',
        'SO CLOSE... 😢',
        'NEXT TIME! 🎮',
        'ALMOST! 💪'
      ];
      
      let gameText;
      if (isWinner) {
        const idx = Math.floor(Math.random() * winMessages.length);
        gameText = winMessages[idx];
      } else {
        const idx = Math.floor(Math.random() * loseMessages.length);
        gameText = loseMessages[idx];
      }
      
      let eloInfo = '';
      if (data.elo_change) {
        const userID = get(state).authedUserID;
        const change = data.elo_change[userID];
        if (change !== undefined) {
          eloInfo = change > 0 ? ` +${change} ELO` : ` ${change} ELO`;
        }
      }
      
      let reasonText = '';
      if (data.reason === 'phrases_exhausted') {
        reasonText = 'Phrases exhausted!';
      } else if (data.reason === 'hp_zero') {
        reasonText = 'HP depleted!';
      }
      
      setState({
        gameOverOpen: true,
        gameOverText: gameText + eloInfo,
        gameOverHP: 'Final HP - ' + a + ' | ' + b,
        gameOverReason: reasonText,
        isGameWinner: isWinner
      });
    }
  };
}

function reconnect() {
  // Reset game state before reconnecting
  setState({
    hp: {},
    elo: {},
    promptText: 'Waiting for round...',
    timerText: '-',
    roundInfo: '-',
    correctCount: 0,
    wrongCount: 0,
    totalDamage: 0,
    totalSpeed: 0,
    speedCount: 0,
    gameOverOpen: false,
    gameEndedNormally: false,
    hitA: false,
    hitB: false,
    attackA: false,
    attackB: false
  });
  connectAndJoin();
}

async function syncAuthFromToken() {
  const s = get(state);
  const token = s.jwtToken.trim();
  if (!token) {
    setState({ authedUsername: '', authedUserID: '' });
    return;
  }
  try {
    const base = httpBaseFromWs(s.wsUrl);
    const res = await fetch(`${base}/me`, {
      headers: { Authorization: 'Bearer ' + token }
    });
    if (!res.ok) {
      setState({ authedUsername: '', authedUserID: '' });
      return;
    }
    const data = await res.json();
    const avatar = data.avatar || 'default';
    setState({ 
      authedUsername: data.username || '', 
      authedUserID: data.user_id || '', 
      userAvatar: avatar, 
      profileUser: data.username,
      coins: data.coins || 0,
      unlockedAvatars: data.unlocked_avatars || ['default'],
    });
    localStorage.setItem(STORAGE_TOKEN_KEY, token);
    localStorage.setItem(STORAGE_USER_KEY, data.username);
    localStorage.setItem(STORAGE_AVATAR_KEY, avatar);
    try {
      const statsRes = await fetch(`${base}/me/stats`, {
          headers: { Authorization: 'Bearer ' + token }
        });
        if (statsRes.ok) {
          const s = await statsRes.json();
          setState({
            profileDuels: String(s.total_duels_played ?? '-'),
            profileWins: String(s.total_duels_won ?? '-'),
            profileAcc: s.overall_accuracy != null ? String(s.overall_accuracy) : '-',
            profileStreak: String(s.best_win_streak ?? '-')
          });
        }
        const duelsRes = await fetch(`${base}/me/duels`, {
          headers: { Authorization: 'Bearer ' + token }
        });
        if (duelsRes.ok) {
          const list = await duelsRes.json();
          const mapped = list.map((d) => {
            const status = (d.status || 'unknown').toUpperCase();
            let badgeClass = 'pending';
            let resultLabel = 'PENDING';
            if (status === 'FINISHED' && d.winner_user_id) {
              if (data.user_id && d.winner_user_id === data.user_id) {
                badgeClass = 'win';
                resultLabel = 'WIN';
              } else {
                badgeClass = 'loss';
                resultLabel = 'LOSS';
              }
            }
            const created = d.created_at ? new Date(d.created_at).toLocaleString() : '-';
            const opponentName = d.opponent_username || null;
            return {
              duelId: d.duel_id || '',
              room: d.room_code || '-',
              status,
              created,
              badgeClass,
              resultLabel,
              opponentName
            };
          });
          setState({ profileDuelsCount: String(list.length || 0), profileDuelsList: mapped });
        }
        await fetchUserRating();
        await fetchAchievements();
      } catch (e) {
        // ignore profile fetch errors
      }
  } catch (e) {
    setState({ authedUsername: '', authedUserID: '' });
  }
}

async function login(username, password) {
  try {
    setState({ authError: '' });
    if (!username || !password) {
      setState({ authError: 'Login and password required' });
      return;
    }
    const base = httpBaseFromWs(get(state).wsUrl);
    const res = await fetch(`${base}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ login: username, password })
    });
    if (!res.ok) {
      const text = await res.text();
      setState({ authError: 'Login failed: ' + text });
      return;
    }
    const data = await res.json();
    setState({ jwtToken: data.token || '', authedUsername: data.username || '' });
    await syncAuthFromToken();
  } catch (e) {
    console.error('Login exception:', e);
    setState({ authError: 'Auth error: ' + (e && e.message ? e.message : e) });
  }
}

async function register(username, email, password, confirm) {
  try {
    setState({ authError: '' });
    if (!username || !password) {
      setState({ authError: 'Username and password required' });
      return;
    }
    if (confirm !== password) {
      setState({ authError: 'Passwords do not match' });
      return;
    }
    const base = httpBaseFromWs(get(state).wsUrl);
    const res = await fetch(`${base}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, email, password })
    });
    if (!res.ok) {
      const text = await res.text();
      setState({ authError: 'Register failed: ' + text });
      return;
    }
    const data = await res.json();
    setState({ jwtToken: data.token || '', authedUsername: data.username || '' });
    await syncAuthFromToken();
  } catch (e) {
    setState({ authError: 'Register error: ' + (e && e.message ? e.message : e) });
  }
}

function createAndConnect() {
  const s = get(state);
  ensureRoomId();
  const room = get(state).createRoom.trim();
  const user = s.authMode === 'auth'
    ? s.authedUsername
    : (s.createUser.trim() || 'Guest-' + Math.random().toString(36).slice(2, 6));
  
  if (!user) {
    setState({ startError: 'Username is required' });
    return;
  }
  
  if (!/^[a-zA-Z][a-zA-Z0-9_\-]*$/.test(user)) {
    setState({ startError: 'Username: 2-30 chars, starts with letter, only letters/numbers/_/-' });
    return;
  }
  
  if (user.length > 30) {
    setState({ startError: 'Username max 30 characters' });
    return;
  }
  
  if (!/^[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$/.test(room)) {
    setState({ startError: 'Room ID: 3-50 chars, letters/numbers/hyphens only' });
    return;
  }
  
  if (room.length > 50) {
    setState({ startError: 'Room ID max 50 characters' });
    return;
  }
  
  setState({
    startError: '',
    isCreator: true,
    currentRoom: room,
    currentUser: user,
    currentLang: s.createLang,
    currentTopic: s.createTopic,
    currentDifficulty: s.createDifficulty,
    createUser: user,
    joinUser: user
  });
  localStorage.setItem(STORAGE_GUEST_KEY, user);
  saveLastSession();
  connectAndJoin();
}

function joinAndConnect() {
  const s = get(state);
  const room = s.joinRoom.trim();
  if (!room) {
    setState({ startError: 'Room ID is required' });
    return;
  }
  
  if (!/^[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$/.test(room)) {
    setState({ startError: 'Room ID: 3-50 chars, letters/numbers/hyphens only' });
    return;
  }
  
  if (room.length > 50) {
    setState({ startError: 'Room ID max 50 characters' });
    return;
  }
  
  const user = s.authMode === 'auth'
    ? s.authedUsername
    : (s.joinUser.trim() || 'Guest-' + Math.random().toString(36).slice(2, 6));
  if (!user) {
    setState({ startError: 'Username is required' });
    return;
  }
  
  if (!/^[a-zA-Z][a-zA-Z0-9_\-]*$/.test(user)) {
    setState({ startError: 'Username: 2-30 chars, starts with letter, only letters/numbers/_/-' });
    return;
  }
  
  setState({
    startError: '',
    isCreator: false,
    currentRoom: room,
    currentUser: user,
    currentLang: 'en',
    currentTopic: 'default',
    createUser: user,
    joinUser: user
  });
  localStorage.setItem(STORAGE_GUEST_KEY, user);
  saveLastSession();
  connectAndJoin();
}

function gateJoin(nick) {
  const user = nick.trim() || 'Guest-' + Math.random().toString(36).slice(2, 6);
  setState({ currentUser: user, joinUser: user, createUser: user });
  localStorage.setItem(STORAGE_GUEST_KEY, user);
  saveLastSession();
  connectAndJoin();
}

function sendAnswer(answer) {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;
  
  // Block answers during halftime
  const s = get(state);
  if (s.promptText && s.promptText.includes('HALF TIME')) return;
  
  const trimmed = answer.trim();
  if (!trimmed || trimmed.length > 200) return;
  
  const speed = roundStartAt ? Date.now() - roundStartAt : 0;
  const msg = {
    type: 'answer',
    room_id: s.currentRoom,
    user_id: s.currentUser,
    answer: trimmed,
    speed
  };
  try {
    ws.send(JSON.stringify(msg));
  } catch (e) {
    console.error('Failed to send answer:', e);
  }
}

function leaveMatch() {
  // Send leave message to server
  const s = get(state);
  if (ws && ws.readyState === WebSocket.OPEN && s.currentRoom) {
    ws.send(JSON.stringify({
      type: 'leave',
      room_id: s.currentRoom,
      user_id: s.currentUser
    }));
  }
  
  try {
    if (ws) ws.close();
  } catch {}
  clearLastSession();
  setState({ opponentAvatar: 'default' });
  goto('/');
}

function logout() {
  try {
    if (ws) ws.close();
  } catch {}
  localStorage.removeItem(STORAGE_TOKEN_KEY);
  localStorage.removeItem(STORAGE_USER_KEY);
  localStorage.removeItem(STORAGE_AVATAR_KEY);
  localStorage.removeItem(STORAGE_GUEST_KEY);
  localStorage.removeItem(STORAGE_AUTH_CHOICE_KEY);
  setState({
    jwtToken: '',
    authedUsername: '',
    authedUserID: '',
    authMode: 'guest',
    authChoiceMade: false,
    profileUser: '-',
    profileDuels: '-',
    profileWins: '-',
    profileAcc: '-',
    profileStreak: '-',
    profileDuelsCount: '0',
    profileDuelsList: [],
    userAvatar: 'default',
    opponentAvatar: 'default',
    createUser: '',
    joinUser: ''
  });
}

function init() {
  if (initialized) return;
  initialized = true;
  if (typeof window !== 'undefined') {
    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
    setState({ wsUrl: `${proto}://${window.location.host}/ws` });
    const storedToken = localStorage.getItem(STORAGE_TOKEN_KEY) || '';
    const storedUser = localStorage.getItem(STORAGE_USER_KEY) || '';
    const storedGuest = localStorage.getItem(STORAGE_GUEST_KEY) || '';
    const storedChoice = localStorage.getItem(STORAGE_AUTH_CHOICE_KEY) || '';
    const storedAvatar = localStorage.getItem(STORAGE_AVATAR_KEY) || 'default';
    if (storedToken) setState({ jwtToken: storedToken });
    if (storedUser) setState({ authedUsername: storedUser, profileUser: storedUser });
    if (storedAvatar) setState({ userAvatar: storedAvatar });
    if (storedGuest) setState({ createUser: storedGuest, joinUser: storedGuest });
    if (storedChoice) {
      setState({ authMode: storedChoice, authChoiceMade: true });
    } else if (storedToken) {
      setState({ authMode: 'auth', authChoiceMade: true });
      localStorage.setItem(STORAGE_AUTH_CHOICE_KEY, 'auth');
    }
    const last = localStorage.getItem(STORAGE_LAST_KEY);
    if (last) {
      try {
        const s = JSON.parse(last);
        setState({
          currentRoom: s.room || '',
          currentUser: s.user || '',
          isCreator: !!s.creator,
          flowMode: s.flow || 'create',
          authMode: s.auth || 'guest',
          createTopic: s.topic || 'default',
          createDifficulty: s.difficulty || 'intermediate'
        });
      } catch {
        // ignore
      }
    }
    syncAuthFromToken();
  }
}

async function updateUsername(newUsername) {
  const s = get(state);
  if (!s.jwtToken) return { error: 'Not authenticated' };
  try {
    const base = httpBaseFromWs(s.wsUrl);
    const res = await fetch(`${base}/me/username`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + s.jwtToken
      },
      body: JSON.stringify({ username: newUsername })
    });
    if (!res.ok) {
      const text = await res.text();
      return { error: text };
    }
    const data = await res.json();
    setState({ authedUsername: data.username, profileUser: data.username });
    localStorage.setItem(STORAGE_USER_KEY, data.username);
    return { success: true };
  } catch (e) {
    return { error: e.message };
  }
}

async function updateAvatar(newAvatar) {
  const s = get(state);
  if (!s.jwtToken) return { error: 'Not authenticated' };
  try {
    const base = httpBaseFromWs(s.wsUrl);
    const res = await fetch(`${base}/me/avatar`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + s.jwtToken
      },
      body: JSON.stringify({ avatar: newAvatar })
    });
    if (!res.ok) {
      const text = await res.text();
      return { error: text };
    }
    const data = await res.json();
    setState({ userAvatar: data.avatar });
    localStorage.setItem(STORAGE_AVATAR_KEY, data.avatar);
    return { success: true };
  } catch (e) {
    return { error: e.message };
  }
}

async function fetchUserRating() {
  const s = get(state);
  if (!s.jwtToken) return;
  try {
    const base = httpBaseFromWs(s.wsUrl);
    const res = await fetch(`${base}/me/rating`, {
      headers: { Authorization: 'Bearer ' + s.jwtToken }
    });
    if (res.ok) {
      const data = await res.json();
      const rankNames = {
        'newbie': '🥉 Newbie',
        'apprentice': '🥈 Apprentice',
        'expert': '🥇 Expert',
        'master': '💎 Master',
        'struggler': '😔 Struggler'
      };
      setState({
        profileElo: data.elo || 1000,
        profileRank: data.rank || 'newbie',
        profileRankName: rankNames[data.rank] || rankNames['newbie'],
        profileCoins: data.coins || 0,
        profileXP: data.xp || 0,
        profileLevel: data.level || 1
      });
    }
  } catch (e) {
    // ignore
  }
}

async function fetchLeaderboard() {
  const s = get(state);
  try {
    const base = httpBaseFromWs(s.wsUrl);
    const res = await fetch(`${base}/api/leaderboard`);
    if (res.ok) {
      const data = await res.json();
      setState({ leaderboard: data || [] });
    }
  } catch (e) {
    setState({ leaderboard: [] });
  }
}

async function fetchAchievements() {
  const s = get(state);
  if (!s.jwtToken) return;
  try {
    const base = httpBaseFromWs(s.wsUrl);
    const res = await fetch(`${base}/me/achievements`, {
      headers: { Authorization: 'Bearer ' + s.jwtToken }
    });
    if (res.ok) {
      const data = await res.json();
      setState({ achievements: data || [] });
    }
  } catch (e) {
    // ignore
  }
}

async function fetchDuelDetails(duelId) {
  const s = get(state);
  if (!s.jwtToken) return null;
  try {
    const base = httpBaseFromWs(s.wsUrl);
    const res = await fetch(`${base}/duels?id=${encodeURIComponent(duelId)}`, {
      headers: { Authorization: 'Bearer ' + s.jwtToken }
    });
    if (res.ok) {
      return await res.json();
    }
  } catch (e) {
    console.error('Failed to fetch duel details:', e);
  }
  return null;
}

async function fetchDuelAnalysis(duelId) {
  const s = get(state);
  if (!duelId) return null;
  try {
    const base = httpBaseFromWs(s.wsUrl);
    const headers = {};
    if (s.jwtToken) {
      headers['Authorization'] = 'Bearer ' + s.jwtToken;
    }
    const res = await fetch(`${base}/analysis?id=${encodeURIComponent(duelId)}`, { headers });
    if (res.ok) {
      return await res.json();
    }
  } catch (e) {
    console.error('Failed to fetch duel analysis:', e);
  }
  return null;
}

async function claimCoins() {
  const s = get(state);
  if (!s.jwtToken) return { coins_awarded: 0 };
  try {
    const base = httpBaseFromWs(s.wsUrl);
    const res = await fetch(`${base}/me/claim-coins`, {
      method: 'POST',
      headers: { Authorization: 'Bearer ' + s.jwtToken }
    });
    if (res.ok) {
      const data = await res.json();
      const newCoins = (s.profileCoins || 0) + (data.coins_awarded || 0);
      setState({ profileCoins: newCoins, coins: newCoins });
      return data;
    }
  } catch (e) {
    console.error('claimCoins error:', e);
  }
  return { coins_awarded: 0 };
}

async function buyAvatar(avatarId) {
  const s = get(state);
  if (!s.jwtToken) return { error: 'not authenticated' };
  
  const price = getAvatarPrice(avatarId);
  const currentCoins = s.profileCoins || s.coins || 0;
  if (currentCoins < price) {
    return { error: 'not enough coins' };
  }
  
  try {
    const base = httpBaseFromWs(s.wsUrl);
    const res = await fetch(`${base}/me/buy-avatar`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + s.jwtToken
      },
      body: JSON.stringify({ avatar_id: avatarId })
    });
    
    if (res.ok) {
      const data = await res.json();
      setState({ 
        coins: data.coins,
        profileCoins: data.coins,
        unlockedAvatars: data.unlocked_avatars
      });
      return data;
    } else {
      const err = await res.text();
      return { error: err };
    }
  } catch (e) {
    return { error: e.message };
  }
}

async function generatePhrases(roomId, topic, difficulty, lang) {
  const s = get(state);
  try {
    setState({ generatingPhrases: true, generationError: '' });
    const base = httpBaseFromWs(s.wsUrl);
    
    // Parse language direction (en-ru or ru-en)
    let langFrom = 'en';
    let langTo = 'ru';
    if (lang === 'ru-en') {
      langFrom = 'ru';
      langTo = 'en';
    }
    
    const res = await fetch(`${base}/api/generate-phrases`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        room_id: roomId,
        topic: topic || s.createTopic,
        difficulty: difficulty || s.createDifficulty,
        lang_from: langFrom,
        lang_to: langTo
      })
    });
    
    if (!res.ok) {
      const text = await res.text();
      setState({ generationError: 'Generation failed: ' + text, generatingPhrases: false });
      return false;
    }
    
    const data = await res.json();
    if (data.success) {
      setState({ phrasesGenerated: true, generatingPhrases: false });
      return true;
    } else {
      setState({ generationError: 'Generation failed', generatingPhrases: false });
      return false;
    }
  } catch (e) {
    setState({ generationError: 'Error: ' + e.message, generatingPhrases: false });
    return false;
  }
}

export const duel = {
  subscribe: state.subscribe,
  init,
  setField: (key, value) => setState({ [key]: value }),
  setAuthMode: (mode) => {
    setState({ authMode: mode, authChoiceMade: true });
    localStorage.setItem(STORAGE_AUTH_CHOICE_KEY, mode);
  },
  selectGuest: () => {
    localStorage.removeItem(STORAGE_TOKEN_KEY);
    localStorage.removeItem(STORAGE_USER_KEY);
    localStorage.removeItem(STORAGE_AVATAR_KEY);
    localStorage.removeItem(STORAGE_GUEST_KEY);
    setState({
      jwtToken: '',
      authedUsername: '',
      authedUserID: '',
      authMode: 'guest',
      authChoiceMade: true,
      profileUser: '-',
      userAvatar: 'default',
      opponentAvatar: 'default',
      profileElo: 1000,
      profileRank: 'newbie',
      profileRankName: '🥉 Newbie',
      createUser: '',
      joinUser: ''
    });
    localStorage.setItem(STORAGE_AUTH_CHOICE_KEY, 'guest');
  },
  setAuthTab: (tab) => setState({ authTab: tab }),
  setFlowMode: (mode) => setState({ flowMode: mode }),
  ensureRoomId,
  buildRoomLink,
  copyLink,
  showNote,
  createAndConnect,
  joinAndConnect,
  reconnect,
  gateJoin,
  sendAnswer,
  leaveMatch,
  logout,
  login,
  register,
  updateUsername,
  updateAvatar,
  fetchUserRating,
  fetchLeaderboard,
  fetchAchievements,
  fetchDuelDetails,
  fetchDuelAnalysis,
  claimCoins,
  generatePhrases,
  getAvatarEmoji,
  getAvatarPrice,
  getAvatarProjectile,
  buyAvatar,
  avgSpeed: () => avgSpeed(get(state)),
  connectionStatus: () => get(state).connectionStatus,
  ping: () => get(state).ping
};

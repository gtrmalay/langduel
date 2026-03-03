import { writable, get } from 'svelte/store';
import { goto } from '$app/navigation';

const initialState = {
  status: 'Disconnected',
  flowMode: 'create',
  authMode: 'guest',
  authChoiceMade: false,
  authTab: 'login',
  wsUrl: '',
  createUser: '',
  joinUser: '',
  createRoom: '',
  joinRoom: '',
  createLang: 'en',
  createTopic: 'default',
  createRounds: 5,
  jwtToken: '',
  authedUsername: '',
  authedUserID: '',
  startError: '',
  authError: '',
  isCreator: false,
  currentRoom: '',
  currentUser: '',
  currentLang: 'en',
  currentTopic: 'default',
  lobbyText: 'Waiting for opponent...',
  lobbyCopyNote: '',
  createCopyNote: '',
  reconnectNote: '',
  playerA: 'Player A',
  playerB: 'Player B',
  hp: {},
  promptText: 'Waiting for round...',
  timerText: '-',
  roundInfo: '-',
  correctCount: 0,
  wrongCount: 0,
  totalDamage: 0,
  totalSpeed: 0,
  speedCount: 0,
  gameOverOpen: false,
  gameOverText: 'Game over',
  gameOverHP: '-',
  profileUser: '-',
  profileDuels: '-',
  profileWins: '-',
  profileAcc: '-',
  profileStreak: '-',
  profileDuelsCount: '0',
  profileDuelsList: [],
  hitA: false,
  hitB: false
};

const STORAGE_TOKEN_KEY = 'langduel_token';
const STORAGE_USER_KEY = 'langduel_user';
const STORAGE_GUEST_KEY = 'langduel_guest';
const STORAGE_LAST_KEY = 'langduel_last_session';
const STORAGE_AUTH_CHOICE_KEY = 'langduel_auth_choice';
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
    totalSpeed: 0,
    speedCount: 0
  });
}

function updateStats(correct, damage, speed) {
  state.update((s) => {
    const next = { ...s };
    if (correct) next.correctCount += 1;
    else next.wrongCount += 1;
    next.totalDamage += damage;
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
    setState({ roundInfo: `Round ${data.round}` });
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
    auth: s.authMode
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
    setState({ status: 'Connected', lobbyText: 'Waiting for opponent...' });
    const msg = {
      type: 'join',
      room_id: s.currentRoom,
      user_id: s.currentUser,
      lang: s.currentLang,
      topic: s.currentTopic
    };
    ws.send(JSON.stringify(msg));
    resetStats();
    goto(`/lobby?room=${encodeURIComponent(s.currentRoom)}`);
  };

  ws.onclose = () => {
    setState({ status: 'Disconnected' });
    stopCountdown();
    const current = get(state).currentRoom;
    if (current) {
      goto(`/reconnect?room=${encodeURIComponent(current)}`);
    } else {
      goto('/');
    }
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

    if (data.type === 'room_state') {
      ensurePlayers(data.players);
      applyHP(data.hp);
      if (data.prompt) setState({ promptText: data.prompt });
      showRoundInfo(data);
    }

    if (data.type === 'player_joined') {
      ensurePlayers(data.players);
      applyHP(data.hp);
      setState({ lobbyText: 'Opponent joined. Starting...' });
    }

    if (data.type === 'player_left') {
      ensurePlayers(data.players);
      applyHP(data.hp);
      setState({ promptText: 'Waiting for opponent...', roundInfo: 'Player left' });
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

    if (data.type === 'round_end') {
      setState({ promptText: 'Time out. Next round...', roundInfo: 'Reason: ' + (data.reason || 'timeout') });
      stopCountdown();
    }

    if (data.type === 'update') {
      applyHP(data.hp);
      updateStats(!!data.correct, data.damage || 0, data.speed || 0);
      const correct = data.correct ? 'correct' : 'wrong';
      setState({
        roundInfo:
          'Attack: ' + (data.attacker_id || '-') +
          ' -> ' + (data.defender_id || '-') +
          ' | damage: ' + (data.damage || 0) +
          ' | ' + correct
      });
      if (data.defender_id) hitEffect(data.defender_id);
    }

    if (data.type === 'game_over') {
      setState({ promptText: 'Winner: ' + data.winner_id, roundInfo: 'Game over' });
      applyHP(data.hp);
      stopCountdown();
      const last = data.hp || {};
      const ids = Object.keys(last);
      const a = ids[0] ? (ids[0] + ': ' + last[ids[0]]) : '-';
      const b = ids[1] ? (ids[1] + ': ' + last[ids[1]]) : '-';
      setState({
        gameOverOpen: true,
        gameOverText: 'Winner: ' + (data.winner_id || '-'),
        gameOverHP: 'Final HP - ' + a + ' | ' + b
      });
    }
  };
}

function reconnect() {
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
    setState({ authedUsername: data.username || '', authedUserID: data.user_id || '' });
    if (data.username) {
      localStorage.setItem(STORAGE_TOKEN_KEY, token);
      localStorage.setItem(STORAGE_USER_KEY, data.username);
      setState({ profileUser: data.username });
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
            const opponent = d.opponent_username ? `vs ${d.opponent_username}` : 'Waiting for opponent';
            return {
              room: d.room_code || '-',
              status,
              created,
              badgeClass,
              resultLabel,
              opponent
            };
          });
          setState({ profileDuelsCount: String(list.length || 0), profileDuelsList: mapped });
        }
      } catch {
        // ignore profile fetch errors
      }
    }
  } catch {
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
    setState({ startError: 'Auth mode enabled, but no user is logged in' });
    return;
  }
  setState({
    startError: '',
    isCreator: true,
    currentRoom: room,
    currentUser: user,
    currentLang: s.createLang,
    currentTopic: s.createTopic,
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
  const user = s.authMode === 'auth'
    ? s.authedUsername
    : (s.joinUser.trim() || 'Guest-' + Math.random().toString(36).slice(2, 6));
  if (!user) {
    setState({ startError: 'Auth mode enabled, but no user is logged in' });
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
  if (!ws) return;
  const speed = roundStartAt ? Date.now() - roundStartAt : 0;
  const s = get(state);
  const msg = {
    type: 'answer',
    room_id: s.currentRoom,
    user_id: s.currentUser,
    answer: answer.trim(),
    speed
  };
  ws.send(JSON.stringify(msg));
}

function leaveMatch() {
  try {
    if (ws) ws.close();
  } catch {}
  clearLastSession();
  goto('/');
}

function logout() {
  try {
    if (ws) ws.close();
  } catch {}
  localStorage.removeItem(STORAGE_TOKEN_KEY);
  localStorage.removeItem(STORAGE_USER_KEY);
  setState({
    jwtToken: '',
    authedUsername: '',
    authedUserID: '',
    profileUser: '-',
    profileDuels: '-',
    profileWins: '-',
    profileAcc: '-',
    profileStreak: '-',
    profileDuelsCount: '0',
    profileDuelsList: []
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
    if (storedToken) setState({ jwtToken: storedToken });
    if (storedUser) setState({ authedUsername: storedUser, profileUser: storedUser });
    if (storedGuest) setState({ createUser: storedGuest, joinUser: storedGuest });
    if (storedChoice) {
      setState({ authMode: storedChoice, authChoiceMade: true });
    } else if (storedToken) {
      setState({ authMode: 'auth', authChoiceMade: true });
      localStorage.setItem(STORAGE_AUTH_CHOICE_KEY, 'auth');
    } else {
      setState({ authMode: 'guest', authChoiceMade: true });
      localStorage.setItem(STORAGE_AUTH_CHOICE_KEY, 'guest');
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
          authMode: s.auth || 'guest'
        });
      } catch {
        // ignore
      }
    }
    syncAuthFromToken();
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
    setState({
      jwtToken: '',
      authedUsername: '',
      authedUserID: '',
      authMode: 'guest',
      authChoiceMade: true,
      profileUser: '-'
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
  avgSpeed: () => avgSpeed(get(state))
};

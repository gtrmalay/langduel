<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import ProfilePanel from '$lib/components/ProfilePanel.svelte';
  import StartPanel from '$lib/components/StartPanel.svelte';
  import LobbyPanel from '$lib/components/LobbyPanel.svelte';
  import JoinGatePanel from '$lib/components/JoinGatePanel.svelte';
  import ReconnectPanel from '$lib/components/ReconnectPanel.svelte';
  import BattleView from '$lib/components/BattleView.svelte';

  export let initialScreen = 'start';
  export let initialFlow = 'create';
  export let initialAuthMode = 'guest';
  export let openProfileOnLoad = false;

  const roundDurationMs = 10000;
  const STORAGE_TOKEN_KEY = 'langduel_token';
  const STORAGE_USER_KEY = 'langduel_user';
  const STORAGE_GUEST_KEY = 'langduel_guest';
  const STORAGE_LAST_KEY = 'langduel_last_session';

  let ws = null;
  let roundStartAt = null;
  let countdownTimer = null;

  let status = 'Disconnected';

  let screen = 'start';
  let flowMode = 'create';
  let authMode = 'guest';
  let authTab = 'login';

  let wsUrl = '';

  let createUser = '';
  let joinUser = '';
  let createRoom = '';
  let joinRoom = '';
  let createLang = 'en';
  let createTopic = 'default';

  let jwtToken = '';
  let authedUsername = '';
  let authedUserID = '';

  let startError = '';
  let authError = '';

  let isCreator = false;
  let currentRoom = '';
  let currentUser = '';
  let currentLang = 'en';
  let currentTopic = 'default';

  let lobbyText = 'Waiting for opponent...';
  let lobbyCopyNote = '';
  let createCopyNote = '';
  let reconnectNote = '';

  let profileOpen = false;

  let playerA = 'Player A';
  let playerB = 'Player B';
  let hp = {};
  let lastHP = {};

  let promptText = 'Waiting for round...';
  let timerText = '-';
  let roundInfo = '-';

  let correctCount = 0;
  let wrongCount = 0;
  let totalDamage = 0;
  let totalSpeed = 0;
  let speedCount = 0;

  let gameOverOpen = false;
  let gameOverText = 'Game over';
  let gameOverHP = '-';

  let profileUser = '-';
  let profileDuels = '-';
  let profileWins = '-';
  let profileAcc = '-';
  let profileStreak = '-';
  let profileDuelsCount = '0';
  let profileDuelsList = [];

  let hitA = false;
  let hitB = false;

  let createCollapsed = false;

  let authLogin = '';
  let authPass = '';
  let authEmail = '';
  let regConfirm = '';
  let gateNick = '';
  let answer = '';

  function showScreen(next) {
    screen = next;
  }

  function setAuthMode(mode) {
    authMode = mode;
    if (mode === 'auth') {
      createUser = authedUsername || createUser;
      joinUser = authedUsername || joinUser;
    }
  }

  function setFlowMode(mode) {
    flowMode = mode;
  }

  function setAuthTab(tab) {
    authTab = tab;
  }

  function ensureRoomId() {
    if (!createRoom.trim()) {
      createRoom = 'room-' + Math.random().toString(36).slice(2, 8);
    }
  }

  function buildRoomLink(roomId) {
    if (typeof window === 'undefined') return '';
    const base = window.location.origin + window.location.pathname;
    return `${base}?room=${encodeURIComponent(roomId)}`;
  }

  function saveLastSession() {
    const payload = {
      room: currentRoom,
      user: currentUser,
      creator: isCreator,
      flow: flowMode,
      auth: authMode
    };
    localStorage.setItem(STORAGE_LAST_KEY, JSON.stringify(payload));
  }

  function clearLastSession() {
    localStorage.removeItem(STORAGE_LAST_KEY);
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

  function showNote(setter, text) {
    setter(text);
    setTimeout(() => setter(''), 1500);
  }

  function resetStats() {
    correctCount = 0;
    wrongCount = 0;
    totalDamage = 0;
    totalSpeed = 0;
    speedCount = 0;
  }

  function updateStats(correct, damage, speed) {
    if (correct) {
      correctCount += 1;
    } else {
      wrongCount += 1;
    }
    totalDamage += damage;
    if (speed > 0) {
      totalSpeed += speed;
      speedCount += 1;
    }
  }

  function avgSpeed() {
    if (!speedCount) return '-';
    return Math.round(totalSpeed / speedCount);
  }

  function applyHP(nextHP) {
    hp = nextHP || {};
  }

  function ensurePlayers(list) {
    if (!list || list.length === 0) return;
    playerA = list[0] || 'Player A';
    playerB = list[1] || 'Player B';
  }

  function showRoundInfo(data) {
    if (data && data.round) {
      roundInfo = `Round ${data.round}`;
    }
  }

  function startCountdown() {
    stopCountdown();
    const end = Date.now() + roundDurationMs;
    countdownTimer = setInterval(() => {
      const left = Math.max(0, end - Date.now());
      timerText = `${Math.ceil(left / 1000)}s`;
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
    if (playerId === playerA) {
      hitA = true;
      setTimeout(() => (hitA = false), 350);
    }
    if (playerId === playerB) {
      hitB = true;
      setTimeout(() => (hitB = false), 350);
    }
  }

  function connectAndJoin() {
    if (!currentRoom || !currentUser) return;

    if (ws) {
      try {
        ws.close();
      } catch {}
    }

    const token = jwtToken.trim();
    const url = token ? `${wsUrl}?token=${encodeURIComponent(token)}` : wsUrl;
    ws = new WebSocket(url);

    ws.onopen = () => {
      status = 'Connected';
      const msg = {
        type: 'join',
        room_id: currentRoom,
        user_id: currentUser,
        lang: currentLang,
        topic: currentTopic
      };
      ws.send(JSON.stringify(msg));
      lobbyText = 'Waiting for opponent...';
      showScreen('lobby');
      if (window.location.pathname !== '/lobby') {
        goto(`/lobby?room=${encodeURIComponent(currentRoom)}`);
      }
      gameOverOpen = false;
      resetStats();
    };

    ws.onclose = () => {
      status = 'Disconnected';
      stopCountdown();
      if (currentRoom) {
        showScreen('reconnect');
      } else {
        showScreen('start');
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
        startError = nice;
        return;
      }

      if (data.type === 'room_state') {
        ensurePlayers(data.players);
        applyHP(data.hp);
        if (data.prompt) promptText = data.prompt;
        showRoundInfo(data);
      }

      if (data.type === 'player_joined') {
        ensurePlayers(data.players);
        applyHP(data.hp);
        lobbyText = 'Opponent joined. Starting...';
      }

      if (data.type === 'player_left') {
        ensurePlayers(data.players);
        applyHP(data.hp);
        promptText = 'Waiting for opponent...';
        roundInfo = 'Player left';
        stopCountdown();
        showScreen('lobby');
      }

      if (data.type === 'round_start') {
        showScreen('battle');
        promptText = data.prompt || 'Round started';
        roundStartAt = Date.now();
        applyHP(data.hp);
        showRoundInfo(data);
        startCountdown();
        if (window.location.pathname !== '/battle') {
          goto(`/battle?room=${encodeURIComponent(currentRoom)}`);
        }
      }

      if (data.type === 'round_end') {
        promptText = 'Time out. Next round...';
        roundInfo = 'Reason: ' + (data.reason || 'timeout');
        stopCountdown();
      }

      if (data.type === 'update') {
        applyHP(data.hp);
        lastHP = data.hp || lastHP;
        updateStats(!!data.correct, data.damage || 0, data.speed || 0);
        const correct = data.correct ? 'correct' : 'wrong';
        roundInfo =
          'Attack: ' + (data.attacker_id || '-') +
          ' -> ' + (data.defender_id || '-') +
          ' | damage: ' + (data.damage || 0) +
          ' | ' + correct;
        if (data.defender_id) hitEffect(data.defender_id);
      }

      if (data.type === 'game_over') {
        promptText = 'Winner: ' + data.winner_id;
        applyHP(data.hp);
        lastHP = data.hp || lastHP;
        roundInfo = 'Game over';
        stopCountdown();
        gameOverText = 'Winner: ' + (data.winner_id || '-');
        if (lastHP) {
          const ids = Object.keys(lastHP);
          const a = ids[0] ? (ids[0] + ': ' + lastHP[ids[0]]) : '-';
          const b = ids[1] ? (ids[1] + ': ' + lastHP[ids[1]]) : '-';
          gameOverHP = 'Final HP - ' + a + ' | ' + b;
        } else {
          gameOverHP = '-';
        }
        gameOverOpen = true;
      }
    };
  }

  async function login() {
    try {
      authError = '';
      const username = authLogin.trim();
      if (!username || !authPass) {
        authError = 'Login and password required';
        return;
      }
      const base = httpBaseFromWs(wsUrl);
      const res = await fetch(`${base}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ login: username, password: authPass })
      });
      if (!res.ok) {
        const text = await res.text();
        authError = 'Login failed: ' + text;
        return;
      }
      const data = await res.json();
      jwtToken = data.token || '';
      if (!jwtToken) {
        authError = 'Token not received';
        return;
      }
      authedUsername = data.username || '';
      localStorage.setItem(STORAGE_TOKEN_KEY, jwtToken);
      if (authedUsername) localStorage.setItem(STORAGE_USER_KEY, authedUsername);
      if (authedUsername) {
        createUser = authedUsername;
        joinUser = authedUsername;
        setAuthMode('auth');
      }
      await syncAuthFromToken();
    } catch (e) {
      authError = 'Auth error: ' + (e && e.message ? e.message : e);
    }
  }

  async function register() {
    try {
      authError = '';
      const username = authLogin.trim();
      const email = authEmail.trim();
      const confirm = regConfirm;
      if (!username || !authPass) {
        authError = 'Username and password required';
        return;
      }
      if (confirm !== authPass) {
        authError = 'Passwords do not match';
        return;
      }
      const base = httpBaseFromWs(wsUrl);
      const res = await fetch(`${base}/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, email, password: authPass })
      });
      if (!res.ok) {
        const text = await res.text();
        authError = 'Register failed: ' + text;
        return;
      }
      const data = await res.json();
      jwtToken = data.token || '';
      if (!jwtToken) {
        authError = 'Token not received';
        return;
      }
      authedUsername = data.username || '';
      localStorage.setItem(STORAGE_TOKEN_KEY, jwtToken);
      if (authedUsername) localStorage.setItem(STORAGE_USER_KEY, authedUsername);
      if (authedUsername) {
        createUser = authedUsername;
        joinUser = authedUsername;
        setAuthMode('auth');
      }
      await syncAuthFromToken();
    } catch (e) {
      authError = 'Register error: ' + (e && e.message ? e.message : e);
    }
  }

  async function syncAuthFromToken() {
    const token = jwtToken.trim();
    if (!token) {
      authedUsername = '';
      authedUserID = '';
      setAuthMode('guest');
      return;
    }
    try {
      const base = httpBaseFromWs(wsUrl);
      const res = await fetch(`${base}/me`, {
        headers: { Authorization: 'Bearer ' + token }
      });
      if (!res.ok) {
        authedUsername = '';
        authedUserID = '';
        setAuthMode('guest');
        return;
      }
      const data = await res.json();
      authedUsername = data.username || '';
      authedUserID = data.user_id || '';
      if (authedUsername) {
        localStorage.setItem(STORAGE_TOKEN_KEY, token);
        localStorage.setItem(STORAGE_USER_KEY, authedUsername);
        setAuthMode('auth');
        profileUser = authedUsername;
        try {
          const statsRes = await fetch(`${base}/me/stats`, {
            headers: { Authorization: 'Bearer ' + token }
          });
          if (statsRes.ok) {
            const s = await statsRes.json();
            profileDuels = String(s.total_duels_played ?? '-');
            profileWins = String(s.total_duels_won ?? '-');
            profileAcc = s.overall_accuracy != null ? String(s.overall_accuracy) : '-';
            profileStreak = String(s.best_win_streak ?? '-');
          }
          const duelsRes = await fetch(`${base}/me/duels`, {
            headers: { Authorization: 'Bearer ' + token }
          });
          if (duelsRes.ok) {
            const list = await duelsRes.json();
            profileDuelsCount = String(list.length || 0);
            profileDuelsList = list.map((d) => {
              const status = (d.status || 'unknown').toUpperCase();
              let badgeClass = 'pending';
              let resultLabel = 'PENDING';
              if (status === 'FINISHED' && d.winner_user_id) {
                if (authedUserID && d.winner_user_id === authedUserID) {
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
                room: d.room_code || '-',
                status,
                created,
                badgeClass,
                resultLabel,
                opponentName
              };
            });
          }
        } catch {}
      } else {
        setAuthMode('guest');
      }
    } catch {
      authedUsername = '';
      authedUserID = '';
      setAuthMode('guest');
    }
  }

  function generateRoom() {
    createRoom = 'room-' + Math.random().toString(36).slice(2, 8);
  }

  async function copyLink(roomId, setter) {
    const link = buildRoomLink(roomId);
    try {
      await navigator.clipboard.writeText(link);
      showNote(setter, 'Link copied');
    } catch {
      showNote(setter, 'Copy failed');
    }
  }

  function connectClick() {
    startError = '';
    if (flowMode === 'create') {
      ensureRoomId();
      currentRoom = createRoom.trim();
      isCreator = true;
      if (authMode === 'auth') {
        if (!authedUsername) {
          startError = 'Auth mode enabled, but no user is logged in';
          return;
        }
        currentUser = authedUsername;
      } else {
        let guest = createUser.trim();
        if (!guest) guest = 'Guest-' + Math.random().toString(36).slice(2, 6);
        currentUser = guest;
        createUser = guest;
        joinUser = guest;
        localStorage.setItem(STORAGE_GUEST_KEY, guest);
      }
      currentLang = createLang;
      currentTopic = createTopic;
      saveLastSession();
      connectAndJoin();
      return;
    }

    currentRoom = joinRoom.trim();
    isCreator = false;
    if (authMode === 'auth') {
      if (!authedUsername) {
        startError = 'Auth mode enabled, but no user is logged in';
        return;
      }
      currentUser = authedUsername;
    } else {
      let guest = joinUser.trim();
      if (!guest) guest = 'Guest-' + Math.random().toString(36).slice(2, 6);
      currentUser = guest;
      joinUser = guest;
      createUser = guest;
      localStorage.setItem(STORAGE_GUEST_KEY, guest);
    }
    currentLang = 'en';
    currentTopic = 'default';
    if (!currentRoom) {
      startError = 'Room ID is required';
      return;
    }
    saveLastSession();
    connectAndJoin();
  }

  function gateJoin() {
    startError = '';
    if (!currentRoom) return;
    isCreator = false;
    let guest = gateNick.trim();
    if (!guest) guest = 'Guest-' + Math.random().toString(36).slice(2, 6);
    currentUser = guest;
    joinUser = guest;
    createUser = guest;
    localStorage.setItem(STORAGE_GUEST_KEY, guest);
    currentLang = 'en';
    currentTopic = 'default';
    saveLastSession();
    connectAndJoin();
  }

  function sendAnswer() {
    if (!ws) return;
    const speed = roundStartAt ? Date.now() - roundStartAt : 0;
    const msg = {
      type: 'answer',
      room_id: currentRoom,
      user_id: currentUser,
      answer: answer.trim(),
      speed
    };
    ws.send(JSON.stringify(msg));
    answer = '';
  }

  function leaveMatch() {
    try {
      if (ws) ws.close();
    } catch {}
    currentRoom = '';
    clearLastSession();
    showScreen('start');
    profileOpen = false;
  }

  function logout() {
    try {
      if (ws) ws.close();
    } catch {}
    jwtToken = '';
    authedUsername = '';
    authedUserID = '';
    localStorage.removeItem(STORAGE_TOKEN_KEY);
    localStorage.removeItem(STORAGE_USER_KEY);
    clearLastSession();
    profileUser = '-';
    profileDuels = '-';
    profileWins = '-';
    profileAcc = '-';
    profileStreak = '-';
    profileDuelsCount = '0';
    profileDuelsList = [];
    setAuthMode('guest');
    showScreen('start');
    profileOpen = false;
  }

  onMount(() => {
    if (initialScreen) {
      screen = initialScreen;
    }
    if (initialFlow) {
      flowMode = initialFlow;
    }
    if (initialAuthMode) {
      authMode = initialAuthMode;
    }
    if (!wsUrl) {
      const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
      wsUrl = `${proto}://${window.location.host}/ws`;
    }
    const storedToken = localStorage.getItem(STORAGE_TOKEN_KEY) || '';
    const storedUser = localStorage.getItem(STORAGE_USER_KEY) || '';
    const storedGuest = localStorage.getItem(STORAGE_GUEST_KEY) || '';

    if (storedToken) jwtToken = storedToken;
    if (storedUser) authedUsername = storedUser;
    if (storedGuest) {
      createUser = storedGuest;
      joinUser = storedGuest;
    }

    syncAuthFromToken();
    if (openProfileOnLoad) {
      profileOpen = true;
    }

    const params = new URLSearchParams(window.location.search);
    const roomFromUrl = params.get('room');
    if (roomFromUrl) {
      joinRoom = roomFromUrl;
      setFlowMode('join');
      currentRoom = roomFromUrl;
      if (jwtToken.trim()) {
        setTimeout(() => connectClick(), 50);
      } else {
        showScreen('joingate');
      }
      return;
    }

    if (initialScreen === 'lobby') {
      showScreen('lobby');
      return;
    }
    if (initialScreen === 'battle') {
      showScreen('battle');
      return;
    }

    const last = localStorage.getItem(STORAGE_LAST_KEY);
    if (last) {
      try {
        const s = JSON.parse(last);
        currentRoom = s.room || '';
        currentUser = s.user || '';
        isCreator = !!s.creator;
        if (s.flow) setFlowMode(s.flow);
        if (s.auth) setAuthMode(s.auth);
      } catch {
        clearLastSession();
      }
      if (currentRoom) {
        if (!currentUser && authMode !== 'auth') {
          joinRoom = currentRoom;
          setFlowMode('join');
          showScreen('joingate');
        } else {
          showScreen('reconnect');
        }
      } else {
        showScreen('start');
      }
    } else {
      showScreen('start');
    }
  });
</script>

<svelte:window on:keydown={(e) => {
  if (e.key === 'Enter' && screen === 'battle' && answer.trim()) {
    sendAnswer();
  }
}} />

<div class="wrap">
  <header>
    <div>
      <div class="title">LangDuel Demo</div>
      <div class="title-badge">DUEL MODE</div>
    </div>
    <div class="status">{status}</div>
  </header>

  {#if authedUsername}
    <button class="profile-fab" on:click={() => (profileOpen = true)}>Profile</button>
  {/if}

  <ProfilePanel
    open={profileOpen}
    {profileUser}
    {profileDuels}
    {profileWins}
    {profileAcc}
    {profileStreak}
    {profileDuelsCount}
    {profileDuelsList}
    onClose={() => (profileOpen = false)}
    onLogout={logout}
  />

  {#if screen === 'start'}
    <StartPanel
      bind:flowMode
      authMode={authMode}
      bind:authTab
      bind:createUser
      bind:joinUser
      bind:createRoom
      bind:joinRoom
      bind:createLang
      bind:createTopic
      bind:createCollapsed
      bind:authLogin
      bind:authPass
      bind:authEmail
      bind:regConfirm
      startError={startError}
      authError={authError}
      createCopyNote={createCopyNote}
      onConnect={connectClick}
      onSetAuthMode={setAuthMode}
      onLogin={login}
      onRegister={register}
      onGenerateRoom={generateRoom}
      onCopyRoom={() => {
        ensureRoomId();
        copyLink(createRoom || 'room', (t) => (createCopyNote = t));
      }}
      buildRoomLink={buildRoomLink}
    />
  {/if}

  {#if screen === 'lobby'}
    <LobbyPanel
      {lobbyText}
      {isCreator}
      {currentRoom}
      {lobbyCopyNote}
      buildRoomLink={buildRoomLink}
      onCopy={() => copyLink(currentRoom, (t) => (lobbyCopyNote = t))}
    />
  {/if}

  {#if screen === 'reconnect'}
    <ReconnectPanel
      {reconnectNote}
      onReconnect={() => {
        reconnectNote = 'Reconnecting...';
        connectAndJoin();
      }}
      onBack={() => {
        currentRoom = '';
        clearLastSession();
        showScreen('start');
      }}
    />
  {/if}

  {#if screen === 'joingate'}
    <JoinGatePanel bind:gateNick onJoin={gateJoin} />
  {/if}

  {#if screen === 'battle'}
    <BattleView
      {playerA}
      {playerB}
      {hp}
      {promptText}
      {timerText}
      {roundInfo}
      {correctCount}
      {wrongCount}
      {totalDamage}
      avgSpeedValue={avgSpeed()}
      bind:answer
      {hitA}
      {hitB}
      {gameOverOpen}
      {gameOverText}
      {gameOverHP}
      onSend={sendAnswer}
      onLeave={leaveMatch}
      onPlayAgain={() => {
        try {
          if (ws) ws.close();
        } catch {}
        setFlowMode('create');
        ensureRoomId();
        showScreen('start');
      }}
    />
  {/if}
</div>

<script>
  export let flowMode = 'create';
  export let authMode = 'guest';
  export let authTab = 'login';
  export let startError = '';
  export let authError = '';
  export let createUser = '';
  export let joinUser = '';
  export let createRoom = '';
  export let joinRoom = '';
  export let createLang = 'en';
  export let createTopic = 'default';
  export let createCollapsed = false;
  export let createCopyNote = '';

  export let onConnect = () => {};
  export let onLogin = () => {};
  export let onRegister = () => {};
  export let onGenerateRoom = () => {};
  export let onCopyRoom = () => {};
  export let onSetAuthMode = () => {};

  export let buildRoomLink = () => '';

  export let authLogin = '';
  export let authPass = '';
  export let authEmail = '';
  export let regConfirm = '';
</script>

<div class="panel">
  <h3>Start</h3>
  <div class="controls flow">
    <div class="toggle" role="tablist" aria-label="Flow">
      <button class:active={flowMode === 'create'} type="button" on:click={() => (flowMode = 'create')}>Create</button>
      <button class:active={flowMode === 'join'} type="button" on:click={() => (flowMode = 'join')}>Join</button>
    </div>
    <button on:click={onConnect}>{flowMode === 'create' ? 'Create & Connect' : 'Join & Connect'}</button>
  </div>

  {#if startError}
    <div class="notice error">{startError}</div>
  {/if}

  <div class="controls two">
    <div class="toggle" role="tablist" aria-label="Mode">
      <button class:active={authMode === 'guest'} type="button" on:click={() => onSetAuthMode('guest')}>Guest</button>
      <button class:active={authMode === 'auth'} type="button" on:click={() => onSetAuthMode('auth')}>Auth</button>
    </div>
    <div class="status-badge">
      <span class={`status-dot ${authMode === 'auth' ? 'on' : ''}`}></span>
      <span>Mode: {authMode === 'auth' ? 'Auth' : 'Guest'}</span>
    </div>
  </div>

  {#if authMode === 'auth'}
    <div class="section">
      <div class="controls two">
        <div class="toggle" role="tablist" aria-label="Auth mode">
          <button class:active={authTab === 'login'} type="button" on:click={() => (authTab = 'login')}>Login</button>
          <button class:active={authTab === 'reg'} type="button" on:click={() => (authTab = 'reg')}>Register</button>
        </div>
        <div class="small">Account access</div>
      </div>

      {#if authTab === 'login'}
        <div class="controls three">
          <input placeholder="Username" bind:value={authLogin} />
          <input type="password" placeholder="Password" bind:value={authPass} />
          <button on:click={onLogin}>Login</button>
        </div>
      {:else}
        <div class="controls three">
          <input placeholder="Email" bind:value={authEmail} />
          <input type="password" placeholder="Confirm password" bind:value={regConfirm} />
          <button on:click={onRegister}>Register</button>
        </div>
      {/if}

      {#if authError}
        <div class="notice error">{authError}</div>
      {/if}
    </div>
  {/if}

  <div class="section" id="blockCreate">
    <div class="controls two">
      <h3 style="margin:0">Create Duel</h3>
      <button on:click={() => (createCollapsed = !createCollapsed)}>{createCollapsed ? 'Expand' : 'Collapse'}</button>
    </div>
    {#if !createCollapsed}
      <div class="controls three">
        <input placeholder="Nickname (Guest)" bind:value={createUser} />
        <select bind:value={createLang}>
          <option value="en">English</option>
        </select>
        <select bind:value={createTopic}>
          <option value="default">Default pack</option>
          <option value="animals">Animals</option>
          <option value="travel">Travel</option>
          <option value="food">Food</option>
        </select>
      </div>
      <div class="controls three">
        <input placeholder="room_id" bind:value={createRoom} />
        <button on:click={onGenerateRoom}>Generate Room</button>
        <div class="small">Create generates a room link.</div>
      </div>
      <div class="controls two">
        <div class="link">Room link: <span>{createRoom ? buildRoomLink(createRoom) : '-'}</span></div>
        <button on:click={onCopyRoom}>Copy Link</button>
      </div>
      {#if createCopyNote}
        <div class="small">{createCopyNote}</div>
      {/if}
    {/if}
  </div>

  {#if flowMode === 'join'}
    <div class="section">
      <h3>Join Duel</h3>
      <div class="controls three">
        <input placeholder="Nickname (Guest)" bind:value={joinUser} />
        <input placeholder="room_id" bind:value={joinRoom} />
        <div class="small">Join by room id or link.</div>
      </div>
    </div>
  {/if}
</div>

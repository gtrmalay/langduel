<script>
  import { _ } from 'svelte-i18n';

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
  <h3>{$_('home.play')}</h3>
  <div class="controls flow">
    <div class="toggle" role="tablist" aria-label="Flow">
      <button class:active={flowMode === 'create'} type="button" on:click={() => (flowMode = 'create')}>{$_('play.create')}</button>
      <button class:active={flowMode === 'join'} type="button" on:click={() => (flowMode = 'join')}>{$_('play.join')}</button>
    </div>
    <button on:click={onConnect}>{flowMode === 'create' ? $_('play.createRoom') : $_('play.joinRoom')}</button>
  </div>

  {#if startError}
    <div class="notice error">{startError}</div>
  {/if}

  <div class="controls two">
    <div class="toggle" role="tablist" aria-label="Mode">
      <button class:active={authMode === 'guest'} type="button" on:click={() => onSetAuthMode('guest')}>{$_('auth.playAsGuestBtn')}</button>
      <button class:active={authMode === 'auth'} type="button" on:click={() => onSetAuthMode('auth')}>{$_('auth.login')}</button>
    </div>
    <div class="status-badge">
      <span class={`status-dot ${authMode === 'auth' ? 'on' : ''}`}></span>
      <span>{$_('auth.guestMode')}: {authMode === 'auth' ? $_('auth.login') : $_('auth.guestMode')}</span>
    </div>
  </div>

  {#if authMode === 'auth'}
    <div class="section">
      <div class="controls two">
        <div class="toggle" role="tablist" aria-label="Auth mode">
          <button class:active={authTab === 'login'} type="button" on:click={() => (authTab = 'login')}>{$_('auth.login')}</button>
          <button class:active={authTab === 'reg'} type="button" on:click={() => (authTab = 'reg')}>{$_('auth.register')}</button>
        </div>
        <div class="small">{$_('play.accountAccess')}</div>
      </div>

      {#if authTab === 'login'}
        <div class="controls three">
          <input placeholder={$_('auth.username')} bind:value={authLogin} />
          <input type="password" placeholder={$_('auth.password')} bind:value={authPass} />
          <button on:click={onLogin}>{$_('auth.loginBtn')}</button>
        </div>
      {:else}
        <div class="controls three">
          <input placeholder={$_('auth.email')} bind:value={authEmail} />
          <input type="password" placeholder={$_('auth.confirmPassword')} bind:value={regConfirm} />
          <button on:click={onRegister}>{$_('auth.registerBtn')}</button>
        </div>
      {/if}

      {#if authError}
        <div class="notice error">{authError}</div>
      {/if}
    </div>
  {/if}

  <div class="section" id="blockCreate">
    <div class="controls two">
      <h3 style="margin:0">{$_('play.createDuel')}</h3>
      <button on:click={() => (createCollapsed = !createCollapsed)}>{createCollapsed ? $_('play.expand') : $_('play.collapse')}</button>
    </div>
    {#if !createCollapsed}
      <div class="controls three">
        <input placeholder={$_('play.nickname')} bind:value={createUser} />
        <select bind:value={createLang}>
          <option value="en">English</option>
        </select>
        <select bind:value={createTopic}>
          <option value="default">{$_('play.defaultPack')}</option>
          <option value="animals">{$_('topics.animals')}</option>
          <option value="travel">{$_('topics.travel')}</option>
          <option value="food">{$_('topics.food')}</option>
        </select>
      </div>
      <div class="controls three">
        <input placeholder={$_('play.room')} bind:value={createRoom} />
        <button on:click={onGenerateRoom}>{$_('play.generateRoom')}</button>
        <div class="small">{$_('play.createRoom')}</div>
      </div>
      <div class="controls two">
        <div class="link">{$_('play.roomLink')}: <span>{createRoom ? buildRoomLink(createRoom) : '-'}</span></div>
        <button on:click={onCopyRoom}>{$_('play.copyLink')}</button>
      </div>
      {#if createCopyNote}
        <div class="small">{createCopyNote}</div>
      {/if}
    {/if}
  </div>

  {#if flowMode === 'join'}
    <div class="section">
      <h3>{$_('play.joinDuel')}</h3>
      <div class="controls three">
        <input placeholder={$_('play.nickname')} bind:value={joinUser} />
        <input placeholder={$_('play.room')} bind:value={joinRoom} />
        <div class="small">{$_('play.joinRoom')}</div>
      </div>
    </div>
  {/if}
</div>

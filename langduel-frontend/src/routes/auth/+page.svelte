<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { duel } from '$lib/stores/duel.js';

  let next = '/play';
  let authLogin = '';
  let authPass = '';
  let authEmail = '';
  let regConfirm = '';

  onMount(() => {
    duel.init();
    const params = new URLSearchParams(window.location.search);
    next = params.get('next') || '/play';
  });

  async function doLogin() {
    await duel.login(authLogin.trim(), authPass);
    if ($duel.authedUsername) {
      duel.setAuthMode('auth');
      goto(next);
    }
  }

  async function doRegister() {
    await duel.register(authLogin.trim(), authEmail.trim(), authPass, regConfirm);
    if ($duel.authedUsername) {
      duel.setAuthMode('auth');
      goto(next);
    }
  }

  function chooseGuest() {
    duel.selectGuest();
    goto(next);
  }

  function chooseAuth() {
    // Stay on this page, just show auth form
  }
</script>

<div class="wrap">
  <div class="hero">
    <h1 class="title">LANGDUEL</h1>
    <p class="subtitle">Choose your path</p>
  </div>

  <div class="choice-section">
    <button class="choice-card guest" on:click={chooseGuest}>
      <span class="choice-icon">🎮</span>
      <span class="choice-title">PLAY AS GUEST</span>
      <span class="choice-desc">Quick access, no account needed</span>
    </button>

    <div class="divider">OR</div>

    <div class="auth-form">
      <div class="tabs">
        <button 
          class="tab" 
          class:active={$duel.authTab === 'login'}
          on:click={() => duel.setAuthTab('login')}
        >
          LOGIN
        </button>
        <button 
          class="tab" 
          class:active={$duel.authTab === 'reg'}
          on:click={() => duel.setAuthTab('reg')}
        >
          REGISTER
        </button>
      </div>

      {#if $duel.authTab === 'login'}
        <div class="form-fields">
          <input 
            type="text" 
            placeholder="Username" 
            bind:value={authLogin} 
          />
          <input 
            type="password" 
            placeholder="Password" 
            bind:value={authPass} 
            on:keydown={(e) => e.key === 'Enter' && doLogin()}
          />
          <button class="submit-btn" on:click={doLogin}>
            LOGIN
          </button>
        </div>
      {:else}
        <div class="form-fields">
          <input 
            type="text" 
            placeholder="Username" 
            bind:value={authLogin} 
          />
          <input 
            type="email" 
            placeholder="Email" 
            bind:value={authEmail} 
          />
          <input 
            type="password" 
            placeholder="Password" 
            bind:value={authPass} 
          />
          <input 
            type="password" 
            placeholder="Confirm password" 
            bind:value={regConfirm}
            on:keydown={(e) => e.key === 'Enter' && doRegister()}
          />
          <button class="submit-btn" on:click={doRegister}>
            REGISTER
          </button>
        </div>
      {/if}

      {#if $duel.authError}
        <div class="error">{$duel.authError}</div>
      {/if}
    </div>
  </div>

  <button class="back-btn" on:click={() => goto('/')}>
    ← BACK TO HOME
  </button>
</div>

<style>
  .wrap {
    max-width: 480px;
    margin: 0 auto;
    padding: 40px 20px;
    display: flex;
    flex-direction: column;
    gap: 28px;
  }

  .hero {
    text-align: center;
  }

  .title {
    font-family: "Press Start 2P", cursive;
    font-size: 24px;
    color: var(--text);
    margin: 0;
    letter-spacing: 3px;
  }

  .subtitle {
    font-size: 13px;
    color: var(--muted);
    margin: 12px 0 0 0;
  }

  .choice-section {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .choice-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 32px 24px;
    border-radius: 16px;
    border: 2px solid var(--outline);
    background: var(--card);
    cursor: pointer;
    transition: all 0.2s;
  }

  .choice-card.guest {
    border-color: rgba(37, 244, 183, 0.3);
  }

  .choice-card.guest:hover {
    border-color: var(--accent);
    box-shadow: 0 0 25px rgba(37, 244, 183, 0.25);
    transform: scale(1.02);
  }

  .choice-icon {
    font-size: 36px;
  }

  .choice-title {
    font-family: "Press Start 2P", cursive;
    font-size: 14px;
    color: var(--text);
    letter-spacing: 1px;
  }

  .choice-desc {
    font-size: 12px;
    color: var(--muted);
  }

  .divider {
    text-align: center;
    font-size: 11px;
    color: var(--muted);
    letter-spacing: 2px;
    position: relative;
  }

  .divider::before,
  .divider::after {
    content: '';
    position: absolute;
    top: 50%;
    width: 40%;
    height: 1px;
    background: var(--outline);
  }

  .divider::before {
    left: 0;
  }

  .divider::after {
    right: 0;
  }

  .auth-form {
    background: var(--card);
    border-radius: 16px;
    padding: 24px;
    border: 1px solid var(--outline);
  }

  .tabs {
    display: flex;
    gap: 8px;
    margin-bottom: 20px;
  }

  .tab {
    flex: 1;
    padding: 12px;
    border-radius: 10px;
    border: 1px solid var(--outline);
    background: transparent;
    color: var(--muted);
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .tab.active {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--bg);
  }

  .tab:hover:not(.active) {
    border-color: var(--accent);
    color: var(--text);
  }

  .form-fields {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  input {
    width: 100%;
    padding: 14px 16px;
    border-radius: 10px;
    border: 1px solid var(--outline);
    background: rgba(11, 15, 31, 0.9);
    color: var(--text);
    font-size: 14px;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(37, 244, 183, 0.15);
  }

  input::placeholder {
    color: var(--muted);
  }

  .submit-btn {
    padding: 14px 24px;
    border-radius: 12px;
    border: 2px solid var(--accent-2);
    background: rgba(246, 193, 68, 0.15);
    color: var(--accent-2);
    font-family: "Space Grotesk", sans-serif;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    margin-top: 8px;
  }

  .submit-btn:hover {
    background: rgba(246, 193, 68, 0.25);
    box-shadow: 0 0 20px rgba(246, 193, 68, 0.3);
  }

  .error {
    margin-top: 12px;
    padding: 12px;
    border-radius: 8px;
    background: rgba(255, 92, 122, 0.15);
    border: 1px solid rgba(255, 92, 122, 0.4);
    color: var(--danger);
    font-size: 12px;
    text-align: center;
  }

  .back-btn {
    padding: 12px 20px;
    border: none;
    border-radius: 10px;
    background: transparent;
    color: var(--muted);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s;
    align-self: center;
  }

  .back-btn:hover {
    color: var(--text);
    background: rgba(255, 255, 255, 0.05);
  }
</style>

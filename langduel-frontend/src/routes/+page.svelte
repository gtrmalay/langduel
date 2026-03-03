<script>
  import { goto } from '$app/navigation';
  import { duel } from '$lib/stores/duel.js';
  import { onMount } from 'svelte';

  onMount(() => {
    duel.init();
  });
</script>

<div class="wrap">
  <div class="hero">
    <h1 class="logo">LANGDUEL</h1>
    <p class="tagline">TRANSLATE BATTLE</p>
  </div>

  <div class="buttons">
    <button class="play-btn" on:click={() => goto('/play')}>
      <span class="btn-icon">►</span>
      PLAY
    </button>
    <button class="profile-btn" on:click={() => goto('/profile')}>
      <span class="btn-icon">👤</span>
      PROFILE
    </button>
  </div>

  <div class="user-info">
    {#if $duel.authMode === 'auth' && $duel.authedUsername}
      <span class="user-badge auth">{$duel.authedUsername}</span>
    {:else}
      <span class="user-badge guest">Guest</span>
    {/if}
  </div>

  <p class="notice">No registration. Just play.</p>
</div>

<style>
  .wrap {
    max-width: 520px;
    margin: 0 auto;
    padding: 60px 20px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 32px;
    min-height: calc(100vh - 120px);
    justify-content: center;
  }

  .hero {
    text-align: center;
  }

  .logo {
    font-family: "Press Start 2P", cursive;
    font-size: 36px;
    color: var(--text);
    margin: 0;
    letter-spacing: 4px;
    text-shadow: 0 0 30px rgba(37, 244, 183, 0.4);
  }

  .tagline {
    font-family: "Press Start 2P", cursive;
    font-size: 12px;
    color: var(--accent);
    margin: 16px 0 0 0;
    letter-spacing: 3px;
  }

  .buttons {
    display: flex;
    flex-direction: column;
    gap: 16px;
    width: 100%;
  }

  .play-btn, .profile-btn {
    width: 100%;
    padding: 24px 32px;
    border-radius: 16px;
    border: 2px solid var(--accent);
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.2), rgba(37, 244, 183, 0.05));
    color: var(--text);
    font-family: "Press Start 2P", cursive;
    font-size: 16px;
    letter-spacing: 2px;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
  }

  .play-btn:hover, .profile-btn:hover {
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.3), rgba(37, 244, 183, 0.1));
    box-shadow: 0 0 30px rgba(37, 244, 183, 0.4);
    transform: scale(1.02);
  }

  .play-btn:active, .profile-btn:active {
    transform: scale(0.98) translateY(2px);
  }

  .profile-btn {
    border-color: var(--outline);
    background: rgba(23, 31, 51, 0.8);
    font-size: 13px;
  }

  .profile-btn:hover {
    border-color: var(--accent-2);
    box-shadow: 0 0 30px rgba(246, 193, 68, 0.3);
    background: rgba(23, 31, 51, 0.9);
  }

  .btn-icon {
    font-size: 18px;
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .user-badge {
    padding: 6px 14px;
    border-radius: 20px;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .user-badge.auth {
    background: rgba(37, 244, 183, 0.15);
    border: 1px solid rgba(37, 244, 183, 0.4);
    color: var(--accent);
  }

  .user-badge.guest {
    background: rgba(154, 164, 178, 0.15);
    border: 1px solid var(--outline);
    color: var(--muted);
  }

  .notice {
    font-size: 13px;
    color: var(--muted);
    margin: 0;
    text-align: center;
  }
</style>

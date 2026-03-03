<script>
  import { goto } from '$app/navigation';
  import { duel } from '$lib/stores/duel.js';

  export let show = true;

  function goProfile() {
    if ($duel.authedUsername) {
      goto('/profile');
    } else {
      goto('/auth?next=/profile');
    }
  }
</script>

{#if show}
  <header class="app-header">
    <button class="brand" on:click={() => goto('/')}>
      <span class="logo">LangDuel</span>
    </button>
    <div class="nav">
      <button class="nav-btn" on:click={() => goto('/play')}>Play</button>
      <button class="nav-btn ghost" on:click={goProfile}>Profile</button>
      {#if $duel.authMode === 'auth' && $duel.authedUsername}
        <div class="user-pill">{$duel.authedUsername}</div>
        <button class="nav-btn ghost" on:click={() => { duel.logout(); goto('/'); }}>Logout</button>
      {:else}
        <button class="nav-btn ghost" on:click={() => goto('/auth')}>
          {$duel.authedUsername ? 'Account' : 'Login'}
        </button>
      {/if}
    </div>
  </header>
{/if}

<style>
  .app-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 24px;
    max-width: 1080px;
    margin: 0 auto;
  }

  .brand {
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
  }

  .logo {
    font-family: "Press Start 2P", cursive;
    font-size: 14px;
    color: var(--text);
    letter-spacing: 2px;
  }

  .logo:hover {
    color: var(--accent);
  }

  .nav {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .nav-btn {
    padding: 10px 18px;
    border-radius: 10px;
    border: 1px solid var(--accent);
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.2), rgba(37, 244, 183, 0.05));
    color: var(--text);
    font-family: "Space Grotesk", sans-serif;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .nav-btn:hover {
    box-shadow: 0 0 15px rgba(37, 244, 183, 0.35);
  }

  .nav-btn.ghost {
    background: transparent;
    border-color: var(--outline);
    color: var(--muted);
  }

  .nav-btn.ghost:hover {
    border-color: var(--accent);
    color: var(--text);
  }

  .user-pill {
    padding: 8px 14px;
    border-radius: 999px;
    border: 1px solid var(--outline);
    background: rgba(11, 15, 31, 0.8);
    font-size: 12px;
    color: var(--accent);
  }
</style>

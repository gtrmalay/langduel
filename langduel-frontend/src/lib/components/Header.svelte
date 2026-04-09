<script>
  import { goto } from '$app/navigation';
  import { duel } from '$lib/stores/duel.js';
  import { _ } from 'svelte-i18n';
  import LangSwitcher from './LangSwitcher.svelte';

  export let show = true;
</script>

{#if show}
  <header class="app-header">
    <button class="brand" on:click={() => goto('/')}>
      <span class="logo">LangDuel</span>
    </button>
    <div class="nav">
      <button class="nav-btn" on:click={() => goto('/play')}>{$_('nav.play')}</button>
      <button class="nav-btn rank" on:click={() => goto('/leaderboard')}>🏆 {$_('nav.rating')}</button>
      
      {#if $duel.authedUsername}
        <button class="nav-btn profile" on:click={() => goto('/profile')}>
          {$duel.authedUsername}
        </button>
        <button class="nav-btn ghost" on:click={() => {
          if (confirm($_('confirm.logout'))) {
            duel.logout();
            goto('/');
          }
        }}>
          {$_('nav.logout')}
        </button>
      {:else}
        <button class="nav-btn ghost" on:click={() => goto('/auth')}>
          {$_('nav.login')}
        </button>
      {/if}
      <LangSwitcher />
    </div>
  </header>
{/if}

<style>
  .app-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 32px;
    width: 100%;
    position: sticky;
    top: 0;
    z-index: 100;
    background: rgba(11, 16, 32, 0.85);
    backdrop-filter: blur(12px);
    border-bottom: 1px solid rgba(43, 52, 74, 0.5);
  }

  .brand {
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px 0;
  }

  .logo {
    font-family: "Press Start 2P", cursive;
    font-size: 13px;
    color: var(--text);
    letter-spacing: 2px;
  }

  .logo:hover {
    color: var(--accent);
  }

  .nav {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .nav-btn {
    padding: 8px 16px;
    border-radius: 8px;
    border: 1px solid var(--accent);
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.15), rgba(37, 244, 183, 0.05));
    color: var(--text);
    font-family: "Space Grotesk", sans-serif;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .nav-btn:hover {
    background: linear-gradient(135deg, rgba(37, 244, 183, 0.25), rgba(37, 244, 183, 0.1));
    box-shadow: 0 0 12px rgba(37, 244, 183, 0.3);
  }

  .nav-btn.profile {
    border-color: var(--accent-2);
    background: rgba(246, 193, 68, 0.12);
    color: var(--accent-2);
  }

  .nav-btn.profile:hover {
    background: rgba(246, 193, 68, 0.2);
    box-shadow: 0 0 12px rgba(246, 193, 68, 0.25);
  }

  .nav-btn.ghost {
    background: transparent;
    border-color: var(--outline);
    color: var(--muted);
  }

  .nav-btn.ghost:hover {
    border-color: var(--text);
    color: var(--text);
    background: rgba(255, 255, 255, 0.05);
  }

  .nav-btn.rank {
    border-color: var(--accent-2);
    background: rgba(246, 193, 68, 0.1);
    color: var(--accent-2);
  }

  .nav-btn.rank:hover {
    background: rgba(246, 193, 68, 0.2);
    box-shadow: 0 0 12px rgba(246, 193, 68, 0.25);
  }
</style>

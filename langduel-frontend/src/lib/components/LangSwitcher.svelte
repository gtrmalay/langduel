<script>
  import { locale } from 'svelte-i18n';
  import { setLocale } from '$lib/i18n';

  const languages = [
    { code: 'en', label: 'EN' },
    { code: 'ru', label: 'RU' }
  ];

  function changeLocale(code) {
    locale.set(code);
    setLocale(code);
  }
</script>

<div class="lang-switcher">
  {#each languages as lang}
    <button 
      class:active={$locale === lang.code}
      on:click={() => changeLocale(lang.code)}
      title={lang.code === 'en' ? 'English' : 'Русский'}
    >
      {lang.label}
    </button>
  {/each}
</div>

<style>
  .lang-switcher {
    display: flex;
    background: rgba(11, 15, 31, 0.9);
    border: 1px solid var(--outline);
    border-radius: 8px;
    padding: 3px;
    gap: 2px;
  }

  .lang-switcher button {
    padding: 6px 12px;
    border-radius: 6px;
    border: none;
    background: transparent;
    color: var(--muted);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s;
    box-shadow: none;
    letter-spacing: 1px;
  }

  .lang-switcher button:hover {
    background: rgba(255, 255, 255, 0.05);
    box-shadow: none;
    color: var(--text);
  }

  .lang-switcher button.active {
    background: rgba(37, 244, 183, 0.2);
    color: var(--accent);
  }
</style>

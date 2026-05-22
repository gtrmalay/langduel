<script>
  import { _ } from 'svelte-i18n';

  export let gateNick = '';
  export let onJoin = () => {};
</script>

<div class="panel">
  <h3>{$_('play.joinDuel')}</h3>
  <div class="small">{$_('play.enterUsername')}</div>
  <div class="controls two">
    <input placeholder={$_('play.nickname')} bind:value={gateNick} />
    <button on:click={onJoin}>{$_('play.join')}</button>
  </div>
  <div class="small">{$_('auth.login')} {$_('auth.or')} {$_('auth.register')}.</div>
</div>

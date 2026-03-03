<script>
  export let gateNick = '';
  export let onJoin = () => {};
</script>

<div class="panel">
  <h3>Join Duel</h3>
  <div class="small">Enter nickname to join the room</div>
  <div class="controls two">
    <input placeholder="Nickname (Guest)" bind:value={gateNick} />
    <button on:click={onJoin}>Join</button>
  </div>
  <div class="small">Or login to use your account.</div>
</div>

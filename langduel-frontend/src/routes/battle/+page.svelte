<script>
  import { onMount } from 'svelte';
  import { duel } from '$lib/stores/duel.js';
  import BattleView from '$lib/components/BattleView.svelte';

  let answer = '';

  $: currentUser = $duel.currentUser || '';
  $: playerA = $duel.playerA || '';
  $: playerB = $duel.playerB || '';
  $: isPlayerA = currentUser && playerA === currentUser;
  $: playerAAvatar = isPlayerA ? ($duel.userAvatar || 'default') : ($duel.opponentAvatar || 'default');
  $: playerBAvatar = isPlayerA ? ($duel.opponentAvatar || 'default') : ($duel.userAvatar || 'default');

  $: ping = $duel.ping;
  $: pingColor = ping < 0 ? '#ff5c7a' : ping < 100 ? '#25f4b7' : ping < 300 ? '#f6c144' : '#ff5c7a';
  $: pingIcon = ping < 0 ? '🔴' : ping < 100 ? '🟢' : ping < 300 ? '🟡' : '🔴';

  onMount(() => {
    duel.init();
    const params = new URLSearchParams(window.location.search);
    const room = params.get('room');
    if (room) {
      duel.setField('currentRoom', room);
    }
  });
</script>

<div class="battle-wrap">
  <BattleView
    playerA={playerA}
    playerB={playerB}
    playerAEmoji={playerAAvatar}
    playerBEmoji={playerBAvatar}
    hp={$duel.hp}
    promptText={$duel.promptText}
    timerText={$duel.timerText}
    roundInfo={$duel.roundInfo}
    correctCount={$duel.correctCount}
    wrongCount={$duel.wrongCount}
    totalDamage={$duel.totalDamage}
    playerADamage={$duel.playerADamage}
    playerBDamage={$duel.playerBDamage}
    avgSpeedValue={duel.avgSpeed()}
    bind:answer
    hitA={$duel.hitA}
    hitB={$duel.hitB}
    lastDamage={$duel.lastDamage}
    lastDamageTo={$duel.lastDamageTo}
    attackA={$duel.attackA}
    attackB={$duel.attackB}
    inputCorrect={$duel.inputCorrect}
    inputWrong={$duel.inputWrong}
    gameOverOpen={$duel.gameOverOpen}
    gameOverText={$duel.gameOverText}
    gameOverHP={$duel.gameOverHP}
    gameOverReason={$duel.gameOverReason}
    isGameWinner={$duel.isGameWinner}
    duelId={$duel.currentDuelId || ''}
    connectionStatus={$duel.connectionStatus}
    rematchWaiting={$duel.rematchWaiting}
    rematchRequested={$duel.rematchRequested}
    ping={$duel.ping}
    onSend={() => {
      duel.sendAnswer(answer);
      answer = '';
    }}
    onLeave={() => duel.leaveMatch()}
    onPlayAgain={() => duel.rematchAndConnect()}
  />
</div>

<style>
  .battle-wrap {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
  }
</style>

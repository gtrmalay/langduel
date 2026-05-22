<script>
  import { onMount } from 'svelte';
  import { duel } from '$lib/stores/duel.js';
  import BattleView from '$lib/components/BattleView.svelte';

  let answer = '';

  $: currentUser = $duel.currentUser || '';
  $: playerA = $duel.playerA || '';
  $: playerB = $duel.playerB || '';
  $: isPlayerA = currentUser && playerA === currentUser;

  // "я" всегда слева (left), соперник — справа (right)
  $: leftPlayer  = isPlayerA ? playerA : playerB;
  $: rightPlayer = isPlayerA ? playerB : playerA;
  $: leftAvatar  = $duel.userAvatar    || 'default';
  $: rightAvatar = $duel.opponentAvatar || 'default';
  $: leftHit     = isPlayerA ? $duel.hitA    : $duel.hitB;
  $: rightHit    = isPlayerA ? $duel.hitB    : $duel.hitA;
  $: leftAttack  = isPlayerA ? $duel.attackA : $duel.attackB;
  $: rightAttack = isPlayerA ? $duel.attackB : $duel.attackA;
  $: leftDamage  = isPlayerA ? $duel.playerADamage : $duel.playerBDamage;
  $: rightDamage = isPlayerA ? $duel.playerBDamage : $duel.playerADamage;

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
    playerA={leftPlayer}
    playerB={rightPlayer}
    playerAEmoji={leftAvatar}
    playerBEmoji={rightAvatar}
    hp={$duel.hp}
    promptText={$duel.promptText}
    timerText={$duel.timerText}
    roundInfo={$duel.roundInfo}
    correctCount={$duel.correctCount}
    wrongCount={$duel.wrongCount}
    totalDamage={$duel.totalDamage}
    playerADamage={leftDamage}
    playerBDamage={rightDamage}
    avgSpeedValue={duel.avgSpeed()}
    bind:answer
    hitA={leftHit}
    hitB={rightHit}
    lastDamage={$duel.lastDamage}
    lastDamageTo={$duel.lastDamageTo}
    attackA={leftAttack}
    attackB={rightAttack}
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

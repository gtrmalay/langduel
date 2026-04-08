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
    avgSpeedValue={duel.avgSpeed()}
    bind:answer
    hitA={$duel.hitA}
    hitB={$duel.hitB}
    gameOverOpen={$duel.gameOverOpen}
    gameOverText={$duel.gameOverText}
    gameOverHP={$duel.gameOverHP}
    gameOverReason={$duel.gameOverReason}
    duelId={$duel.currentDuelId || ''}
    onSend={() => {
      duel.sendAnswer(answer);
      answer = '';
    }}
    onLeave={() => duel.leaveMatch()}
    onPlayAgain={() => duel.createAndConnect()}
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

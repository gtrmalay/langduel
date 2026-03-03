<script>
  export let items = [];
  export let value = '';
  export let onChange = () => {};
</script>

<div class="tabs" role="tablist">
  {#each items as item}
    <button
      type="button"
      class:active={value === item.value}
      on:click={() => onChange(item.value)}
    >
      {item.label}
    </button>
  {/each}
</div>

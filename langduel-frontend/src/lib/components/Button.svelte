<script>
  export let variant = 'primary';
  export let type = 'button';
  export let disabled = false;
</script>

<button class={`btn ${variant}`} {type} {disabled}>
  <slot />
</button>

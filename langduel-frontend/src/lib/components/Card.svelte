<script>
  export let title = '';
</script>

<section class="card">
  {#if title}
    <div class="card-title">{title}</div>
  {/if}
  <div class="card-body">
    <slot />
  </div>
</section>

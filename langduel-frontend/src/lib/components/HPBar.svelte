<script>
  export let value = 100;
</script>

<div class="hp-bar">
  <span style={`width: ${Math.max(0, Math.min(100, value))}%`}></span>
</div>

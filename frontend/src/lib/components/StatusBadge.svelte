<script>
  import { statusInfo, PRINT_STATUS, ORDER_STATUS } from '$lib/format.js';

  let { map, status, tone = 'cream' } = $props();
</script>

<span class="badge {tone === 'primary' ? 'badge-primary' : tone === 'dark' ? 'badge-dark' : ''}">
  {#if tone === 'primary'}<span class="dot"></span>{/if}
  {map ? statusInfo(map, status).label : status}
</span>

<style>
  .badge { font-weight: 500; }
  .badge-primary { background: var(--primary); color: var(--on-primary); }
  .badge-dark { background: var(--ink); color: var(--on-primary); }
  .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--on-primary); flex-shrink: 0; }
</style>
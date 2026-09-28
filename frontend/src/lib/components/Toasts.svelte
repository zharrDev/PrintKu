<script>
  import { toasts } from '$lib/stores.js';
  import { CircleCheckBig, CircleAlert, Info } from 'lucide-svelte';
</script>

<div class="toast-stack">
  {#each $toasts as t (t.id)}
    <div class="toast toast-{t.type}">
      {#if t.type === 'error'}
        <CircleAlert size={18} />
      {:else if t.type === 'success'}
        <CircleCheckBig size={18} />
      {:else}
        <Info size={18} />
      {/if}
      <span>{t.message}</span>
    </div>
  {/each}
</div>

<style>
  .toast-stack {
    position: fixed;
    bottom: var(--space-xl);
    right: var(--space-xl);
    z-index: var(--z-toast);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    max-width: 380px;
  }
  /* Warna semantik datang dari app.css (.toast-success/-error/-info) */
</style>
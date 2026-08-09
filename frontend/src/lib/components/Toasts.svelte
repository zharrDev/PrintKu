<script>
  import { toasts } from '$lib/stores.js';
  import { CircleCheckBig, CircleAlert, Info } from 'lucide-svelte';
</script>

<div class="toast-stack">
  {#each $toasts as t (t.id)}
    <div class="toast {t.type}">
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
    z-index: 100;
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    max-width: 380px;
  }
  .toast {
    display: flex;
    align-items: flex-start;
    gap: var(--space-sm);
    background: var(--canvas);
    border: 1px solid var(--ink);
    border-radius: var(--radius-md);
    padding: var(--space-md) var(--space-lg);
    font-size: 16px;
    line-height: 24px;
    color: var(--ink);
    box-shadow: var(--shadow-soft);
    animation: slide-in 0.2s ease;
  }
  .toast.error { border-color: var(--primary); }
  .toast.error svg { color: var(--primary); flex-shrink: 0; }
  .toast svg { color: var(--primary); flex-shrink: 0; margin-top: 2px; }
  @keyframes slide-in {
    from { transform: translateY(8px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  }
</style>
<script>
  import { onMount } from 'svelte';
  import { ShoppingCart, Menu, X } from 'lucide-svelte';
  import { userStore, cartStore, loadSession } from '$lib/stores.js';
  import { api } from '$lib/api.js';

  let open = $state(false);
  let notifications = $state([]);

  $effect(() => {
    const u = $userStore;
    if (u) loadNotifications();
  });

  async function loadNotifications() {
    const r = await api('/notifications');
    if (r.ok) notifications = r.json.notifications;
  }

  async function refreshCart() {
    const r = await api('/cart');
    if (r.ok) cartStore.set({ items: r.json.items, total: r.json.total });
  }

  onMount(() => {
    loadSession();
    if (typeof localStorage !== 'undefined' && localStorage.getItem('printmart_token')) refreshCart();
  });
</script>

<nav class="nav">
  <div class="container nav-inner">
    <a href="/" class="brand" aria-label="PrintMart beranda">
      <span class="brand-mark">P</span>
      <span class="brand-name display-xs">PrintMart</span>
    </a>

    <div class="nav-links" class:open>
      <a href="/" onclick={() => (open = false)}>Beranda</a>
      <a href="/katalog" onclick={() => (open = false)}>Katalog</a>
      <a href="/print" onclick={() => (open = false)}>Jasa Print</a>
      <a href="/pesanan" onclick={() => (open = false)}>Pesanan</a>
    </div>

    <div class="nav-actions">
      {#if $userStore}
        <a href={$userStore.role === 'admin' ? '/admin' : '/profil'} class={$userStore.role === 'admin' ? 'badge badge-primary' : 'badge'}>
          {$userStore.name.split(' ')[0]}
        </a>
      {:else}
        <a href="/masuk" class="btn-text">Masuk</a>
        <a href="/daftar" class="btn btn-primary btn-sm">Daftar</a>
      {/if}
      <a href="/keranjang" class="cart-btn" aria-label="Keranjang">
        <ShoppingCart size={20} strokeWidth={2} />
        {#if $cartStore.items.length > 0}
          <span class="cart-count">{$cartStore.items.reduce((n, it) => n + it.quantity, 0)}</span>
        {/if}
      </a>
      <button class="hamburger" aria-label="Menu" onclick={() => (open = !open)}>
        {#if open}<X size={22} />{:else}<Menu size={22} />{/if}
      </button>
    </div>
  </div>
</nav>

<style>
  .nav {
    position: sticky;
    top: 0;
    z-index: 50;
    background: var(--canvas);
    border-bottom: 1px solid rgba(32, 21, 21, 0.1);
  }
  .nav-inner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-xl);
    padding-top: var(--space-md);
    padding-bottom: var(--space-md);
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: var(--space-sm);
  }
  .brand-mark {
    width: 32px;
    height: 32px;
    border-radius: var(--radius-sm);
    background: var(--primary);
    color: var(--on-primary);
    font-weight: 700;
    font-size: 18px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  .brand-name { font-size: 18px; }
  .nav-links {
    display: flex;
    gap: var(--space-2xl);
  }
  .nav-links a {
    font-size: 16px;
    line-height: 24px;
    color: var(--ink);
    position: relative;
  }
  .nav-links a:hover { color: var(--primary); }
  .nav-actions {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }
  .cart-btn {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: var(--radius-md);
    color: var(--ink);
    border: 1px solid var(--ink);
    background: var(--canvas);
  }
  .cart-btn:hover { background: var(--canvas-soft); }
  .cart-count {
    position: absolute;
    top: -6px;
    right: -6px;
    min-width: 18px;
    height: 18px;
    border-radius: var(--radius-pill);
    background: var(--primary);
    color: var(--on-primary);
    font-size: 12px;
    font-weight: 700;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0 4px;
  }
  .hamburger {
    display: none;
    background: none;
    border: none;
    padding: var(--space-sm);
  }
  @media (max-width: 767px) {
    .hamburger { display: inline-flex; }
    .nav-links {
      display: none;
      position: absolute;
      top: 100%;
      left: 0;
      right: 0;
      background: var(--canvas);
      flex-direction: column;
      padding: var(--space-xl) var(--space-lg);
      border-bottom: 1px solid rgba(32, 21, 21, 0.1);
      gap: var(--space-lg);
      box-shadow: var(--shadow-soft);
    }
    .nav-links.open { display: flex; }
    .btn-text { display: none; }
  }
</style>
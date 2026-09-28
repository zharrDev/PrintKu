<script>
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { ShoppingCart, Menu, X } from 'lucide-svelte';
  import { userStore, cartStore, loadSession } from '$lib/stores.js';
  import { api } from '$lib/api.js';

  let open = $state(false);
  let notifications = $state([]);
  let path = $derived(page.url.pathname);

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
    if (typeof localStorage !== 'undefined' && localStorage.getItem('printku_token')) refreshCart();
  });
</script>

<nav class="nav">
  <div class="container nav-inner">
    <a href="/" class="brand" aria-label="PrintKu beranda">
      <span class="brand-mark">P</span>
      <span class="brand-name display-xs">PrintKu</span>
    </a>

    <div class="nav-links" class:open>
      <a href="/" class="nav-link" aria-current={path === '/' ? 'page' : undefined} onclick={() => (open = false)}>Beranda</a>
      <a href="/katalog" class="nav-link" aria-current={path.startsWith('/katalog') || path.startsWith('/produk') ? 'page' : undefined} onclick={() => (open = false)}>Katalog</a>
      <a href="/print" class="nav-link" aria-current={path === '/print' ? 'page' : undefined} onclick={() => (open = false)}>Jasa Print</a>
      <a href="/pesanan" class="nav-link" aria-current={path.startsWith('/pesanan') || path.startsWith('/bayar') ? 'page' : undefined} onclick={() => (open = false)}>Pesanan</a>
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
    z-index: var(--z-nav);
    background: var(--canvas);
    border-bottom: var(--border);
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
    width: 36px;
    height: 36px;
    border-radius: var(--radius-sm);
    background: var(--primary);
    color: var(--on-primary);
    border: var(--border);
    box-shadow: var(--shadow-sm);
    font-weight: 700;
    font-size: 20px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  .brand-name { font-size: 20px; }
  .nav-links {
    display: flex;
    gap: var(--space-sm);
  }
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
    width: var(--touch-min);
    height: var(--touch-min);
    border-radius: var(--radius-md);
    color: var(--ink);
    border: var(--border);
    background: var(--canvas);
    box-shadow: var(--shadow-sm);
    transition: transform var(--dur) var(--ease),
                box-shadow var(--dur) var(--ease),
                background-color var(--dur) var(--ease);
  }
  .cart-btn:hover {
    transform: translate(-2px, -2px);
    box-shadow: var(--shadow-md);
    background: var(--canvas-soft);
  }
  .cart-btn:active { transform: translate(2px, 2px); box-shadow: 0 0 0 var(--ink); }
  .cart-count {
    position: absolute;
    top: -8px;
    right: -8px;
    min-width: 22px;
    height: 22px;
    border-radius: var(--radius-pill);
    background: var(--primary);
    color: var(--on-primary);
    border: var(--border);
    font-size: 12px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0 var(--space-xs);
  }
  .hamburger {
    display: none;
    background: var(--canvas);
    border: var(--border);
    border-radius: var(--radius-md);
    color: var(--ink);
    min-width: var(--touch-min);
    min-height: var(--touch-min);
    place-content: center;
    box-shadow: var(--shadow-sm);
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
      padding: var(--space-lg);
      border-bottom: var(--border);
      gap: var(--space-sm);
      box-shadow: var(--shadow-md);
    }
    .nav-links.open { display: flex; }
    .btn-text { display: none; }
  }
</style>
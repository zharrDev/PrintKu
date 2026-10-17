<script>
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { ShoppingCart, Menu, X, Printer, Bell, CheckCheck } from 'lucide-svelte';
  import { userStore, cartStore, loadSession } from '$lib/stores.js';
  import { api } from '$lib/api.js';
  import { formatDate } from '$lib/format.js';

  let open = $state(false);
  let notifications = $state([]);
  let notifOpen = $state(false);
  let notifWrap;
  let path = $derived(page.url.pathname);
  let unread = $derived(notifications.filter((n) => !n.is_read).length);

  $effect(() => {
    const u = $userStore;
    if (u) loadNotifications();
  });

  async function loadNotifications() {
    const r = await api('/notifications');
    if (r.ok) notifications = r.json.notifications;
  }

  async function toggleNotif() {
    notifOpen = !notifOpen;
    if (notifOpen) loadNotifications();
  }

  async function openNotification(n) {
    notifOpen = false;
    if (!n.is_read) {
      await api(`/notifications/${n.id}/read`, { method: 'POST' });
      notifications = notifications.map((x) => (x.id === n.id ? { ...x, is_read: 1 } : x));
    }
    if (n.order_id) goto(`/pesanan/${n.order_id}`);
  }

  async function markAllRead() {
    await api('/notifications/read-all', { method: 'POST' });
    notifications = notifications.map((x) => ({ ...x, is_read: 1 }));
  }

  function onWindowClick(e) {
    if (notifOpen && notifWrap && !notifWrap.contains(e.target)) notifOpen = false;
  }

  function onWindowKeydown(e) {
    if (e.key === 'Escape') notifOpen = false;
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

<svelte:window onclick={onWindowClick} onkeydown={onWindowKeydown} />

<nav class="nav">
  <div class="container nav-inner">
    <a href="/" class="brand" aria-label="PrintKu beranda">
      <span class="brand-mark" aria-hidden="true"><Printer size={18} strokeWidth={2.4} /></span>
      <span class="brand-name display-xs">PrintKu</span>
    </a>

    <div class="nav-links" class:open id="nav-menu">
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
        <div class="notif-wrap" bind:this={notifWrap}>
          <button
            class="bell-btn"
            type="button"
            aria-label={`Notifikasi${unread ? `, ${unread} belum dibaca` : ''}`}
            aria-expanded={notifOpen}
            aria-controls="notif-panel"
            onclick={toggleNotif}
          >
            <Bell size={20} strokeWidth={2} />
            {#if unread > 0}
              <span class="cart-count">{unread > 9 ? '9+' : unread}</span>
            {/if}
          </button>

          {#if notifOpen}
            <div class="notif-panel" id="notif-panel">
              <div class="notif-head">
                <span class="body-sm-strong">Notifikasi</span>
                <button
                  class="btn-text notif-clear"
                  type="button"
                  onclick={markAllRead}
                  disabled={unread === 0}
                ><CheckCheck size={14} /> Tandai dibaca</button>
              </div>
              {#if notifications.length === 0}
                <p class="notif-empty caption text-body-mid">Belum ada notifikasi.</p>
              {:else}
                <ul class="notif-list">
                  {#each notifications.slice(0, 8) as n (n.id)}
                    <li>
                      <button
                        class="notif-item"
                        class:notif-unread={!n.is_read}
                        type="button"
                        onclick={() => openNotification(n)}
                      >
                        <span class="notif-dot" aria-hidden="true"></span>
                        <span class="notif-body">
                          <span class="body-sm">{n.message}</span>
                          <span class="caption text-body-mid">{formatDate(n.created_at)}</span>
                        </span>
                      </button>
                    </li>
                  {/each}
                </ul>
                <a href="/pesanan" class="notif-more caption" onclick={() => (notifOpen = false)}>Lihat semua pesanan</a>
              {/if}
            </div>
          {/if}
        </div>
      {:else}
        <a href="/masuk" class="btn-text">Masuk</a>
        <a href="/daftar" class="btn btn-primary btn-sm">Daftar</a>
      {/if}
      <a href="/keranjang" class="cart-btn" aria-label={`Keranjang${$cartStore.items.length ? `, ${$cartStore.items.reduce((n, it) => n + it.quantity, 0)} item` : ''}`}>
        <ShoppingCart size={20} strokeWidth={2} />
        {#if $cartStore.items.length > 0}
          <span class="cart-count">{$cartStore.items.reduce((n, it) => n + it.quantity, 0)}</span>
        {/if}
      </a>
      <button
        class="hamburger"
        type="button"
        aria-label="Menu navigasi"
        aria-expanded={open}
        aria-controls="nav-menu"
        onclick={() => (open = !open)}
      >
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
    color: var(--on-accent);
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
  .cart-btn,
  .bell-btn {
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
  .cart-btn:hover,
  .bell-btn:hover {
    transform: translate(-2px, -2px);
    box-shadow: var(--shadow-md);
    background: var(--canvas-soft);
  }
  .cart-btn:active,
  .bell-btn:active { transform: translate(2px, 2px); box-shadow: 0 0 0 var(--ink); }
  .cart-count {
    position: absolute;
    top: -8px;
    right: -8px;
    min-width: 22px;
    height: 22px;
    border-radius: var(--radius-pill);
    background: var(--primary);
    color: var(--on-accent);
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

  /* ── Notifikasi ─────────────────────────────────────────────── */
  .notif-wrap { position: relative; }
  .notif-panel {
    position: absolute;
    top: calc(100% + var(--space-sm));
    right: 0;
    width: min(340px, calc(100vw - 2 * var(--space-lg)));
    background: var(--canvas);
    border: var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    z-index: var(--z-nav);
    overflow: hidden;
  }
  .notif-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-sm);
    padding: var(--space-sm) var(--space-md);
    border-bottom: var(--border-thin);
  }
  .notif-clear {
    display: inline-flex;
    align-items: center;
    gap: var(--space-xs);
    font-size: 13px;
  }
  .notif-clear:disabled { opacity: 0.45; cursor: not-allowed; }
  .notif-empty { padding: var(--space-md); margin: 0; }
  .notif-list {
    list-style: none;
    margin: 0;
    padding: 0;
    max-height: 320px;
    overflow-y: auto;
  }
  .notif-item {
    display: flex;
    align-items: flex-start;
    gap: var(--space-sm);
    width: 100%;
    padding: var(--space-sm) var(--space-md);
    background: transparent;
    border: 0;
    border-bottom: var(--border-thin);
    text-align: left;
    font-family: inherit;
    color: var(--ink);
    cursor: pointer;
  }
  .notif-item:hover { background: var(--canvas-soft); }
  .notif-unread { background: var(--primary-soft); }
  .notif-unread:hover { background: var(--primary-soft); }
  .notif-dot {
    flex-shrink: 0;
    width: 8px;
    height: 8px;
    margin-top: 8px;
    border-radius: 50%;
    background: var(--mute);
  }
  .notif-unread .notif-dot { background: var(--primary); }
  .notif-body {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .notif-body .body-sm {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .notif-more {
    display: block;
    padding: var(--space-sm) var(--space-md);
    text-align: center;
  }
  .notif-more:hover { background: var(--canvas-soft); }
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
<script>
  import { page } from '$app/state';
  import { userStore, clearSession } from '$lib/stores.js';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import {
    LayoutDashboard, Package, ShoppingBag, FileText, Ticket, LogOut, Printer, Store,
  } from 'lucide-svelte';

  let { children } = $props();

  onMount(() => {
    if (!$userStore || $userStore.role !== 'admin') {
      goto('/masuk');
    }
  });

  const links = [
    { href: '/admin', label: 'Dashboard', icon: LayoutDashboard, exact: true },
    { href: '/admin/produk', label: 'Produk', icon: Package },
    { href: '/admin/pesanan', label: 'Pesanan', icon: ShoppingBag },
    { href: '/admin/print-jobs', label: 'Print Jobs', icon: FileText },
    { href: '/admin/vouchers', label: 'Voucher', icon: Ticket },
  ];

  function isCurrent(l) {
    return l.exact ? page.url.pathname === l.href : page.url.pathname.startsWith(l.href);
  }
</script>

<div class="admin-shell">
  <aside class="sidebar">
    <a href="/admin" class="brand">
      <span class="brand-mark" aria-hidden="true"><Printer size={18} strokeWidth={2.4} /></span>
      <span class="brand-name">PrintKu</span>
      <span class="brand-tag">Admin</span>
    </a>

    <nav class="side-nav" aria-label="Navigasi admin">
      {#each links as l (l.href)}
        <a href={l.href} class="side-link" aria-current={isCurrent(l) ? 'page' : undefined}>
          <l.icon size={18} aria-hidden="true" />
          {l.label}
        </a>
      {/each}
    </nav>

    <div class="side-bottom">
      <a href="/" class="side-link">
        <Store size={18} aria-hidden="true" /> Lihat toko
      </a>
      <div class="side-user">
        <span class="caption">Masuk sebagai</span>
        <span class="body-sm-strong">{$userStore?.name ?? '—'}</span>
      </div>
      <button class="side-link" type="button" onclick={() => { clearSession(); goto('/'); }}>
        <LogOut size={18} aria-hidden="true" /> Keluar
      </button>
    </div>
  </aside>

  <main class="admin-main">{@render children()}</main>
</div>

<style>
  .admin-shell {
    display: grid;
    grid-template-columns: 264px 1fr;
    min-height: calc(100vh - 100px);
  }
  .sidebar {
    background: var(--canvas-soft);
    border-right: var(--border);
    padding: var(--space-xl) 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-xl);
    position: sticky;
    top: 0;
    height: 100vh;
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: var(--space-sm);
    padding: 0 var(--space-xl);
  }
  .brand-mark {
    width: 36px;
    height: 36px;
    border-radius: var(--radius-sm);
    background: var(--primary);
    color: var(--on-primary);
    border: var(--border);
    box-shadow: var(--shadow-xs);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .brand-name { font-family: var(--font-display); font-size: 20px; font-weight: 700; }
  .brand-tag {
    margin-left: auto;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 1px;
    text-transform: uppercase;
    background: var(--ink);
    color: var(--on-primary);
    border-radius: var(--radius-pill);
    padding: 2px var(--space-sm);
  }
  .side-nav {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
    padding: 0 var(--space-lg);
  }
  .side-bottom {
    margin-top: auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    padding: 0 var(--space-lg) var(--space-sm);
  }
  .side-user {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--space-sm) var(--space-lg);
    background: var(--canvas);
    border: var(--border-thin);
    border-radius: var(--radius-md);
  }
  .admin-main { padding: var(--space-3xl) var(--space-2xl); min-width: 0; }

  @media (max-width: 767px) {
    .admin-shell { grid-template-columns: 1fr; }
    .sidebar {
      position: static;
      height: auto;
      flex-direction: row;
      align-items: center;
      gap: var(--space-md);
      border-right: none;
      border-bottom: var(--border);
      overflow-x: auto;
      padding: var(--space-md);
    }
    .side-nav { flex-direction: row; padding: 0; }
    .side-link { white-space: nowrap; }
    .brand-tag, .side-bottom { display: none; }
    .admin-main { padding: var(--space-xl) var(--space-lg); }
  }
</style>

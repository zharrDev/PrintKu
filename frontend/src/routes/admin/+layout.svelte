<script>
  import { page } from '$app/state';
  import { userStore, clearSession } from '$lib/stores.js';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { LayoutDashboard, Package, ShoppingBag, FileText, Ticket, LogOut } from 'lucide-svelte';

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
</script>

<div class="admin-shell">
  <aside class="sidebar">
    <a href="/admin" class="brand">
      <span class="brand-mark">P</span>
      <span class="brand-name">PrintMart</span>
    </a>
    <nav class="side-nav">
      {#each links as l (l.href)}
        <a
          href={l.href}
          class="side-link"
          class:side-active={l.exact ? page.url.pathname === l.href : page.url.pathname.startsWith(l.href)}
        >
          <l.icon size={18} />
          {l.label}
        </a>
      {/each}
    </nav>
    <div class="side-bottom">
      <span class="caption text-body-mid" style="padding: 0 var(--space-lg)">
        {$userStore?.name}
      </span>
      <button class="side-link" onclick={() => { clearSession(); goto('/'); }}>
        <LogOut size={18} /> Keluar
      </button>
    </div>
  </aside>

  <main class="admin-main">
    {@render children()}
  </main>
</div>

<style>
  .admin-shell {
    display: grid;
    grid-template-columns: 240px 1fr;
    min-height: calc(100vh - 100px);
  }
  .sidebar {
    background: var(--canvas-soft);
    border-right: 1px solid var(--mute);
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
  .brand-name { font-size: 18px; font-weight: 700; }
  .side-nav { display: flex; flex-direction: column; gap: var(--space-xs); padding: 0 var(--space-md); }
  .side-link {
    display: flex;
    align-items: center;
    gap: var(--space-md);
    padding: var(--space-md) var(--space-lg);
    border-radius: var(--radius-sm);
    color: var(--ink);
    font-size: 16px;
    line-height: 24px;
    background: var(--canvas);
    border: 1px solid transparent;
  }
  .side-link:hover { background: var(--canvas); border-color: var(--mute); }
  .side-active {
    background: var(--canvas);
    border-left: 3px solid var(--primary);
    font-weight: 600;
  }
  .side-bottom {
    margin-top: auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }
  .admin-main { padding: var(--space-3xl) var(--space-2xl); }
  @media (max-width: 767px) {
    .admin-shell { grid-template-columns: 1fr; }
    .sidebar {
      position: static;
      height: auto;
      flex-direction: row;
      align-items: center;
      overflow-x: auto;
      padding: var(--space-md);
    }
    .side-nav { flex-direction: row; padding: 0; }
    .side-link { white-space: nowrap; }
    .side-bottom { display: none; }
    .admin-main { padding: var(--space-xl) var(--space-lg); }
  }
</style>
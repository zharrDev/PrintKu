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
      <span class="brand-name">PrintKu</span>
    </a>
    <nav class="side-nav">
      {#each links as l (l.href)}
        <a
          href={l.href}
          class="side-link"
          aria-current={l.exact ? page.url.pathname === l.href : page.url.pathname.startsWith(l.href) ? 'page' : undefined}
        >
          <l.icon size={18} />
          {l.label}
        </a>
      {/each}
    </nav>
    <div class="side-bottom">
      <span class="caption side-user">{$userStore?.name}</span>
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
    grid-template-columns: 260px 1fr;
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
    box-shadow: var(--shadow-sm);
    font-weight: 700;
    font-size: 20px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  .brand-name { font-size: 20px; font-weight: 700; }
  .side-nav {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    padding: 0 var(--space-lg);
  }
  .side-bottom {
    margin-top: auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }
  .side-user {
    padding: 0 var(--space-lg);
    font-weight: 600;
    color: var(--ink);
  }
  .admin-main { padding: var(--space-3xl) var(--space-2xl); }
  @media (max-width: 767px) {
    .admin-shell { grid-template-columns: 1fr; }
    .sidebar {
      position: static;
      height: auto;
      flex-direction: row;
      align-items: center;
      border-right: none;
      border-bottom: var(--border);
      overflow-x: auto;
      padding: var(--space-md);
    }
    .side-nav { flex-direction: row; padding: 0; }
    .side-link { white-space: nowrap; }
    .side-bottom { display: none; }
    .admin-main { padding: var(--space-xl) var(--space-lg); }
  }
</style>
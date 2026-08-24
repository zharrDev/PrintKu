<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import { formatRupiah, formatDate, shortId, ORDER_STATUS } from '$lib/format.js';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { TrendingUp, ShoppingBag, Package, Users, FileText } from 'lucide-svelte';
  import { goto } from '$app/navigation';

  let stats = $state(null);
  let loading = $state(true);

  onMount(load);

  async function load() {
    const r = await api('/admin/stats');
    loading = false;
    if (r.ok) stats = r.json;
  }

  const cards = [
    { key: 'totalRevenue', label: 'Omzet terbayar', icon: TrendingUp, fmt: formatRupiah },
    { key: 'orderCounts.total', label: 'Total Pesanan', icon: ShoppingBag, fmt: (v) => v ?? 0 },
    { key: 'productCount', label: 'Produk Aktif', icon: Package, fmt: (v) => v ?? 0 },
    { key: 'userCount', label: 'Pelanggan', icon: Users, fmt: (v) => v ?? 0 },
  ];
</script>

<svelte:head><title>Dashboard Admin — PrintKu</title></svelte:head>

<div class="stack-xl">
  <div class="stack-sm">
    <span class="eyebrow">Admin</span>
    <h1 class="display-md">Dashboard</h1>
  </div>

  {#if loading}
    <div class="empty-state"><span class="spinner"></span> Memuat…</div>
  {:else if stats}
    <div class="grid grid-4">
      {#each cards as c (c.label)}
        <div class="card stack-sm">
          <c.icon size={20} style="color: var(--primary)" />
          <span class="caption text-body-mid">{c.label}</span>
          <span class="display-sub-sm">
            {#if c.key === 'totalRevenue'}
              {c.fmt(stats.stats.totalRevenue)}
            {:else if c.key === 'orderCounts.total'}
              {c.fmt(stats.stats.orderCounts?.total)}
            {:else}
              {c.fmt(stats.stats[c.key])}
            {/if}
          </span>
        </div>
      {/each}
    </div>

    <div class="grid-2" style="display: grid; gap: var(--space-xl)">
      <div class="pricing-card stack-md">
        <div class="row">
          <FileText size={18} />
          <h3 class="display-sub-sm">Antrian print</h3>
        </div>
        <div class="row-between">
          <span class="body-sm text-body">Menunggu</span>
          <span class="body-md-strong">{stats.stats.printJobs?.menunggu ?? 0}</span>
        </div>
        <div class="row-between">
          <span class="body-sm text-body">Sedang diproses</span>
          <span class="body-md-strong">{stats.stats.printJobs?.diproses ?? 0}</span>
        </div>
        <div class="row-between">
          <span class="body-sm text-body">Pesanan menunggu bayar</span>
          <span class="body-md-strong">{stats.stats.orderCounts?.pending ?? 0}</span>
        </div>
      </div>

      <div class="pricing-card stack-md">
        <div class="row">
          <TrendingUp size={18} />
          <h3 class="display-sub-sm">Pendapatan 7 hari terakhir</h3>
        </div>
        {#if stats.revenueByDay.length === 0}
          <span class="caption text-body-mid">Belum ada transaksi sukses.</span>
        {:else}
          {#each stats.revenueByDay as d (d.day)}
            <div class="row-between">
              <span class="body-sm text-body">{formatDate(`${d.day}T00:00:00`)}</span>
              <span class="body-md-strong">{formatRupiah(d.total)}</span>
            </div>
          {/each}
        {/if}
      </div>
    </div>

    <div class="stack-md">
      <h3 class="display-sub-sm">Pesanan terbaru</h3>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Pelanggan</th>
              <th>Tipe</th>
              <th>Total</th>
              <th>Status</th>
              <th>Waktu</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {#each stats.recentOrders as o (o.id)}
              <tr>
                <td class="body-sm-strong">#{shortId(o.id)}</td>
                <td>{o.user_name}</td>
                <td>{o.order_type === 'print' ? 'Print' : 'Produk'}</td>
                <td>{formatRupiah(o.total_price)}</td>
                <td><StatusBadge map={ORDER_STATUS} status={o.status} tone={['diproses', 'menunggu_pembayaran'].includes(o.status) ? 'primary' : o.status === 'dibatalkan' ? 'dark' : 'cream'} /></td>
                <td class="caption text-body-mid">{formatDate(o.created_at)}</td>
                <td><button class="btn-text" onclick={() => goto('/admin/pesanan')}>Kelola</button></td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>
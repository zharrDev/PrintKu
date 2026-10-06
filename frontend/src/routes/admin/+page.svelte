<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import { formatRupiah, formatDate, shortId, ORDER_STATUS } from '$lib/format.js';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import {
    TrendingUp, ShoppingBag, Package, Users, FileText, ArrowRight, Clock, Printer,
  } from 'lucide-svelte';

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

  function cardValue(key) {
    if (!stats) return '—';
    const s = stats.stats ?? {};
    if (key === 'totalRevenue') return s.totalRevenue;
    if (key === 'orderCounts.total') return s.orderCounts?.total;
    return s[key];
  }
</script>

<svelte:head><title>Dashboard Admin — PrintKu</title></svelte:head>

<div class="stack-xl">
  <div class="row-between wrap">
    <div class="stack-sm">
      <span class="eyebrow eyebrow-bar">Admin</span>
      <h1 class="display-md">Dashboard</h1>
    </div>
    <div class="row wrap">
      <a href="/admin/produk" class="btn btn-tertiary btn-sm"><Package size={16} aria-hidden="true" /> Produk</a>
      <a href="/admin/print-jobs" class="btn btn-primary btn-sm"><Printer size={16} aria-hidden="true" /> Antrian print</a>
    </div>
  </div>

  {#if loading}
    <div class="empty-state" role="status"><span class="spinner"></span> Memuat…</div>
  {:else if stats}
    <div class="stats">
      {#each cards as c (c.label)}
        <div class="stat">
          <span class="stat-icon" aria-hidden="true"><c.icon size={20} /></span>
          <span class="caption text-body-mid">{c.label}</span>
          <span class="stat-value">{c.fmt(cardValue(c.key))}</span>
        </div>
      {/each}
    </div>

    <div class="grid grid-2 panes">
      <section class="pricing-card stack-md" aria-labelledby="judul-antrian">
        <div class="row">
          <FileText size={18} aria-hidden="true" />
          <h2 class="display-sub-sm" id="judul-antrian">Antrian print</h2>
        </div>
        <div class="row-between">
          <span class="body-sm text-body">Menunggu</span>
          <span class="body-md-strong">{stats.stats?.printJobs?.menunggu ?? 0}</span>
        </div>
        <div class="row-between">
          <span class="body-sm text-body">Sedang diproses</span>
          <span class="body-md-strong">{stats.stats?.printJobs?.diproses ?? 0}</span>
        </div>
        <div class="row-between">
          <span class="body-sm text-body">Pesanan menunggu bayar</span>
          <span class="body-md-strong">{stats.stats?.orderCounts?.pending ?? 0}</span>
        </div>
        <a href="/admin/print-jobs" class="btn btn-tertiary btn-sm" style="align-self: flex-start">
          Buka antrian <ArrowRight size={16} aria-hidden="true" />
        </a>
      </section>

      <section class="pricing-card stack-md" aria-labelledby="judul-pendapatan">
        <div class="row">
          <TrendingUp size={18} aria-hidden="true" />
          <h2 class="display-sub-sm" id="judul-pendapatan">Pendapatan 7 hari terakhir</h2>
        </div>
        {#if (stats.revenueByDay ?? []).length === 0}
          <p class="body-sm text-body-mid">Belum ada transaksi sukses.</p>
        {:else}
          {@const peak = Math.max(...stats.revenueByDay.map((d) => Number(d.total) || 0), 1)}
          <ul class="bars">
            {#each stats.revenueByDay as d (d.day)}
              <li class="bar-row">
                <span class="caption bar-day">{formatDate(`${d.day}T00:00:00`)}</span>
                <span class="bar-track">
                  <span
                    class="bar-fill"
                    style={`width: ${Math.max(6, Math.round((Number(d.total) / peak) * 100))}%`}
                  ></span>
                </span>
                <span class="body-sm-strong bar-total">{formatRupiah(d.total)}</span>
              </li>
            {/each}
          </ul>
        {/if}
      </section>
    </div>

    <section class="stack-md" aria-labelledby="judul-terbaru">
      <div class="row-between wrap">
        <h2 class="display-sub-sm" id="judul-terbaru">Pesanan terbaru</h2>
        <a href="/admin/pesanan" class="btn-text">Lihat semua <ArrowRight size={14} aria-hidden="true" /></a>
      </div>

      {#if (stats.recentOrders ?? []).length === 0}
        <div class="empty-state stack-sm">
          <Clock size={28} strokeWidth={1.6} aria-hidden="true" />
          <span class="body-md-strong">Belum ada pesanan</span>
          <span class="body-sm text-body">Pesanan baru akan muncul di sini.</span>
        </div>
      {:else}
        <div class="table-wrap">
          <table class="data-table">
            <caption class="sr-only">Daftar pesanan terbaru</caption>
            <thead>
              <tr>
                <th scope="col">ID</th>
                <th scope="col">Pelanggan</th>
                <th scope="col">Tipe</th>
                <th scope="col">Total</th>
                <th scope="col">Status</th>
                <th scope="col">Waktu</th>
                <th scope="col"><span class="sr-only">Aksi</span></th>
              </tr>
            </thead>
            <tbody>
              {#each stats.recentOrders as o (o.id)}
                <tr>
                  <td class="body-sm-strong">#{shortId(o.id)}</td>
                  <td>{o.user_name}</td>
                  <td>{o.order_type === 'print' ? 'Print' : 'Produk'}</td>
                  <td>{formatRupiah(o.total_price)}</td>
                  <td>
                    <StatusBadge
                      map={ORDER_STATUS}
                      status={o.status}
                      tone={['diproses', 'menunggu_pembayaran'].includes(o.status) ? 'primary' : o.status === 'dibatalkan' ? 'dark' : 'cream'}
                    />
                  </td>
                  <td class="caption text-body-mid">{formatDate(o.created_at)}</td>
                  <td class="cell-action">
                    <a href="/admin/pesanan" class="btn-text">Kelola</a>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </section>
  {/if}
</div>

<style>
  .panes { align-items: start; }
  .stat-icon {
    width: 40px;
    height: 40px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--primary-soft);
    border: var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-xs);
    margin-bottom: var(--space-xs);
  }
  /* Grafik batang sederhana tanpa dependensi. */
  .bars { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: var(--space-sm); }
  .bar-row {
    display: grid;
    grid-template-columns: minmax(80px, auto) 1fr auto;
    align-items: center;
    gap: var(--space-md);
  }
  .bar-track {
    display: block;
    height: 12px;
    background: var(--canvas-soft);
    border: 1px solid var(--mute);
    border-radius: var(--radius-pill);
    overflow: hidden;
  }
  .bar-fill {
    display: block;
    height: 100%;
    background: var(--primary);
    border-right: 1px solid var(--ink);
  }
  .bar-total { text-align: right; white-space: nowrap; }
  .cell-action { text-align: right; }

  @media (max-width: 767px) {
    .bar-row { grid-template-columns: 1fr auto; }
    .bar-day { grid-column: 1 / -1; }
  }
</style>

<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import { formatRupiah, formatDate, shortId, ORDER_STATUS } from '$lib/format.js';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { toast } from '$lib/stores.js';

  let orders = $state([]);
  let loading = $state(true);
  let filter = $state('');

  const FLOW = [
    { value: 'menunggu_pembayaran', label: 'Menunggu Pembayaran' },
    { value: 'diproses', label: 'Diproses' },
    { value: 'siap', label: 'Siap' },
    { value: 'selesai', label: 'Selesai' },
    { value: 'dibatalkan', label: 'Dibatalkan' },
  ];

  onMount(load);

  async function load() {
    loading = true;
    const params = filter ? `?status=${filter}` : '';
    const r = await api(`/admin/orders${params}`);
    loading = false;
    if (r.ok) orders = r.json.orders;
  }

  async function setStatus(o, status) {
    const r = await api(`/orders/${o.id}/status`, { method: 'PATCH', body: { status } });
    if (r.ok) {
      toast(`Pesanan #${shortId(o.id)} → ${FLOW.find((f) => f.value === status)?.label}`, 'success');
      load();
    }
  }

  function nextStatus(o) {
    if (o.status === 'menunggu_pembayaran') return 'diproses';
    if (o.status === 'diproses') return 'siap';
    if (o.status === 'siap') return 'selesai';
    return null;
  }
</script>

<svelte:head><title>Pesanan — Admin PrintMart</title></svelte:head>

<div class="stack-xl">
  <div class="stack-sm">
    <span class="eyebrow">Admin</span>
    <h1 class="display-md">Kelola Pesanan</h1>
  </div>

  <div class="row wrap">
    <button class="chip" class:chip-active={!filter} onclick={() => { filter = ''; load(); }}>Semua</button>
    {#each FLOW as f (f.value)}
      <button class="chip" class:chip-active={filter === f.value} onclick={() => { filter = f.value; load(); }}>{f.label}</button>
    {/each}
  </div>

  {#if loading}
    <div class="empty-state"><span class="spinner"></span> Memuat…</div>
  {:else if orders.length === 0}
    <div class="empty-state body-md">Tidak ada pesanan dengan filter ini.</div>
  {:else}
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Pelanggan</th>
            <th>Tipe</th>
            <th>Total</th>
            <th>Status</th>
            <th>Pembayaran</th>
            <th>Waktu</th>
            <th style="text-align: right">Aksi</th>
          </tr>
        </thead>
        <tbody>
          {#each orders as o (o.id)}
            <tr>
              <td class="body-sm-strong">#{shortId(o.id)}</td>
              <td>{o.user_name}</td>
              <td>{o.order_type === 'print' ? 'Print' : 'Produk'}</td>
              <td class="body-sm-strong">{formatRupiah(o.total_price)}</td>
              <td><StatusBadge map={ORDER_STATUS} status={o.status} tone={['diproses', 'menunggu_pembayaran'].includes(o.status) ? 'primary' : o.status === 'dibatalkan' ? 'dark' : 'cream'} /></td>
              <td class="caption">{o.payment_status || '-'}</td>
              <td class="caption text-body-mid">{formatDate(o.created_at)}</td>
              <td>
                <div class="row" style="justify-content: flex-end">
                  {#if o.status === 'menunggu_pembayaran'}
                    <select class="input mini-select" onchange={(e) => setStatus(o, e.currentTarget.value)}>
                      <option value="">Aksi…</option>
                      <option value="diproses">Tandai Dibayar / Proses</option>
                      <option value="dibatalkan">Batalkan</option>
                    </select>
                  {:else if nextStatus(o)}
                    <button class="btn btn-primary btn-sm" onclick={() => setStatus(o, nextStatus(o))}>
                      {o.status === 'diproses' ? 'Tandai Siap' : 'Tandai Selesai'}
                    </button>
                  {/if}
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .chip {
    border: 1px solid var(--mute);
    background: var(--canvas);
    color: var(--ink);
    border-radius: var(--radius-pill);
    padding: var(--space-xs) var(--space-lg);
    font-size: 15px;
  }
  .chip-active { background: var(--ink); color: var(--on-primary); border-color: var(--ink); }
  .mini-select { width: auto; padding: var(--space-xs) var(--space-xl) var(--space-xs) var(--space-md); font-size: 15px; min-height: 0; }
</style>
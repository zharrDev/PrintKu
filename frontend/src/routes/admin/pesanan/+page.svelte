<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import { formatRupiah, formatDate, shortId, ORDER_STATUS } from '$lib/format.js';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { toast } from '$lib/stores.js';
  import { ShoppingBag } from 'lucide-svelte';

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

<svelte:head><title>Pesanan — Admin PrintKu</title></svelte:head>

<div class="stack-xl">
  <div class="stack-sm">
    <span class="eyebrow eyebrow-bar">Admin</span>
    <h1 class="display-md">Kelola Pesanan</h1>
  </div>

  <div class="row wrap" role="group" aria-label="Filter status pesanan">
    <button class="chip" type="button" class:chip-active={!filter} aria-pressed={!filter} onclick={() => { filter = ''; load(); }}>
      Semua
    </button>
    {#each FLOW as f (f.value)}
      <button
        class="chip"
        type="button"
        class:chip-active={filter === f.value}
        aria-pressed={filter === f.value}
        onclick={() => { filter = f.value; load(); }}
      >{f.label}</button>
    {/each}
  </div>

  {#if loading}
    <div class="empty-state" role="status"><span class="spinner"></span> Memuat…</div>
  {:else if orders.length === 0}
    <div class="empty-state stack-sm">
      <ShoppingBag size={28} strokeWidth={1.6} aria-hidden="true" />
      <span class="body-md-strong">Tidak ada pesanan</span>
      <span class="body-sm text-body">Coba ubah filter status di atas.</span>
    </div>
  {:else}
    <div class="table-wrap">
      <table class="data-table">
        <caption class="sr-only">Daftar pesanan pelanggan</caption>
        <thead>
          <tr>
            <th scope="col">ID</th>
            <th scope="col">Pelanggan</th>
            <th scope="col">Tipe</th>
            <th scope="col">Total</th>
            <th scope="col">Status</th>
            <th scope="col">Pembayaran</th>
            <th scope="col">Waktu</th>
            <th scope="col"><span class="sr-only">Aksi</span></th>
          </tr>
        </thead>
        <tbody>
          {#each orders as o (o.id)}
            <tr>
              <td class="body-sm-strong">#{shortId(o.id)}</td>
              <td>{o.user_name}</td>
              <td>{o.order_type === 'print' ? 'Print' : 'Produk'}</td>
              <td class="body-sm-strong">{formatRupiah(o.total_price)}</td>
              <td>
                <StatusBadge
                  map={ORDER_STATUS}
                  status={o.status}
                  tone={['diproses', 'menunggu_pembayaran'].includes(o.status) ? 'primary' : o.status === 'dibatalkan' ? 'dark' : 'cream'}
                />
              </td>
              <td class="caption">{o.payment_status || '-'}</td>
              <td class="caption text-body-mid">{formatDate(o.created_at)}</td>
              <td class="cell-action">
                <div class="row" style="justify-content: flex-end">
                  {#if o.status === 'menunggu_pembayaran'}
                    <label class="sr-only" for={`aksi-${o.id}`}>Aksi untuk pesanan #{shortId(o.id)}</label>
                    <select
                      id={`aksi-${o.id}`}
                      class="input mini-select"
                      value=""
                      onchange={(e) => { if (e.currentTarget.value) setStatus(o, e.currentTarget.value); }}
                    >
                      <option value="">Aksi…</option>
                      <option value="diproses">Tandai Dibayar / Proses</option>
                      <option value="dibatalkan">Batalkan</option>
                    </select>
                  {:else if nextStatus(o)}
                    <button class="btn btn-primary btn-sm" type="button" onclick={() => setStatus(o, nextStatus(o))}>
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
  .mini-select {
    width: auto;
    padding: var(--space-xs) var(--space-xl) var(--space-xs) var(--space-md);
    font-size: 15px;
    min-height: 0;
  }
  .cell-action { text-align: right; }
</style>

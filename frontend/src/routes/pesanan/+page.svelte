<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api.js';
  import { userStore } from '$lib/stores.js';
import { formatRupiah, formatDate, shortId, ORDER_STATUS, PRINT_STATUS, statusInfo } from '$lib/format.js';
import StatusBadge from '$lib/components/StatusBadge.svelte';
import { PackageOpen, FileText, ArrowRight } from 'lucide-svelte';

  let orders = $state([]);
  let loading = $state(true);

  onMount(async () => {
    if (!$userStore) {
      goto('/masuk');
      return;
    }
    const r = await api('/orders');
    loading = false;
    if (r.ok) orders = r.json.orders;
  });

  const TYPE_LABEL = { product: 'Produk', print: 'Jasa Print' };
</script>

<svelte:head><title>Pesanan Saya — PrintKu</title></svelte:head>

<section class="band-hero">
  <div class="container stack-xl">
    <div class="stack-sm">
      <span class="eyebrow">Pesanan</span>
      <h1 class="display-lg">Riwayat pesanan</h1>
    </div>

    {#if loading}
      <div class="empty-state"><span class="spinner"></span> Memuat…</div>
    {:else if orders.length === 0}
      <div class="empty-state stack-md" style="align-items: center">
        <PackageOpen size={32} strokeWidth={1.5} />
        <span class="body-md-strong">Belum ada pesanan</span>
        <span class="body-sm text-body-mid">Mulai dari belanja ATK atau upload file untuk dicetak.</span>
        <div class="row">
          <a href="/katalog" class="btn btn-primary">Belanja Produk</a>
          <a href="/print" class="btn btn-tertiary">Upload File</a>
        </div>
      </div>
    {:else}
      <div class="stack-md">
        {#each orders as o (o.id)}
          <div class="card order-row">
            <div class="row-between wrap">
              <div class="stack-sm">
                <div class="row wrap">
                  <span class="badge badge-dark">#{shortId(o.id)}</span>
                  <StatusBadge map={ORDER_STATUS} status={o.status} tone={['diproses', 'menunggu_pembayaran'].includes(o.status) ? 'primary' : o.status === 'dibatalkan' ? 'dark' : 'cream'} />
                  <span class="badge">{TYPE_LABEL[o.order_type]}</span>
                </div>
                <span class="caption text-body-mid">{formatDate(o.created_at)} · {o.delivery_method === 'kirim' ? 'Diantar' : 'Ambil di toko'}</span>
              </div>
              <div class="stack-sm" style="align-items: flex-end">
                <span class="body-md-strong">{formatRupiah(o.total_price)}</span>
                <a href={`/pesanan/${o.id}`} class="btn btn-tertiary btn-sm">Lacak Pesanan <ArrowRight size={14} /></a>
              </div>
            </div>
            {#if o.order_type === 'print' && o.printJobs?.length}
              <hr class="divider" style="margin: var(--space-md) 0" />
              <div class="row wrap">
                {#each o.printJobs as j (j.id)}
                  <span class="badge"><FileText size={13} /> {j.file_name} · {statusInfo(PRINT_STATUS, j.status).label}</span>
                {/each}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</section>

<style>
  .order-row { transition: transform var(--dur) var(--ease), box-shadow var(--dur) var(--ease); }
.order-row:hover { transform: translate(-2px, -2px); box-shadow: var(--shadow-lg); }
</style>
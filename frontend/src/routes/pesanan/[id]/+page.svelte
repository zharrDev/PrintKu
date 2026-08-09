<script>
  import { onMount, onDestroy } from 'svelte';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api.js';
  import { userStore, toast } from '$lib/stores.js';
  import {
    formatRupiah, formatDate, shortId, ORDER_STATUS, PRINT_STATUS,
    PAPER_SIZE_NAME, COLOR_MODE_NAME,
  } from '$lib/format.js';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { ArrowLeft, FileText, Bell, CreditCard, Clock } from 'lucide-svelte';

  let order = $state(null);
  let liveEvents = $state([]);
  let ws = null;
  let loading = $state(true);

  const STEP_LABELS = ['Menunggu Pembayaran', 'Diproses', 'Siap Diambil / Dikirim', 'Selesai'];

  onMount(async () => {
    if (!$userStore) {
      goto('/masuk');
      return;
    }
    await load();
    if (order) connectWs();
  });

  onDestroy(() => {
    ws?.close();
  });

  async function load() {
    const r = await api(`/orders/${page.params.id}`);
    loading = false;
    if (r.ok) order = r.json.order;
  }

  function connectWs() {
    try {
      ws = new WebSocket(`ws://${location.host}/ws/orders/${$userStore.id}`);
      ws.onmessage = (e) => {
        try {
          const msg = JSON.parse(e.data);
          liveEvents = [{ ...msg, at: new Date() }, ...liveEvents].slice(0, 6);
          if (msg.orderId === order.id) load();
          toast(msg.message || 'Status pesanan diperbarui');
        } catch { /* ignore */ }
      };
      ws.onclose = () => {
        setTimeout(() => {
          if (order && !['selesai', 'dibatalkan'].includes(order.status)) connectWs();
        }, 3000);
      };
    } catch { /* ignore */ }
  }

  function orderStep(o) {
    const s = ORDER_STATUS[o.status];
    return s ? s.step : -1;
  }
</script>

<svelte:head><title>Lacak Pesanan — PrintMart</title></svelte:head>

<section class="band-hero">
  <div class="container stack-xl">
    {#if loading}
      <div class="empty-state"><span class="spinner"></span> Memuat…</div>
    {:else if order}
      <a href="/pesanan" class="btn-text" style="align-self: flex-start"><ArrowLeft size={16} /> Semua pesanan</a>

      <div class="row-between wrap">
        <div class="stack-sm">
          <div class="row wrap">
            <h1 class="display-md">Pesanan #{shortId(order.id)}</h1>
            <StatusBadge
              map={ORDER_STATUS}
              status={order.status}
              tone={['diproses', 'menunggu_pembayaran'].includes(order.status) ? 'primary' : order.status === 'dibatalkan' ? 'dark' : 'cream'}
            />
          </div>
          <span class="body-sm text-body-mid">{formatDate(order.created_at)} · {order.delivery_method === 'kirim' ? 'Diantar' : 'Ambil di toko'}</span>
        </div>
        {#if order.status === 'menunggu_pembayaran' && order.payment?.status !== 'success'}
          <a href={`/bayar/${order.id}`} class="btn btn-primary">Bayar Sekarang</a>
        {/if}
      </div>

      <!-- Timeline -->
      <div class="card stack-md">
        <h3 class="display-sub-sm">Lacak status</h3>
        {#if order.status === 'dibatalkan'}
          <div class="empty-state">Pesanan ini dibatalkan. Hubungi toko jika ada pertanyaan.</div>
        {:else}
          <div class="timeline">
            {#each STEP_LABELS as label, i (label)}
              <div class="step" class:done={orderStep(order) >= i} class:current={orderStep(order) === i}>
                <div class="step-dot"></div>
                <span class="caption step-label">{label}</span>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Live events -->
      {#if liveEvents.length}
        <div class="card stack-sm">
          <div class="row"><Bell size={18} /><h3 class="display-sub-sm">Update real-time</h3></div>
          {#each liveEvents as ev (ev.at)}
            <div class="row-between">
              <span class="body-sm">{ev.message}</span>
              <span class="caption text-body-mid">{formatDate(ev.at.toISOString())}</span>
            </div>
          {/each}
        </div>
      {/if}

      <div class="track-grid">
        <!-- Items / print jobs -->
        <div class="card stack-md">
          <h3 class="display-sub-sm">{order.order_type === 'print' ? 'Detail print' : 'Detail produk'}</h3>
          {#if order.order_type === 'print'}
            {#each order.printJobs as j (j.id)}
              <div class="job-row stack-sm">
                <div class="row">
                  <FileText size={18} />
                  <span class="body-md-strong">{j.file_name}</span>
                  <StatusBadge map={PRINT_STATUS} status={j.status} tone={['diproses', 'menunggu'].includes(j.status) ? 'primary' : j.status === 'dibatalkan' ? 'dark' : 'cream'} />
                </div>
                <div class="row wrap">
                  <span class="badge">{j.page_count} halaman</span>
                  <span class="badge">{COLOR_MODE_NAME[j.color_mode]}</span>
                  <span class="badge">{PAPER_SIZE_NAME[j.paper_size]}</span>
                  <span class="badge">{j.copies}x rangkap</span>
                  {#if j.duplex}<span class="badge">Dua sisi</span>{/if}
                  {#if j.finishing_name}<span class="badge">{j.finishing_name}</span>{/if}
                </div>
                <span class="body-md-strong">{formatRupiah(j.estimated_price)}</span>
                <div class="prog"><div class="prog-bar" style="width: {((PRINT_STATUS[j.status]?.step ?? 0) / 3) * 100}%"></div></div>
              </div>
            {/each}
          {:else}
            {#each order.items as it (it.id)}
              <div class="row-between">
                <span class="body-sm">{it.product_name} × {it.quantity}</span>
                <span class="body-sm">{formatRupiah(it.subtotal)}</span>
              </div>
            {/each}
          {/if}
          <hr class="divider" />
          {#if order.discount > 0}
            <div class="row-between">
              <span class="body-sm text-body">Subtotal</span>
              <span class="body-sm">{formatRupiah(order.subtotal)}</span>
            </div>
            <div class="row-between">
              <span class="body-sm text-body">Diskon{order.voucher_code ? ` (${order.voucher_code})` : ''}</span>
              <span class="body-sm" style="color: var(--primary)">−{formatRupiah(order.discount)}</span>
            </div>
          {/if}
          <div class="row-between">
            <span class="body-md-strong">Total</span>
            <span class="display-sub-sm" style="color: var(--primary)">{formatRupiah(order.total_price)}</span>
          </div>
        </div>

        <!-- Info side -->
        <div class="stack-md">
          <div class="pricing-card stack-sm">
            <div class="row"><CreditCard size={18} /><h3 class="display-sub-sm">Pembayaran</h3></div>
            {#if order.payment}
              <div class="row-between">
                <span class="caption text-body-mid">Provider</span>
                <span class="body-sm">{order.payment.provider}</span>
              </div>
              <div class="row-between">
                <span class="caption text-body-mid">Status</span>
                <StatusBadge
                  map={{ pending: { label: 'Pending' }, success: { label: 'Berhasil' }, failed: { label: 'Gagal' } }}
                  status={order.payment.status}
                  tone={order.payment.status === 'success' ? 'dark' : order.payment.status === 'pending' ? 'primary' : 'cream'}
                />
              </div>
              {#if order.payment.va_number}
                <div class="row-between">
                  <span class="caption text-body-mid">VA</span>
                  <span class="body-sm">{order.payment.va_number}</span>
                </div>
              {/if}
              {#if order.payment.paid_at}
                <div class="row-between">
                  <span class="caption text-body-mid">Dibayar</span>
                  <span class="body-sm">{formatDate(order.payment.paid_at)}</span>
                </div>
              {/if}
            {:else}
              <span class="caption text-body-mid">Belum ada transaksi.</span>
            {/if}
          </div>

          <div class="pricing-card stack-sm">
            <div class="row"><Clock size={18} /><h3 class="display-sub-sm">Pengiriman</h3></div>
            <div class="row-between">
              <span class="caption text-body-mid">Metode</span>
              <span class="body-sm">{order.delivery_method === 'kirim' ? 'Diantar' : 'Ambil di toko'}</span>
            </div>
            {#if order.delivery_address}
              <span class="body-sm text-body">{order.delivery_address}</span>
            {/if}
            {#if order.notes}
              <span class="caption text-body-mid">Catatan: {order.notes}</span>
            {/if}
          </div>

          <div class="pricing-card stack-sm">
            <span class="caption text-body-mid">Butuh bantuan? Hubungi toko via WhatsApp 0812-3456-7890.</span>
          </div>
        </div>
      </div>
    {:else}
      <div class="empty-state body-md">Pesanan tidak ditemukan.</div>
    {/if}
  </div>
</section>

<style>
  .timeline {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-sm);
    margin-top: var(--space-md);
  }
  .step {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-sm);
    position: relative;
  }
  .step:not(:last-child)::after {
    content: '';
    position: absolute;
    top: 9px;
    left: 22px;
    right: -8px;
    height: 2px;
    background: var(--mute);
  }
  .step.done:not(:last-child)::after { background: var(--primary); }
  .step-dot {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--canvas);
    border: 2px solid var(--mute);
    z-index: 1;
  }
  .step.done .step-dot { background: var(--primary); border-color: var(--primary); }
  .step.current .step-dot { box-shadow: 0 0 0 4px rgba(255, 79, 0, 0.25); }
  .step-label { color: var(--body); font-weight: 500; }
  .step.done .step-label { color: var(--ink); font-weight: 600; }
  .track-grid {
    display: grid;
    grid-template-columns: 1fr 320px;
    gap: var(--space-2xl);
    align-items: start;
  }
  .job-row { border-bottom: 1px solid var(--mute); padding-bottom: var(--space-md); }
  .prog { height: 6px; background: var(--canvas-soft); border-radius: var(--radius-pill); overflow: hidden; }
  .prog-bar { height: 100%; background: var(--primary); border-radius: var(--radius-pill); transition: width 0.4s ease; }
  @media (max-width: 767px) {
    .track-grid { grid-template-columns: 1fr; }
    .timeline { flex-direction: column; gap: var(--space-md); }
    .step::after { display: none; }
  }
</style>
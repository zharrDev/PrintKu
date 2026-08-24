<script>
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api.js';
  import { formatRupiah, shortId, formatDate } from '$lib/format.js';
  import { toast } from '$lib/stores.js';
  import { CreditCard, TriangleAlert, CircleCheckBig, Copy, LoaderCircle } from 'lucide-svelte';

  let order = $state(null);
  let payment = $state(null);
  let simulating = $state(false);
  let loading = $state(true);

  onMount(load);

  async function load() {
    loading = true;
    const r = await api(`/orders/${page.params.id}`);
    loading = false;
    if (r.ok) {
      order = r.json.order;
      if (order.payment?.status === 'success') {
        goto(`/pesanan/${order.id}`);
        return;
      }
      const p = await api(`/payments/${order.id}`);
      if (p.ok) payment = p.json.payment;
      if (!payment?.va_number) {
        const c = await api('/payments/create', { method: 'POST', body: { orderId: order.id } });
        if (c.ok) payment = c.json.payment;
      }
    }
  }

  function copyVa() {
    if (navigator.clipboard) navigator.clipboard.writeText(payment.va_number);
    toast('Nomor VA disalin');
  }

  async function simulate(status) {
    simulating = true;
    const r = await api('/payments/webhook', { method: 'POST', body: { orderId: order.id, status } });
    simulating = false;
    if (r.ok) {
      if (status === 'success') {
        toast('Pembayaran berhasil! Pesanan diproses.', 'success');
      } else {
        toast('Pembayaran gagal — pesanan dibatalkan.', 'error');
      }
      goto(`/pesanan/${order.id}`);
    }
  }
</script>

<svelte:head><title>Pembayaran — PrintKu</title></svelte:head>

<section class="band-hero">
  <div class="container">
    <div class="pay-wrap">
      {#if loading}
        <div class="empty-state"><span class="spinner"></span> Menyiapkan pembayaran…</div>
      {:else if order && payment}
        <div class="pay-card stack-lg">
          <div class="stack-sm text-center" style="align-items: center">
            <span class="badge badge-primary">Sandbox / Uji Coba</span>
            <h1 class="display-md">Pembayaran Pesanan #{shortId(order.id)}</h1>
            <p class="body-sm text-body">Simulasikan alur pembayaran Midtrans dalam mode testing.</p>
          </div>

          <div class="pricing-card stack-md">
            <div class="row">
              <CreditCard size={22} />
              <h3 class="display-sub-sm">Virtual Account — {payment.provider}</h3>
            </div>
            <div class="row-between va-row">
              <span class="caption text-body-mid">Nomor VA</span>
              <span class="row">
                <strong class="body-md-strong">{payment.va_number}</strong>
                <button class="btn-text" onclick={copyVa}><Copy size={14} /> Salin</button>
              </span>
            </div>
            <div class="row-between">
              <span class="caption text-body-mid">Total tagihan</span>
              <span class="display-sub-sm" style="color: var(--primary)">{formatRupiah(payment.amount)}</span>
            </div>
            <div class="row-between">
              <span class="caption text-body-mid">Batas pembayaran</span>
              <span class="body-sm">{formatDate(new Date(Date.now() + 24 * 3600 * 1000).toISOString())}</span>
            </div>
            <p class="caption text-body-mid">
              Di mode sandbox, transfer tidak benar-benar dilakukan. Klik tombol di bawah untuk
              mensimulasikan status dari payment gateway.
            </p>
          </div>

          <button class="btn btn-primary" onclick={() => simulate('success')} disabled={simulating}>
            <LoaderCircle size={18} class={simulating ? 'spin' : ''} /> Simulasikan Pembayaran Berhasil
          </button>
          <button class="btn btn-tertiary" onclick={() => simulate('failed')} disabled={simulating}>
            <TriangleAlert size={18} /> Simulasikan Gagal / Batal
          </button>
          <span class="caption text-body-mid text-center" style="display: flex; align-items: center; gap: 6px; justify-content: center">
            <CircleCheckBig size={14} /> Setelah berhasil, kamu akan diarahkan ke halaman pelacakan pesanan.
          </span>
        </div>
      {:else}
        <div class="empty-state body-md">Order tidak ditemukan.</div>
      {/if}
    </div>
  </div>
</section>

<style>
  .pay-wrap { max-width: 520px; margin: 0 auto; }
  .pay-card {
    background: var(--canvas-soft);
    border-radius: var(--radius-md);
    padding: var(--space-2xl);
    box-shadow: var(--shadow-soft);
  }
  .va-row { border-bottom: 1px solid var(--mute); padding-bottom: var(--space-md); }
  .spin { animation: spin2 0.9s linear infinite; }
  @keyframes spin2 { to { transform: rotate(360deg); } }
</style>
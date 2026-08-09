<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api.js';
  import { cartStore, toast, userStore } from '$lib/stores.js';
  import { formatRupiah, PAPER_SIZE_NAME, COLOR_MODE_NAME } from '$lib/format.js';
  import { Store, Truck, MapPin, Ticket, X } from 'lucide-svelte';

  let items = $state([]);
  let total = $state(0);
  let addresses = $state([]);
  let selectedAddressId = $state('');
  let printOrderMode = $state(false);
  let printJobs = $state([]);

  let deliveryMethod = $state('ambil');
  let customerName = $state('');
  let deliveryAddress = $state('');
  let city = $state('');
  let postalCode = $state('');
  let notes = $state('');
  let placing = $state(false);
  let user = $state($userStore);

  // Voucher / diskon
  let voucherInput = $state('');
  let appliedVoucher = $state(null);
  let discount = $state(0);
  let checkingVoucher = $state(false);
  let grandTotal = $derived(Math.max(0, total - discount));

  onMount(async () => {
    const cart = await api('/cart');
    if (cart.ok) {
      items = cart.json.items;
      total = cart.json.total;
      cartStore.set({ items, total });
    }
    const addrs = await api('/auth/addresses');
    if (addrs.ok) {
      addresses = addrs.json.addresses;
      if (addresses.length) {
        selectedAddressId = addresses[0].id;
        applyAddress(addresses[0]);
      }
    }
    const session = sessionStorage.getItem('printmart_draft_job');
    if (session) {
      try {
        const job = JSON.parse(session);
        printOrderMode = true;
        printJobs = [job.printJob];
        total = job.pricing.total;
      } catch { /* ignore */ }
    }
    if (user) customerName = user.name;
  });

  function applyAddress(a) {
    deliveryAddress = a.full_address;
    city = a.city;
    postalCode = a.postal_code || '';
  }

  async function applyVoucher() {
    const code = voucherInput.trim();
    if (!code) return;
    if (total <= 0) {
      toast('Tidak ada yang bisa didiskon', 'error');
      return;
    }
    checkingVoucher = true;
    const r = await api('/vouchers/validate', { method: 'POST', body: { code, subtotal: total } });
    checkingVoucher = false;
    if (r.ok && r.json.valid) {
      appliedVoucher = r.json.voucher;
      discount = r.json.discount;
      toast(`Voucher ${appliedVoucher.code} dipakai — hemat ${formatRupiah(discount)}`, 'success');
    }
    // pesan error sudah ditampilkan otomatis oleh api()
  }

  function removeVoucher() {
    appliedVoucher = null;
    discount = 0;
    voucherInput = '';
  }

  async function placeOrder() {
    if (!user) {
      toast('Silakan masuk untuk checkout', 'error');
      goto('/masuk');
      return;
    }
    placing = true;
    const base = {
      deliveryMethod,
      delivery_address: deliveryMethod === 'kirim' ? `${deliveryAddress}, ${city} ${postalCode}`.trim() : '',
      notes,
      customer_name: customerName,
      voucher_code: appliedVoucher ? appliedVoucher.code : '',
    };
    const body = printOrderMode
      ? { ...base, orderType: 'print', printJobIds: printJobs.map((j) => j.id) }
      : { ...base, orderType: 'product' };

    const r = await api('/orders', { method: 'POST', body });
    placing = false;
    if (r.ok) {
      sessionStorage.removeItem('printmart_draft_job');
      cartStore.set({ items: [], total: 0 });
      toast('Pesanan dibuat! Lanjutkan ke pembayaran.', 'success');
      goto(`/bayar/${r.json.order.id}`);
    }
  }

  function rollbackToPrint() {
    sessionStorage.removeItem('printmart_draft_job');
    printOrderMode = false;
    printJobs = [];
    total = items.reduce((s, it) => s + it.price * it.quantity, 0);
    removeVoucher();
  }
</script>

<svelte:head><title>Checkout — PrintMart</title></svelte:head>

<section class="band-hero">
  <div class="container stack-xl">
    <div class="stack-sm">
      <span class="eyebrow">Checkout</span>
      <h1 class="display-lg">Konfirmasi pesanan</h1>
    </div>

    <div class="checkout-grid">
      <div class="stack-xl">
        <!-- Ringkasan pesanan -->
        <div class="card stack-md">
          <h3 class="display-sub-sm">{printOrderMode ? 'Jasa print' : 'Produk'}</h3>
          {#if printOrderMode}
            {#each printJobs as j (j.id)}
              <div class="row-between">
                <div class="stack-sm" style="min-width: 0">
                  <span class="body-sm-strong">{j.file_name}</span>
                  <span class="caption text-body-mid">
                    {j.page_count} hal · {COLOR_MODE_NAME[j.color_mode]} · {PAPER_SIZE_NAME[j.paper_size]} · {j.copies}x rangkap
                    {#if j.finishing} · {j.finishing.name}{/if}
                  </span>
                </div>
                <span class="body-md-strong">{formatRupiah(j.estimated_price)}</span>
              </div>
            {/each}
            <button class="btn-text" style="align-self: flex-start" onclick={rollbackToPrint}>Ubah spesifikasi print</button>
          {:else if items.length === 0}
            <p class="body-sm text-body">Keranjang kosong. <a href="/katalog" class="muted-link" style="font-weight: 600">Lihat produk >></a></p>
          {:else}
            {#each items as it (it.cart_item_id)}
              <div class="row-between">
                <span class="body-sm">{it.name} × {it.quantity}</span>
                <span class="body-sm">{formatRupiah(it.price * it.quantity)}</span>
              </div>
            {/each}
          {/if}
          <hr class="divider" />
          <div class="row-between">
            <span class="body-md-strong">Total</span>
            <span class="display-sub-sm" style="color: var(--primary)">{formatRupiah(total)}</span>
          </div>
        </div>

        <!-- Metode penerimaan -->
        <div class="card stack-md">
          <h3 class="display-sub-sm">Cara menerima pesanan</h3>
          <div class="grid grid-2">
            <button class="method-card" class:method-active={deliveryMethod === 'ambil'} onclick={() => (deliveryMethod = 'ambil')}>
              <Store size={22} />
              <span class="body-md-strong">Ambil di toko</span>
              <span class="caption text-body-mid">Gratis, siap 30-60 menit</span>
            </button>
            <button class="method-card" class:method-active={deliveryMethod === 'kirim'} onclick={() => (deliveryMethod = 'kirim')}>
              <Truck size={22} />
              <span class="body-md-strong">Diantar</span>
              <span class="caption text-body-mid">Ongkir dihitung oleh admin</span>
            </button>
          </div>
        </div>

        <!-- Data pemesan / alamat -->
        <div class="card stack-md">
          <h3 class="display-sub-sm">Data pemesan</h3>
          <label class="field">
            <span class="label-text">Nama</span>
            <input class="input" bind:value={customerName} placeholder="Nama kamu" />
          </label>

          {#if deliveryMethod === 'kirim'}
            {#if addresses.length > 0}
              <label class="field">
                <span class="label-text">Pilih alamat tersimpan</span>
                <select class="input" bind:value={selectedAddressId} onchange={(e) => {
                  const a = addresses.find((x) => x.id === e.currentTarget.value);
                  if (a) applyAddress(a);
                }}>
                  {#each addresses as a (a.id)}
                    <option value={a.id}>{a.label} — {a.full_address}, {a.city}</option>
                  {/each}
                </select>
              </label>
            {/if}
            <div class="grid grid-2">
              <label class="field">
                <span class="label-text">Alamat lengkap</span>
                <input class="input" bind:value={deliveryAddress} placeholder="Jl. Merdeka No. 88" />
              </label>
              <label class="field">
                <span class="label-text">Kota</span>
                <input class="input" bind:value={city} placeholder="Bandung" />
              </label>
            </div>
            <label class="field">
              <span class="label-text">Kode pos (opsional)</span>
              <input class="input" bind:value={postalCode} placeholder="40111" />
            </label>
            <p class="caption text-body-mid"><MapPin size={13} style="vertical-align: -2px" /> Pastikan alamat benar — admin akan menghubungi untuk konfirmasi.</p>
          {/if}

          <label class="field">
            <span class="label-text">Catatan (opsional)</span>
            <textarea class="input" rows="2" bind:value={notes} placeholder="Contoh: kirim sore hari / jilid warna biru"></textarea>
          </label>
        </div>
      </div>

      <!-- Summary side -->
      <div class="pricing-card stack-md summary">
        <h3 class="display-sub-sm">Total bayar</h3>

        <!-- Voucher -->
        <div class="stack-sm">
          <span class="label-text"><Ticket size={13} style="vertical-align: -2px" /> Kode voucher</span>
          {#if appliedVoucher}
            <div class="voucher-applied row-between">
              <div class="stack-xs" style="min-width: 0">
                <span class="body-sm-strong">{appliedVoucher.code}</span>
                {#if appliedVoucher.description}<span class="caption text-body-mid clamp">{appliedVoucher.description}</span>{/if}
              </div>
              <button class="icon-btn" title="Lepas voucher" onclick={removeVoucher}><X size={15} /></button>
            </div>
          {:else}
            <div class="row">
              <input
                class="input"
                placeholder="mis. HEMAT10"
                bind:value={voucherInput}
                onkeydown={(e) => e.key === 'Enter' && applyVoucher()}
                style="text-transform: uppercase"
              />
              <button class="btn btn-tertiary" onclick={applyVoucher} disabled={checkingVoucher || !voucherInput.trim()}>
                {checkingVoucher ? '…' : 'Pakai'}
              </button>
            </div>
          {/if}
        </div>

        <hr class="divider" />
        <div class="row-between">
          <span class="body-sm text-body">Subtotal</span>
          <span class="body-md">{formatRupiah(total)}</span>
        </div>
        {#if discount > 0}
          <div class="row-between">
            <span class="body-sm text-body">Diskon voucher</span>
            <span class="body-md" style="color: var(--primary)">−{formatRupiah(discount)}</span>
          </div>
        {/if}
        <div class="row-between">
          <span class="body-sm text-body">Ongkir</span>
          <span class="body-md">{deliveryMethod === 'ambil' ? 'Gratis' : 'Ditentukan admin'}</span>
        </div>
        <hr class="divider" />
        <div class="row-between">
          <span class="body-md-strong">Total</span>
          <span class="display-sub-sm" style="color: var(--primary)">{formatRupiah(grandTotal)}</span>
        </div>
        <button class="btn btn-primary" onclick={placeOrder} disabled={placing || (!printOrderMode && items.length === 0)}>
          {placing ? 'Membuat pesanan…' : 'Buat Pesanan & Bayar'}
        </button>
        <p class="caption text-body-mid">Pembayaran memakai mode sandbox — tidak ada uang asli.</p>
      </div>
    </div>
  </div>
</section>

<style>
  .checkout-grid {
    display: grid;
    grid-template-columns: 1fr 340px;
    gap: var(--space-2xl);
    align-items: start;
  }
  .method-card {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-xs);
    text-align: left;
    background: var(--canvas-soft);
    border: 1px solid var(--mute);
    border-radius: var(--radius-md);
    padding: var(--space-lg);
    color: var(--ink);
    transition: all 0.12s ease;
  }
  .method-card:hover { border-color: var(--ink); }
  .method-active { border-color: var(--ink); background: var(--canvas); box-shadow: inset 0 0 0 1px var(--ink); }
  .summary { position: sticky; top: 96px; }
  .stack-xs { display: flex; flex-direction: column; gap: 2px; }
  .clamp { max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .voucher-applied {
    background: var(--canvas-soft);
    border: 1px dashed var(--primary);
    border-radius: var(--radius-sm);
    padding: var(--space-sm) var(--space-md);
  }
  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    flex-shrink: 0;
    border: 1px solid var(--mute);
    background: var(--canvas);
    border-radius: var(--radius-sm);
    color: var(--ink);
  }
  .icon-btn:hover { border-color: var(--ink); }
  @media (max-width: 767px) {
    .checkout-grid { grid-template-columns: 1fr; }
    .summary { position: static; }
  }
</style>
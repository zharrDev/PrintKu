<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api.js';
  import { cartStore, toast } from '$lib/stores.js';
  import { formatRupiah } from '$lib/format.js';
  import { Plus, Minus, Trash2, ShoppingBag, ArrowRight } from 'lucide-svelte';

  let items = $state([]);
  let total = $state(0);
  let loading = $state(true);

  onMount(load);

  async function load() {
    const r = await api('/cart');
    loading = false;
    if (r.ok) {
      items = r.json.items;
      total = r.json.total;
      cartStore.set({ items, total });
    }
  }

  async function changeQty(cartItemId, qty) {
    const r = await api(`/cart/${cartItemId}`, { method: 'PATCH', body: { quantity: qty } });
    if (r.ok) {
      items = r.json.items;
      total = r.json.total;
      cartStore.set({ items, total });
    }
  }

  async function removeItem(cartItemId) {
    const r = await api(`/cart/${cartItemId}`, { method: 'DELETE' });
    if (r.ok) {
      toast('Item dihapus dari keranjang');
      items = r.json.items;
      total = r.json.total;
      cartStore.set({ items, total });
    }
  }
</script>

<svelte:head><title>Keranjang — PrintKu</title></svelte:head>

<section class="band-hero">
  <div class="container stack-xl">
    <div class="stack-sm">
      <span class="eyebrow">Keranjang</span>
      <h1 class="display-lg">Belanjaan kamu</h1>
    </div>

    {#if loading}
      <div class="empty-state"><span class="spinner"></span> Memuat…</div>
    {:else if items.length === 0}
      <div class="empty-state stack-md" style="align-items: center">
        <ShoppingBag size={32} strokeWidth={1.5} />
        <span class="body-md-strong">Keranjang masih kosong</span>
        <span class="body-sm text-body-mid">Yuk isi dengan kebutuhan tulis kamu.</span>
        <a href="/katalog" class="btn btn-primary">Lihat Katalog <ArrowRight size={16} /></a>
      </div>
    {:else}
      <div class="cart-grid">
        <div class="card stack-md">
          {#each items as it (it.cart_item_id)}
            <div class="cart-item">
              <div class="item-thumb">
                <svg viewBox="0 0 64 48" aria-hidden="true">
                  <rect width="64" height="48" rx="8" class="thumb-bg" />
                  <text x="32" y="34" text-anchor="middle" font-size="20" font-weight="700" class="thumb-fg">{it.name[0]}</text>
                </svg>
              </div>
              <div class="stack-sm" style="flex: 1; min-width: 0">
                <a href={`/produk/${it.id}`} class="body-md-strong">{it.name}</a>
                <span class="caption text-body-mid">{formatRupiah(it.price)} / unit</span>
              </div>
              <div class="qty">
                <button class="qty-btn" onclick={() => changeQty(it.cart_item_id, it.quantity - 1)}><Minus size={16} /></button>
                <input class="qty-input" type="number" min="1" bind:value={it.quantity} onchange={(e) => changeQty(it.cart_item_id, Math.max(1, Number(e.currentTarget.value) || 1))} />
                <button class="qty-btn" onclick={() => changeQty(it.cart_item_id, it.quantity + 1)}><Plus size={16} /></button>
              </div>
              <span class="body-md-strong item-total">{formatRupiah(it.price * it.quantity)}</span>
              <button class="remove-btn" aria-label="Hapus" onclick={() => removeItem(it.cart_item_id)}><Trash2 size={18} /></button>
            </div>
          {/each}
        </div>

        <div class="pricing-card stack-md summary" style="align-self: start; position: sticky; top: 96px">
          <h3 class="display-sub-sm">Ringkasan</h3>
          <div class="row-between">
            <span class="body-sm text-body">Jumlah item</span>
            <span class="body-md">{items.reduce((n, it) => n + it.quantity, 0)}</span>
          </div>
          <div class="row-between">
            <span class="body-md-strong">Subtotal</span>
            <span class="display-sub-sm" style="color: var(--primary)">{formatRupiah(total)}</span>
          </div>
          <p class="caption text-body-mid">Belum termasuk metode pengiriman. Ditentukan di checkout.</p>
          <button class="btn btn-primary" onclick={() => goto('/checkout')}>Lanjut ke Checkout <ArrowRight size={18} /></button>
          <a href="/katalog" class="btn-text" style="justify-content: center">Tambahkan barang lain</a>
        </div>
      </div>
    {/if}
  </div>
</section>

<style>
  .cart-grid {
    display: grid;
    grid-template-columns: 1fr 340px;
    gap: var(--space-2xl);
    align-items: start;
  }
  .cart-item {
    display: flex;
    align-items: center;
    gap: var(--space-lg);
    padding: var(--space-md) 0;
    border-bottom: 1px solid var(--mute);
  }
  .cart-item:last-child { border-bottom: none; }
  .item-thumb svg { width: 56px; border: 1px solid var(--mute); border-radius: var(--radius-sm); }
  .qty { display: inline-flex; align-items: center; border: 1px solid var(--ink); border-radius: var(--radius-sm); overflow: hidden; }
  .qty-btn {
    background: var(--canvas);
    border: none;
    padding: var(--space-xs) var(--space-sm);
    display: inline-flex;
    color: var(--ink);
  }
  .qty-btn:hover { background: var(--canvas-soft); }
  .qty-input {
    width: 40px;
    border: none;
    border-left: 1px solid var(--ink);
    border-right: 1px solid var(--ink);
    text-align: center;
    font-size: 15px;
    padding: var(--space-xs) 0;
    background: var(--canvas);
    color: var(--ink);
  }
  .item-total { min-width: 90px; text-align: right; }
  .remove-btn {
    background: none;
    border: none;
    color: var(--body-mid);
    padding: var(--space-xs);
    border-radius: var(--radius-sm);
  }
  .remove-btn:hover { color: var(--primary); background: var(--canvas-soft); }
  @media (max-width: 767px) {
    .cart-grid { grid-template-columns: 1fr; }
    .cart-item { flex-wrap: wrap; }
    .item-total { width: 100%; text-align: left; }
    .summary { position: static !important; }
  }
</style>
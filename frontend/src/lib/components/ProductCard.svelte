<script>
  import { onMount } from 'svelte';
  import { toast } from '$lib/stores.js';
  import { api } from '$lib/api.js';
  import { formatRupiah } from '$lib/format.js';
  import { ShoppingCart, Plus, Minus } from 'lucide-svelte';
  import { goto } from '$app/navigation';

  let { product } = $props();
  let qty = $state(1);
  let adding = $state(false);

  async function addToCart() {
    if (!localStorage.getItem('printku_token')) {
      toast('Silakan masuk dulu untuk berbelanja', 'error');
      goto('/masuk');
      return;
    }
    adding = true;
    const r = await api('/cart', { method: 'POST', body: { product_id: product.id, quantity: qty } });
    adding = false;
    if (r.ok) {
      toast(`"${product.name}" ditambahkan ke keranjang`);
      goto('/keranjang');
    }
  }
</script>

<div class="card card-hover product-card">
  <a href={`/produk/${product.id}`}>
    <div class="thumb">
      <svg viewBox="0 0 64 64" aria-hidden="true">
        <rect width="64" height="64" rx="8" class="thumb-bg" />
        <text x="32" y="44" text-anchor="middle" font-size="26" font-weight="700" class="thumb-fg">{product.name[0]}</text>
      </svg>
    </div>
    <div class="stack-sm body">
      <span class="badge">{product.category_name || 'Produk'}</span>
      <h3 class="display-xs">{product.name}</h3>
      <p class="caption text-body-mid clamp">{product.description}</p>
      <div class="row-between">
        <span class="body-md-strong price">{formatRupiah(product.price)}</span>
        <span class="caption text-body-mid">Stok {product.stock}</span>
      </div>
    </div>
  </a>
  <div class="row actions">
    <div class="qty">
      <button class="qty-btn" aria-label="Kurangi" onclick={() => (qty = Math.max(1, qty - 1))}><Minus size={16} /></button>
      <input type="number" min="1" max={product.stock} bind:value={qty} class="qty-input" />
      <button class="qty-btn" aria-label="Tambah" onclick={() => (qty = Math.min(product.stock, qty + 1))}><Plus size={16} /></button>
    </div>
    <button class="btn btn-primary btn-sm flex-grow" onclick={addToCart} disabled={adding || product.stock < 1}>
      <ShoppingCart size={16} /> {product.stock < 1 ? 'Stok Habis' : adding ? '...' : 'Keranjang'}
    </button>
  </div>
</div>

<style>
  .product-card { display: flex; flex-direction: column; gap: var(--space-lg); }
  .thumb svg { width: 100%; aspect-ratio: 1.6; border-radius: var(--radius-sm); }
  /* Warna SVG lewat CSS supaya ikut token, bukan hard-coded hex. */
  .thumb-bg { fill: var(--canvas); }
  .thumb-fg { fill: var(--ink); }
  .clamp {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .price { color: var(--ink); }
  .actions { margin-top: auto; }
  .qty {
    display: inline-flex;
    align-items: center;
    border: var(--border);
    border-radius: var(--radius-sm);
    overflow: hidden;
    background: var(--canvas);
  }
  .qty-btn {
    background: var(--canvas);
    border: none;
    padding: var(--space-sm);
    min-width: 36px;
    min-height: 36px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--ink);
    transition: background-color var(--dur) var(--ease);
  }
  .qty-btn:hover { background: var(--canvas-soft); }
  .qty-input {
    width: 48px;
    border: none;
    border-left: var(--border-thin);
    border-right: var(--border-thin);
    text-align: center;
    font-size: 16px;
    font-variant-numeric: tabular-nums;
    padding: var(--space-sm) 0;
    background: var(--canvas);
    color: var(--ink);
  }
  .flex-grow { flex: 1; }
</style>
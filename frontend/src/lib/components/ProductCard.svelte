<script>
  import { toast } from '$lib/stores.js';
  import { api } from '$lib/api.js';
  import { formatRupiah } from '$lib/format.js';
  import { ShoppingCart, Plus, Minus } from 'lucide-svelte';
  import { goto } from '$app/navigation';

  let { product } = $props();
  let qty = $state(1);
  let adding = $state(false);
  // Bila gambar produk gagal dimuat, tampilkan stiker inisial sebagai gantinya.
  let imgBroken = $state(false);
  let showImage = $derived(Boolean(product?.image_url) && !imgBroken);

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
  <a href={`/produk/${product.id}`} class="thumb-link">
    <div class="thumb">
      {#if showImage}
        <img
          class="thumb-img"
          src={product.image_url}
          alt={product.name}
          loading="lazy"
          decoding="async"
          onerror={() => (imgBroken = true)}
        />
      {:else}
        <span class="thumb-letter" aria-hidden="true">{product.name[0]}</span>
      {/if}
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
      <input type="number" min="1" max={product.stock} bind:value={qty} class="qty-input" aria-label={`Jumlah ${product.name}`} />
      <button class="qty-btn" aria-label="Tambah" onclick={() => (qty = Math.min(product.stock, qty + 1))}><Plus size={16} /></button>
    </div>
    <button class="btn btn-primary btn-sm flex-grow" onclick={addToCart} disabled={adding || product.stock < 1}>
      <ShoppingCart size={16} /> {product.stock < 1 ? 'Stok Habis' : adding ? '...' : 'Keranjang'}
    </button>
  </div>
</div>

<style>
  .product-card { display: flex; flex-direction: column; gap: var(--space-lg); }
  /* Placeholder produk: kertas berbintik + "stiker" inisial yang miring,
     jadi tetap menarik walau produk belum punya foto asli. */
  .thumb {
    position: relative;
    aspect-ratio: 1.6;
    border: var(--border);
    border-radius: var(--radius-md);
    background-color: var(--canvas);
    background-image: radial-gradient(circle at 1px 1px, rgba(32, 21, 21, 0.09) 1px, transparent 0);
    background-size: 14px 14px;
    display: grid;
    place-items: center;
    overflow: hidden;
  }
  .thumb-letter {
    font-family: var(--font-display);
    font-size: 40px;
    font-weight: 600;
    line-height: 1;
    color: var(--ink);
    background: var(--primary-soft);
    border: var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-sm) var(--space-xl);
    box-shadow: var(--shadow-xs);
    transform: rotate(-3deg);
    transition: transform var(--dur) var(--ease);
  }
  .card-hover:hover .thumb-letter { transform: rotate(-3deg) scale(1.06); }
  .thumb-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 320ms var(--ease);
  }
  .card-hover:hover .thumb-img { transform: scale(1.05); }
  .clamp {
    display: -webkit-box;
    line-clamp: 2;
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
<script>
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { api } from '$lib/api.js';
  import { formatRupiah } from '$lib/format.js';
  import { toast } from '$lib/stores.js';
  import { ShoppingCart, Minus, Plus, ArrowLeft } from 'lucide-svelte';
  import { goto } from '$app/navigation';

  let product = $state(null);
  let qty = $state(1);
  let adding = $state(false);
  // Fallback ke stiker inisial bila gambar produk tidak bisa dimuat.
  let imgBroken = $state(false);
  let showImage = $derived(Boolean(product?.image_url) && !imgBroken);

  onMount(async () => {
    const r = await api(`/products/${page.params.id}`);
    if (r.ok) product = r.json.product;
  });

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
      toast(`"${product.name}" masuk keranjang`, 'success');
      goto('/keranjang');
    }
  }
</script>

<svelte:head>
  <title>{product?.name ? `${product.name} — PrintKu` : 'Produk — PrintKu'}</title>
</svelte:head>

<section class="band-hero">
  <div class="container stack-xl">
    {#if product}
      <a href="/katalog" class="btn-text" style="align-self: flex-start"><ArrowLeft size={16} /> Kembali ke katalog</a>
      <div class="detail-grid">
        <div class="thumb">
          {#if showImage}
            <img
              class="thumb-img"
              src={product.image_url}
              alt={`Foto produk ${product.name}`}
              decoding="async"
              onerror={() => (imgBroken = true)}
            />
          {:else}
            <span class="thumb-letter" aria-hidden="true">{product.name[0]}</span>
          {/if}
        </div>
        <div class="card stack-xl">
          <div class="row wrap">
            <span class="badge">{product.category_name || 'Produk'}</span>
            <span class="badge">Stok {product.stock}</span>
          </div>
          <h1 class="display-lg">{product.name}</h1>
          <p class="body-md text-body">{product.description}</p>
          <span class="display-md" style="color: var(--primary)">{formatRupiah(product.price)}</span>
          <hr class="divider" />
          <div class="row">
            <div class="qty">
              <button class="qty-btn" aria-label="Kurangi" onclick={() => (qty = Math.max(1, qty - 1))} disabled={product.stock < 1}><Minus size={16} /></button>
              <input type="number" min="1" max={product.stock} bind:value={qty} class="qty-input" aria-label={`Jumlah ${product.name}`} />
              <button class="qty-btn" aria-label="Tambah" onclick={() => (qty = Math.min(product.stock, qty + 1))} disabled={product.stock < 1}><Plus size={16} /></button>
            </div>
            <button class="btn btn-primary" style="flex: 1" onclick={addToCart} disabled={adding || product.stock < 1}>
              <ShoppingCart size={18} />
              {product.stock < 1 ? 'Stok Habis' : adding ? 'Menambahkan…' : 'Tambah ke Keranjang'}
            </button>
          </div>
          <p class="caption text-body-mid">
            Subtotal: <strong class="text-ink">{formatRupiah(product.price * qty)}</strong> — bisa ambil di toko atau dikirim.
          </p>
        </div>
      </div>
    {:else}
      <div class="empty-state"><span class="spinner"></span> Memuat produk…</div>
    {/if}
  </div>
</section>

<style>
  .detail-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3xl);
    align-items: start;
  }
  /* Placeholder produk: kertas berbintik + "stiker" inisial yang miring. */
  .thumb {
    position: relative;
    aspect-ratio: 4 / 3;
    border: var(--border);
    border-radius: var(--radius-lg);
    background-color: var(--canvas-soft);
    background-image: radial-gradient(circle at 1px 1px, rgba(32, 21, 21, 0.09) 1px, transparent 0);
    background-size: 18px 18px;
    display: grid;
    place-items: center;
    box-shadow: var(--shadow-md);
    overflow: hidden;
  }
  .thumb-letter {
    font-family: var(--font-display);
    font-size: 96px;
    font-weight: 600;
    line-height: 1;
    color: var(--ink);
    background: var(--primary-soft);
    border: var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-lg) var(--space-3xl);
    box-shadow: var(--shadow-sm);
    transform: rotate(-3deg);
  }
  .thumb-img { width: 100%; height: 100%; object-fit: cover; }
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
    padding: var(--space-md);
    min-width: 44px;
    min-height: 44px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--ink);
    transition: background-color var(--dur) var(--ease);
  }
  .qty-btn:disabled { background: var(--mute); color: var(--canvas); cursor: not-allowed; }
  .qty-btn:hover:not(:disabled) { background: var(--canvas-soft); }
  .qty-input {
    width: 56px;
    border: none;
    border-left: var(--border-thin);
    border-right: var(--border-thin);
    text-align: center;
    font-size: 18px;
    font-variant-numeric: tabular-nums;
    padding: var(--space-md) 0;
    background: var(--canvas);
    color: var(--ink);
  }
  @media (max-width: 767px) {
    .detail-grid { grid-template-columns: 1fr; }
  }
</style>
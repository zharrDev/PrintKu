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
          <svg viewBox="0 0 400 300" aria-hidden="true">
            <rect width="400" height="300" rx="12" fill="#f8f4f0" />
            <text x="200" y="175" text-anchor="middle" font-size="100" font-weight="700" fill="#201515">
              {product.name[0]}
            </text>
          </svg>
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
              <input type="number" min="1" max={product.stock} bind:value={qty} class="qty-input" />
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
  .thumb svg { width: 100%; border-radius: var(--radius-md); }
  .qty { display: inline-flex; align-items: center; border: 1px solid var(--ink); border-radius: var(--radius-sm); overflow: hidden; }
  .qty-btn {
    background: var(--canvas);
    border: none;
    padding: var(--space-md);
    display: inline-flex;
    color: var(--ink);
  }
  .qty-btn:disabled { opacity: 0.4; }
  .qty-btn:hover:not(:disabled) { background: var(--canvas-soft); }
  .qty-input {
    width: 56px;
    border: none;
    border-left: 1px solid var(--ink);
    border-right: 1px solid var(--ink);
    text-align: center;
    font-size: 18px;
    padding: var(--space-md) 0;
    background: var(--canvas);
    color: var(--ink);
  }
  @media (max-width: 767px) {
    .detail-grid { grid-template-columns: 1fr; }
  }
</style>
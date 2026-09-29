<script>
  import { onMount } from 'svelte';
  import { Search, SlidersHorizontal } from 'lucide-svelte';
  import { api } from '$lib/api.js';
  import ProductCard from '$lib/components/ProductCard.svelte';

  let products = $state([]);
  let categories = $state([]);
  let loading = $state(true);
  let search = $state('');
  let category = $state('');
  let sort = $state('name');

  onMount(async () => {
    loadProducts();
    const c = await api('/categories');
    if (c.ok) categories = c.json.categories;
  });

  async function loadProducts() {
    loading = true;
    const params = new URLSearchParams();
    if (search.trim()) params.set('search', search.trim());
    if (category) params.set('category', category);
    if (sort) params.set('sort', sort);
    const r = await api(`/products?${params.toString()}`);
    loading = false;
    if (r.ok) products = r.json.products;
  }

  function applyFilters() {
    loadProducts();
  }
</script>

<svelte:head>
  <title>Katalog — PrintKu</title>
</svelte:head>

<section class="band-hero">
  <div class="container stack-xl">
    <div class="stack-sm">
      <span class="eyebrow">Katalog</span>
      <h1 class="display-lg">Belanja kebutuhan tulis & kantor</h1>
    </div>

    <div class="card stack-md">
      <div class="row wrap">
        <div class="search-box">
          <Search size={18} />
          <input
            class="input search-input"
            placeholder="Cari produk… (misal: kertas, pulpen)"
            bind:value={search}
            onkeydown={(e) => e.key === 'Enter' && applyFilters()}
          />
        </div>
        <button class="btn btn-primary" onclick={applyFilters}>Cari</button>
      </div>
      <div class="row wrap filter-row">
        <button class:chip-active={!category} class="chip" onclick={() => { category = ''; loadProducts(); }}>Semua</button>
        {#each categories as c (c.id)}
          <button
            class:chip-active={category === c.id}
            class="chip"
            onclick={() => { category = c.id; loadProducts(); }}
          >{c.name}</button>
        {/each}
        <span style="flex: 1"></span>
        <label class="row sort-label">
          <SlidersHorizontal size={16} />
          <select class="input sort-select" bind:value={sort} onchange={loadProducts}>
            <option value="name">Urut: Nama</option>
            <option value="price_asc">Harga terendah</option>
            <option value="price_desc">Harga tertinggi</option>
            <option value="newest">Terbaru</option>
          </select>
        </label>
      </div>
    </div>

    {#if loading}
      <div class="empty-state"><span class="spinner"></span> Memuat produk…</div>
    {:else if products.length === 0}
      <div class="empty-state stack-sm">
        <span class="body-md-strong">Tidak ada produk ditemukan</span>
        <span class="body-sm text-body-mid">Coba ubah kata kunci atau filter kategori.</span>
      </div>
    {:else}
      <div class="grid grid-4">
        {#each products as product (product.id)}
          <ProductCard {product} />
        {/each}
      </div>
    {/if}
  </div>
</section>

<style>
  .search-box {
    position: relative;
    flex: 1;
    min-width: 260px;
  }
  .search-box :global(svg) {
    position: absolute;
    left: var(--space-lg);
    top: 50%;
    transform: translateY(-50%);
    color: var(--body-mid);
  }
  .search-input { padding-left: var(--space-3xl); }
  .filter-row { gap: var(--space-sm); }
  .chip {
    border: 1px solid var(--mute);
    background: var(--canvas);
    color: var(--ink);
    border-radius: var(--radius-pill);
    padding: var(--space-xs) var(--space-lg);
    font-size: 16px;
    line-height: 24px;
    transition: background-color var(--dur) var(--ease),
                border-color var(--dur) var(--ease),
                color var(--dur) var(--ease),
                transform var(--dur) var(--ease);
  }
  .chip:hover { border-color: var(--ink); }
  .chip-active {
    background: var(--ink);
    color: var(--on-primary);
    border-color: var(--ink);
  }
  .sort-label { margin-left: auto; }
  .sort-select { width: auto; min-width: 180px; padding: var(--space-sm) var(--space-xl); }
</style>
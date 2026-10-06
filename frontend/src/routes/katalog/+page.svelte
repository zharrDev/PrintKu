<script>
  import { onMount } from 'svelte';
  import { Search, SlidersHorizontal, ChevronLeft, ChevronRight } from 'lucide-svelte';
  import { api } from '$lib/api.js';
  import ProductCard from '$lib/components/ProductCard.svelte';

  const PAGE_SIZE = 12;

  let products = $state([]);
  let categories = $state([]);
  let loading = $state(true);
  let search = $state('');
  let category = $state('');
  let sort = $state('name');
  let page = $state(1);
  let total = $state(0);
  let totalPages = $state(1);

  // Nomor halaman untuk navigasi (maksimal 5 tombol agar tidak penuh).
  let pageNumbers = $derived.by(() => {
    const max = 5;
    let start = Math.max(1, page - 2);
    const end = Math.min(totalPages, start + max - 1);
    start = Math.max(1, end - max + 1);
    const out = [];
    for (let i = start; i <= end; i++) out.push(i);
    return out;
  });

  onMount(async () => {
    loadProducts();
    const c = await api('/categories');
    if (c.ok) categories = c.json.categories;
  });

  async function loadProducts(targetPage = 1) {
    loading = true;
    page = targetPage;
    const params = new URLSearchParams();
    if (search.trim()) params.set('search', search.trim());
    if (category) params.set('category', category);
    if (sort) params.set('sort', sort);
    params.set('page', String(page));
    params.set('limit', String(PAGE_SIZE));
    const r = await api(`/products?${params.toString()}`);
    loading = false;
    if (r.ok) {
      products = r.json.products;
      total = r.json.total ?? r.json.products.length;
      totalPages = r.json.totalPages || 1;
      // Jaga-jaga bila server mengembalikan halaman kosong (mis. data
      // berubah saat user masih di halaman terakhir).
      if (products.length === 0 && page > 1) {
        await loadProducts(page - 1);
      }
    }
  }

  // Setiap perubahan filter/spencarian kembali ke halaman pertama.
  function applyFilters() {
    loadProducts(1);
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
        <button class:chip-active={!category} class="chip" onclick={() => { category = ''; applyFilters(); }}>Semua</button>
        {#each categories as c (c.id)}
          <button
            class:chip-active={category === c.id}
            class="chip"
            onclick={() => { category = c.id; applyFilters(); }}
          >{c.name}</button>
        {/each}
        <span style="flex: 1"></span>
        <label class="row sort-label">
          <SlidersHorizontal size={16} />
          <select class="input sort-select" bind:value={sort} onchange={applyFilters}>
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

      {#if totalPages > 1}
        <nav class="pager" aria-label="Navigasi halaman katalog">
          <button
            class="pager-btn"
            onclick={() => loadProducts(page - 1)}
            disabled={page <= 1 || loading}
            aria-label="Halaman sebelumnya"
          >
            <ChevronLeft size={16} /> Sebelumnya
          </button>

          <div class="pager-numbers">
            {#each pageNumbers as n (n)}
              <button
                class="pager-num"
                class:pager-num-active={n === page}
                onclick={() => loadProducts(n)}
                disabled={loading}
                aria-current={n === page ? 'page' : undefined}
                aria-label="Halaman {n}"
              >{n}</button>
            {/each}
          </div>

          <button
            class="pager-btn"
            onclick={() => loadProducts(page + 1)}
            disabled={page >= totalPages || loading}
            aria-label="Halaman berikutnya"
          >
            Berikutnya <ChevronRight size={16} />
          </button>
        </nav>
        <p class="pager-info caption text-body-mid">
          Menampilkan {products.length} dari {total} produk
        </p>
      {/if}
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

  /* ── Pagination (token neubrutalisme yang sama) ─────────────── */
  .pager {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-wrap: wrap;
    gap: var(--space-sm);
    margin-top: var(--space-2xl);
  }
  .pager-numbers { display: flex; gap: var(--space-xs); }
  .pager-btn,
  .pager-num {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-xs);
    min-height: var(--touch-min);
    padding: var(--space-sm) var(--space-lg);
    background: var(--canvas);
    color: var(--ink);
    border: var(--border);
    border-radius: var(--radius-md);
    font-size: 16px;
    line-height: 24px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    box-shadow: var(--shadow-sm);
    transition: transform var(--dur) var(--ease),
                box-shadow var(--dur) var(--ease),
                background-color var(--dur) var(--ease);
  }
  .pager-num { min-width: var(--touch-min); padding: var(--space-sm) var(--space-md); }
  .pager-btn:hover:not(:disabled),
  .pager-num:hover:not(:disabled) {
    transform: translate(-2px, -2px);
    box-shadow: var(--shadow-md);
    background: var(--canvas-soft);
  }
  .pager-btn:active:not(:disabled),
  .pager-num:active:not(:disabled) {
    transform: translate(2px, 2px);
    box-shadow: 0 0 0 var(--ink);
  }
  .pager-btn:disabled,
  .pager-num:disabled {
    background: var(--mute);
    color: var(--canvas);
    border-color: var(--mute);
    box-shadow: none;
    cursor: not-allowed;
    transform: none;
  }
  .pager-num-active,
  .pager-num-active:hover:not(:disabled) {
    background: var(--primary);
    color: var(--on-primary);
    border-color: var(--ink);
  }
  .pager-info { text-align: center; margin-top: var(--space-md); }
  @media (max-width: 767px) {
    .pager-btn { font-size: 0; gap: 0; padding: var(--space-sm) var(--space-md); }
    .pager-btn :global(svg) { width: 20px; height: 20px; }
  }
</style>
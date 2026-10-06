<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import { formatRupiah } from '$lib/format.js';
  import { toast } from '$lib/stores.js';
  import Modal from '$lib/components/Modal.svelte';
  import { Plus, Pencil, Trash2, Search, Package, ImageOff } from 'lucide-svelte';

  let products = $state([]);
  let categories = $state([]);
  let loading = $state(true);
  let search = $state('');
  let modal = $state(false);
  let editing = $state(null);
  let saving = $state(false);
  // ID produk yang gambarnya gagal dimuat → tampilkan stiker inisial.
  let broken = $state([]);

  const emptyForm = () => ({
    category_id: '',
    name: '',
    description: '',
    price: 0,
    stock: 0,
    image_url: '',
    is_active: 1,
  });
  let form = $state(emptyForm());

  onMount(async () => {
    load();
    const c = await api('/categories');
    if (c.ok) categories = c.json.categories;
  });

  async function load() {
    loading = true;
    const params = search.trim() ? `?search=${encodeURIComponent(search.trim())}` : '';
    const r = await api(`/products${params}`);
    loading = false;
    if (r.ok) products = r.json.products;
  }

  function openCreate() {
    editing = null;
    form = { ...emptyForm(), category_id: categories[0]?.id || '' };
    modal = true;
  }

  function openEdit(p) {
    editing = p;
    form = {
      category_id: p.category_id,
      name: p.name,
      description: p.description ?? '',
      price: p.price,
      stock: p.stock,
      image_url: p.image_url ?? '',
      is_active: p.is_active,
    };
    modal = true;
  }

  async function save() {
    if (!form.name.trim()) {
      toast('Nama produk wajib diisi', 'error');
      return;
    }
    saving = true;
    const r = editing
      ? await api(`/products/${editing.id}`, { method: 'PUT', body: form })
      : await api('/products', { method: 'POST', body: form });
    saving = false;
    if (r.ok) {
      toast(editing ? 'Produk diperbarui' : 'Produk ditambahkan', 'success');
      modal = false;
      load();
    }
  }

  async function remove(p) {
    if (!confirm(`Hapus "${p.name}"?`)) return;
    const r = await api(`/products/${p.id}`, { method: 'DELETE' });
    if (r.ok) {
      toast('Produk dihapus');
      load();
    }
  }
</script>

<svelte:head><title>Produk — Admin PrintKu</title></svelte:head>

<div class="stack-xl">
  <div class="row-between wrap">
    <div class="stack-sm">
      <span class="eyebrow eyebrow-bar">Admin</span>
      <h1 class="display-md">Kelola Produk</h1>
    </div>
    <button class="btn btn-primary" type="button" onclick={openCreate}>
      <Plus size={18} aria-hidden="true" /> Tambah Produk
    </button>
  </div>

  <form class="row tool-row" onsubmit={(e) => { e.preventDefault(); load(); }}>
    <div class="search-box">
      <Search size={18} aria-hidden="true" />
      <label class="sr-only" for="cari-produk">Cari produk</label>
      <input id="cari-produk" class="input search-input" placeholder="Cari produk…" bind:value={search} />
    </div>
    <button class="btn btn-tertiary" type="submit">Cari</button>
  </form>

  {#if loading}
    <div class="empty-state" role="status"><span class="spinner"></span> Memuat…</div>
  {:else if products.length === 0}
    <div class="empty-state stack-sm">
      <Package size={28} strokeWidth={1.6} aria-hidden="true" />
      <span class="body-md-strong">Belum ada produk</span>
      <span class="body-sm text-body">Tambahkan produk pertama lewat tombol di atas.</span>
    </div>
  {:else}
    <div class="table-wrap">
      <table class="data-table">
        <caption class="sr-only">Daftar produk</caption>
        <thead>
          <tr>
            <th scope="col">Produk</th>
            <th scope="col">Kategori</th>
            <th scope="col">Harga</th>
            <th scope="col">Stok</th>
            <th scope="col">Status</th>
            <th scope="col"><span class="sr-only">Aksi</span></th>
          </tr>
        </thead>
        <tbody>
          {#each products as p (p.id)}
            <tr>
              <td>
                <div class="product-cell">
                  <span class="thumb-sm">
                    {#if p.image_url && !broken.includes(p.id)}
                      <img
                        src={p.image_url}
                        alt=""
                        loading="lazy"
                        decoding="async"
                        onerror={() => (broken = [...broken, p.id])}
                      />
                    {:else}
                      <span aria-hidden="true">{p.name[0]}</span>
                    {/if}
                  </span>
                  <span class="stack-xs">
                    <span class="body-sm-strong">{p.name}</span>
                    <span class="caption text-body-mid clamp">{p.description}</span>
                  </span>
                </div>
              </td>
              <td>{p.category_name}</td>
              <td class="body-sm-strong">{formatRupiah(p.price)}</td>
              <td class={p.stock === 0 ? 'body-sm-strong' : 'body-sm'}>
                {#if p.stock === 0}
                  <span class="badge badge-danger">Habis</span>
                {:else}
                  {p.stock}
                {/if}
              </td>
              <td>
                <span class="badge" class:badge-primary={!p.is_active}>
                  {p.is_active ? 'Aktif' : 'Nonaktif'}
                </span>
              </td>
              <td class="cell-action">
                <div class="row" style="justify-content: flex-end">
                  <button class="icon-btn" type="button" aria-label={`Edit ${p.name}`} onclick={() => openEdit(p)}>
                    <Pencil size={16} aria-hidden="true" />
                  </button>
                  <button class="icon-btn" type="button" aria-label={`Hapus ${p.name}`} onclick={() => remove(p)}>
                    <Trash2 size={16} aria-hidden="true" />
                  </button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

{#if modal}
  <Modal title={editing ? 'Edit Produk' : 'Tambah Produk'} onclose={() => (modal = false)}>
    <form class="stack-md" onsubmit={(e) => { e.preventDefault(); save(); }}>
      <label class="field">
        <span class="label-text">Nama produk</span>
        <input class="input" bind:value={form.name} required />
      </label>

      <div class="grid grid-2">
        <label class="field">
          <span class="label-text">Kategori</span>
          <select class="input" bind:value={form.category_id}>
            {#each categories as c (c.id)}
              <option value={c.id}>{c.name}</option>
            {/each}
          </select>
        </label>
        <label class="field">
          <span class="label-text">Stok</span>
          <input class="input" type="number" min="0" bind:value={form.stock} />
        </label>
      </div>

      <div class="grid grid-2">
        <label class="field">
          <span class="label-text">Harga (Rp)</span>
          <input class="input" type="number" min="0" bind:value={form.price} required />
        </label>
        <label class="field">
          <span class="label-text">Status</span>
          <select class="input" bind:value={form.is_active}>
            <option value={1}>Aktif</option>
            <option value={0}>Nonaktif</option>
          </select>
        </label>
      </div>

      <label class="field">
        <span class="label-text">Deskripsi</span>
        <textarea class="input" rows="3" bind:value={form.description}></textarea>
      </label>

      <label class="field">
        <span class="label-text">URL gambar (opsional)</span>
        <input class="input" bind:value={form.image_url} placeholder="/products/nama-produk.jpg" />
        <span class="caption text-body-mid">Contoh bawaan: /products/kertas-hvs-a4.jpg</span>
      </label>

      {#if form.image_url}
        <div class="preview">
          <img src={form.image_url} alt="Pratinjau gambar produk" />
        </div>
      {:else}
        <p class="caption text-body-mid preview-hint">
          <ImageOff size={14} aria-hidden="true" /> Tanpa gambar, kartu produk memakai stiker inisial.
        </p>
      {/if}

      <div class="row">
        <button class="btn btn-primary" type="submit" disabled={saving}>
          {saving ? 'Menyimpan…' : 'Simpan'}
        </button>
        <button class="btn btn-tertiary" type="button" onclick={() => (modal = false)}>Batal</button>
      </div>
    </form>
  </Modal>
{/if}

<style>
  .tool-row { flex-wrap: wrap; }
  .search-box { position: relative; flex: 1; max-width: 420px; }
  .search-box :global(svg) {
    position: absolute;
    left: var(--space-lg);
    top: 50%;
    transform: translateY(-50%);
    color: var(--body-mid);
  }
  .search-input { padding-left: var(--space-3xl); width: 100%; }
  .clamp { max-width: 260px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .cell-action { text-align: right; }
  .product-cell { display: flex; align-items: center; gap: var(--space-md); }
  .thumb-sm {
    width: 44px;
    height: 44px;
    flex-shrink: 0;
    display: grid;
    place-items: center;
    overflow: hidden;
    font-family: var(--font-display);
    font-weight: 600;
    background: var(--primary-soft);
    border: var(--border);
    border-radius: var(--radius-sm);
  }
  .thumb-sm img { width: 100%; height: 100%; object-fit: cover; }
  .preview {
    border: var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
    box-shadow: var(--shadow-xs);
  }
  .preview img { width: 100%; max-height: 220px; object-fit: cover; }
  .preview-hint { display: inline-flex; align-items: center; gap: var(--space-xs); }
</style>

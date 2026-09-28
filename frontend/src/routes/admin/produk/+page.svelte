<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import { formatRupiah } from '$lib/format.js';
  import { toast } from '$lib/stores.js';
  import { Plus, Pencil, Trash2, X, Search } from 'lucide-svelte';

  let products = $state([]);
  let categories = $state([]);
  let loading = $state(true);
  let search = $state('');
  let modal = $state(false);
  let editing = $state(null);
  let saving = $state(false);

  let form = $state({ category_id: '', name: '', description: '', price: 0, stock: 0, image_url: '', is_active: 1 });

  onMount(async () => {
    load();
    const c = await api('/categories');
    if (c.ok) categories = c.json.categories;
  });

  async function load() {
    const params = search.trim() ? `?search=${encodeURIComponent(search.trim())}` : '';
    const r = await api(`/products${params}`);
    loading = false;
    if (r.ok) products = r.json.products;
  }

  function openCreate() {
    editing = null;
    form = { category_id: categories[0]?.id || '', name: '', description: '', price: 0, stock: 0, image_url: '', is_active: 1 };
    modal = true;
  }

  function openEdit(p) {
    editing = p;
    form = { category_id: p.category_id, name: p.name, description: p.description, price: p.price, stock: p.stock, image_url: p.image_url, is_active: p.is_active };
    modal = true;
  }

  async function save() {
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
      <span class="eyebrow">Admin</span>
      <h1 class="display-md">Kelola Produk</h1>
    </div>
    <button class="btn btn-primary" onclick={openCreate}><Plus size={18} /> Tambah Produk</button>
  </div>

  <div class="row">
    <div class="search-box">
      <Search size={18} />
      <input class="input search-input" placeholder="Cari produk…" bind:value={search} onkeydown={(e) => e.key === 'Enter' && load()} />
    </div>
    <button class="btn btn-tertiary" onclick={load}>Cari</button>
  </div>

  {#if loading}
    <div class="empty-state"><span class="spinner"></span> Memuat…</div>
  {:else}
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>Produk</th>
            <th>Kategori</th>
            <th>Harga</th>
            <th>Stok</th>
            <th>Status</th>
            <th style="text-align: right">Aksi</th>
          </tr>
        </thead>
        <tbody>
          {#each products as p (p.id)}
            <tr>
              <td>
                <div class="stack-xs">
                  <span class="body-sm-strong">{p.name}</span>
                  <span class="caption text-body-mid clamp">{p.description}</span>
                </div>
              </td>
              <td>{p.category_name}</td>
              <td class="body-sm-strong">{formatRupiah(p.price)}</td>
              <td class={p.stock === 0 ? 'caption' : 'body-sm'} style={p.stock === 0 ? 'color: var(--primary)' : ''}>{p.stock}</td>
              <td>
                <span class="badge" class:badge-primary={!p.is_active}>{p.is_active ? 'Aktif' : 'Nonaktif'}</span>
              </td>
              <td>
                <div class="row" style="justify-content: flex-end">
                  <button class="icon-btn" title="Edit" onclick={() => openEdit(p)}><Pencil size={16} /></button>
                  <button class="icon-btn" title="Hapus" onclick={() => remove(p)}><Trash2 size={16} /></button>
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
    <div class="overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) modal = false; }}>
      <div class="modal card stack-md" role="dialog" aria-modal="true" tabindex="-1">
      <div class="row-between">
        <h3 class="display-sub-sm">{editing ? 'Edit Produk' : 'Tambah Produk'}</h3>
        <button class="icon-btn" onclick={() => (modal = false)}><X size={18} /></button>
      </div>
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
      <div class="row">
        <button class="btn btn-primary" onclick={save} disabled={saving}>{saving ? 'Menyimpan…' : 'Simpan'}</button>
        <button class="btn btn-tertiary" onclick={() => (modal = false)}>Batal</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .search-box { position: relative; flex: 1; max-width: 420px; }
  .search-box :global(svg) { position: absolute; left: var(--space-lg); top: 50%; transform: translateY(-50%); color: var(--body-mid); }
  .search-input { padding-left: var(--space-3xl); }
  .stack-xs { display: flex; flex-direction: column; gap: 2px; }
  .clamp { max-width: 260px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  /* .icon-btn, .overlay, .modal sekarang global di app.css */
</style>
<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import { formatRupiah } from '$lib/format.js';
  import { toast } from '$lib/stores.js';
  import { Plus, Pencil, Trash2, X, Ticket } from 'lucide-svelte';

  let vouchers = $state([]);
  let loading = $state(true);
  let modal = $state(false);
  let editing = $state(null);
  let saving = $state(false);

  const empty = () => ({
    code: '',
    description: '',
    discount_type: 'percent',
    discount_value: 10,
    min_spend: 0,
    valid_until: '', // YYYY-MM-DD (kosong = tanpa batas)
    usage_limit: 0,
    is_active: 1,
  });
  let form = $state(empty());

  onMount(load);

  async function load() {
    const r = await api('/admin/vouchers');
    loading = false;
    if (r.ok) vouchers = r.json.vouchers;
  }

  function openCreate() {
    editing = null;
    form = empty();
    modal = true;
  }

  function openEdit(v) {
    editing = v;
    form = {
      code: v.code,
      description: v.description || '',
      discount_type: v.discount_type,
      discount_value: v.discount_value,
      min_spend: v.min_spend,
      valid_until: v.valid_until ? String(v.valid_until).slice(0, 10) : '',
      usage_limit: v.usage_limit,
      is_active: v.is_active,
    };
    modal = true;
  }

  async function save() {
    if (!form.code.trim()) {
      toast('Kode voucher wajib diisi', 'error');
      return;
    }
    saving = true;
    // date-only -> RFC3339 akhir hari agar perbandingan expiry di server konsisten
    const payload = {
      ...form,
      code: form.code.trim().toUpperCase(),
      discount_value: Number(form.discount_value),
      min_spend: Number(form.min_spend),
      usage_limit: Number(form.usage_limit),
      is_active: Number(form.is_active),
      valid_until: form.valid_until ? `${form.valid_until}T23:59:59Z` : '',
    };
    const r = editing
      ? await api(`/admin/vouchers/${editing.id}`, { method: 'PATCH', body: payload })
      : await api('/admin/vouchers', { method: 'POST', body: payload });
    saving = false;
    if (r.ok) {
      toast(editing ? 'Voucher diperbarui' : 'Voucher ditambahkan', 'success');
      modal = false;
      load();
    }
  }

  async function remove(v) {
    if (!confirm(`Hapus voucher "${v.code}"?`)) return;
    const r = await api(`/admin/vouchers/${v.id}`, { method: 'DELETE' });
    if (r.ok) {
      toast('Voucher dihapus');
      load();
    }
  }

  function discountLabel(v) {
    return v.discount_type === 'percent' ? `${v.discount_value}%` : formatRupiah(v.discount_value);
  }
</script>

<svelte:head><title>Voucher — Admin PrintMart</title></svelte:head>

<div class="stack-xl">
  <div class="row-between wrap">
    <div class="stack-sm">
      <span class="eyebrow">Admin</span>
      <h1 class="display-md">Kelola Voucher</h1>
    </div>
    <button class="btn btn-primary" onclick={openCreate}><Plus size={18} /> Tambah Voucher</button>
  </div>

  {#if loading}
    <div class="empty-state"><span class="spinner"></span> Memuat…</div>
  {:else if vouchers.length === 0}
    <div class="empty-state body-md"><Ticket size={18} /> Belum ada voucher. Tambahkan yang pertama.</div>
  {:else}
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>Kode</th>
            <th>Diskon</th>
            <th>Min. Belanja</th>
            <th>Pemakaian</th>
            <th>Berlaku s/d</th>
            <th>Status</th>
            <th style="text-align: right">Aksi</th>
          </tr>
        </thead>
        <tbody>
          {#each vouchers as v (v.id)}
            <tr>
              <td>
                <div class="stack-xs">
                  <span class="body-sm-strong">{v.code}</span>
                  <span class="caption text-body-mid clamp">{v.description}</span>
                </div>
              </td>
              <td class="body-sm-strong">{discountLabel(v)}</td>
              <td class="body-sm">{v.min_spend > 0 ? formatRupiah(v.min_spend) : '—'}</td>
              <td class="body-sm">{v.used_count}{v.usage_limit > 0 ? ` / ${v.usage_limit}` : ''}</td>
              <td class="body-sm">{v.valid_until ? String(v.valid_until).slice(0, 10) : '∞'}</td>
              <td>
                <span class="badge" class:badge-primary={!v.is_active}>{v.is_active ? 'Aktif' : 'Nonaktif'}</span>
              </td>
              <td>
                <div class="row" style="justify-content: flex-end">
                  <button class="icon-btn" title="Edit" onclick={() => openEdit(v)}><Pencil size={16} /></button>
                  <button class="icon-btn" title="Hapus" onclick={() => remove(v)}><Trash2 size={16} /></button>
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
  <div class="overlay" role="presentation" onclick={() => (modal = false)}>
    <div class="modal card stack-md" role="dialog" aria-modal="true" onclick={(e) => e.stopPropagation()}>
      <div class="row-between">
        <h3 class="display-sub-sm">{editing ? 'Edit Voucher' : 'Tambah Voucher'}</h3>
        <button class="icon-btn" onclick={() => (modal = false)}><X size={18} /></button>
      </div>
      <label class="field">
        <span class="label-text">Kode voucher</span>
        <input class="input" bind:value={form.code} placeholder="HEMAT10" style="text-transform: uppercase" disabled={!!editing} required />
        {#if editing}<span class="caption text-body-mid">Kode tidak dapat diubah.</span>{/if}
      </label>
      <label class="field">
        <span class="label-text">Deskripsi</span>
        <input class="input" bind:value={form.description} placeholder="Diskon 10% untuk semua pesanan" />
      </label>
      <div class="grid grid-2">
        <label class="field">
          <span class="label-text">Tipe diskon</span>
          <select class="input" bind:value={form.discount_type}>
            <option value="percent">Persentase (%)</option>
            <option value="fixed">Nominal (Rp)</option>
          </select>
        </label>
        <label class="field">
          <span class="label-text">Nilai diskon {form.discount_type === 'percent' ? '(%)' : '(Rp)'}</span>
          <input class="input" type="number" min="1" bind:value={form.discount_value} required />
        </label>
      </div>
      <div class="grid grid-2">
        <label class="field">
          <span class="label-text">Min. belanja (Rp, 0 = bebas)</span>
          <input class="input" type="number" min="0" bind:value={form.min_spend} />
        </label>
        <label class="field">
          <span class="label-text">Kuota pemakaian (0 = tak terbatas)</span>
          <input class="input" type="number" min="0" bind:value={form.usage_limit} />
        </label>
      </div>
      <div class="grid grid-2">
        <label class="field">
          <span class="label-text">Berlaku s/d (opsional)</span>
          <input class="input" type="date" bind:value={form.valid_until} />
        </label>
        <label class="field">
          <span class="label-text">Status</span>
          <select class="input" bind:value={form.is_active}>
            <option value={1}>Aktif</option>
            <option value={0}>Nonaktif</option>
          </select>
        </label>
      </div>
      <div class="row">
        <button class="btn btn-primary" onclick={save} disabled={saving}>{saving ? 'Menyimpan…' : 'Simpan'}</button>
        <button class="btn btn-tertiary" onclick={() => (modal = false)}>Batal</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .stack-xs { display: flex; flex-direction: column; gap: 2px; }
  .clamp { max-width: 260px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    border: 1px solid var(--mute);
    background: var(--canvas);
    border-radius: var(--radius-sm);
    color: var(--ink);
  }
  .icon-btn:hover { border-color: var(--ink); }
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(32, 21, 21, 0.45);
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding: var(--space-4xl) var(--space-xl);
    z-index: 80;
    overflow-y: auto;
  }
  .modal { width: 100%; max-width: 560px; box-shadow: var(--shadow-soft); }
</style>

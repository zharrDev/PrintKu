<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import { formatRupiah } from '$lib/format.js';
  import { toast } from '$lib/stores.js';
  import Modal from '$lib/components/Modal.svelte';
  import { Plus, Pencil, Trash2, Ticket } from 'lucide-svelte';

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
    loading = true;
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

<svelte:head><title>Voucher — Admin PrintKu</title></svelte:head>

<div class="stack-xl">
  <div class="row-between wrap">
    <div class="stack-sm">
      <span class="eyebrow eyebrow-bar">Admin</span>
      <h1 class="display-md">Kelola Voucher</h1>
    </div>
    <button class="btn btn-primary" type="button" onclick={openCreate}>
      <Plus size={18} aria-hidden="true" /> Tambah Voucher
    </button>
  </div>

  {#if loading}
    <div class="empty-state" role="status"><span class="spinner"></span> Memuat…</div>
  {:else if vouchers.length === 0}
    <div class="empty-state stack-sm">
      <Ticket size={28} strokeWidth={1.6} aria-hidden="true" />
      <span class="body-md-strong">Belum ada voucher</span>
      <span class="body-sm text-body">Tambahkan yang pertama lewat tombol di atas.</span>
    </div>
  {:else}
    <div class="table-wrap">
      <table class="data-table">
        <caption class="sr-only">Daftar voucher</caption>
        <thead>
          <tr>
            <th scope="col">Kode</th>
            <th scope="col">Diskon</th>
            <th scope="col">Min. Belanja</th>
            <th scope="col">Pemakaian</th>
            <th scope="col">Berlaku s/d</th>
            <th scope="col">Status</th>
            <th scope="col"><span class="sr-only">Aksi</span></th>
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
              <td class="body-sm">{v.valid_until ? String(v.valid_until).slice(0, 10) : '—'}</td>
              <td>
                <span class="badge" class:badge-primary={!v.is_active}>
                  {v.is_active ? 'Aktif' : 'Nonaktif'}
                </span>
              </td>
              <td class="cell-action">
                <div class="row" style="justify-content: flex-end">
                  <button class="icon-btn" type="button" aria-label={`Edit voucher ${v.code}`} onclick={() => openEdit(v)}>
                    <Pencil size={16} aria-hidden="true" />
                  </button>
                  <button class="icon-btn" type="button" aria-label={`Hapus voucher ${v.code}`} onclick={() => remove(v)}>
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
  <Modal title={editing ? 'Edit Voucher' : 'Tambah Voucher'} onclose={() => (modal = false)}>
    <form class="stack-md" onsubmit={(e) => { e.preventDefault(); save(); }}>
      <label class="field">
        <span class="label-text">Kode voucher</span>
        <input
          class="input"
          bind:value={form.code}
          placeholder="HEMAT10"
          style="text-transform: uppercase"
          disabled={!!editing}
          required
        />
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
        <button class="btn btn-primary" type="submit" disabled={saving}>
          {saving ? 'Menyimpan…' : 'Simpan'}
        </button>
        <button class="btn btn-tertiary" type="button" onclick={() => (modal = false)}>Batal</button>
      </div>
    </form>
  </Modal>
{/if}

<style>
  .clamp { max-width: 260px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .cell-action { text-align: right; }
</style>

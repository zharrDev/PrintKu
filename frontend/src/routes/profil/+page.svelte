<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api.js';
  import { userStore, clearSession, toast } from '$lib/stores.js';
  import { formatDate } from '$lib/format.js';
  import { MapPin, Plus, LogOut, Trash2 } from 'lucide-svelte';

  let user = $state($userStore);
  let addresses = $state([]);
  let saving = $state(false);

  let form = $state({ label: 'Rumah', full_address: '', city: '', postal_code: '' });

  onMount(async () => {
    if (!user) {
      goto('/masuk');
      return;
    }
    const r = await api('/auth/addresses');
    if (r.ok) addresses = r.json.addresses;
  });

  async function addAddress() {
    saving = true;
    const r = await api('/auth/addresses', { method: 'POST', body: form });
    saving = false;
    if (r.ok) {
      addresses = [r.json.address, ...addresses];
      form = { label: 'Rumah', full_address: '', city: '', postal_code: '' };
      toast('Alamat tersimpan', 'success');
    }
  }

  async function deleteAddress(a) {
    if (!confirm(`Hapus alamat "${a.label}"?`)) return;
    const r = await api(`/auth/addresses/${a.id}`, { method: 'DELETE' });
    if (r.ok) {
      addresses = addresses.filter((x) => x.id !== a.id);
      toast('Alamat dihapus');
    }
  }

  function logout() {
    clearSession();
    toast('Sampai jumpa lagi!');
    goto('/');
  }
</script>

<svelte:head><title>Profil — PrintKu</title></svelte:head>

<section class="band-hero">
  <div class="container stack-xl">
    <div class="row-between wrap">
      <div class="stack-sm">
        <span class="eyebrow">Profil</span>
        <h1 class="display-lg">Halo, {user?.name?.split(' ')[0]}</h1>
      </div>
      <button class="btn btn-tertiary" onclick={logout}><LogOut size={16} /> Keluar</button>
    </div>

    {#if user}
      <div class="card stack-md">
        <h3 class="display-sub-sm">Data akun</h3>
        <div class="grid grid-2">
          <div class="stack-sm">
            <span class="caption text-body-mid">Nama</span>
            <span class="body-md-strong">{user.name}</span>
          </div>
          <div class="stack-sm">
            <span class="caption text-body-mid">Email</span>
            <span class="body-md">{user.email}</span>
          </div>
          <div class="stack-sm">
            <span class="caption text-body-mid">No. HP</span>
            <span class="body-md">{user.phone || '-'}</span>
          </div>
          <div class="stack-sm">
            <span class="caption text-body-mid">Terdaftar sejak</span>
            <span class="body-md">{formatDate(user.created_at)}</span>
          </div>
        </div>
      </div>

      <div class="card stack-md">
        <div class="row">
          <MapPin size={20} />
          <h3 class="display-sub-sm">Alamat pengiriman</h3>
        </div>
        {#if addresses.length === 0}
          <p class="body-sm text-body-mid">Belum ada alamat tersimpan.</p>
        {:else}
          {#each addresses as a (a.id)}
            <div class="row-between address-row">
              <div class="stack-sm">
                <span class="body-md-strong">{a.label}</span>
                <span class="body-sm text-body">{a.full_address}, {a.city} {a.postal_code}</span>
              </div>
              <button class="icon-btn" title="Hapus alamat" onclick={() => deleteAddress(a)}><Trash2 size={16} /></button>
            </div>
          {/each}
        {/if}
        <hr class="divider" />
        <form class="stack-md" onsubmit={(e) => { e.preventDefault(); addAddress(); }}>
          <div class="grid grid-2">
            <label class="field">
              <span class="label-text">Label</span>
              <input class="input" bind:value={form.label} placeholder="Rumah / Kantor" />
            </label>
            <label class="field">
              <span class="label-text">Kota</span>
              <input class="input" required bind:value={form.city} placeholder="Bandung" />
            </label>
          </div>
          <label class="field">
            <span class="label-text">Alamat lengkap</span>
            <input class="input" required bind:value={form.full_address} placeholder="Jl. Merdeka No. 88, RT 01 RW 02" />
          </label>
          <label class="field">
            <span class="label-text">Kode pos (opsional)</span>
            <input class="input" bind:value={form.postal_code} placeholder="40111" />
          </label>
          <button class="btn btn-secondary" style="align-self: flex-start" disabled={saving}><Plus size={16} /> Tambah Alamat</button>
        </form>
      </div>
    {/if}
  </div>
</section>

<style>
  .address-row { border-bottom: 1px solid var(--mute); padding-bottom: var(--space-md); }
  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    flex-shrink: 0;
    border: 1px solid var(--mute);
    background: var(--canvas);
    border-radius: var(--radius-sm);
    color: var(--ink);
  }
  .icon-btn:hover { border-color: var(--primary); color: var(--primary); }
</style>
<script>
  import { api } from '$lib/api.js';
  import { setSession, toast } from '$lib/stores.js';
  import { goto } from '$app/navigation';
  import { UserPlus } from 'lucide-svelte';

  let name = $state('');
  let email = $state('');
  let phone = $state('');
  let password = $state('');
  let loading = $state(false);
  let error = $state('');

  async function submit() {
    error = '';
    loading = true;
    const r = await api('/auth/register', { method: 'POST', body: { name, email, phone, password } });
    loading = false;
    if (r.ok) {
      setSession(r.json.token, r.json.user);
      toast(`Halo, ${r.json.user.name}! Akun berhasil dibuat`, 'success');
      goto('/');
    } else {
      error = r.json?.error || 'Gagal daftar';
    }
  }
</script>

<svelte:head><title>Daftar — PrintKu</title></svelte:head>

<section class="band-hero">
  <div class="container">
    <div class="auth-wrap">
      <div class="auth-card card stack-lg">
        <div class="stack-sm">
          <span class="eyebrow">Akun Baru</span>
          <h1 class="display-md">Daftar gratis</h1>
          <p class="body-sm text-body">Satu akun untuk belanja ATK dan jasa print.</p>
        </div>
        <form class="stack-md" onsubmit={(e) => { e.preventDefault(); submit(); }}>
          <label class="field">
            <span class="label-text">Nama lengkap</span>
            <input class="input" required placeholder="Budi Santoso" bind:value={name} />
          </label>
          <label class="field">
            <span class="label-text">Email</span>
            <input class="input" type="email" required placeholder="nama@email.com" bind:value={email} />
          </label>
          <label class="field">
            <span class="label-text">No. HP (opsional)</span>
            <input class="input" placeholder="0812-3456-7890" bind:value={phone} />
          </label>
          <label class="field">
            <span class="label-text">Password</span>
            <input class="input" type="password" required minlength="6" placeholder="Minimal 6 karakter" bind:value={password} />
          </label>
          {#if error}<p class="caption" style="color: var(--primary)">{error}</p>{/if}
          <button class="btn btn-primary" disabled={loading}><UserPlus size={16} /> {loading ? 'Mendaftar…' : 'Daftar'}</button>
        </form>
        <hr class="divider" />
        <p class="body-sm text-body text-center">
          Sudah punya akun? <a href="/masuk" class="muted-link" style="font-weight: 600">Masuk di sini</a>
        </p>
      </div>
    </div>
  </div>
</section>

<style>
  .auth-wrap { max-width: 440px; margin: 0 auto; }
  .auth-card { box-shadow: var(--shadow-soft); }
</style>
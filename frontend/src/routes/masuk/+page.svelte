<script>
  import { api } from '$lib/api.js';
  import { setSession, toast } from '$lib/stores.js';
  import { goto } from '$app/navigation';
  import { Lock } from 'lucide-svelte';

  let email = $state('');
  let password = $state('');
  let loading = $state(false);
  let error = $state('');

  async function submit() {
    error = '';
    loading = true;
    const r = await api('/auth/login', { method: 'POST', body: { email, password } });
    loading = false;
    if (r.ok) {
      setSession(r.json.token, r.json.user);
      toast(`Selamat datang kembali, ${r.json.user.name}!`, 'success');
      goto(r.json.user.role === 'admin' ? '/admin' : '/');
    } else {
      error = r.json?.error || 'Gagal masuk';
    }
  }
</script>

<svelte:head><title>Masuk — PrintKu</title></svelte:head>

<section class="band-hero">
  <div class="container">
    <div class="auth-wrap">
      <div class="auth-card card stack-lg">
        <div class="stack-sm">
          <span class="eyebrow">Akun</span>
          <h1 class="display-md">Masuk ke PrintKu</h1>
          <p class="body-sm text-body">Lanjutkan belanja atau cek status print-mu.</p>
        </div>
        <form class="stack-md" onsubmit={(e) => { e.preventDefault(); submit(); }}>
          <label class="field">
            <span class="label-text">Email</span>
            <input class="input" type="email" required placeholder="nama@email.com" bind:value={email} />
          </label>
          <label class="field">
            <span class="label-text">Password</span>
            <input class="input" type="password" required placeholder="••••••••" bind:value={password} />
          </label>
          {#if error}<p class="caption" style="color: var(--primary)">{error}</p>{/if}
          <button class="btn btn-primary" disabled={loading}><Lock size={16} /> {loading ? 'Memeriksa…' : 'Masuk'}</button>
        </form>
        <hr class="divider" />
        <p class="body-sm text-body text-center">
          Belum punya akun? <a href="/daftar" class="muted-link" style="font-weight: 600">Daftar gratis</a>
        </p>
      </div>
    </div>
  </div>
</section>

<style>
  .auth-wrap { max-width: 440px; margin: 0 auto; }
  .auth-card { box-shadow: var(--shadow-soft); }
</style>
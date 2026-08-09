<script>
  import { onMount } from 'svelte';
  import { Printer, BookCopy, Layers, Scan, ArrowRight, Upload, PackageCheck, Clock, ShieldCheck } from 'lucide-svelte';
  import { api } from '$lib/api.js';
  import ProductCard from '$lib/components/ProductCard.svelte';
  import { formatRupiah } from '$lib/format.js';

  let featured = $state([]);

  onMount(async () => {
    const r = await api('/products?sort=price_asc');
    if (r.ok) featured = r.json.products.slice(0, 4);
  });

  const services = [
    { icon: Printer, title: 'Print & Fotocopy', desc: 'Upload file (PDF, DOCX, JPG, PNG), sistem hitung harga otomatis per halaman.', href: '/print' },
    { icon: BookCopy, title: 'Jilid', desc: 'Spiral, lakban, hingga hardcover untuk skripsi dan laporan.', href: '/print' },
    { icon: Layers, title: 'Laminating', desc: 'Berbagai ukuran: A4, A3, sampai kartu — dokumen awet dan rapi.', href: '/print' },
    { icon: Scan, title: 'Scan Dokumen', desc: 'Scan fisik ke PDF dan siap dikirim ke email kamu.', href: '/print' },
  ];

  const pricing = [
    { name: 'Print Hitam Putih', price: 500, unit: '/ halaman', desc: 'A4 mulai dari Rp 500', featured: false },
    { name: 'Print Warna', price: 1500, unit: '/ halaman', desc: 'A4 mulai dari Rp 1.500', featured: true },
    { name: 'Fotocopy', price: 300, unit: '/ halaman', desc: 'A4 mulai dari Rp 300', featured: false },
  ];

  const steps = [
    { icon: Upload, title: 'Upload file', desc: 'PDF, DOCX, JPG, atau PNG — maksimal 20MB.' },
    { icon: PackageCheck, title: 'Pilih spesifikasi', desc: 'Warna, ukuran kertas, rangkap, dan finishing.' },
    { icon: Clock, title: 'Ambil atau diantar', desc: 'Lacak status pesanan secara real-time.' },
  ];
</script>

<svelte:head>
  <title>PrintMart — Toko ATK & Jasa Fotocopy Online</title>
  <meta
    name="description"
    content="Toko alat tulis kantor & jasa print, fotocopy, jilid, laminating online. Upload file dan hitung harga otomatis."
  />
</svelte:head>

<!-- HERO -->
<section class="band-hero">
  <div class="container hero-grid">
    <div class="stack-xl hero-copy">
      <span class="eyebrow">Toko ATK & Jasa Cetak Online</span>
      <h1 class="display-xl">Semua kebutuhan tulis & cetak, satu tempat.</h1>
      <p class="body-lg text-body">
        Belanja alat tulis kantor, sekolah, dan kertas — atau upload file untuk
        print, jilid, dan laminating. Harga dihitung otomatis, pesanan bisa
        diambil di toko atau diantar ke rumah.
      </p>
      <div class="row wrap">
        <a href="/katalog" class="btn btn-primary">Mulai Belanja <ArrowRight size={18} /></a>
        <a href="/print" class="btn btn-secondary">Upload File Print <Upload size={18} /></a>
      </div>
      <div class="row wrap hero-badges">
        <span class="badge"><ShieldCheck size={14} /> Pembayaran sandbox (uji coba)</span>
        <span class="badge">Tracking real-time</span>
        <span class="badge">Ambil di toko / antar</span>
      </div>
    </div>
    <div class="hero-visual">
      <div class="mock-card">
        <span class="eyebrow">Ringkasan Print</span>
        <h3 class="display-sub-sm" style="margin: 8px 0">makalah-sistem.pdf</h3>
        <div class="mock-row">
          <span class="caption text-body-mid">4 halaman · Hitam Putih · A4</span>
        </div>
        <div class="mock-row row-between">
          <span class="body-md">Estimasi</span>
          <span class="body-md-strong" style="color: var(--primary)">Rp 2.000</span>
        </div>
        <button class="btn btn-primary" disabled style="width: 100%">Lanjut ke Pembayaran</button>
        <span class="caption text-body-mid" style="margin-top: 8px">Status: siap diambil dalam 30 menit</span>
      </div>
    </div>
  </div>
</section>

<!-- SERVICES -->
<section class="band-cream">
  <div class="container stack-xl">
    <div class="row-between wrap">
      <div class="stack-sm">
        <span class="eyebrow">Jasa Cetak</span>
        <h2 class="display-lg">Layanan yang kami siapkan untukmu</h2>
      </div>
    </div>
    <div class="grid grid-4">
      {#each services as s (s.title)}
        <a href={s.href} class="card stack-md service-card">
          <s.icon size={28} strokeWidth={1.75} />
          <h3 class="display-sub-sm">{s.title}</h3>
          <p class="body-sm text-body">{s.desc}</p>
        </a>
      {/each}
    </div>
  </div>
</section>

<!-- PIPELINE -->
<section class="band-hero">
  <div class="container stack-xl">
    <span class="eyebrow">Cara Kerja</span>
    <h2 class="display-lg">Dari file jadi pesanan, kurang dari satu menit</h2>
    <div class="grid grid-3">
      {#each steps as st, i (st.title)}
        <div class="pricing-card stack-md">
          <div class="row">
            <span class="badge badge-primary">{i + 1}</span>
            <st.icon size={24} strokeWidth={1.75} />
          </div>
          <h3 class="display-sub-sm">{st.title}</h3>
          <p class="body-sm text-body">{st.desc}</p>
        </div>
      {/each}
    </div>
  </div>
</section>

<!-- FEATURED PRODUCTS -->
<section class="band-cream">
  <div class="container stack-xl">
    <div class="row-between wrap">
      <div class="stack-sm">
        <span class="eyebrow">Katalog</span>
        <h2 class="display-lg">Produk pilihan minggu ini</h2>
      </div>
      <a href="/katalog" class="btn btn-tertiary">Lihat Semua <ArrowRight size={18} /></a>
    </div>
    {#if featured.length}
      <div class="grid grid-4">
        {#each featured as p (p.id)}
          <ProductCard {p} />
        {/each}
      </div>
    {:else}
      <div class="empty-state body-md">Memuat produk…</div>
    {/if}
  </div>
</section>

<!-- PRICING -->
<section class="band-hero">
  <div class="container stack-xl">
    <div class="stack-sm text-center" style="align-items: center">
      <span class="eyebrow">Tarif Cetak</span>
      <h2 class="display-lg">Harga transparan, tanpa biaya tersembunyi</h2>
    </div>
    <div class="grid grid-3">
      {#each pricing as pc (pc.name)}
        <div class={pc.featured ? 'pricing-card-featured stack-md' : 'pricing-card stack-md'}>
          <span class={pc.featured ? 'eyebrow on-primary' : 'eyebrow'}>{pc.name}</span>
          <div class="row">
            <span class={pc.featured ? 'display-md on-primary' : 'display-md'}>{formatRupiah(pc.price)}</span>
            <span class={pc.featured ? 'caption on-primary' : 'caption'}>{pc.unit}</span>
          </div>
          <p class={pc.featured ? 'body-sm on-primary' : 'body-sm'}>{pc.desc}</p>
          <a href="/print" class={pc.featured ? 'btn btn-primary' : 'btn btn-tertiary'} style="align-self: flex-start">Mulai Cetak</a>
        </div>
      {/each}
    </div>
  </div>
</section>

<!-- CTA DARK -->
<section class="band-dark">
  <div class="container stack-lg text-center" style="align-items: center">
    <h2 class="display-lg">Punya file yang harus dicetak hari ini?</h2>
    <p class="body-lg" style="color: var(--canvas-soft); max-width: 640px">
      Upload sekarang, dapatkan estimasi harga instan, dan status pesanan
      yang bisa kamu pantau real-time sampai siap diambil.
    </p>
    <a href="/print" class="btn btn-primary" style="font-size: 20px; padding: var(--space-lg) var(--space-3xl)">
      Upload File Sekarang <Upload size={20} />
    </a>
  </div>
</section>

<style>
  .hero-grid {
    display: grid;
    grid-template-columns: 1.15fr 0.85fr;
    gap: var(--space-3xl);
    align-items: center;
  }
  .hero-copy h1 { margin-bottom: var(--space-lg); }
  .hero-badges { gap: var(--space-sm); }
  .hero-visual { display: flex; justify-content: center; }
  .mock-card {
    width: 100%;
    max-width: 380px;
    background: var(--canvas-soft);
    border-radius: var(--radius-md);
    padding: var(--space-xl);
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    box-shadow: var(--shadow-soft);
    transform: rotate(1.5deg);
  }
  .mock-row { padding: var(--space-sm) 0; border-bottom: 1px solid var(--mute); }
  .service-card { transition: transform 0.15s ease, box-shadow 0.15s ease; }
  .service-card:hover { transform: translateY(-3px); box-shadow: var(--shadow-soft); }
  .strong-price { color: var(--ink); }
  .on-primary { color: var(--on-primary) !important; }
  @media (max-width: 1023px) {
    .hero-grid { grid-template-columns: 1fr; }
  }
</style>
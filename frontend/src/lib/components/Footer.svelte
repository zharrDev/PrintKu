<footer class="footer">
  <div class="container">
    <div class="footer-grid">
      <div class="stack-md">
        <a href="/" class="brand">
          <span class="brand-mark">P</span>
          <span class="brand-name">PrintMart</span>
        </a>
        <p class="body-sm">
          Toko ATK & jasa fotocopy online. Upload file, pilih spesifikasi,
          dan pesanan siap diambil atau diantar.
        </p>
      </div>
      <div class="stack-md">
        <h4 class="footer-title">Belanja</h4>
        <a href="/katalog" class="footer-link">Katalog Produk</a>
        <a href="/print" class="footer-link">Jasa Print & Fotocopy</a>
        <a href="/keranjang" class="footer-link">Keranjang</a>
      </div>
      <div class="stack-md">
        <h4 class="footer-title">Akun</h4>
        <a href="/masuk" class="footer-link">Masuk</a>
        <a href="/daftar" class="footer-link">Daftar</a>
        <a href="/pesanan" class="footer-link">Pesanan Saya</a>
      </div>
      <div class="stack-md">
        <h4 class="footer-title">Kontak</h4>
        <span class="footer-link">Jl. Merdeka No. 88, Bandung</span>
        <span class="footer-link">0812-3456-7890</span>
        <span class="footer-link">halo@printmart.id</span>
      </div>
    </div>
    <div class="footer-bottom">
      <span class="caption">© {new Date().getFullYear()} PrintMart — Toko ATK & Jasa Fotocopy Online</span>
      <span class="caption">Pembayaran: sandbox mode (midtrans-sandbox)</span>
    </div>
  </div>
</footer>

<style>
  .footer {
    background: var(--ink);
    color: var(--canvas-soft);
    padding: var(--space-3xl) var(--space-xl);
    margin-top: var(--space-4xl);
  }
  .footer-grid {
    display: grid;
    grid-template-columns: 2fr 1fr 1fr 1fr;
    gap: var(--space-2xl);
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: var(--space-sm);
    color: var(--canvas-soft);
  }
  .brand-mark {
    width: 32px;
    height: 32px;
    border-radius: var(--radius-sm);
    background: var(--primary);
    color: var(--on-primary);
    font-weight: 700;
    font-size: 18px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  .brand-name { font-size: 20px; font-weight: 600; }
  .footer-title {
    color: var(--canvas-soft);
    font-size: 14px;
    font-weight: 600;
    letter-spacing: 1px;
    text-transform: uppercase;
    margin-bottom: var(--space-sm);
  }
  .footer-link { color: var(--canvas-soft); opacity: 0.8; font-size: 16px; line-height: 24px; }
  a.footer-link:hover { opacity: 1; color: var(--primary); }
  .footer-bottom {
    display: flex;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: var(--space-md);
    border-top: 1px solid rgba(248, 244, 240, 0.15);
    margin-top: var(--space-3xl);
    padding-top: var(--space-xl);
  }
  @media (max-width: 767px) {
    .footer-grid { grid-template-columns: 1fr; }
  }
</style>
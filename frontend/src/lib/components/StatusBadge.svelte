<script>
  import { statusInfo, PRINT_STATUS, ORDER_STATUS } from '$lib/format.js';

  let { map, status, tone = 'cream' } = $props();

  // Warna semantik ikut status (bukan hanya warna — label teks selalu
  // ikut tampil, jadi status tidak bergantung pada warna saja).
  const TONE_BY_STATUS = {
    menunggu_pembayaran: 'warning',
    dibayar: 'info',
    diproses: 'info',
    siap: 'success',
    selesai: 'success',
    menunggu: 'warning',
    dibatalkan: 'danger',
  };

  let resolved = $derived(
    tone !== 'cream' ? tone : TONE_BY_STATUS[status] ?? 'cream'
  );
  let info = $derived(map ? statusInfo(map, status) : null);
</script>

<span class="badge badge-{resolved}">
  {#if resolved === 'primary' || resolved === 'dark'}<span class="dot"></span>{/if}
  {info ? info.label : status}
</span>

<style>
  /* Semua warna comes dari app.css (.badge-success/-warning/-danger/-info).
     Teks di atas pastel selalu --ink. */
</style>
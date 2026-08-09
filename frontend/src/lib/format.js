export function formatRupiah(value) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(Number(value || 0));
}

export function formatDate(iso) {
  if (!iso) return '-';
  return new Date(iso).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function shortId(id) {
  return id ? id.slice(0, 8).toUpperCase() : '';
}

export const ORDER_STATUS = {
  menunggu_pembayaran: { label: 'Menunggu Pembayaran', step: 0 },
  dibayar: { label: 'Dibayar', step: 1 },
  diproses: { label: 'Diproses', step: 2 },
  siap: { label: 'Siap Diambil / Dikirim', step: 3 },
  selesai: { label: 'Selesai', step: 4 },
  dibatalkan: { label: 'Dibatalkan', step: -1 },
};

export const PRINT_STATUS = {
  menunggu: { label: 'Menunggu', step: 0 },
  diproses: { label: 'Diproses', step: 1 },
  siap: { label: 'Siap Diambil / Dikirim', step: 2 },
  selesai: { label: 'Selesai', step: 3 },
  dibatalkan: { label: 'Dibatalkan', step: -1 },
};

export const PAPER_SIZE_NAME = { a4: 'A4', f4: 'F4 (Folio)', a3: 'A3' };
export const COLOR_MODE_NAME = { bw: 'Hitam Putih', color: 'Warna' };

export function statusInfo(map, key) {
  return map[key] || { label: key, step: -99 };
}
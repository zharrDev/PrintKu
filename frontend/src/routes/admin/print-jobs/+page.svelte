<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import { formatRupiah, formatDate, PRINT_STATUS, PAPER_SIZE_NAME, COLOR_MODE_NAME } from '$lib/format.js';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { toast } from '$lib/stores.js';
  import { Printer } from 'lucide-svelte';

  let jobs = $state([]);
  let loading = $state(true);
  let filter = $state('');

  const FLOW = [
    { value: 'menunggu', label: 'Menunggu' },
    { value: 'diproses', label: 'Diproses' },
    { value: 'siap', label: 'Siap' },
    { value: 'selesai', label: 'Selesai' },
    { value: 'dibatalkan', label: 'Dibatalkan' },
  ];

  onMount(load);

  async function load() {
    loading = true;
    const r = await api('/admin/print-jobs');
    loading = false;
    if (r.ok) {
      jobs = filter
        ? r.json.printJobs.filter((j) => j.status === filter)
        : r.json.printJobs;
    }
  }

  async function setStatus(job, status) {
    const r = await api(`/print-jobs/${job.id}/status`, { method: 'PATCH', body: { status } });
    if (r.ok) {
      toast(`Print "${job.file_name}" → ${FLOW.find((f) => f.value === status)?.label}`, 'success');
      load();
    }
  }

  function nextJob(job) {
    if (job.status === 'menunggu') return 'diproses';
    if (job.status === 'diproses') return 'siap';
    if (job.status === 'siap') return 'selesai';
    return null;
  }
</script>

<svelte:head><title>Print Jobs — Admin PrintKu</title></svelte:head>

<div class="stack-xl">
  <div class="stack-sm">
    <span class="eyebrow eyebrow-bar">Admin</span>
    <h1 class="display-md">Antrian Print Jobs</h1>
    <p class="body-sm text-body">Update status di sini — pelanggan akan menerima notifikasi real-time.</p>
  </div>

  <div class="row wrap" role="group" aria-label="Filter status print job">
    <button class="chip" type="button" class:chip-active={!filter} aria-pressed={!filter} onclick={() => { filter = ''; load(); }}>
      Semua
    </button>
    {#each FLOW as f (f.value)}
      <button
        class="chip"
        type="button"
        class:chip-active={filter === f.value}
        aria-pressed={filter === f.value}
        onclick={() => { filter = f.value; load(); }}
      >{f.label}</button>
    {/each}
  </div>

  {#if loading}
    <div class="empty-state" role="status"><span class="spinner"></span> Memuat…</div>
  {:else if jobs.length === 0}
    <div class="empty-state stack-sm">
      <Printer size={28} strokeWidth={1.6} aria-hidden="true" />
      <span class="body-md-strong">Tidak ada print job</span>
      <span class="body-sm text-body">Coba ubah filter status di atas.</span>
    </div>
  {:else}
    <div class="table-wrap">
      <table class="data-table">
        <caption class="sr-only">Daftar print job pelanggan</caption>
        <thead>
          <tr>
            <th scope="col">File</th>
            <th scope="col">Spesifikasi</th>
            <th scope="col">Harga</th>
            <th scope="col">Status</th>
            <th scope="col">Order</th>
            <th scope="col">Waktu</th>
            <th scope="col"><span class="sr-only">Aksi</span></th>
          </tr>
        </thead>
        <tbody>
          {#each jobs as j (j.id)}
            <tr>
              <td>
                <div class="stack-xs">
                  <span class="body-sm-strong">{j.file_name}</span>
                  <span class="caption text-body-mid">{j.page_count} halaman</span>
                </div>
              </td>
              <td>
                <span class="caption">
                  {COLOR_MODE_NAME[j.color_mode]} · {PAPER_SIZE_NAME[j.paper_size]} · {j.copies}x{#if j.duplex} · 2 sisi{/if}
                </span>
              </td>
              <td class="body-sm-strong">{formatRupiah(j.estimated_price)}</td>
              <td>
                <StatusBadge
                  map={PRINT_STATUS}
                  status={j.status}
                  tone={['diproses', 'menunggu'].includes(j.status) ? 'primary' : j.status === 'dibatalkan' ? 'dark' : 'cream'}
                />
              </td>
              <td class="caption">{j.order_id ? j.order_id.slice(0, 8).toUpperCase() : 'Draft'}</td>
              <td class="caption text-body-mid">{formatDate(j.created_at)}</td>
              <td class="cell-action">
                <div class="row" style="justify-content: flex-end">
                  {#if nextJob(j)}
                    <button class="btn btn-primary btn-sm" type="button" onclick={() => setStatus(j, nextJob(j))}>
                      {j.status === 'menunggu' ? 'Mulai Proses' : j.status === 'diproses' ? 'Tandai Siap' : 'Selesai'}
                    </button>
                  {:else if j.status === 'dibatalkan'}
                    <button class="btn btn-tertiary btn-sm" type="button" onclick={() => setStatus(j, 'menunggu')}>
                      Kembalikan
                    </button>
                  {/if}
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .cell-action { text-align: right; }
</style>

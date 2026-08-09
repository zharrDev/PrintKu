<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import { formatRupiah, formatDate, PRINT_STATUS, PAPER_SIZE_NAME, COLOR_MODE_NAME } from '$lib/format.js';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { toast } from '$lib/stores.js';

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

<svelte:head><title>Print Jobs — Admin PrintMart</title></svelte:head>

<div class="stack-xl">
  <div class="stack-sm">
    <span class="eyebrow">Admin</span>
    <h1 class="display-md">Antrian Print Jobs</h1>
    <p class="body-sm text-body">Update status di sini — pelanggan akan mendapat notifikasi real-time via WebSocket.</p>
  </div>

  <div class="row wrap">
    <button class="chip" class:chip-active={!filter} onclick={() => { filter = ''; load(); }}>Semua</button>
    {#each FLOW as f (f.value)}
      <button class="chip" class:chip-active={filter === f.value} onclick={() => { filter = f.value; load(); }}>{f.label}</button>
    {/each}
  </div>

  {#if loading}
    <div class="empty-state"><span class="spinner"></span> Memuat…</div>
  {:else if jobs.length === 0}
    <div class="empty-state body-md">Tidak ada print job.</div>
  {:else}
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>File</th>
            <th>Spesifikasi</th>
            <th>Harga</th>
            <th>Status</th>
            <th>Order</th>
            <th>Waktu</th>
            <th style="text-align: right">Aksi</th>
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
                  {COLOR_MODE_NAME[j.color_mode]} · {PAPER_SIZE_NAME[j.paper_size]} · {j.copies}x
                  {#if j.duplex} · 2 sisi{/if}
                </span>
              </td>
              <td class="body-sm-strong">{formatRupiah(j.estimated_price)}</td>
              <td><StatusBadge map={PRINT_STATUS} status={j.status} tone={['diproses', 'menunggu'].includes(j.status) ? 'primary' : j.status === 'dibatalkan' ? 'dark' : 'cream'} /></td>
              <td class="caption">{j.order_id ? j.order_id.slice(0, 8).toUpperCase() : 'Draft'}</td>
              <td class="caption text-body-mid">{formatDate(j.created_at)}</td>
              <td>
                <div class="row" style="justify-content: flex-end">
                  {#if nextJob(j)}
                    <button class="btn btn-primary btn-sm" onclick={() => setStatus(j, nextJob(j))}>
                      {j.status === 'menunggu' ? 'Mulai Proses' : j.status === 'diproses' ? 'Tandai Siap' : 'Selesai'}
                    </button>
                  {:else if j.status === 'dibatalkan'}
                    <button class="btn btn-tertiary btn-sm" onclick={() => setStatus(j, 'menunggu')}>Kembalikan</button>
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
  .chip {
    border: 1px solid var(--mute);
    background: var(--canvas);
    color: var(--ink);
    border-radius: var(--radius-pill);
    padding: var(--space-xs) var(--space-lg);
    font-size: 15px;
  }
  .chip-active { background: var(--ink); color: var(--on-primary); border-color: var(--ink); }
  .stack-xs { display: flex; flex-direction: column; gap: 2px; }
</style>
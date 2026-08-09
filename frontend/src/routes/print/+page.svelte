<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api.js';
  import { userStore, toast } from '$lib/stores.js';
  import { formatRupiah, PAPER_SIZE_NAME, COLOR_MODE_NAME } from '$lib/format.js';
  import {
    CloudUpload, FileText, CircleCheckBig, ArrowRight, ArrowLeft, Printer,
    X, Minus, Plus, LoaderCircle,
  } from 'lucide-svelte';

  let step = $state(1);
  let upload = $state(null);
  let uploading = $state(false);
  let options = $state({ paperSizes: [], colorModes: [], finishingOptions: [] });

  let spec = $state({ colorMode: 'bw', paperSize: 'a4', copies: 1, duplex: false, finishingOptionId: '' });
  let quote = $state(null);
  let creating = $state(false);
  let dragOver = $state(false);
  let fileInput = $state(null);

  const ACCEPTED = '.pdf,.doc,.docx,.jpg,.jpeg,.png';

  onMount(async () => {
    const r = await api('/print-jobs/options');
    if (r.ok) options = r.json;
    if (!$userStore) {
      toast('Silakan masuk dulu untuk upload file', 'error');
      goto('/masuk');
    }
  });

  function pickFile() {
    fileInput?.click();
  }

  async function handleFile(e) {
    const file = e.target.files?.[0] || e.dataTransfer?.files?.[0];
    if (!file) return;
    if (file.size > 20 * 1024 * 1024) {
      toast('File maksimal 20MB', 'error');
      return;
    }
    uploading = true;
    step = 1;
    const form = new FormData();
    form.append('file', file);
    const r = await api('/print-jobs/upload', { method: 'POST', form });
    uploading = false;
    if (r.ok) {
      upload = r.json.upload;
      toast(`File diterima â€” ${upload.pageCount} halaman terdeteksi`, 'success');
      step = 2;
    }
  }

  async function requestQuote() {
    creating = true;
    const r = await api('/print-jobs', {
      method: 'POST',
      body: {
        fileUrl: upload.fileUrl,
        fileName: upload.fileName,
        pageCount: upload.pageCount,
        colorMode: spec.colorMode,
        paperSize: spec.paperSize,
        copies: spec.copies,
        duplex: spec.duplex ? 1 : 0,
        finishingOptionId: spec.finishingOptionId || null,
      },
    });
    creating = false;
    if (r.ok) {
      quote = r.json;
      step = 3;
    }
  }

  async function updateQuote() {
    if (!quote) return;
    const r = await api(`/print-jobs/${quote.printJob.id}`, {
      method: 'PATCH',
      body: {
        colorMode: spec.colorMode,
        paperSize: spec.paperSize,
        copies: spec.copies,
        duplex: spec.duplex ? 1 : 0,
        finishingOptionId: spec.finishingOptionId || null,
      },
    });
    if (r.ok) quote = r.json;
  }

  function finish() {
    sessionStorage.setItem('printmart_draft_job', JSON.stringify({ printJob: quote.printJob, pricing: quote.pricing }));
    goto('/checkout');
  }

  function reset() {
    upload = null;
    quote = null;
    spec = { colorMode: 'bw', paperSize: 'a4', copies: 1, duplex: false, finishingOptionId: '' };
    step = 1;
  }
</script>

<svelte:head><title>Jasa Print â€” PrintMart</title></svelte:head>

<section class="band-hero">
  <div class="container stack-xl">
    <div class="stack-sm">
      <span class="eyebrow">Jasa Print / Fotocopy</span>
      <h1 class="display-lg">Upload file, kami urus sisanya</h1>
      <p class="body-lg text-body">PDF, DOCX, JPG, atau PNG. Jumlah halaman & harga dihitung otomatis.</p>
    </div>

    <!-- Steps indicator -->
    <div class="row steps">
      <span class="step-item" class:active={step >= 1}>1 Â· Upload</span>
      <span class="step-item" class:active={step >= 2}>2 Â· Spesifikasi</span>
      <span class="step-item" class:active={step >= 3}>3 Â· Ringkasan</span>
    </div>

    {#if step === 1}
      <!-- UPLOAD -->
      <div
        class="dropzone"
        class:dragover={dragOver}
        ondragover={(e) => { e.preventDefault(); dragOver = true; }}
        ondragleave={() => (dragOver = false)}
        ondrop={(e) => { e.preventDefault(); dragOver = false; handleFile(e); }}
        onclick={pickFile}
        role="button"
        tabindex="0"
        onkeydown={(e) => e.key === 'Enter' && pickFile()}
      >
        <input bind:this={fileInput} type="file" accept={ACCEPTED} class="hidden-input" onchange={handleFile} />
        {#if uploading}
          <LoaderCircle size={40} class="spin" style="color: var(--primary)" />
          <span class="display-sub-sm">Memproses fileâ€¦</span>
          <span class="caption text-body-mid">Menghitung jumlah halaman</span>
        {:else}
          <CloudUpload size={44} strokeWidth={1.5} style="color: var(--primary)" />
          <span class="display-sub-sm">Tarik & letakkan file di sini</span>
          <span class="body-sm text-body-mid">atau klik untuk memilih â€” PDF, DOCX, JPG, PNG (maks 20MB)</span>
          <button class="btn btn-primary btn-sm">Pilih File</button>
        {/if}
      </div>

      <div class="grid grid-3">
        <div class="pricing-card stack-xs">
          <Printer size={20} />
          <span class="caption text-body-mid">Jenis file</span>
          <span class="body-md-strong">PDF Â· Word Â· Gambar</span>
        </div>
        <div class="pricing-card stack-xs">
          <FileText size={20} />
          <span class="caption text-body-mid">Halaman</span>
          <span class="body-md-strong">Dihitung otomatis</span>
        </div>
        <div class="pricing-card stack-xs">
          <CircleCheckBig size={20} />
          <span class="caption text-body-mid">Harga</span>
          <span class="body-md-strong">Estimasi instan</span>
        </div>
      </div>
    {:else if step === 2}
      <!-- SPEK -->
      <div class="spec-grid">
        <div class="stack-xl">
          <div class="card stack-md">
            <h3 class="display-sub-sm">Warna & ukuran kertas</h3>
            <div class="grid grid-2">
              {#each options.colorModes as cm (cm.id)}
                <button class="opt-card" class:opt-active={spec.colorMode === cm.id} onclick={() => { spec.colorMode = cm.id; updateQuote(); }}>
                  <span class="body-md-strong">{cm.name}</span>
                  <span class="caption text-body-mid">
                    {cm.id === 'bw' ? 'Rp 500' : 'Rp 1.500'} / halaman (A4)
                  </span>
                </button>
              {/each}
            </div>
            <div class="row wrap mt-1">
              {#each options.paperSizes as ps (ps.id)}
                <button class="chip" class:chip-active={spec.paperSize === ps.id} onclick={() => { spec.paperSize = ps.id; updateQuote(); }}>
                  {ps.name}
                </button>
              {/each}
            </div>
          </div>

          <div class="card stack-md">
            <h3 class="display-sub-sm">Jumlah & finishing</h3>
            <div class="row">
              <span class="body-md">Jumlah rangkap</span>
              <div class="qty">
                <button class="qty-btn" onclick={() => { spec.copies = Math.max(1, spec.copies - 1); updateQuote(); }}><Minus size={16} /></button>
                <input class="qty-input" type="number" min="1" bind:value={spec.copies} onchange={updateQuote} />
                <button class="qty-btn" onclick={() => { spec.copies = Math.min(50, spec.copies + 1); updateQuote(); }}><Plus size={16} /></button>
              </div>
            </div>
            <button class="duplex-btn" class:opt-active={spec.duplex} onclick={() => { spec.duplex = !spec.duplex; updateQuote(); }}>
              <span class="body-md-strong">Cetak dua sisi (duplex)</span>
              <span class="caption text-body-mid">Hemat kertas â€” berlaku untuk dokumen genap ganjil</span>
            </button>
            <label class="field">
              <span class="label-text">Finishing (opsional)</span>
              <select class="input" bind:value={spec.finishingOptionId} onchange={updateQuote}>
                <option value="">Tanpa finishing</option>
                {#each options.finishingOptions as fo (fo.id)}
                  <option value={fo.id}>{fo.name} â€” {formatRupiah(fo.price)}</option>
                {/each}
              </select>
            </label>
          </div>
        </div>

        <!-- Live quote -->
        <div class="pricing-card stack-md quote-card">
          {#if quote}
            <span class="eyebrow">Estimasi</span>
            <h3 class="display-sub-sm">{upload.fileName}</h3>
            <div class="row-between">
              <span class="caption text-body-mid">Halaman</span>
              <span class="body-sm">{upload.pageCount} hal{#if quote.printJob.totalPages !== upload.pageCount} (total {quote.printJob.totalPages} cetak){/if}</span>
            </div>
            <div class="row-between">
              <span class="caption text-body-mid">Warna</span>
              <span class="body-sm">{COLOR_MODE_NAME[spec.colorMode]}</span>
            </div>
            <div class="row-between">
              <span class="caption text-body-mid">Ukuran</span>
              <span class="body-sm">{PAPER_SIZE_NAME[spec.paperSize]}</span>
            </div>
            <div class="row-between">
              <span class="caption text-body-mid">Rangkap</span>
              <span class="body-sm">{spec.copies}x</span>
            </div>
            {#if spec.finishingOptionId}
              <div class="row-between">
                <span class="caption text-body-mid">Finishing</span>
                <span class="body-sm">{quote.printJob.finishing?.name}</span>
              </div>
            {/if}
            <hr class="divider" />
            <div class="row-between">
              <span class="body-md-strong">Total</span>
              <span class="display-sub-sm" style="color: var(--primary)">{formatRupiah(quote.pricing.total)}</span>
            </div>
            <span class="caption text-body-mid">Estimasi selesai: Â±{quote.estimateMinutes} menit</span>
            <button class="btn btn-primary" onclick={() => (step = 3)}>Lanjut <ArrowRight size={18} /></button>
          {:else}
            <div class="empty-state">Isi spesifikasi untuk melihat harga.</div>
          {/if}
        </div>
      </div>
    {:else if step === 3}
      <!-- RINGKASAN -->
      <div class="summary-wrap">
        <div class="pricing-card stack-md">
          <div class="row">
            <FileText size={22} />
            <h3 class="display-sub-sm">{upload.fileName}</h3>
          </div>
          <div class="row wrap">
            <span class="badge badge-primary">{upload.pageCount} halaman</span>
            <span class="badge">{COLOR_MODE_NAME[spec.colorMode]}</span>
            <span class="badge">{PAPER_SIZE_NAME[spec.paperSize]}</span>
            <span class="badge">{spec.copies}x rangkap</span>
            {#if spec.duplex}<span class="badge">Dua sisi</span>{/if}
            {#if quote?.printJob.finishing}<span class="badge">{quote.printJob.finishing.name}</span>{/if}
          </div>
          <hr class="divider" />
          <div class="row-between">
            <span class="body-sm text-body">Halaman per lembar</span>
            <span class="body-sm">{formatRupiah(quote.pricing.perPage)}</span>
          </div>
          <div class="row-between">
            <span class="body-sm text-body">Subtotal ({upload.pageCount} Ã— {spec.copies})</span>
            <span class="body-sm">{formatRupiah(quote.pricing.subtotal)}</span>
          </div>
          {#if quote.pricing.finishingPrice > 0}
            <div class="row-between">
              <span class="body-sm text-body">Finishing</span>
              <span class="body-sm">{formatRupiah(quote.pricing.finishingPrice)}</span>
            </div>
          {/if}
          <div class="row-between">
            <span class="body-md-strong">Total</span>
            <span class="display-md" style="color: var(--primary)">{formatRupiah(quote.pricing.total)}</span>
          </div>
          <div class="row wrap mt-1">
            <button class="btn btn-primary" onclick={finish} disabled={creating}>
              <CircleCheckBig size={18} /> Lanjut ke Checkout
            </button>
            <button class="btn btn-tertiary" onclick={() => (step = 2)}><ArrowLeft size={16} /> Ubah Spesifikasi</button>
            <button class="btn-text" onclick={reset}><X size={14} /> Mulai ulang</button>
          </div>
          <p class="caption text-body-mid">Rincian spesifikasi akan dikonfirmasi admin sebelum cetak.</p>
        </div>
      </div>
    {/if}
  </div>
</section>

<style>
  .steps { gap: var(--space-md); }
  .step-item {
    font-size: 14px;
    font-weight: 600;
    letter-spacing: 0.5px;
    color: var(--body-mid);
    text-transform: uppercase;
  }
  .step-item.active { color: var(--ink); }
  .step-item.active::before {
    content: '';
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--primary);
    margin-right: 6px;
  }
  .dropzone {
    border: 2px dashed var(--ink);
    border-radius: var(--radius-md);
    background: var(--canvas-soft);
    padding: var(--space-4xl) var(--space-xl);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-md);
    text-align: center;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .dropzone:hover { background: var(--canvas); }
  .dropzone.dragover { background: var(--canvas); border-color: var(--primary); }
  .hidden-input { display: none; }
  .spec-grid {
    display: grid;
    grid-template-columns: 1.2fr 0.8fr;
    gap: var(--space-2xl);
    align-items: start;
  }
  .opt-card {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-xs);
    text-align: left;
    background: var(--canvas);
    border: 1px solid var(--mute);
    border-radius: var(--radius-md);
    padding: var(--space-lg);
    color: var(--ink);
    transition: all 0.12s ease;
  }
  .opt-card:hover { border-color: var(--ink); }
  .opt-active { border-color: var(--primary); box-shadow: inset 0 0 0 1px var(--primary); }
  .duplex-btn {
    width: 100%;
    text-align: left;
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
    background: var(--canvas);
    border: 1px solid var(--mute);
    border-radius: var(--radius-md);
    padding: var(--space-md) var(--space-lg);
    color: var(--ink);
    margin-bottom: var(--space-lg);
  }
  .duplex-btn.opt-active { border-color: var(--primary); box-shadow: inset 0 0 0 1px var(--primary); }
  .qty { display: inline-flex; align-items: center; border: 1px solid var(--ink); border-radius: var(--radius-sm); overflow: hidden; margin-left: auto; }
  .qty-btn {
    background: var(--canvas);
    border: none;
    padding: var(--space-sm);
    display: inline-flex;
    color: var(--ink);
  }
  .qty-btn:hover { background: var(--canvas-soft); }
  .qty-input {
    width: 48px;
    border: none;
    border-left: 1px solid var(--ink);
    border-right: 1px solid var(--ink);
    text-align: center;
    font-size: 16px;
    padding: var(--space-sm) 0;
    background: var(--canvas);
    color: var(--ink);
  }
  .chip {
    border: 1px solid var(--mute);
    background: var(--canvas);
    color: var(--ink);
    border-radius: var(--radius-pill);
    padding: var(--space-xs) var(--space-lg);
    font-size: 16px;
  }
  .chip-active { background: var(--ink); color: var(--on-primary); border-color: var(--ink); }
  .quote-card { position: sticky; top: 96px; }
  .summary-wrap { max-width: 560px; margin: 0 auto; }
  .spin { animation: rot 0.9s linear infinite; }
  @keyframes rot { to { transform: rotate(360deg); } }
  @media (max-width: 767px) {
    .spec-grid { grid-template-columns: 1fr; }
    .quote-card { position: static; }
  }
</style>
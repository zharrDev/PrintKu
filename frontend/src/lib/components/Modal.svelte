<script>
  import { onMount } from 'svelte';
  import { X } from 'lucide-svelte';

  let { title = '', onclose, children } = $props();

  const titleId = `modal-${Math.random().toString(36).slice(2, 9)}`;
  let panel = $state(null);

  const FOCUSABLE =
    'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

  // Escape menutup dialog; Tab berputar di dalam dialog (focus trap).
  function onKeydown(e) {
    if (e.key === 'Escape') {
      e.preventDefault();
      onclose();
      return;
    }
    if (e.key !== 'Tab' || !panel) return;
    const nodes = [...panel.querySelectorAll(FOCUSABLE)].filter(
      (el) => el.offsetParent !== null
    );
    if (nodes.length === 0) return;
    const first = nodes[0];
    const last = nodes[nodes.length - 1];
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault();
      first.focus();
    }
  }

  // Pindahkan fokus ke dialog saat dibuka, kunci scroll halaman,
  // lalu kembalikan fokus ke elemen pemicu saat ditutup.
  onMount(() => {
    const previous = document.activeElement;
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    panel?.focus();
    return () => {
      document.body.style.overflow = prevOverflow;
      if (previous instanceof HTMLElement) previous.focus();
    };
  });
</script>

<svelte:window onkeydown={onKeydown} />

<div class="overlay">
  <button class="backdrop" type="button" aria-label="Tutup dialog" onclick={onclose}></button>
  <div
    class="modal"
    role="dialog"
    aria-modal="true"
    aria-labelledby={title ? titleId : undefined}
    tabindex="-1"
    bind:this={panel}
  >
    <div class="modal-head">
      <h2 class="display-sub-sm" id={titleId}>{title}</h2>
      <button class="icon-btn" type="button" aria-label="Tutup dialog" onclick={onclose}>
        <X size={18} />
      </button>
    </div>
    {@render children()}
  </div>
</div>

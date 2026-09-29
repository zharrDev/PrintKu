<script>
  // Bare import (bukan path relatif ke node_modules) — aman untuk npm & pnpm.
  import '@fontsource/inter/400.css';
  import '@fontsource/inter/500.css';
  import '@fontsource/inter/600.css';
  import '@fontsource/inter/700.css';
  import '@fontsource/space-grotesk/600.css';
  import '@fontsource/space-grotesk/700.css';
  import '../app.css';
  import Navbar from '$lib/components/Navbar.svelte';
  import Footer from '$lib/components/Footer.svelte';
  import Toasts from '$lib/components/Toasts.svelte';

  let { children } = $props();
</script>

<Navbar />
<main>
  {@render children()}
</main>
<Footer />
<Toasts />

<style>
  main { min-height: 60vh; }
</style>
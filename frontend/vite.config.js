import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

// Target backend bisa diatur lewat env BACKEND_URL (dipakai di Docker
// Compose: http://backend:8080). Default ke localhost untuk dev lokal.
const backend = process.env.BACKEND_URL || 'http://localhost:8080';
const wsBackend = backend.replace(/^http/, 'ws');

export default defineConfig({
  plugins: [sveltekit()],
  server: {
    host: true,
    proxy: {
      '/api': backend,
      '/uploads': backend,
      '/ws': { target: wsBackend, ws: true },
    },
  },
});

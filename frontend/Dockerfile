# syntax=docker/dockerfile:1

# Frontend SvelteKit — dijalankan via Vite dev server sehingga proxy
# /api, /uploads, /ws otomatis diteruskan ke service backend (lihat
# BACKEND_URL). Cocok untuk lingkungan sandbox/dev PrintKu.
FROM node:22-slim
WORKDIR /app

# Aktifkan pnpm via corepack (bundled dengan Node), lalu install sesuai lockfile.
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .

ENV BACKEND_URL=http://backend:8080
EXPOSE 5173
CMD ["pnpm", "dev", "--", "--host", "0.0.0.0"]

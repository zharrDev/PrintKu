// Wrapper lintas-platform untuk memanggil toolchain Go yang dibundel di
// .tools/go/bin. Dipakai oleh script npm supaya perintah `go` tetap jalan
// tanpa perlu Go terpasang di PATH, apa pun shell-nya (cmd/PowerShell/bash).
import { spawn } from 'node:child_process';
import { existsSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const backendDir = join(here, '..');
const projectRoot = join(backendDir, '..');
const isWin = process.platform === 'win32';

const bundled = join(projectRoot, '.tools', 'go', 'bin', isWin ? 'go.exe' : 'go');
const goBin = existsSync(bundled) ? bundled : 'go';

const args = process.argv.slice(2);
const child = spawn(goBin, args, { stdio: 'inherit', cwd: backendDir });
child.on('error', (err) => {
  console.error(`[go] gagal menjalankan '${goBin}':`, err.message);
  if (goBin === 'go') {
    console.error('[go] Toolchain bundel tidak ditemukan dan Go tidak ada di PATH.');
  }
  process.exit(1);
});
child.on('exit', (code) => process.exit(code ?? 0));

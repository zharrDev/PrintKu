import { getToken, toast } from './stores.js';

export async function api(path, { method = 'GET', body, form, token } = {}) {
  const headers = {};
  const t = token !== undefined ? token : getToken();
  if (t) headers.Authorization = `Bearer ${t}`;
  if (body !== undefined && !form) headers['Content-Type'] = 'application/json';

  const res = await fetch(`/api${path}`, {
    method,
    headers,
    body: form ? form : body !== undefined ? JSON.stringify(body) : undefined,
  });

  let json = null;
  try {
    json = await res.json();
  } catch {
    /* no body */
  }

  if (!res.ok && json?.error) {
    toast(json.error, 'error');
  }
  return { status: res.status, ok: res.ok, json };
}

export async function apiThrow(path, opts) {
  const r = await api(path, opts);
  if (!r.ok) throw new Error(r.json?.error || 'Terjadi kesalahan');
  return r.json;
}
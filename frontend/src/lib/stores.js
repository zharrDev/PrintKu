import { writable } from 'svelte/store';

export const TOKEN_KEY = 'printmart_token';
export const USER_KEY = 'printmart_user';

export const userStore = writable(null);
export const cartStore = writable({ items: [], total: 0 });
export const toasts = writable([]);

let toastId = 0;

export function toast(message, type = 'info') {
  const id = ++toastId;
  toasts.update((list) => [...list, { id, message, type }]);
  setTimeout(() => {
    toasts.update((list) => list.filter((t) => t.id !== id));
  }, 4200);
}

export function getToken() {
  if (typeof localStorage === 'undefined') return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setSession(token, user) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify(user));
  userStore.set(user);
}

export function loadSession() {
  if (typeof localStorage === 'undefined') return;
  const token = localStorage.getItem(TOKEN_KEY);
  const raw = localStorage.getItem(USER_KEY);
  if (token && raw) {
    try {
      userStore.set(JSON.parse(raw));
    } catch {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
    }
  }
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
  userStore.set(null);
}
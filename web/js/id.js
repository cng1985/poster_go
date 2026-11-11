export function uid(prefix = 'l') {
  const bytes = crypto.getRandomValues(new Uint8Array(6));
  return prefix + [...bytes].map((b) => b.toString(16).padStart(2, '0')).join('');
}

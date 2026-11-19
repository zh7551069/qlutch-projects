export function appEnv(): string {
  return import.meta.env.VITE_QLUTCH_ENV || 'local'
}

export function apiBase(): string {
  return import.meta.env.VITE_QLUTCH_API_BASE || 'http://127.0.0.1:4710'
}

export function envLabel(value: string): string {
  if (value === 'local') return '本地'
  if (value === 'staging') return '预发'
  if (value === 'production') return '生产'
  return value
}

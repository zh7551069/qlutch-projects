export type Role = 'duty' | 'admin'

export interface Session {
  role: Role
}

const storageKey = 'qlutch.admin.session'

export function readSession(): Session | null {
  const raw = window.sessionStorage.getItem(storageKey)
  if (raw === 'duty' || raw === 'admin') return { role: raw }
  return null
}

export function writeSession(session: Session): void {
  window.sessionStorage.setItem(storageKey, session.role)
}

export function clearSession(): void {
  window.sessionStorage.removeItem(storageKey)
}

import type { Role } from './session'

export interface NavItem {
  id: string
  label: string
  href: string
  roles: Role[]
}

export const NAV: NavItem[] = [
  { id: 'home', label: '工作台', href: '/', roles: ['duty', 'admin'] },
  { id: 'access', label: '权限占位', href: '/access', roles: ['admin'] },
]

export function visibleNav(role: Role): NavItem[] {
  return NAV.filter((item) => item.roles.includes(role))
}

export function canAccess(role: Role, href: string): boolean {
  const item = NAV.find((entry) => entry.href === href)
  if (!item) return false
  return item.roles.includes(role)
}

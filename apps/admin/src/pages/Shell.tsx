import { QlutchClient } from '@qlutch/contracts'
import { type ReactNode, useEffect, useState } from 'react'
import { EnvBanner } from '../components/EnvBanner'
import { apiBase } from '../env'
import type { NavItem } from '../nav'

export function Shell({
  env,
  items,
  path,
  onNavigate,
  onLogout,
  children,
}: {
  env: string
  items: NavItem[]
  path: string
  onNavigate: (href: string) => void
  onLogout: () => void
  children: ReactNode
}) {
  return (
    <div className="app">
      <EnvBanner env={env} />
      <div className="frame">
        <nav aria-label="空导航">
          {items.map((item) => (
            <a
              key={item.id}
              href={item.href}
              aria-current={path === item.href ? 'page' : undefined}
              onClick={(event) => {
                event.preventDefault()
                onNavigate(item.href)
              }}
            >
              {item.label}
            </a>
          ))}
        </nav>
        <div className="main">
          {children}
          <Probe />
        </div>
      </div>
      <footer>
        <button type="button" onClick={onLogout}>
          结束会话
        </button>
      </footer>
    </div>
  )
}

function Probe() {
  const [text, setText] = useState('尚未检查')
  useEffect(() => {
    const client = new QlutchClient(apiBase())
    let cancelled = false
    client
      .getHealth()
      .then((result) => {
        if (!cancelled) setText(`${result.status} ${result.body.status}`)
      })
      .catch(() => {
        if (!cancelled) setText('不可达')
      })
    return () => {
      cancelled = true
    }
  }, [])
  return <p className="probe">探针 {text}</p>
}

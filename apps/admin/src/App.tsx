import { useEffect, useState } from 'react'
import { EnvBanner } from './components/EnvBanner'
import { appEnv } from './env'
import { canAccess, visibleNav } from './nav'
import { EmptyPage } from './pages/Empty'
import { LoginPage } from './pages/LoginPage'
import { Shell } from './pages/Shell'
import { Unauthorized } from './pages/Unauthorized'
import {
  clearSession,
  type Role,
  readSession,
  type Session,
  writeSession,
} from './session'

function usePath(): string {
  const [path, setPath] = useState(window.location.pathname)
  useEffect(() => {
    const onPop = () => setPath(window.location.pathname)
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [])
  return path
}

export function navigate(href: string): void {
  window.history.pushState(null, '', href)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

function normalize(path: string): string {
  if (path === '' || path === '/login') return '/'
  return path
}

export function App() {
  const path = usePath()
  const [session, setSession] = useState<Session | null>(() => readSession())
  const env = appEnv()

  if (!session) {
    return (
      <>
        <EnvBanner env={env} />
        <LoginPage
          onEnter={(role: Role) => {
            const next = { role }
            writeSession(next)
            setSession(next)
            navigate('/')
          }}
        />
      </>
    )
  }

  const route = normalize(path)
  const page = canAccess(session.role, route) ? (
    <EmptyPage
      title={
        visibleNav(session.role).find((item) => item.href === route)?.label ??
        '工作台'
      }
    />
  ) : (
    <Unauthorized />
  )

  return (
    <Shell
      env={env}
      items={visibleNav(session.role)}
      path={route}
      onNavigate={navigate}
      onLogout={() => {
        clearSession()
        setSession(null)
        navigate('/login')
      }}
    >
      {page}
    </Shell>
  )
}

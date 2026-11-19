import { envLabel } from '../env'

export function EnvBanner({ env }: { env: string }) {
  return (
    <div className="env-banner" data-env={env} role="status">
      <span>环境 {envLabel(env)}</span>
      <code>{env}</code>
    </div>
  )
}

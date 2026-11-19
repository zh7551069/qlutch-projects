import type { FormEvent } from 'react'
import type { Role } from '../session'

export function LoginPage({ onEnter }: { onEnter: (role: Role) => void }) {
  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const role =
      new FormData(event.currentTarget).get('role') === 'admin'
        ? 'admin'
        : 'duty'
    onEnter(role)
  }

  return (
    <main className="login">
      <form onSubmit={onSubmit}>
        <p className="kicker">Qlutch 后管</p>
        <h1>占位登录</h1>
        <p>登录渠道未接。口令不会离开这台浏览器，只用来进入空壳。</p>
        <label>
          账号
          <input
            name="account"
            autoComplete="username"
            placeholder="本地占位"
          />
        </label>
        <label>
          口令
          <input
            name="secret"
            type="password"
            autoComplete="current-password"
          />
        </label>
        <label>
          角色
          <select name="role" defaultValue="duty">
            <option value="duty">值班</option>
            <option value="admin">管理</option>
          </select>
        </label>
        <button type="submit">进入</button>
      </form>
    </main>
  )
}

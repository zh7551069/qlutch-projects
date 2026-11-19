export function Unauthorized() {
  return (
    <section className="empty denied">
      <p className="kicker">无权</p>
      <h1>这个位置没有对当前角色开放</h1>
      <p>这是空态占位，没有业务数据。</p>
    </section>
  )
}

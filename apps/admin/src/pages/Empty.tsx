export function EmptyPage({ title }: { title: string }) {
  return (
    <section className="empty">
      <p className="kicker">空骨架</p>
      <h1>{title}</h1>
      <p>这里还没有业务列表。</p>
    </section>
  )
}

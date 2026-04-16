import { NavLink } from 'react-router-dom'

const navItems = [
  { to: '/', label: 'Dashboard', icon: '⌂' },
  { to: '/stacks', label: 'Stacks', icon: '▦' },
  { to: '/deployments', label: 'Deployments', icon: '⇪' },
  { to: '/schedules', label: 'Schedules', icon: '⌚' },
  { to: '/dependencies', label: 'Dependencies', icon: '⇄' },
]

export default function Sidebar() {
  return (
    <aside className="w-60 bg-[var(--color-surface)] border-r border-[var(--color-border)] flex flex-col">
      <div className="p-4 border-b border-[var(--color-border)]">
        <h1 className="text-xl font-bold text-[var(--color-primary)]">
          Composarr
        </h1>
        <p className="text-xs text-[var(--color-text-muted)] mt-1">
          Stack Lifecycle Manager
        </p>
      </div>

      <nav className="flex-1 p-3 space-y-1">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                isActive
                  ? 'bg-[var(--color-primary)] text-white'
                  : 'text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)]'
              }`
            }
          >
            <span className="text-base">{item.icon}</span>
            {item.label}
          </NavLink>
        ))}
      </nav>

      <div className="p-4 border-t border-[var(--color-border)] text-xs text-[var(--color-text-muted)]">
        v0.1.0
      </div>
    </aside>
  )
}

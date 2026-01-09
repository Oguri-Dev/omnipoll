import { NavLink } from 'react-router-dom'
import { LayoutDashboard, Settings, FileText, Database } from 'lucide-react'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/events', icon: Database, label: 'Events' },
  { to: '/config', icon: Settings, label: 'Configuration' },
  { to: '/logs', icon: FileText, label: 'Logs' },
]

export default function Sidebar() {
  return (
    <aside className="w-64 bg-gray-900 text-white">
      <div className="p-6">
        <h1 className="text-2xl font-bold">Omnipoll</h1>
        <p className="text-gray-400 text-sm">Admin Panel</p>
      </div>
      <nav className="mt-6">
        {navItems.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            className={({ isActive }) =>
              `flex items-center gap-3 px-6 py-3 hover:bg-gray-800 transition-colors ${
                isActive ? 'bg-gray-800 border-l-4 border-blue-500' : ''
              }`
            }
          >
            <Icon size={20} />
            {label}
          </NavLink>
        ))}
      </nav>
    </aside>
  )
}

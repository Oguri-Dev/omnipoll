import { ReactNode } from 'react'

interface StatusCardProps {
  title: string
  value: string | number
  icon: ReactNode
  status?: 'success' | 'warning' | 'error' | 'neutral'
}

const statusColors = {
  success: 'bg-green-100 text-green-800',
  warning: 'bg-yellow-100 text-yellow-800',
  error: 'bg-red-100 text-red-800',
  neutral: 'bg-gray-100 text-gray-800',
}

export default function StatusCard({ title, value, icon, status = 'neutral' }: StatusCardProps) {
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-500">{title}</p>
          <p className="text-2xl font-bold mt-1">{value}</p>
        </div>
        <div className={`p-3 rounded-full ${statusColors[status]}`}>{icon}</div>
      </div>
    </div>
  )
}

import { useTranslation } from 'react-i18next'
import { LockClosedIcon } from '@heroicons/react/24/outline'
import { useAuth } from '@/hooks/use-auth'
import { AppHeader } from '@/components/app-header'
import { clsx } from 'clsx'

export function DashboardPage() {
  const { t } = useTranslation()
  const { user } = useAuth()

  const initials = user.username?.charAt(0).toUpperCase() ?? '?'
  const isAdmin = user.role === 'ADMIN'

  return (
    <div className="min-h-screen bg-zinc-100 dark:bg-zinc-900">
      <AppHeader />

      <main className="mx-auto max-w-6xl px-4 py-8">
        <div className="mb-8 flex items-center gap-4 rounded-xl border border-zinc-200 bg-white px-5 py-4 dark:border-zinc-700/60 dark:bg-zinc-800">
          <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-indigo-100 text-indigo-600 dark:bg-indigo-950 dark:text-indigo-400">
            <span className="text-sm font-semibold">{initials}</span>
          </div>
          <div className="min-w-0">
            <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">{user.username}</p>
            <span
              className={clsx(
                'mt-0.5 inline-flex items-center rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset',
                isAdmin
                  ? 'bg-amber-50 text-amber-700 ring-amber-600/20 dark:bg-amber-950/60 dark:text-amber-400 dark:ring-amber-400/20'
                  : 'bg-zinc-100 text-zinc-600 ring-zinc-500/20 dark:bg-zinc-700 dark:text-zinc-400 dark:ring-zinc-600/30',
              )}
            >
              {t(`role.${user.role?.toLowerCase()}`)}
            </span>
          </div>
        </div>

        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-zinc-300 py-16 text-center dark:border-zinc-700">
          <LockClosedIcon className="mb-3 size-8 text-zinc-300 dark:text-zinc-600" aria-hidden="true" />
          <p className="text-sm font-medium text-zinc-500 dark:text-zinc-400">
            {t('dashboard.noGates')}
          </p>
          <p className="mt-1 text-xs text-zinc-400 dark:text-zinc-600">
            {t('dashboard.noGatesHint')}
          </p>
        </div>
      </main>
    </div>
  )
}

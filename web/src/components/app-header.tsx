import { Link } from '@tanstack/react-router'
import type { LinkProps } from '@tanstack/react-router'
import { UsersIcon, ArrowRightOnRectangleIcon, HomeModernIcon } from '@heroicons/react/24/outline'
import { useMutation } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { authApi } from '@/lib/api'
import { clearAuth } from '@/lib/auth'
import { useAuth } from '@/hooks/use-auth'
import { Button } from '@/components/ui/button'
import { ThemeToggle } from '@/components/theme-toggle'
import { LangToggle } from '@/components/lang-toggle'

const navLinkClasses =
  'flex shrink-0 items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-sm text-zinc-500 transition-colors hover:bg-zinc-200 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100 [&.active]:bg-zinc-200 [&.active]:text-zinc-900 dark:[&.active]:bg-zinc-800 dark:[&.active]:text-zinc-100'

function NavLink({ children, ...props }: LinkProps & { children: React.ReactNode }) {
  return (
    <Link className={navLinkClasses} activeProps={{ className: 'active' }} {...props}>
      {children}
    </Link>
  )
}

export function AppHeader() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { user } = useAuth()

  const logoutMutation = useMutation({
    mutationFn: () => authApi.logout(),
    onSettled: () => {
      clearAuth()
      navigate({ to: '/login' })
    },
  })

  return (
    <header className="border-b border-zinc-200 bg-zinc-100 dark:border-zinc-800 dark:bg-zinc-900">
      <div className="mx-auto flex h-14 max-w-6xl items-center gap-3 px-4">
        <Link
          to="/"
          className="shrink-0 text-xs font-semibold tracking-[0.25em] text-zinc-900 uppercase dark:text-zinc-100"
        >
          GATIE
        </Link>

        {user.role === 'ADMIN' && (
          <nav className="flex min-w-0 items-center gap-0.5 overflow-x-auto">
            <NavLink to="/gates">
              <HomeModernIcon className="size-4 shrink-0" aria-hidden="true" />
              <span className="hidden sm:inline">{t('nav.gates')}</span>
            </NavLink>
            <NavLink to="/members">
              <UsersIcon className="size-4 shrink-0" aria-hidden="true" />
              <span className="hidden sm:inline">{t('nav.members')}</span>
            </NavLink>
          </nav>
        )}

        <div className="ml-auto flex shrink-0 items-center gap-1">
          <LangToggle />
          <ThemeToggle />
          <div className="mx-1 h-5 w-px bg-zinc-200 dark:bg-zinc-700" />
          <Button
            variant="ghost"
            size="sm"
            onClick={() => logoutMutation.mutate()}
            loading={logoutMutation.isPending}
          >
            <ArrowRightOnRectangleIcon className="size-4 shrink-0" aria-hidden="true" />
            <span className="ml-1.5 hidden sm:inline">{t('nav.logout')}</span>
          </Button>
        </div>
      </div>
    </header>
  )
}

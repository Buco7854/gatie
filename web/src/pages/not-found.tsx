import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export function NotFoundPage() {
  const { t } = useTranslation()

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-100 px-4 dark:bg-zinc-900">
      <div className="text-center">
        <p className="text-4xl font-bold text-zinc-300 dark:text-zinc-600">404</p>
        <h1 className="mt-3 text-sm font-semibold text-zinc-900 dark:text-zinc-100">
          {t('error.notFound')}
        </h1>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          {t('error.notFoundHint')}
        </p>
        <Link
          to="/"
          className="mt-5 inline-block rounded-lg bg-indigo-600 px-3.5 py-2 text-sm font-medium text-white hover:bg-indigo-500"
        >
          {t('error.backHome')}
        </Link>
      </div>
    </div>
  )
}

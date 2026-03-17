import { Component } from 'react'
import type { ReactNode, ErrorInfo } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
}

function ErrorFallback() {
  const { t } = useTranslation()

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-100 px-4 dark:bg-zinc-900">
      <div className="text-center">
        <p className="text-4xl font-bold text-zinc-300 dark:text-zinc-600">!</p>
        <h1 className="mt-3 text-sm font-semibold text-zinc-900 dark:text-zinc-100">
          {t('error.boundary')}
        </h1>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          {t('error.boundaryHint')}
        </p>
        <Button className="mt-5" onClick={() => window.location.reload()}>
          {t('error.reload')}
        </Button>
      </div>
    </div>
  )
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false }

  static getDerivedStateFromError(): State {
    return { hasError: true }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('ErrorBoundary caught:', error, info.componentStack)
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback />
    }
    return this.props.children
  }
}

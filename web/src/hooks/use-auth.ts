import { useEffect, useState, useSyncExternalStore } from 'react'
import { silentRefresh } from '@/lib/api'
import { isAuthenticated, subscribe, getSnapshot } from '@/lib/auth'

type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated'

export function useAuth() {
  const auth = useSyncExternalStore(subscribe, getSnapshot)
  const [status, setStatus] = useState<AuthStatus>('loading')

  useEffect(() => {
    if (isAuthenticated()) {
      setStatus('authenticated')
      return
    }
    silentRefresh().then((ok) => {
      setStatus(ok ? 'authenticated' : 'unauthenticated')
    })
  }, [])

  return { status, user: auth }
}

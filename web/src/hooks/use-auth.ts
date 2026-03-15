import { useEffect, useState } from 'react'
import { silentRefresh } from '@/lib/api'
import { isAuthenticated, getAuthState } from '@/lib/auth'

type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated'

export function useAuth() {
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

  return { status, user: getAuthState() }
}

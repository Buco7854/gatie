import { useSyncExternalStore } from "react"
import { auth } from "@/lib/auth"

export function useAuth() {
  const state = useSyncExternalStore(
    (cb) => auth.subscribe(cb),
    () => auth.getState(),
  )

  return {
    ...state,
    logout: () => auth.logout(),
  }
}

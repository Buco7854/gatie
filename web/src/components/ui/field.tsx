import { useId } from 'react'
import { clsx } from 'clsx'

interface FieldProps {
  label: string
  hint?: string
  error?: string
  children: React.ReactElement<{ id?: string; 'aria-describedby'?: string; 'aria-invalid'?: boolean }>
  className?: string
}

export function Field({ label, hint, error, children, className }: FieldProps) {
  const id = useId()
  const errorId = `${id}-error`
  const hintId = `${id}-hint`

  return (
    <div className={clsx('space-y-1.5', className)}>
      <label htmlFor={id} className="block text-sm font-medium text-zinc-700 dark:text-zinc-300">
        {label}
      </label>
      {children}
      {hint && !error && <p id={hintId} className="text-xs text-zinc-500 dark:text-zinc-400">{hint}</p>}
      {error && <p id={errorId} role="alert" className="text-xs text-red-600 dark:text-red-400">{error}</p>}
    </div>
  )
}

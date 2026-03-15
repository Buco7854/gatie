import { forwardRef } from 'react'
import { clsx } from 'clsx'

type SelectProps = React.SelectHTMLAttributes<HTMLSelectElement>

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, children, ...props }, ref) => (
    <select
      ref={ref}
      className={clsx(
        'block w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900',
        'focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20',
        'dark:border-zinc-700 dark:bg-zinc-800/50 dark:text-zinc-100',
        'dark:focus:border-indigo-400 dark:focus:ring-indigo-400/20',
        'disabled:cursor-not-allowed disabled:opacity-50',
        className,
      )}
      {...props}
    >
      {children}
    </select>
  ),
)
Select.displayName = 'Select'

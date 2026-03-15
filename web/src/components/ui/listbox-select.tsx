import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/24/outline'
import { clsx } from 'clsx'

export interface SelectOption {
  value: string
  label: string
}

interface ListboxSelectProps {
  value: string
  onChange: (value: string) => void
  options: SelectOption[]
  className?: string
  disabled?: boolean
}

export function ListboxSelect({ value, onChange, options, className, disabled }: ListboxSelectProps) {
  const selected = options.find((o) => o.value === value) ?? options[0]

  return (
    <Listbox value={value} onChange={onChange} disabled={disabled}>
      <ListboxButton
        className={clsx(
          'flex w-full cursor-pointer items-center justify-between rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900',
          'focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20',
          'dark:border-zinc-700 dark:bg-zinc-800/50 dark:text-zinc-100',
          'dark:focus:border-indigo-400 dark:focus:ring-indigo-400/20',
          'disabled:cursor-not-allowed disabled:opacity-50',
          className,
        )}
      >
        <span>{selected?.label}</span>
        <ChevronUpDownIcon className="size-4 shrink-0 text-zinc-400 dark:text-zinc-500" aria-hidden="true" />
      </ListboxButton>

      <ListboxOptions
        anchor="bottom start"
        transition
        className={clsx(
          'z-50 mt-1 w-[var(--button-width)] rounded-xl border border-zinc-200 bg-white p-1 shadow-lg',
          'dark:border-zinc-700 dark:bg-zinc-800',
          'focus:outline-none',
          'transition duration-100 ease-in data-[closed]:opacity-0 data-[closed]:scale-95',
        )}
      >
        {options.map((option) => (
          <ListboxOption
            key={option.value}
            value={option.value}
            className={clsx(
              'group flex cursor-pointer select-none items-center gap-2 rounded-lg px-3 py-2 text-sm',
              'text-zinc-700 dark:text-zinc-300',
              'data-[focus]:bg-zinc-100 dark:data-[focus]:bg-zinc-700/60',
            )}
          >
            <span className="flex-1 group-data-[selected]:font-medium group-data-[selected]:text-indigo-600 dark:group-data-[selected]:text-indigo-400">
              {option.label}
            </span>
            <CheckIcon
              className="size-3.5 text-indigo-600 opacity-0 group-data-[selected]:opacity-100 dark:text-indigo-400"
              aria-hidden="true"
            />
          </ListboxOption>
        ))}
      </ListboxOptions>
    </Listbox>
  )
}

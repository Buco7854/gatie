import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from '@headlessui/react'
import { GlobeAltIcon, CheckIcon } from '@heroicons/react/24/outline'
import { useTranslation } from 'react-i18next'
import { clsx } from 'clsx'

const languages = [
  { code: 'en', label: 'English' },
  { code: 'fr', label: 'Français' },
] as const

type LangCode = (typeof languages)[number]['code']

export function LangToggle() {
  const { i18n } = useTranslation()
  const currentCode = (languages.find((l) => l.code === i18n.language)?.code ?? 'en') as LangCode

  return (
    <Listbox value={currentCode} onChange={(code) => i18n.changeLanguage(code)}>
      <ListboxButton
        className={clsx(
          'flex cursor-pointer items-center gap-1.5 rounded-lg px-2 py-2',
          'text-zinc-500 hover:bg-zinc-200 hover:text-zinc-700',
          'dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-200',
          'transition-colors text-xs font-medium uppercase tracking-wide',
          'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500',
        )}
      >
        <GlobeAltIcon className="size-4 shrink-0" aria-hidden="true" />
        <span>{currentCode}</span>
      </ListboxButton>

      <ListboxOptions
        anchor="bottom end"
        transition
        className={clsx(
          'z-50 mt-1 w-36 rounded-xl border border-zinc-200 bg-white p-1 shadow-lg',
          'dark:border-zinc-700 dark:bg-zinc-800',
          'focus:outline-none',
          'transition duration-100 ease-in data-[closed]:opacity-0 data-[closed]:scale-95',
        )}
      >
        {languages.map(({ code, label }) => (
          <ListboxOption
            key={code}
            value={code}
            className={clsx(
              'group flex cursor-pointer select-none items-center gap-2 rounded-lg px-3 py-2 text-sm',
              'text-zinc-700 dark:text-zinc-300',
              'data-[focus]:bg-zinc-100 dark:data-[focus]:bg-zinc-700/60',
            )}
          >
            <span className="flex-1 group-data-[selected]:font-medium group-data-[selected]:text-indigo-600 dark:group-data-[selected]:text-indigo-400">
              {label}
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

import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from '@headlessui/react'
import {
  SunIcon,
  MoonIcon,
  ComputerDesktopIcon,
  CheckIcon,
} from '@heroicons/react/24/solid'
import { useTheme } from 'next-themes'
import { useTranslation } from 'react-i18next'
import { clsx } from 'clsx'

const themes = [
  { id: 'light', Icon: SunIcon },
  { id: 'dark', Icon: MoonIcon },
  { id: 'system', Icon: ComputerDesktopIcon },
] as const

type ThemeId = (typeof themes)[number]['id']

function CurrentIcon({ themeId }: { themeId: string | undefined }) {
  const match = themes.find((t) => t.id === themeId) ?? themes[2]
  return <match.Icon className="size-4" aria-hidden="true" />
}

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  const { t } = useTranslation()

  return (
    <Listbox value={(theme ?? 'system') as ThemeId} onChange={setTheme}>
      <ListboxButton
        className={clsx(
          'flex cursor-pointer items-center justify-center rounded-lg p-2',
          'text-zinc-500 hover:bg-zinc-200 hover:text-zinc-700',
          'dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-200',
          'transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500',
        )}
      >
        <CurrentIcon themeId={theme} />
        <span className="sr-only">{t('theme.toggle')}</span>
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
        {themes.map(({ id, Icon }) => (
          <ListboxOption
            key={id}
            value={id}
            className={clsx(
              'group flex cursor-pointer select-none items-center gap-2.5 rounded-lg px-3 py-2 text-sm',
              'text-zinc-700 dark:text-zinc-300',
              'data-[focus]:bg-zinc-100 dark:data-[focus]:bg-zinc-700/60',
            )}
          >
            <Icon
              className={clsx(
                'size-4 text-zinc-400 dark:text-zinc-500',
                'group-data-[selected]:text-indigo-600 dark:group-data-[selected]:text-indigo-400',
              )}
              aria-hidden="true"
            />
            <span
              className={clsx(
                'flex-1',
                'group-data-[selected]:font-medium group-data-[selected]:text-indigo-600 dark:group-data-[selected]:text-indigo-400',
              )}
            >
              {t(`theme.${id}`)}
            </span>
            <CheckIcon
              className={clsx(
                'size-3.5 text-indigo-600 opacity-0 dark:text-indigo-400',
                'group-data-[selected]:opacity-100',
              )}
            />
          </ListboxOption>
        ))}
      </ListboxOptions>
    </Listbox>
  )
}

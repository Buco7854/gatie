import { useNavigate, Link } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { apiFetch, type AuthTokens } from '@/lib/api'
import { setAuth } from '@/lib/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field } from '@/components/ui/field'
import { ThemeToggle } from '@/components/theme-toggle'

const schema = z.object({
  username: z.string().min(1),
  password: z.string().min(1),
})

type FormData = z.infer<typeof schema>

export function LoginPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({ resolver: zodResolver(schema) })

  const mutation = useMutation({
    mutationFn: (data: FormData) =>
      apiFetch<AuthTokens>('/auth/login', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: (data) => {
      setAuth(data)
      navigate({ to: '/' })
    },
  })

  const onSubmit = handleSubmit((data) => mutation.mutate(data))

  return (
    <div className="flex min-h-screen flex-col bg-zinc-100 dark:bg-zinc-900">
      <div className="flex justify-end p-4">
        <ThemeToggle />
      </div>

      <div className="flex flex-1 items-center justify-center px-4 pb-16">
        <div className="w-full max-w-sm">
          <div className="mb-8 text-center">
            <span className="text-xs font-semibold tracking-[0.25em] text-zinc-400 uppercase dark:text-zinc-500">
              GATIE
            </span>
            <h1 className="mt-2 text-xl font-semibold text-zinc-900 dark:text-zinc-100">
              {t('login.title')}
            </h1>
            <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
              {t('login.subtitle')}
            </p>
          </div>

          <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-sm dark:border-zinc-700/60 dark:bg-zinc-800">
            <form onSubmit={onSubmit} className="space-y-4">
              <Field
                label={t('field.username')}
                error={errors.username ? t('validation.required') : undefined}
              >
                <Input
                  {...register('username')}
                  placeholder={t('field.usernamePlaceholder')}
                  autoComplete="username"
                  autoFocus
                />
              </Field>

              <Field
                label={t('field.password')}
                error={errors.password ? t('validation.required') : undefined}
              >
                <Input
                  {...register('password')}
                  type="password"
                  placeholder="••••••••"
                  autoComplete="current-password"
                />
              </Field>

              {mutation.isError && (
                <p className="text-xs text-red-600 dark:text-red-400">
                  {t('login.invalidCredentials')}
                </p>
              )}

              <Button type="submit" loading={mutation.isPending} className="w-full mt-2">
                {t('login.submit')}
              </Button>
            </form>
          </div>

          <p className="mt-5 text-center text-sm text-zinc-500 dark:text-zinc-400">
            {t('login.setupLink')}{' '}
            <Link
              to="/setup"
              className="text-indigo-600 hover:text-indigo-700 hover:underline dark:text-indigo-400 dark:hover:text-indigo-300"
            >
              {t('login.setupLinkAnchor')}
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}

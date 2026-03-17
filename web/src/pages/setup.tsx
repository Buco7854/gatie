import { useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { authApi } from '@/lib/api'
import { setAuth } from '@/lib/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field } from '@/components/ui/field'
import { ThemeToggle } from '@/components/theme-toggle'
import { Spinner } from '@/components/ui/spinner'

const schema = z
  .object({
    username: z.string().min(3),
    password: z.string().min(8),
    confirmPassword: z.string(),
  })
  .refine((d) => d.password === d.confirmPassword, {
    path: ['confirmPassword'],
    message: 'mismatch',
  })

type FormData = z.infer<typeof schema>

export function SetupPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  const { data: setupStatus, isError, isLoading } = useQuery({
    queryKey: ['setup-status'],
    queryFn: () => authApi.getSetupStatus(),
  })

  useEffect(() => {
    if (setupStatus && !setupStatus.needs_setup) {
      navigate({ to: '/login' })
    }
  }, [setupStatus, navigate])

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({ resolver: zodResolver(schema) })

  const mutation = useMutation({
    mutationFn: (data: Omit<FormData, 'confirmPassword'>) =>
      authApi.setup(data.username, data.password),
    onSuccess: (data) => {
      setAuth(data)
      navigate({ to: '/' })
    },
  })

  const onSubmit = handleSubmit(({ username, password }) => {
    mutation.mutate({ username, password })
  })

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-100 dark:bg-zinc-900">
        <Spinner />
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-100 dark:bg-zinc-900">
        <p className="text-sm text-red-600 dark:text-red-400">{t('error.generic')}</p>
      </div>
    )
  }

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
              {t('setup.title')}
            </h1>
            <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
              {t('setup.subtitle')}
            </p>
          </div>

          <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-sm dark:border-zinc-700/60 dark:bg-zinc-800">
            <form onSubmit={onSubmit} className="space-y-4">
              <Field
                label={t('field.username')}
                error={errors.username ? t('validation.minLength', { min: 3 }) : undefined}
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
                error={errors.password ? t('validation.minLength', { min: 8 }) : undefined}
              >
                <Input
                  {...register('password')}
                  type="password"
                  placeholder="••••••••"
                  autoComplete="new-password"
                />
              </Field>

              <Field
                label={t('field.confirmPassword')}
                error={errors.confirmPassword ? t('setup.passwordMismatch') : undefined}
              >
                <Input
                  {...register('confirmPassword')}
                  type="password"
                  placeholder="••••••••"
                  autoComplete="new-password"
                />
              </Field>

              {mutation.isError && (
                <p className="text-xs text-red-600 dark:text-red-400">{t('error.generic')}</p>
              )}

              <Button type="submit" loading={mutation.isPending} className="w-full mt-2">
                {t('setup.submit')}
              </Button>
            </form>
          </div>
        </div>
      </div>
    </div>
  )
}

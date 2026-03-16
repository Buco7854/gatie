import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Dialog, DialogPanel, DialogTitle } from '@headlessui/react'
import {
  PlusIcon,
  PencilSquareIcon,
  TrashIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline'
import { useTranslation } from 'react-i18next'
import { useAuth } from '@/hooks/use-auth'
import { apiFetch, ApiError } from '@/lib/api'
import type { Member, MembersPage } from '@/lib/types'
import { AppHeader } from '@/components/app-header'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field } from '@/components/ui/field'
import { ListboxSelect } from '@/components/ui/listbox-select'
import { Spinner } from '@/components/ui/spinner'
import { Pagination } from '@/components/ui/pagination'
import { clsx } from 'clsx'

// --- Schemas ---

const createSchema = z.object({
  username: z.string().min(3).max(100),
  display_name: z.string().max(200).optional(),
  password: z.string().min(8).max(128),
  role: z.enum(['ADMIN', 'MEMBER']),
})

const updateSchema = z.object({
  username: z.string().min(3).max(100),
  display_name: z.string().max(200).optional(),
  role: z.enum(['ADMIN', 'MEMBER']),
})

type CreateFormData = z.infer<typeof createSchema>
type UpdateFormData = z.infer<typeof updateSchema>

// --- Role badge ---

function RoleBadge({ role }: { role: Member['role'] }) {
  const { t } = useTranslation()
  return (
    <span
      className={clsx(
        'inline-flex items-center rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset',
        role === 'ADMIN'
          ? 'bg-amber-50 text-amber-700 ring-amber-600/20 dark:bg-amber-950/60 dark:text-amber-400 dark:ring-amber-400/20'
          : 'bg-zinc-100 text-zinc-600 ring-zinc-500/20 dark:bg-zinc-700 dark:text-zinc-400 dark:ring-zinc-600/30',
      )}
    >
      {t(`role.${role.toLowerCase()}`)}
    </span>
  )
}

// --- Action buttons ---

interface ActionButtonsProps {
  member: Member
  isSelf: boolean
  setModal: (state: ModalState) => void
  setMemberToDelete: (member: Member) => void
}

function ActionButtons({ member, isSelf, setModal, setMemberToDelete }: ActionButtonsProps) {
  const { t } = useTranslation()
  return (
    <div className="flex items-center gap-1">
      <button
        onClick={() => setModal({ type: 'edit', member })}
        className="cursor-pointer rounded-lg p-1.5 text-zinc-400 transition-all hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-700 dark:hover:text-zinc-200"
        title={t('action.edit')}
      >
        <PencilSquareIcon className="size-4" aria-hidden="true" />
      </button>
      {!isSelf && (
        <button
          onClick={() => setMemberToDelete(member)}
          className="cursor-pointer rounded-lg p-1.5 text-zinc-400 transition-all hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950 dark:hover:text-red-400"
          title={t('action.delete')}
        >
          <TrashIcon className="size-4" aria-hidden="true" />
        </button>
      )}
    </div>
  )
}

// --- Role options hook ---

function useRoleOptions() {
  const { t } = useTranslation()
  return [
    { value: 'MEMBER', label: t('role.member') },
    { value: 'ADMIN', label: t('role.admin') },
  ]
}

// --- Create form ---

function CreateForm({ onSuccess }: { onSuccess: () => void }) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const roleOptions = useRoleOptions()

  const {
    register,
    handleSubmit,
    control,
    setError,
    formState: { errors },
  } = useForm<CreateFormData>({
    resolver: zodResolver(createSchema),
    defaultValues: { role: 'MEMBER' },
  })

  const mutation = useMutation({
    mutationFn: (data: CreateFormData) =>
      apiFetch<Member>('/members', { method: 'POST', body: JSON.stringify(data) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['members'] })
      onSuccess()
    },
    onError: (err: Error) => {
      if (err instanceof ApiError && err.hasFieldError('body.username')) {
        setError('username', { message: t('validation.usernameTaken') })
      }
    },
  })

  return (
    <form onSubmit={handleSubmit((data) => mutation.mutate(data))} className="space-y-4">
      <Field
        label={t('field.username')}
        error={errors.username ? (errors.username.message ?? t('validation.minLength', { min: 3 })) : undefined}
      >
        <Input
          {...register('username')}
          placeholder={t('field.usernamePlaceholder')}
          autoComplete="off"
          autoFocus
        />
      </Field>

      <Field label={t('field.displayName')} error={errors.display_name?.message}>
        <Input
          {...register('display_name')}
          placeholder={t('field.displayNamePlaceholder')}
          autoComplete="off"
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

      <Field label={t('field.role')}>
        <Controller
          control={control}
          name="role"
          render={({ field }) => (
            <ListboxSelect value={field.value} onChange={field.onChange} options={roleOptions} />
          )}
        />
      </Field>

      {mutation.isError && !errors.username && (
        <p className="text-xs text-red-600 dark:text-red-400">{t('error.generic')}</p>
      )}

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onSuccess}>
          {t('action.cancel')}
        </Button>
        <Button type="submit" loading={mutation.isPending}>
          {t('members.create')}
        </Button>
      </div>
    </form>
  )
}

// --- Edit form ---

function EditForm({ member, isSelf, onSuccess }: { member: Member; isSelf: boolean; onSuccess: () => void }) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const roleOptions = useRoleOptions()

  const {
    register,
    handleSubmit,
    control,
    setError,
    formState: { errors },
  } = useForm<UpdateFormData>({
    resolver: zodResolver(updateSchema),
    defaultValues: {
      username: member.username,
      display_name: member.display_name ?? '',
      role: member.role,
    },
  })

  const mutation = useMutation({
    mutationFn: (data: UpdateFormData) =>
      apiFetch<Member>(`/members/${member.id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['members'] })
      onSuccess()
    },
    onError: (err: Error) => {
      if (err instanceof ApiError && err.hasFieldError('body.username')) {
        setError('username', { message: t('validation.usernameTaken') })
      }
    },
  })

  return (
    <form onSubmit={handleSubmit((data) => mutation.mutate(data))} className="space-y-4">
      <Field
        label={t('field.username')}
        error={errors.username ? (errors.username.message ?? t('validation.minLength', { min: 3 })) : undefined}
      >
        <Input
          {...register('username')}
          placeholder={t('field.usernamePlaceholder')}
          autoComplete="off"
          autoFocus
        />
      </Field>

      <Field label={t('field.displayName')} error={errors.display_name?.message}>
        <Input
          {...register('display_name')}
          placeholder={t('field.displayNamePlaceholder')}
          autoComplete="off"
        />
      </Field>

      <Field label={t('field.role')}>
        <Controller
          control={control}
          name="role"
          render={({ field }) => (
            <ListboxSelect
              value={field.value}
              onChange={field.onChange}
              options={roleOptions}
              disabled={isSelf}
            />
          )}
        />
        {isSelf && (
          <p className="mt-1 text-xs text-zinc-400 dark:text-zinc-500">{t('members.cannotEditOwnRole')}</p>
        )}
      </Field>

      {mutation.isError && !errors.username && (
        <p className="text-xs text-red-600 dark:text-red-400">{t('error.generic')}</p>
      )}

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onSuccess}>
          {t('action.cancel')}
        </Button>
        <Button type="submit" loading={mutation.isPending}>
          {t('action.save')}
        </Button>
      </div>
    </form>
  )
}

// --- Main page ---

type ModalState = { type: 'create' } | { type: 'edit'; member: Member } | null

const PER_PAGE = 20

export function MembersPage() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [modal, setModal] = useState<ModalState>(null)
  const [memberToDelete, setMemberToDelete] = useState<Member | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['members', page],
    queryFn: () => apiFetch<MembersPage>(`/members?page=${page}&per_page=${PER_PAGE}`),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiFetch<undefined>(`/members/${id}`, { method: 'DELETE' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['members'] })
      setMemberToDelete(null)
    },
  })

  const totalPages = data ? Math.ceil(data.total / PER_PAGE) : 1

  const currentUserId = user.memberId
  const actionProps = { setModal, setMemberToDelete }

  return (
    <div className="min-h-screen bg-zinc-100 dark:bg-zinc-900">
      <AppHeader />

      <main className="mx-auto max-w-6xl px-4 py-6 sm:py-8">
        <div className="mb-5 flex items-center justify-between">
          <div>
            <h1 className="text-base font-semibold text-zinc-900 dark:text-zinc-100">
              {t('members.title')}
            </h1>
            {data && (
              <p className="mt-0.5 text-sm text-zinc-500 dark:text-zinc-400">
                {t('members.count', { count: data.total })}
              </p>
            )}
          </div>
          <Button size="sm" onClick={() => setModal({ type: 'create' })}>
            <PlusIcon className="mr-1.5 size-4" aria-hidden="true" />
            {t('members.add')}
          </Button>
        </div>

        <div className="rounded-xl border border-zinc-200 bg-white dark:border-zinc-700/60 dark:bg-zinc-800">
          {isLoading ? (
            <div className="flex items-center justify-center py-16">
              <Spinner className="size-6" />
            </div>
          ) : data?.items.length === 0 ? (
            <div className="py-16 text-center">
              <p className="text-sm text-zinc-500 dark:text-zinc-400">{t('members.empty')}</p>
            </div>
          ) : (
            <>
              {/* Mobile: card list */}
              <ul className="divide-y divide-zinc-100 dark:divide-zinc-700/60 sm:hidden">
                {data?.items.map((member) => (
                  <li key={member.id} className="flex items-center justify-between gap-3 px-4 py-3">
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium text-zinc-900 dark:text-zinc-100">
                        {member.username}
                      </p>
                      {member.display_name && (
                        <p className="truncate text-xs text-zinc-500 dark:text-zinc-400">
                          {member.display_name}
                        </p>
                      )}
                      <div className="mt-1.5">
                        <RoleBadge role={member.role} />
                      </div>
                    </div>
                    <div className="shrink-0">
                      <ActionButtons member={member} isSelf={member.id === currentUserId} {...actionProps} />
                    </div>
                  </li>
                ))}
              </ul>

              {/* Desktop: table */}
              <table className="hidden w-full text-sm sm:table">
                <thead>
                  <tr className="border-b border-zinc-200 dark:border-zinc-700">
                    <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500 dark:text-zinc-400">
                      {t('field.username')}
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500 dark:text-zinc-400">
                      {t('field.displayName')}
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500 dark:text-zinc-400">
                      {t('field.role')}
                    </th>
                    <th className="px-4 py-3" />
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-100 dark:divide-zinc-700/60">
                  {data?.items.map((member) => (
                    <tr key={member.id} className="group">
                      <td className="px-4 py-3 font-medium text-zinc-900 dark:text-zinc-100">
                        {member.username}
                      </td>
                      <td className="px-4 py-3 text-zinc-500 dark:text-zinc-400">
                        {member.display_name ?? (
                          <span className="italic text-zinc-300 dark:text-zinc-600">—</span>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <RoleBadge role={member.role} />
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex justify-end">
                          <ActionButtons member={member} isSelf={member.id === currentUserId} {...actionProps} />
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}

          <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
        </div>
      </main>

      {/* Create / Edit modal */}
      <Dialog open={modal !== null} onClose={() => setModal(null)} className="relative z-50">
        <div className="fixed inset-0 bg-black/25 backdrop-blur-sm dark:bg-black/40" aria-hidden="true" />
        <div className="fixed inset-0 flex items-center justify-center p-4">
          <DialogPanel className="w-full max-w-md max-h-[90vh] overflow-y-auto rounded-xl border border-zinc-200 bg-white p-6 shadow-xl dark:border-zinc-700/60 dark:bg-zinc-800">
            <div className="mb-5 flex items-center justify-between">
              <DialogTitle className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                {modal?.type === 'create' ? t('members.modalTitleCreate') : t('members.modalTitleEdit')}
              </DialogTitle>
              <button
                onClick={() => setModal(null)}
                className="cursor-pointer rounded-lg p-1 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-600 dark:hover:bg-zinc-700 dark:hover:text-zinc-200"
              >
                <XMarkIcon className="size-4" aria-hidden="true" />
              </button>
            </div>

            {modal?.type === 'create' && <CreateForm onSuccess={() => setModal(null)} />}
            {modal?.type === 'edit' && (
              <EditForm
                member={modal.member}
                isSelf={modal.member.id === currentUserId}
                onSuccess={() => setModal(null)}
              />
            )}
          </DialogPanel>
        </div>
      </Dialog>

      {/* Delete confirmation modal */}
      <Dialog open={memberToDelete !== null} onClose={() => setMemberToDelete(null)} className="relative z-50">
        <div className="fixed inset-0 bg-black/25 backdrop-blur-sm dark:bg-black/40" aria-hidden="true" />
        <div className="fixed inset-0 flex items-center justify-center p-4">
          <DialogPanel className="w-full max-w-sm rounded-xl border border-zinc-200 bg-white p-6 shadow-xl dark:border-zinc-700/60 dark:bg-zinc-800">
            <DialogTitle className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">
              {t('members.deleteTitle')}
            </DialogTitle>
            <p className="mt-2 text-sm text-zinc-500 dark:text-zinc-400">
              {t('members.deleteConfirm', { username: memberToDelete?.username })}
            </p>
            <div className="mt-5 flex justify-end gap-2">
              <Button variant="ghost" onClick={() => setMemberToDelete(null)}>
                {t('action.cancel')}
              </Button>
              <Button
                variant="danger"
                loading={deleteMutation.isPending}
                onClick={() => memberToDelete && deleteMutation.mutate(memberToDelete.id)}
              >
                {t('action.delete')}
              </Button>
            </div>
          </DialogPanel>
        </div>
      </Dialog>
    </div>
  )
}

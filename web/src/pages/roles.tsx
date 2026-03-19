import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  PlusIcon,
  PencilSquareIcon,
  TrashIcon,
  AdjustmentsHorizontalIcon,
} from '@heroicons/react/24/outline'
import { useTranslation } from 'react-i18next'
import { ApiError, rolesApi } from '@/lib/api'
import type { Role, Permission } from '@/lib/types'
import { AppHeader } from '@/components/app-header'
import { Modal } from '@/components/ui/modal'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'

// --- Constants ---

const PROTECTED_ROLE = 'ADMIN'
const PROTECTED_PERMISSION = '*'

// --- Schemas ---

const roleSchema = z.object({
  id: z.string().min(1).max(20),
  description: z.string().min(1).max(500),
})

const roleEditSchema = z.object({
  description: z.string().min(1).max(500),
})

const permissionSchema = z.object({
  id: z.string().min(1).max(50),
  description: z.string().min(1).max(500),
})

const permissionEditSchema = z.object({
  description: z.string().min(1).max(500),
})

type RoleFormData = z.infer<typeof roleSchema>
type RoleEditFormData = z.infer<typeof roleEditSchema>
type PermissionFormData = z.infer<typeof permissionSchema>
type PermissionEditFormData = z.infer<typeof permissionEditSchema>

// --- Permission badge ---

function PermBadge({ perm }: { perm: string }) {
  return (
    <span className="inline-flex items-center rounded-md bg-zinc-100 px-1.5 py-0.5 text-xs font-mono text-zinc-600 ring-1 ring-inset ring-zinc-500/20 dark:bg-zinc-700 dark:text-zinc-300 dark:ring-zinc-600/30">
      {perm}
    </span>
  )
}

// --- Role create form ---

function RoleCreateForm({ onSuccess }: { onSuccess: () => void }) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<RoleFormData>({ resolver: zodResolver(roleSchema) })

  const mutation = useMutation({
    mutationFn: (data: RoleFormData) => rolesApi.createRole(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      onSuccess()
    },
    onError: (err: Error) => {
      if (err instanceof ApiError && err.status === 409) {
        setError('id', { message: t('roles.alreadyExists') })
      }
    },
  })

  return (
    <form onSubmit={handleSubmit((data) => mutation.mutate(data))} className="space-y-4">
      <Field label={t('roles.fieldId')} error={errors.id?.message}>
        <Input {...register('id')} placeholder="OPERATOR" autoComplete="off" autoFocus />
      </Field>
      <Field label={t('roles.fieldDescription')} error={errors.description?.message}>
        <Input {...register('description')} placeholder={t('roles.descriptionPlaceholder')} autoComplete="off" />
      </Field>
      {mutation.isError && !errors.id && (
        <p className="text-xs text-red-600 dark:text-red-400">{t('error.generic')}</p>
      )}
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onSuccess}>{t('action.cancel')}</Button>
        <Button type="submit" loading={mutation.isPending}>{t('roles.create')}</Button>
      </div>
    </form>
  )
}

// --- Role edit form ---

function RoleEditForm({ role, onSuccess }: { role: Role; onSuccess: () => void }) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RoleEditFormData>({
    resolver: zodResolver(roleEditSchema),
    defaultValues: { description: role.description },
  })

  const mutation = useMutation({
    mutationFn: (data: RoleEditFormData) => rolesApi.updateRole(role.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      onSuccess()
    },
  })

  return (
    <form onSubmit={handleSubmit((data) => mutation.mutate(data))} className="space-y-4">
      <Field label={t('roles.fieldId')}>
        <Input value={role.id} disabled />
      </Field>
      <Field label={t('roles.fieldDescription')} error={errors.description?.message}>
        <Input {...register('description')} autoFocus />
      </Field>
      {mutation.isError && (
        <p className="text-xs text-red-600 dark:text-red-400">{t('error.generic')}</p>
      )}
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onSuccess}>{t('action.cancel')}</Button>
        <Button type="submit" loading={mutation.isPending}>{t('action.save')}</Button>
      </div>
    </form>
  )
}

// --- Role permissions form ---

function RolePermissionsForm({
  role,
  allPermissions,
  onSuccess,
}: {
  role: Role
  allPermissions: Permission[]
  onSuccess: () => void
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [selected, setSelected] = useState<Set<string>>(new Set(role.permissions))

  const mutation = useMutation({
    mutationFn: (perms: string[]) => rolesApi.setRolePermissions(role.id, perms),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      onSuccess()
    },
  })

  function toggle(id: string) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const assignable = allPermissions.filter((p) => p.id !== PROTECTED_PERMISSION)

  return (
    <div className="space-y-4">
      <p className="text-sm text-zinc-500 dark:text-zinc-400">
        {t('roles.permissionsHint', { role: role.id })}
      </p>
      <div className="space-y-2">
        {assignable.map((perm) => (
          <label
            key={perm.id}
            className="flex cursor-pointer items-center gap-3 rounded-lg border border-zinc-200 px-3 py-2.5 transition-colors hover:bg-zinc-50 dark:border-zinc-700 dark:hover:bg-zinc-700/50"
          >
            <input
              type="checkbox"
              checked={selected.has(perm.id)}
              onChange={() => toggle(perm.id)}
              className="size-4 rounded border-zinc-300 text-zinc-900 focus:ring-zinc-500 dark:border-zinc-600 dark:bg-zinc-800 dark:text-zinc-100"
            />
            <div className="min-w-0">
              <p className="text-sm font-mono text-zinc-900 dark:text-zinc-100">{perm.id}</p>
              <p className="text-xs text-zinc-500 dark:text-zinc-400">{perm.description}</p>
            </div>
          </label>
        ))}
      </div>
      {mutation.isError && (
        <p className="text-xs text-red-600 dark:text-red-400">{t('error.generic')}</p>
      )}
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onSuccess}>{t('action.cancel')}</Button>
        <Button
          loading={mutation.isPending}
          onClick={() => mutation.mutate([...selected])}
        >
          {t('action.save')}
        </Button>
      </div>
    </div>
  )
}

// --- Permission create form ---

function PermissionCreateForm({ onSuccess }: { onSuccess: () => void }) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<PermissionFormData>({ resolver: zodResolver(permissionSchema) })

  const mutation = useMutation({
    mutationFn: (data: PermissionFormData) => rolesApi.createPermission(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] })
      onSuccess()
    },
    onError: (err: Error) => {
      if (err instanceof ApiError && err.status === 409) {
        setError('id', { message: t('permissions.alreadyExists') })
      }
    },
  })

  return (
    <form onSubmit={handleSubmit((data) => mutation.mutate(data))} className="space-y-4">
      <Field label={t('permissions.fieldId')} error={errors.id?.message}>
        <Input {...register('id')} placeholder="gate:view_logs" autoComplete="off" autoFocus />
      </Field>
      <Field label={t('permissions.fieldDescription')} error={errors.description?.message}>
        <Input {...register('description')} placeholder={t('permissions.descriptionPlaceholder')} autoComplete="off" />
      </Field>
      {mutation.isError && !errors.id && (
        <p className="text-xs text-red-600 dark:text-red-400">{t('error.generic')}</p>
      )}
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onSuccess}>{t('action.cancel')}</Button>
        <Button type="submit" loading={mutation.isPending}>{t('permissions.create')}</Button>
      </div>
    </form>
  )
}

// --- Permission edit form ---

function PermissionEditForm({ permission, onSuccess }: { permission: Permission; onSuccess: () => void }) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<PermissionEditFormData>({
    resolver: zodResolver(permissionEditSchema),
    defaultValues: { description: permission.description },
  })

  const mutation = useMutation({
    mutationFn: (data: PermissionEditFormData) => rolesApi.updatePermission(permission.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] })
      onSuccess()
    },
  })

  return (
    <form onSubmit={handleSubmit((data) => mutation.mutate(data))} className="space-y-4">
      <Field label={t('permissions.fieldId')}>
        <Input value={permission.id} disabled />
      </Field>
      <Field label={t('permissions.fieldDescription')} error={errors.description?.message}>
        <Input {...register('description')} autoFocus />
      </Field>
      {mutation.isError && (
        <p className="text-xs text-red-600 dark:text-red-400">{t('error.generic')}</p>
      )}
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onSuccess}>{t('action.cancel')}</Button>
        <Button type="submit" loading={mutation.isPending}>{t('action.save')}</Button>
      </div>
    </form>
  )
}

// --- Main page ---

type RoleModal =
  | { type: 'create-role' }
  | { type: 'edit-role'; role: Role }
  | { type: 'role-permissions'; role: Role }
  | { type: 'create-permission' }
  | { type: 'edit-permission'; permission: Permission }
  | null

export function RolesPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const [modal, setModal] = useState<RoleModal>(null)
  const [roleToDelete, setRoleToDelete] = useState<Role | null>(null)
  const [permToDelete, setPermToDelete] = useState<Permission | null>(null)

  const { data: roles, isLoading: rolesLoading } = useQuery({
    queryKey: ['roles'],
    queryFn: () => rolesApi.listRoles(),
  })

  const { data: permissions, isLoading: permsLoading } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => rolesApi.listPermissions(),
  })

  const deleteRoleMutation = useMutation({
    mutationFn: (id: string) => rolesApi.deleteRole(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      setRoleToDelete(null)
    },
  })

  const deletePermMutation = useMutation({
    mutationFn: (id: string) => rolesApi.deletePermission(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] })
      setPermToDelete(null)
    },
  })

  const isLoading = rolesLoading || permsLoading

  return (
    <div className="min-h-screen bg-zinc-100 dark:bg-zinc-900">
      <AppHeader />

      <main className="mx-auto max-w-6xl space-y-8 px-4 py-6 sm:py-8">
        {/* --- Roles section --- */}
        <section>
          <div className="mb-5 flex items-center justify-between">
            <div>
              <h1 className="text-base font-semibold text-zinc-900 dark:text-zinc-100">
                {t('roles.title')}
              </h1>
              {roles && (
                <p className="mt-0.5 text-sm text-zinc-500 dark:text-zinc-400">
                  {t('roles.count', { count: roles.length })}
                </p>
              )}
            </div>
            <Button size="sm" onClick={() => setModal({ type: 'create-role' })}>
              <PlusIcon className="mr-1.5 size-4" aria-hidden="true" />
              {t('roles.add')}
            </Button>
          </div>

          <div className="rounded-xl border border-zinc-200 bg-white dark:border-zinc-700/60 dark:bg-zinc-800">
            {isLoading ? (
              <div className="flex items-center justify-center py-16">
                <Spinner className="size-6" />
              </div>
            ) : roles?.length === 0 ? (
              <div className="py-16 text-center">
                <p className="text-sm text-zinc-500 dark:text-zinc-400">{t('roles.empty')}</p>
              </div>
            ) : (
              <ul className="divide-y divide-zinc-100 dark:divide-zinc-700/60">
                {roles?.map((role) => {
                  const isProtected = role.id === PROTECTED_ROLE
                  return (
                    <li key={role.id} className="group px-4 py-3">
                      <div className="flex items-center justify-between gap-3">
                        <div className="min-w-0">
                          <p className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                            {role.id}
                            {isProtected && (
                              <span className="ml-2 text-xs font-normal text-amber-600 dark:text-amber-400">
                                {t('roles.protected')}
                              </span>
                            )}
                          </p>
                          <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
                            {role.description}
                          </p>
                        </div>
                        <div className="flex shrink-0 items-center gap-1">
                          {!isProtected && (
                            <>
                              <button
                                onClick={() => setModal({ type: 'role-permissions', role })}
                                className="cursor-pointer rounded-lg p-1.5 text-zinc-400 transition-all hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-700 dark:hover:text-zinc-200"
                                title={t('roles.managePermissions')}
                                aria-label={t('roles.managePermissions')}
                              >
                                <AdjustmentsHorizontalIcon className="size-4" aria-hidden="true" />
                              </button>
                              <button
                                onClick={() => setModal({ type: 'edit-role', role })}
                                className="cursor-pointer rounded-lg p-1.5 text-zinc-400 transition-all hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-700 dark:hover:text-zinc-200"
                                title={t('action.edit')}
                                aria-label={t('action.edit')}
                              >
                                <PencilSquareIcon className="size-4" aria-hidden="true" />
                              </button>
                              <button
                                onClick={() => setRoleToDelete(role)}
                                className="cursor-pointer rounded-lg p-1.5 text-zinc-400 transition-all hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950 dark:hover:text-red-400"
                                title={t('action.delete')}
                                aria-label={t('action.delete')}
                              >
                                <TrashIcon className="size-4" aria-hidden="true" />
                              </button>
                            </>
                          )}
                        </div>
                      </div>
                      {role.permissions.length > 0 && (
                        <div className="mt-2 flex flex-wrap gap-1">
                          {role.permissions.map((p) => (
                            <PermBadge key={p} perm={p} />
                          ))}
                        </div>
                      )}
                    </li>
                  )
                })}
              </ul>
            )}
          </div>
        </section>

        {/* --- Permissions section --- */}
        <section>
          <div className="mb-5 flex items-center justify-between">
            <div>
              <h1 className="text-base font-semibold text-zinc-900 dark:text-zinc-100">
                {t('permissions.title')}
              </h1>
              {permissions && (
                <p className="mt-0.5 text-sm text-zinc-500 dark:text-zinc-400">
                  {t('permissions.count', { count: permissions.length })}
                </p>
              )}
            </div>
            <Button size="sm" onClick={() => setModal({ type: 'create-permission' })}>
              <PlusIcon className="mr-1.5 size-4" aria-hidden="true" />
              {t('permissions.add')}
            </Button>
          </div>

          <div className="rounded-xl border border-zinc-200 bg-white dark:border-zinc-700/60 dark:bg-zinc-800">
            {isLoading ? (
              <div className="flex items-center justify-center py-16">
                <Spinner className="size-6" />
              </div>
            ) : permissions?.length === 0 ? (
              <div className="py-16 text-center">
                <p className="text-sm text-zinc-500 dark:text-zinc-400">{t('permissions.empty')}</p>
              </div>
            ) : (
              <ul className="divide-y divide-zinc-100 dark:divide-zinc-700/60">
                {permissions?.map((perm) => {
                  const isProtected = perm.id === PROTECTED_PERMISSION
                  return (
                    <li key={perm.id} className="flex items-center justify-between gap-3 px-4 py-3">
                      <div className="min-w-0">
                        <p className="text-sm font-mono text-zinc-900 dark:text-zinc-100">
                          {perm.id}
                          {isProtected && (
                            <span className="ml-2 font-sans text-xs font-normal text-amber-600 dark:text-amber-400">
                              {t('permissions.protected')}
                            </span>
                          )}
                        </p>
                        <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
                          {perm.description}
                        </p>
                      </div>
                      {!isProtected && (
                        <div className="flex shrink-0 items-center gap-1">
                          <button
                            onClick={() => setModal({ type: 'edit-permission', permission: perm })}
                            className="cursor-pointer rounded-lg p-1.5 text-zinc-400 transition-all hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-700 dark:hover:text-zinc-200"
                            title={t('action.edit')}
                            aria-label={t('action.edit')}
                          >
                            <PencilSquareIcon className="size-4" aria-hidden="true" />
                          </button>
                          <button
                            onClick={() => setPermToDelete(perm)}
                            className="cursor-pointer rounded-lg p-1.5 text-zinc-400 transition-all hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950 dark:hover:text-red-400"
                            title={t('action.delete')}
                            aria-label={t('action.delete')}
                          >
                            <TrashIcon className="size-4" aria-hidden="true" />
                          </button>
                        </div>
                      )}
                    </li>
                  )
                })}
              </ul>
            )}
          </div>
        </section>
      </main>

      {/* --- Modals --- */}

      <Modal
        open={modal?.type === 'create-role'}
        onClose={() => setModal(null)}
        title={t('roles.modalTitleCreate')}
      >
        <RoleCreateForm onSuccess={() => setModal(null)} />
      </Modal>

      <Modal
        open={modal?.type === 'edit-role'}
        onClose={() => setModal(null)}
        title={t('roles.modalTitleEdit')}
      >
        {modal?.type === 'edit-role' && (
          <RoleEditForm role={modal.role} onSuccess={() => setModal(null)} />
        )}
      </Modal>

      <Modal
        open={modal?.type === 'role-permissions'}
        onClose={() => setModal(null)}
        title={t('roles.modalTitlePermissions')}
        size="lg"
      >
        {modal?.type === 'role-permissions' && permissions && (
          <RolePermissionsForm
            role={modal.role}
            allPermissions={permissions}
            onSuccess={() => setModal(null)}
          />
        )}
      </Modal>

      <Modal
        open={modal?.type === 'create-permission'}
        onClose={() => setModal(null)}
        title={t('permissions.modalTitleCreate')}
      >
        <PermissionCreateForm onSuccess={() => setModal(null)} />
      </Modal>

      <Modal
        open={modal?.type === 'edit-permission'}
        onClose={() => setModal(null)}
        title={t('permissions.modalTitleEdit')}
      >
        {modal?.type === 'edit-permission' && (
          <PermissionEditForm permission={modal.permission} onSuccess={() => setModal(null)} />
        )}
      </Modal>

      {/* Delete role confirmation */}
      <Modal
        open={roleToDelete !== null}
        onClose={() => setRoleToDelete(null)}
        title={t('roles.deleteTitle')}
        size="sm"
      >
        <p className="text-sm text-zinc-500 dark:text-zinc-400">
          {t('roles.deleteConfirm', { role: roleToDelete?.id })}
        </p>
        {deleteRoleMutation.isError && (
          <p className="mt-2 text-xs text-red-600 dark:text-red-400">
            {deleteRoleMutation.error instanceof ApiError && deleteRoleMutation.error.status === 422
              ? t('roles.inUseError')
              : t('error.generic')}
          </p>
        )}
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="ghost" onClick={() => setRoleToDelete(null)}>{t('action.cancel')}</Button>
          <Button
            variant="danger"
            loading={deleteRoleMutation.isPending}
            onClick={() => roleToDelete && deleteRoleMutation.mutate(roleToDelete.id)}
          >
            {t('action.delete')}
          </Button>
        </div>
      </Modal>

      {/* Delete permission confirmation */}
      <Modal
        open={permToDelete !== null}
        onClose={() => setPermToDelete(null)}
        title={t('permissions.deleteTitle')}
        size="sm"
      >
        <p className="text-sm text-zinc-500 dark:text-zinc-400">
          {t('permissions.deleteConfirm', { permission: permToDelete?.id })}
        </p>
        {deletePermMutation.isError && (
          <p className="mt-2 text-xs text-red-600 dark:text-red-400">
            {deletePermMutation.error instanceof ApiError && deletePermMutation.error.status === 422
              ? t('permissions.inUseError')
              : t('error.generic')}
          </p>
        )}
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="ghost" onClick={() => setPermToDelete(null)}>{t('action.cancel')}</Button>
          <Button
            variant="danger"
            loading={deletePermMutation.isPending}
            onClick={() => permToDelete && deletePermMutation.mutate(permToDelete.id)}
          >
            {t('action.delete')}
          </Button>
        </div>
      </Modal>
    </div>
  )
}

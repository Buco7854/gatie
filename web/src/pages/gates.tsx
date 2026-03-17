import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  PlusIcon,
  PencilSquareIcon,
  TrashIcon,
  ArrowPathIcon,
  ClipboardDocumentIcon,
  CheckIcon,
} from '@heroicons/react/24/outline'
import { useTranslation } from 'react-i18next'
import { useAuth } from '@/hooks/use-auth'
import { gatesApi } from '@/lib/api'
import type { Gate } from '@/lib/types'
import { AppHeader } from '@/components/app-header'
import { Modal } from '@/components/ui/modal'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { Pagination } from '@/components/ui/pagination'

// --- Schemas ---

const gateSchema = z.object({
  name: z.string().min(1).max(100),
  status_ttl_seconds: z.coerce.number().int().min(1).max(86400).default(60),
})

type GateFormData = z.infer<typeof gateSchema>

// --- Token reveal panel ---

function TokenReveal({ token, onClose }: { token: string; onClose: () => void }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  function copy() {
    navigator.clipboard.writeText(token).then(
      () => {
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
      },
      () => {
        // Clipboard API not available (e.g. non-HTTPS)
      },
    )
  }

  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-amber-200 bg-amber-50 p-3 dark:border-amber-800/50 dark:bg-amber-950/40">
        <p className="text-xs text-amber-700 dark:text-amber-300">{t('gates.tokenWarning')}</p>
      </div>

      <div className="flex items-center gap-2">
        <code className="min-w-0 flex-1 overflow-x-auto rounded-lg bg-zinc-100 px-3 py-2 font-mono text-xs text-zinc-900 dark:bg-zinc-900 dark:text-zinc-100">
          {token}
        </code>
        <button
          onClick={copy}
          className="shrink-0 cursor-pointer rounded-lg p-2 text-zinc-400 transition-all hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-700 dark:hover:text-zinc-200"
          title={copied ? t('gates.tokenCopied') : t('gates.tokenCopy')}
        >
          {copied ? (
            <CheckIcon className="size-4 text-green-500" aria-hidden="true" />
          ) : (
            <ClipboardDocumentIcon className="size-4" aria-hidden="true" />
          )}
        </button>
      </div>

      <div className="flex justify-end">
        <Button onClick={onClose}>{t('action.confirm')}</Button>
      </div>
    </div>
  )
}

// --- Gate form ---

function GateForm({
  defaultValues,
  submitLabel,
  onSubmit,
  isPending,
  error,
  onCancel,
}: {
  defaultValues?: Partial<GateFormData>
  submitLabel: string
  onSubmit: (data: GateFormData) => void
  isPending: boolean
  error?: boolean
  onCancel: () => void
}) {
  const { t } = useTranslation()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<GateFormData>({
    resolver: zodResolver(gateSchema),
    defaultValues: { status_ttl_seconds: 60, ...defaultValues },
  })

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <Field
        label={t('gates.fieldName')}
        error={errors.name ? t('validation.required') : undefined}
      >
        <Input
          {...register('name')}
          placeholder={t('gates.fieldNamePlaceholder')}
          autoComplete="off"
          autoFocus
        />
      </Field>

      <Field
        label={t('gates.fieldStatusTTL')}
        hint={t('gates.fieldStatusTTLHint')}
        error={errors.status_ttl_seconds ? t('validation.minLength', { min: 1 }) : undefined}
      >
        <Input {...register('status_ttl_seconds')} type="number" min={1} max={86400} />
      </Field>

      {error && <p className="text-xs text-red-600 dark:text-red-400">{t('error.generic')}</p>}

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="ghost" onClick={onCancel}>
          {t('action.cancel')}
        </Button>
        <Button type="submit" loading={isPending}>
          {submitLabel}
        </Button>
      </div>
    </form>
  )
}

// --- Main page ---

type ModalState =
  | { type: 'create' }
  | { type: 'edit'; gate: Gate }
  | { type: 'token'; token: string; gateName: string }
  | { type: 'regenerate-confirm'; gate: Gate }
  | null

const PER_PAGE = 20

export function GatesPage() {
  const { t } = useTranslation()
  useAuth()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [modal, setModal] = useState<ModalState>(null)
  const [gateToDelete, setGateToDelete] = useState<Gate | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['gates', page],
    queryFn: () => gatesApi.listGates(page, PER_PAGE),
  })

  const createMutation = useMutation({
    mutationFn: (data: GateFormData) => gatesApi.createGate(data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['gates'] })
      setModal({ type: 'token', token: result.token, gateName: result.name })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: GateFormData }) =>
      gatesApi.updateGate(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['gates'] })
      setModal(null)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => gatesApi.deleteGate(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['gates'] })
      setGateToDelete(null)
    },
  })

  const regenerateMutation = useMutation({
    mutationFn: (id: string) => gatesApi.regenerateGateToken(id),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['gates'] })
      setModal({ type: 'token', token: result.token, gateName: result.name })
    },
  })

  const totalPages = data ? Math.ceil(data.total / PER_PAGE) : 1

  return (
    <div className="min-h-screen bg-zinc-100 dark:bg-zinc-900">
      <AppHeader />

      <main className="mx-auto max-w-6xl px-4 py-6 sm:py-8">
        <div className="mb-5 flex items-center justify-between">
          <div>
            <h1 className="text-base font-semibold text-zinc-900 dark:text-zinc-100">
              {t('gates.title')}
            </h1>
            {data && (
              <p className="mt-0.5 text-sm text-zinc-500 dark:text-zinc-400">
                {t('gates.count', { count: data.total })}
              </p>
            )}
          </div>
          <Button size="sm" onClick={() => setModal({ type: 'create' })}>
            <PlusIcon className="mr-1.5 size-4" aria-hidden="true" />
            {t('gates.add')}
          </Button>
        </div>

        <div className="rounded-xl border border-zinc-200 bg-white dark:border-zinc-700/60 dark:bg-zinc-800">
          {isLoading ? (
            <div className="flex items-center justify-center py-16">
              <Spinner className="size-6" />
            </div>
          ) : data?.items.length === 0 ? (
            <div className="py-16 text-center">
              <p className="text-sm text-zinc-500 dark:text-zinc-400">{t('gates.empty')}</p>
            </div>
          ) : (
            <>
              {/* Mobile: card list */}
              <ul className="divide-y divide-zinc-100 dark:divide-zinc-700/60 sm:hidden">
                {data?.items.map((gate) => (
                  <li key={gate.id} className="flex items-center justify-between gap-3 px-4 py-3">
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium text-zinc-900 dark:text-zinc-100">
                        {gate.name}
                      </p>
                      <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
                        TTL {gate.status_ttl_seconds}s
                      </p>
                    </div>
                    <GateActions gate={gate} setModal={setModal} setGateToDelete={setGateToDelete} />
                  </li>
                ))}
              </ul>

              {/* Desktop: table */}
              <table className="hidden w-full text-sm sm:table">
                <thead>
                  <tr className="border-b border-zinc-200 dark:border-zinc-700">
                    <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500 dark:text-zinc-400">
                      {t('gates.fieldName')}
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500 dark:text-zinc-400">
                      {t('gates.fieldStatusTTL')}
                    </th>
                    <th className="px-4 py-3" />
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-100 dark:divide-zinc-700/60">
                  {data?.items.map((gate) => (
                    <tr key={gate.id} className="group">
                      <td className="px-4 py-3 font-medium text-zinc-900 dark:text-zinc-100">
                        {gate.name}
                      </td>
                      <td className="px-4 py-3 text-zinc-500 dark:text-zinc-400">
                        {gate.status_ttl_seconds}s
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex justify-end">
                          <GateActions gate={gate} setModal={setModal} setGateToDelete={setGateToDelete} />
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

      {/* Create modal */}
      <Modal open={modal?.type === 'create'} onClose={() => setModal(null)} title={t('gates.modalTitleCreate')}>
        <GateForm
          submitLabel={t('gates.create')}
          onSubmit={(data) => createMutation.mutate(data)}
          isPending={createMutation.isPending}
          error={createMutation.isError}
          onCancel={() => setModal(null)}
        />
      </Modal>

      {/* Edit modal */}
      <Modal open={modal?.type === 'edit'} onClose={() => setModal(null)} title={t('gates.modalTitleEdit')}>
        {modal?.type === 'edit' && (
          <GateForm
            defaultValues={{ name: modal.gate.name, status_ttl_seconds: modal.gate.status_ttl_seconds }}
            submitLabel={t('action.save')}
            onSubmit={(data) => updateMutation.mutate({ id: modal.gate.id, data })}
            isPending={updateMutation.isPending}
            error={updateMutation.isError}
            onCancel={() => setModal(null)}
          />
        )}
      </Modal>

      {/* Token modal */}
      <Modal open={modal?.type === 'token'} onClose={() => setModal(null)} title={t('gates.tokenTitle')} size="lg">
        {modal?.type === 'token' && (
          <TokenReveal token={modal.token} onClose={() => setModal(null)} />
        )}
      </Modal>

      {/* Regenerate confirm modal */}
      <Modal
        open={modal?.type === 'regenerate-confirm'}
        onClose={() => setModal(null)}
        title={t('gates.tokenRegenerateConfirmTitle')}
        size="sm"
      >
        <p className="text-sm text-zinc-500 dark:text-zinc-400">
          {t('gates.tokenRegenerateConfirmBody')}
        </p>
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="ghost" onClick={() => setModal(null)}>
            {t('action.cancel')}
          </Button>
          <Button
            variant="danger"
            loading={regenerateMutation.isPending}
            onClick={() => modal?.type === 'regenerate-confirm' && regenerateMutation.mutate(modal.gate.id)}
          >
            {t('gates.tokenRegenerate')}
          </Button>
        </div>
      </Modal>

      {/* Delete confirm modal */}
      <Modal
        open={gateToDelete !== null}
        onClose={() => setGateToDelete(null)}
        title={t('gates.deleteTitle')}
        size="sm"
      >
        <p className="text-sm text-zinc-500 dark:text-zinc-400">
          {t('gates.deleteConfirm', { name: gateToDelete?.name })}
        </p>
        {deleteMutation.isError && (
          <p className="mt-2 text-xs text-red-600 dark:text-red-400">{t('error.generic')}</p>
        )}
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="ghost" onClick={() => setGateToDelete(null)}>
            {t('action.cancel')}
          </Button>
          <Button
            variant="danger"
            loading={deleteMutation.isPending}
            onClick={() => gateToDelete && deleteMutation.mutate(gateToDelete.id)}
          >
            {t('action.delete')}
          </Button>
        </div>
      </Modal>
    </div>
  )
}

// --- Sub-components ---

function GateActions({
  gate,
  setModal,
  setGateToDelete,
}: {
  gate: Gate
  setModal: (s: ModalState) => void
  setGateToDelete: (g: Gate) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="flex items-center gap-1">
      <button
        onClick={() => setModal({ type: 'edit', gate })}
        className="cursor-pointer rounded-lg p-1.5 text-zinc-400 transition-all hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-700 dark:hover:text-zinc-200"
        title={t('action.edit')}
      >
        <PencilSquareIcon className="size-4" aria-hidden="true" />
      </button>
      <button
        onClick={() => setModal({ type: 'regenerate-confirm', gate })}
        className="cursor-pointer rounded-lg p-1.5 text-zinc-400 transition-all hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-700 dark:hover:text-zinc-200"
        title={t('gates.tokenRegenerate')}
      >
        <ArrowPathIcon className="size-4" aria-hidden="true" />
      </button>
      <button
        onClick={() => setGateToDelete(gate)}
        className="cursor-pointer rounded-lg p-1.5 text-zinc-400 transition-all hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950 dark:hover:text-red-400"
        title={t('action.delete')}
      >
        <TrashIcon className="size-4" aria-hidden="true" />
      </button>
    </div>
  )
}

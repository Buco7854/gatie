import { createRouter, createRoute, createRootRoute, Outlet, Navigate } from '@tanstack/react-router'
import { useAuth } from '@/hooks/use-auth'
import { Spinner } from '@/components/ui/spinner'
import { SetupPage } from '@/pages/setup'
import { LoginPage } from '@/pages/login'
import { DashboardPage } from '@/pages/dashboard'
import { MembersPage } from '@/pages/members'
import { GatesPage } from '@/pages/gates'
import { RolesPage } from '@/pages/roles'
import { NotFoundPage } from '@/pages/not-found'

const rootRoute = createRootRoute({
  component: () => <Outlet />,
  notFoundComponent: NotFoundPage,
})

// --- Public routes (no auth required) ---

const setupRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/setup',
  component: SetupPage,
})

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginPage,
})

// --- Authenticated layout ---

function AuthLayout() {
  const { status } = useAuth()

  if (status === 'loading') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-100 dark:bg-zinc-900">
        <Spinner />
      </div>
    )
  }

  if (status === 'unauthenticated') {
    return <Navigate to="/login" />
  }

  return <Outlet />
}

const authLayout = createRoute({
  getParentRoute: () => rootRoute,
  id: 'auth',
  component: AuthLayout,
})

const dashboardRoute = createRoute({
  getParentRoute: () => authLayout,
  path: '/',
  component: DashboardPage,
})

// --- Admin layout ---

function AdminLayout() {
  const { user } = useAuth()

  if (user.role !== 'ADMIN') {
    return <Navigate to="/" />
  }

  return <Outlet />
}

const adminLayout = createRoute({
  getParentRoute: () => authLayout,
  id: 'admin',
  component: AdminLayout,
})

const membersRoute = createRoute({
  getParentRoute: () => adminLayout,
  path: '/members',
  component: MembersPage,
})

const gatesRoute = createRoute({
  getParentRoute: () => adminLayout,
  path: '/gates',
  component: GatesPage,
})

const rolesRoute = createRoute({
  getParentRoute: () => adminLayout,
  path: '/roles',
  component: RolesPage,
})

const routeTree = rootRoute.addChildren([
  setupRoute,
  loginRoute,
  authLayout.addChildren([
    dashboardRoute,
    adminLayout.addChildren([membersRoute, gatesRoute, rolesRoute]),
  ]),
])

export const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

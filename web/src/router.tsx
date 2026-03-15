import { createRouter, createRoute, createRootRoute, Outlet } from '@tanstack/react-router'
import { SetupPage } from '@/pages/setup'
import { LoginPage } from '@/pages/login'
import { DashboardPage } from '@/pages/dashboard'
import { MembersPage } from '@/pages/members'

const rootRoute = createRootRoute({
  component: () => <Outlet />,
})

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

const dashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage,
})

const membersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/members',
  component: MembersPage,
})

const routeTree = rootRoute.addChildren([setupRoute, loginRoute, dashboardRoute, membersRoute])

export const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

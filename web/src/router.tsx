import {
  createRouter,
  createRootRoute,
  createRoute,
  redirect,
  Outlet,
} from "@tanstack/react-router"
import { ThemeProvider } from "next-themes"
import { auth } from "@/lib/auth"
import { SetupPage } from "@/pages/setup"
import { LoginPage } from "@/pages/login"
import { DashboardPage } from "@/pages/dashboard"

const rootRoute = createRootRoute({
  component: () => (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
      <Outlet />
    </ThemeProvider>
  ),
})

const setupRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/setup",
  component: SetupPage,
})

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/login",
  component: LoginPage,
})

const dashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  beforeLoad: async () => {
    const state = auth.getState()
    if (!state.isAuthenticated) {
      const refreshed = await auth.tryRefresh()
      if (!refreshed) {
        throw redirect({ to: "/login" })
      }
    }
  },
  component: DashboardPage,
})

const routeTree = rootRoute.addChildren([setupRoute, loginRoute, dashboardRoute])

export const router = createRouter({ routeTree })

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}

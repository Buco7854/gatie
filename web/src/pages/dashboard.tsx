import { useAuth } from "@/hooks/use-auth"
import { ThemeToggle } from "@/components/theme-toggle"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { LogOut, Shield, User } from "lucide-react"
import { useNavigate } from "@tanstack/react-router"

export function DashboardPage() {
  const { username, role, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate({ to: "/login" })
  }

  return (
    <div className="min-h-screen">
      <header className="border-b border-border">
        <div className="max-w-5xl mx-auto flex items-center justify-between px-4 h-14">
          <h1 className="text-lg font-semibold">GATIE</h1>
          <div className="flex items-center gap-2">
            <ThemeToggle />
            <Button variant="ghost" size="sm" onClick={handleLogout}>
              <LogOut className="h-4 w-4 mr-2" />
              Logout
            </Button>
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto p-4 mt-8">
        <h2 className="text-2xl font-bold mb-6">Dashboard</h2>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Account</CardTitle>
              <User className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{username}</div>
              <p className="text-xs text-muted-foreground mt-1">Logged in</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Role</CardTitle>
              <Shield className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{role}</div>
              <p className="text-xs text-muted-foreground mt-1">Permission level</p>
            </CardContent>
          </Card>
        </div>

        <p className="text-sm text-muted-foreground mt-8">
          Gate management will be available in the next iteration.
        </p>
      </main>
    </div>
  )
}

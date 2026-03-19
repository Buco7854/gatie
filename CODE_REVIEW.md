# Code Review — GATIE

**Date:** 2026-03-19
**Scope:** Full codebase (Go backend, React frontend, infrastructure)
**Approach:** Manual review + automated analysis — only confirmed issues reported

---

## Summary

| Severity | Count |
|----------|-------|
| Critical | 1 |
| High | 6 |
| Medium | 5 |
| Low | 4 |

Overall the codebase is well-structured with clean layer separation, consistent patterns, and good security fundamentals (bcrypt, SHA-256 token hashing, parameterized SQL, HttpOnly cookies). The issues below are real and actionable.

---

## CRITICAL

### 1. Race condition in Setup endpoint (TOCTOU)

**File:** `server/internal/service/auth.go:59-83`

`Setup()` calls `CountMembers()` then `CreateMember()` without a transaction. Two concurrent POST requests to `/api/setup` can both read `count == 0` and both create an admin account.

**Impact:** An attacker can create a rogue admin account by racing the legitimate first setup.

**Fix:** Wrap in `s.tx.WithinTransaction` and use `SELECT count(*) FROM members FOR UPDATE` to serialize:

```go
func (s *AuthService) Setup(ctx context.Context, input SetupInput) (*AuthResult, error) {
    hash, err := auth.HashPassword(input.Password)
    if err != nil {
        return nil, err
    }

    var result *AuthResult
    err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
        count, err := s.repo.CountMembersForUpdate(txCtx) // new repo method with FOR UPDATE
        if err != nil {
            return fmt.Errorf("counting members: %w", err)
        }
        if count > 0 {
            return ErrSetupAlreadyCompleted
        }
        row, err := s.repo.CreateMember(txCtx, repository.CreateMemberParams{...})
        if err != nil {
            return fmt.Errorf("creating admin: %w", err)
        }
        result, err = s.generateTokens(txCtx, row)
        return err
    })
    return result, err
}
```

---

## HIGH

### 2. `MustParseURL` panics on invalid Valkey URL

**File:** `server/cmd/server/main.go:51`

```go
vkClient, err := valkey.NewClient(valkey.MustParseURL(opts.ValkeyURL))
```

`MustParseURL` panics if the URL is malformed. A typo in `SERVICE_VALKEY_URL` crashes the process with an unrecoverable panic instead of a clean error log.

**Fix:** Use `valkey.ParseURL` and handle the error:

```go
vkOpts, err := valkey.ParseURL(opts.ValkeyURL)
if err != nil {
    slog.Error("invalid Valkey URL", "error", err)
    os.Exit(1)
}
vkClient, err := valkey.NewClient(vkOpts)
```

### 3. FK violation errors not mapped — returns HTTP 500 instead of 422

**File:** `server/internal/repository/postgres/errors.go:21`

`mapError` only handles pgx error code `23505` (unique violation). FK violations (`23503`) — triggered when creating a gate membership with an invalid `role_id` or a member with an invalid `role` — propagate as generic 500 errors.

**Fix:** Add FK violation handling:

```go
if errors.As(err, &pgErr) {
    switch pgErr.Code {
    case "23505":
        return repository.ErrConflict
    case "23503":
        return repository.ErrForeignKeyViolation // new sentinel
    }
}
```

Then map it to 422 in handlers.

### 4. Refresh cookie duration hardcoded, diverges from token TTL

**File:** `server/internal/handler/auth.go:192`

```go
cookie := buildRefreshCookie(result.RefreshToken, 7*24*time.Hour)
```

The cookie Max-Age is hardcoded to 7 days while the actual token TTL comes from `JWTManager.refreshDuration`. If someone changes the refresh duration, the cookie and token lifetimes will silently diverge.

**Fix:** Pass the refresh duration from the auth service result or inject it into the handler.

### 5. `Field` component breaks label-input association (WCAG failure)

**File:** `web/src/components/ui/field.tsx:22`

The component creates `<label htmlFor={id}>` but never passes `id` to the child input. The label is not programmatically associated with the input. Screen readers won't announce the label, `aria-describedby` for errors/hints is never applied.

**Fix:** Clone the child element and inject props:

```tsx
{React.cloneElement(children, {
  id,
  'aria-describedby': error ? errorId : hint ? hintId : undefined,
  'aria-invalid': error ? true : undefined,
})}
```

### 6. Token reveal modal can be dismissed accidentally

**File:** `web/src/pages/gates.tsx` (token reveal modal)

The gate token is shown once in a standard modal that can be closed by clicking the backdrop or pressing Escape. If the user hasn't copied the token yet, it's lost permanently.

**Fix:** Either disable backdrop/escape closing when showing a token (`static` prop on HeadlessUI Dialog), or require the user to explicitly acknowledge they've copied the token before allowing close.

### 7. `setAuth` mutates state before calling `emitChange`

**File:** `web/src/lib/auth.ts:40-44`

```ts
state.accessToken = data.access_token
state.memberId = data.member_id
state.username = data.username
state.role = data.role
emitChange()
```

`state` is mutated in-place before `emitChange` creates a new reference. If any code calls `getSnapshot()` between the first mutation and `emitChange()`, it sees a partially updated state (e.g., new token but old role). This violates useSyncExternalStore's contract that the snapshot must be immutable between `subscribe` notifications.

**Fix:** Build a new object atomically:

```ts
export function setAuth(data: { ... }) {
  state = {
    accessToken: data.access_token,
    memberId: data.member_id,
    username: data.username,
    role: data.role,
  }
  emitChange()
}
```

Same for `clearAuth`.

---

## MEDIUM

### 8. `updated_at = now()` duplicated between SQL queries and DB trigger

**Files:** `server/internal/repository/postgres/member.go:122`, `gate.go:93`

Both `PatchMember` and `PatchGate` explicitly set `updated_at = now()` in the UPDATE query, but migration 007 already installs a `BEFORE UPDATE` trigger doing the same thing. It works today but creates confusion about the source of truth.

**Fix:** Remove `updated_at = now()` from the SQL queries — let the trigger handle it.

### 9. `goose` migrations use a separate `database/sql` connection

**File:** `server/internal/database/migrate.go`

`RunMigrations` opens a new `database/sql` connection via `goose.OpenDBWithDriver` while `main.go` already has a `pgxpool`. This means two connection pools exist during startup, and the goose connection may not inherit all pgxpool config (timeouts, pool size, etc.).

**Fix:** Use `goose.SetDialect("postgres")` + `goose.Up(db, dir)` with a `*sql.DB` obtained from `pgxpool.Config().ConnString()`, or use pgx's stdlib adapter.

### 10. `ListGates` / `ListMembers` do COUNT + SELECT without transaction

**Files:** `server/internal/service/gate.go:59-79`, `member.go:59-79`

Both list methods call `CountX()` then `ListX()` in two separate queries. Between the two calls, rows can be inserted or deleted, leading to inconsistent `total` vs `items` (e.g., `total: 5` but 6 items returned). This is a minor UI inconsistency.

**Fix:** Run both queries in a read-only transaction, or use a single `SELECT *, count(*) OVER() FROM ... LIMIT ... OFFSET ...` window query.

### 11. No logout cookie clearing

**File:** `server/internal/handler/auth.go:180-187`

The `logout` handler revokes the refresh token in the DB but doesn't send a `Set-Cookie` header to clear the `refresh_token` cookie in the browser. The browser keeps sending the (now-revoked) cookie on subsequent requests until it naturally expires.

**Fix:** Return a `Set-Cookie` with `Max-Age=0` to delete the cookie:

```go
func (h *AuthHandler) logout(ctx context.Context, input *LogoutInput) (*LogoutOutput, error) {
    // ... revoke token ...
    return &LogoutOutput{
        SetCookie: buildRefreshCookie("", 0), // clears cookie
    }, nil
}
```

### 12. `useAuth` doesn't react to auth state changes after mount

**File:** `web/src/hooks/use-auth.ts:11-21`

The `useEffect` that triggers `silentRefresh` has an empty dependency array `[]`. If the user logs out and the auth state is cleared, `status` remains stuck at its last value because the effect only runs on mount. The `useSyncExternalStore` subscription updates `auth` but not `status`.

**Fix:** Derive `status` from the auth snapshot instead of maintaining separate state:

```ts
export function useAuth() {
  const auth = useSyncExternalStore(subscribe, getSnapshot)
  const [status, setStatus] = useState<AuthStatus>(() =>
    auth.accessToken ? 'authenticated' : 'loading'
  )
  // ...rest stays the same, but also sync status when auth changes
}
```

Or simpler: compute status directly from auth state after the initial refresh attempt.

---

## LOW

### 13. `gates.name` missing UNIQUE constraint

**File:** `server/internal/database/migrations/002_create_gates.sql`

The `name` column has no UNIQUE constraint in the schema, but the service layer expects `ErrConflict` on duplicate names. This means duplicate gate names will succeed at the DB level — the unique check in the service layer has no enforcement.

**Fix:** Add `UNIQUE` to the `name` column (or add it in a new migration):

```sql
ALTER TABLE gates ADD CONSTRAINT uq_gates_name UNIQUE (name);
```

### 14. `members.created_at` pagination index missing

**File:** `server/internal/database/migrations/008_add_missing_indexes_and_constraints.sql`

There's an index on `gates.created_at` for pagination but no equivalent on `members.created_at`, even though `ListMembers` orders by `created_at ASC LIMIT/OFFSET`.

**Fix:** Add: `CREATE INDEX idx_members_created_at ON members(created_at ASC);`

### 15. Pagination shows when there's only one page

**File:** `web/src/components/ui/pagination.tsx`

The pagination component renders even when `totalPages <= 1`. Shows "Page 1 of 1" with both buttons disabled, which is unnecessary UI clutter.

**Fix:** Return `null` when `totalPages <= 1`.

### 16. `ActionButtons` always visible on desktop (design intent says hover-only)

**File:** `web/src/pages/members.tsx:434-438`

Per CLAUDE.md: "Boutons d'action: visibles au survol uniquement sur desktop." But the action buttons in the table rows are always visible, not hidden behind `group-hover`. Same issue in `gates.tsx`.

**Fix:** Add `opacity-0 group-hover:opacity-100` or `invisible group-hover:visible` on desktop.

---

## Architecture Notes (not bugs)

These are not issues to fix now, but design observations for future awareness:

- **Large page components** (`members.tsx` 494 lines, `gates.tsx` 428 lines, `roles.tsx` 643 lines): each contains schemas, sub-components, forms, modals, and the main page. Consider extracting forms into separate files as complexity grows.
- **No password change endpoint**: members are created with passwords but there's no way to change them. Will need to be added.
- **No refresh token invalidation on role change**: when an admin changes a member's role, existing JWTs retain the old role until they expire (15 min). The refresh flow re-reads the role from DB, which is correct, but the 15-minute window could be a concern.
- **No CSRF protection**: the app relies on `SameSite=Strict` cookies, which is adequate for same-origin requests but won't protect against subdomain attacks if deployed on a shared domain.

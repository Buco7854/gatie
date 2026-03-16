# GATIE — Conventions & Protocole de développement

## Identité du projet

GATIE est une application de contrôle de portails/barrières IoT, conçue pour le self-hosting.
Une installation = une instance. Pas de multi-tenant applicatif.
L'expression de besoin complète est dans `EXPRESSION_OF_NEED.md`.

---

## Stack technique

### Backend — Go
| Composant | Choix | Package |
|-----------|-------|---------|
| Framework HTTP | **Huma v2** | `github.com/danielgtaylor/huma/v2` |
| Router | **Chi** | `github.com/go-chi/chi/v5` (via `humachi`) |
| DB access | **pgx** | SQL pur via `github.com/jackc/pgx/v5` |
| Migrations | **goose** | `github.com/pressly/goose/v3` |
| Base de données | **PostgreSQL 18** | `github.com/jackc/pgx/v5` |
| Cache | **Valkey** (Redis-compatible) | `github.com/valkey-io/valkey-go` |
| MQTT client | **paho.mqtt.golang** | `github.com/eclipse/paho.mqtt.golang` |
| Auth JWT | **golang-jwt** | `github.com/golang-jwt/jwt/v5` |
| Auth OIDC | **go-oidc** | `github.com/coreos/go-oidc/v3` |
| Validation | **Huma built-in** (struct tags + resolvers) | — |
| Config | **humacli** + env vars | `SERVICE_` prefix |

### Frontend — React + TypeScript
| Composant | Choix |
|-----------|-------|
| Build | Vite 7 |
| Framework | React 19 |
| Routing | TanStack Router |
| Data fetching | TanStack Query |
| UI | Tailwind CSS v4 + HeadlessUI v2 |
| i18n | i18next |
| Thème | next-themes |
| Formulaires | React Hook Form + Zod |

### Frontend — Règles de design
- **Mobile-first** : tout composant est conçu pour mobile en priorité, adapté au desktop via breakpoints Tailwind (`sm:`, `md:`, `lg:`).
- **Pas de `<select>` natif** : utiliser `ListboxSelect` (`@/components/ui/listbox-select`) pour tout champ de sélection dans les formulaires — cohérence visuelle avec le reste de l'UI HeadlessUI.
- **Tableaux** : sur mobile (`< sm`), les tableaux sont remplacés par une vue en cartes/liste. Sur desktop, tableau classique.
- **Boutons d'action** (édition/suppression dans une liste) : toujours visibles sur mobile (touch ≠ hover), visibles au survol uniquement sur desktop.
- **Breakpoints standards** : `sm = 640px`, `md = 768px`, `lg = 1024px`.

### Infrastructure (Docker Compose)
| Service | Image |
|---------|-------|
| PostgreSQL | `postgres:18` |
| Valkey | `valkey/valkey:8` |
| Mosquitto | `eclipse-mosquitto:2` |
| Caddy | `caddy:2` (reverse proxy, TLS auto) |

### Structure du projet
```
gatie/
├── server/
│   ├── cmd/server/          # Point d'entrée (main.go, humacli)
│   ├── internal/
│   │   ├── database/        # Migrations (embed + goose)
│   │   │   └── migrations/  # Fichiers SQL de migration
│   │   ├── handler/         # Routes HTTP (Huma operations)
│   │   ├── repository/      # Erreurs domaine (ErrNotFound, ErrConflict)
│   │   │   └── postgres/    # Accès DB (SQL pur via pgx)
│   │   ├── service/         # Logique métier
│   │   ├── mqtt/            # Client MQTT + handlers
│   │   ├── middleware/      # Auth, rate-limit, etc.
│   │   └── model/           # Structs partagés (domain)
│   ├── go.mod
│   └── go.sum
├── web/                     # React SPA
├── mosquitto/               # Config Mosquitto
├── docker-compose.yml
├── CLAUDE.md
└── EXPRESSION_OF_NEED.md
```

---

## MQTT — Deux modes d'authentification

Le gate token contient les infos de la gate (ID, nom, etc.) et est vérifié en BDD à chaque usage.

### Mode 1 : Mosquitto Dynamic Security
- Le broker gère l'auth via son plugin Dynamic Security
- GATIE provisionne les clients/ACLs dans Mosquitto à la création/modification d'une gate
- Le broker refuse les connexions/publications non autorisées

### Mode 2 : Broker-agnostic
- Le broker accepte toutes les connexions (ou auth basique)
- GATIE vérifie à chaque message MQTT reçu que :
  1. Le token correspond exactement à celui en BDD
  2. Les infos du token (gate ID, etc.) correspondent au topic actuel
- Plus portable, fonctionne avec n'importe quel broker

---

## Référence Huma v2 — Patterns clés

### Opérations (routes)
```go
huma.Register(api, huma.Operation{
    OperationID: "get-gate",
    Method:      http.MethodGet,
    Path:        "/gates/{gate-id}",
    Summary:     "Get a gate",
    Tags:        []string{"Gates"},
    Security:    []map[string][]string{{"bearer": {}}},
    Middlewares:  huma.Middlewares{AuthMiddleware},
}, func(ctx context.Context, input *GetGateInput) (*GetGateOutput, error) {
    // ...
})
```
Raccourcis : `huma.Get`, `huma.Post`, `huma.Put`, `huma.Patch`, `huma.Delete`.

### Input (requête)
```go
type GetGateInput struct {
    ID   string `path:"gate-id"`                           // toujours requis
    Auth string `header:"Authorization" required:"true"`   // header
}
type ListGatesInput struct {
    Page    int `query:"page" minimum:"1" default:"1"`     // query, optionnel
    PerPage int `query:"per_page" minimum:"1" maximum:"100" default:"20"`
}
type CreateGateInput struct {
    Body struct {
        Name string `json:"name" minLength:"1" maxLength:"100"` // body, requis par défaut
    }
}
```

### Validation par struct tags
| Tag | Usage |
|-----|-------|
| `required:"true"` / `required:"false"` | Forcer requis/optionnel |
| `minLength` / `maxLength` | Longueur string |
| `pattern` | Regex |
| `format` | `email`, `uuid`, `uri`, `date-time`, `ipv4` |
| `minimum` / `maximum` | Bornes numériques |
| `enum` | Valeurs autorisées : `enum:"open,closed,offline"` |
| `default` | Valeur par défaut serveur |
| `doc` | Description du champ |
| `nullable:"true"` | Accepte null |

### Nullable vs Optional
- Body fields : **requis par défaut**. `omitempty` ou `omitzero` → optionnel.
- Query/header/cookie : **optionnels par défaut**. `required:"true"` → requis.
- Path : **toujours requis**.
- `*string` = nullable + requis. `*string` + `omitempty` = nullable + optionnel.

### Resolvers (validation custom)
```go
func (m *MyInput) Resolve(ctx huma.Context) []error {
    if m.Body.End.Before(m.Body.Start) {
        return []error{&huma.ErrorDetail{
            Location: "body.end",
            Message:  "end must be after start",
            Value:    m.Body.End,
        }}
    }
    return nil
}
var _ huma.Resolver = (*MyInput)(nil)
```

### Output (réponse)
```go
type GetGateOutput struct {
    Body GateResponse                    // corps de réponse
}
type CreateGateOutput struct {
    Header string `header:"Location"`   // header de réponse
    Body   GateResponse
}
```
Status par défaut : 200 avec body, 204 sans. Override via `DefaultStatus` dans Operation.

### Erreurs (RFC 9457)
```go
return nil, huma.Error404NotFound("gate not found")
return nil, huma.Error422UnprocessableEntity("invalid", &huma.ErrorDetail{
    Location: "body.name", Message: "already exists", Value: input.Body.Name,
})
```

### Middleware Huma
```go
func AuthMiddleware(ctx huma.Context, next func(huma.Context)) {
    token := ctx.Header("Authorization")
    if token == "" {
        huma.WriteErr(api, ctx, 401, "unauthorized")
        return
    }
    ctx = huma.WithValue(ctx, "user", parsedUser)
    next(ctx)
}
api.UseMiddleware(AuthMiddleware)
```

### SSE (Server-Sent Events)
```go
import "github.com/danielgtaylor/huma/v2/sse"

sse.Register(api, huma.Operation{
    OperationID: "gate-events",
    Method:      http.MethodGet,
    Path:        "/gates/{gate-id}/events",
}, map[string]any{
    "status": GateStatusEvent{},
    "metadata": GateMetadataEvent{},
}, func(ctx context.Context, input *GateEventsInput, send sse.Sender) {
    send.Data(GateStatusEvent{Status: "open"})
})
```

### Groupes
```go
grp := huma.NewGroup(api, "/api/v1")
grp.UseMiddleware(AuthMiddleware)
huma.Get(grp, "/gates", listGatesHandler)
```

### Tests
```go
_, api := humatest.New(t)
addRoutes(api)
resp := api.Get("/gates")
assert.Equal(t, 200, resp.Code)
resp = api.Post("/gates", map[string]any{"name": "Main Gate"})
resp = api.Get("/gates/1", "Authorization: Bearer token")
```

### CLI (humacli)
```go
type Options struct {
    Host string `doc:"Hostname" default:"0.0.0.0"`
    Port int    `doc:"Port" short:"p" default:"8888"`
}
cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
    router := chi.NewMux()
    api := humachi.New(router, huma.DefaultConfig("GATIE", "1.0.0"))
    addRoutes(api)
    server := &http.Server{Addr: fmt.Sprintf(":%d", opts.Port), Handler: router}
    hooks.OnStart(func() { server.ListenAndServe() })
    hooks.OnStop(func() { server.Shutdown(context.Background()) })
})
cli.Run()
```

---

## Protocole de développement (STRICT)

### Règle 1 — Plan avant code
Avant chaque feature, un plan d'implémentation est proposé en français :
- **Quoi** : ce qu'on construit
- **Pourquoi** : lien avec l'expression de besoin (section §)
- **Comment** : approche technique, fichiers, choix
- **Ce que tu dois comprendre** : concepts résumés simplement
**AUCUNE ligne de code n'est écrite avant validation explicite du plan par l'utilisateur.**

### Règle 2 — Petits blocs expliqués
Chaque commit/bloc est cohérent et petit. Après chaque bloc :
- Explication de ce qui a été fait et pourquoi
- L'utilisateur peut poser des questions
- On avance seulement quand c'est clair

### Règle 3 — Checkpoints compréhension
Après chaque feature complétée, un résumé est proposé avec des questions simples pour vérifier que l'utilisateur comprend le code. Si blocage, on revient dessus.

### Règle 4 — L'utilisateur choisit le rythme
À tout moment :
- "Explique-moi ce fichier" → décomposition ligne par ligne
- "Pourquoi ce choix ?" → justification
- "Propose une alternative" → comparaison
- "Ralentis" → plus de détails, blocs plus petits

### Règle 5 — Avancement persistant
Le fichier `PROGRESS.md` à la racine est mis à jour après chaque session avec :
- Ce qui a été fait
- Ce qui reste à faire
- Le prochain bloc prévu
- Les décisions techniques prises

---

## Conventions de code

### Go (backend)
- Noms de packages en minuscules, un mot
- Erreurs wrappées avec `fmt.Errorf("context: %w", err)`
- Pas de `panic` en dehors de `main`
- Tests dans le même package avec `_test.go`
- SQL dans `migrations/` (goose)
- **Commentaires minimaux** : ne commenter que ce qui est réellement difficile à comprendre ou non-intuitif. Pas de commentaires évidents ou redondants.
- **UUIDs** : utiliser UUID v7 (support natif PostgreSQL 18 via `uuidv7()`) pour l'ordre chronologique

### TypeScript (frontend)
- Composants en PascalCase, fichiers en kebab-case
- Hooks custom préfixés `use`
- Pas de `any` sauf nécessité absolue
- Imports absolus via alias `@/`

### Git
- Commits en anglais, conventionnels : `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`
- Un commit = un changement logique
- Branches : `feat/xxx`, `fix/xxx`

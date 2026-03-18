# GATIE — Avancement

## Décision architecture — PBAC dynamique (2026-03-18)

### Modèle d'accès retenu : Permission-Based Access Control (PBAC)

**Contexte :** un père déploie GATIE, veut que sa femme gère les PINs/plannings, et que son enfant puisse juste ouvrir. Avec seulement ADMIN/MEMBER, la femme devrait être ADMIN — trop large. D'où les rôles intermédiaires.

Les rôles et permissions sont stockés dynamiquement en BDD (pas de booléens hardcodés). Le scope (global vs local) est défini implicitement par l'endroit où le rôle est assigné :
- **Scope global** : rôle assigné dans `members.role_id`
- **Scope local** : rôle assigné dans `gate_memberships.role_id`

Les rôles/permissions sont **pré-remplis au démarrage** (upsert), non éditables par l'utilisateur. L'utilisateur peut uniquement **lister les rôles** (avec descriptions et permissions) pour les assigner.

### Schéma BDD

```sql
-- Rôles disponibles (pré-remplis, lecture seule)
CREATE TABLE roles (
    id          VARCHAR PRIMARY KEY,  -- 'ADMIN', 'MANAGER', 'MEMBER', 'VIEWER'
    description TEXT NOT NULL
);

-- Permissions disponibles (pré-remplis, lecture seule)
CREATE TABLE permissions (
    id          VARCHAR PRIMARY KEY,  -- 'gate:open', 'gate:manage_members', '*'
    description TEXT NOT NULL
);

-- Association rôle → permissions (pré-rempli, lecture seule)
CREATE TABLE role_permissions (
    role_id       VARCHAR REFERENCES roles(id) ON DELETE CASCADE,
    permission_id VARCHAR REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Membres : rôle global via FK
ALTER TABLE members ... role_id VARCHAR REFERENCES roles(id);

-- Accès par portail : rôle local
CREATE TABLE gate_memberships (
    gate_id   UUID REFERENCES gates(id) ON DELETE CASCADE,
    member_id UUID REFERENCES members(id) ON DELETE CASCADE,
    role_id   VARCHAR REFERENCES roles(id) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (gate_id, member_id)
);
```

L'ancienne table `permissions` (booléens) est supprimée.

### Rôles et permissions par défaut (seedés)

| Rôle | Permissions |
|------|------------|
| ADMIN | `*` (bypass toutes les vérifications) |
| MANAGER | `gate:open`, `gate:close`, `gate:view_status`, `gate:configure`, `gate:manage_members` |
| MEMBER | `gate:open`, `gate:close`, `gate:view_status` |
| VIEWER | `gate:view_status` |

### Bootstrapping (upsert au démarrage)

Le backend lance un `EnsureDefaultRolesAndPermissions(ctx)` au démarrage qui fait des upserts. Si une future version ajoute une permission, le redémarrage la crée automatiquement. La migration seed l'état initial, le bootstrap maintient la cohérence.

### Logique d'autorisation — `can(member, membership, permission)`

```
1. Vérifier si le rôle global (member.role_id) possède la permission demandée (ou '*')
   → Si oui : true
2. Si non, vérifier si l'utilisateur a une entrée membership pour cette gate
   → Si non : false
3. Si oui, vérifier si le rôle local (membership.role_id) possède la permission demandée
   → Retourner le résultat
```

> Note : les ADMIN n'ont pas besoin d'entrée dans `gate_memberships` — la permission `*` dans leur rôle global suffit.

### Endpoints prévus

- `GET /roles` — liste les rôles avec descriptions et permissions (lecture seule)
- `GET /gates/{id}/members` — liste les membres d'une gate avec leur rôle local
- `POST /gates/{id}/members` — ajouter un membre à une gate
- `PATCH /gates/{id}/members/{member-id}` — changer le rôle d'un membre sur une gate
- `DELETE /gates/{id}/members/{member-id}` — retirer un membre d'une gate

---

## Session 6 — 2026-03-16

### Ce qui a été fait
- **CRUD portails — backend** :
  - `service/gate.go` : `GateService` (List, Get, Create, Update, Delete, RegenerateToken)
  - `handler/gate.go` : 6 routes Huma (`GET/POST /api/gates`, `GET/PATCH/DELETE /api/gates/{id}`, `POST /api/gates/{id}/token`)
  - `main.go` : wiring `GateService` + `GateHandler`
- **CRUD portails — frontend** :
  - `lib/types.ts` : types `Gate`, `GateWithToken`, `GatesPage`
  - `pages/gates.tsx` : liste paginée + modale create/edit + suppression + panel "token reveal" + confirmation regénération
  - `router.tsx` : route `/gates`
  - `app-header.tsx` : lien nav "Portails" (HomeModernIcon)
  - `components/ui/field.tsx` : prop `hint` ajoutée
  - i18n FR + EN mis à jour (portails, nav, actions)

### Décisions techniques prises
- Token de portail : 32 octets aléatoires (`crypto/rand`) → hex (64 chars) → hashé SHA-256 en BDD
- Le token brut n'est retourné qu'une seule fois (création ou regénération) avec une alerte "sauvegardez-le"
- Regénération : confirmation modale + invalidation immédiate de l'ancien token
- Copie du token : bouton clipboard avec feedback visuel (icône check verte 2s)

### Prochain bloc prévu
- **Permissions (gate_memberships)** :
  - Nouvelle migration remplaçant `permissions` par `gate_memberships` avec colonne `role`
  - Queries sqlc mises à jour
  - `service/permission.go` : logique `can()` centralisée + CRUD des accès par portail
  - `handler/permission.go` : routes pour gérer les accès d'un portail
  - Frontend : onglet/section "Accès" dans la page portail

### Ce qui reste à faire (grandes étapes)
1. ~~Init projet + Docker Compose + health check~~
2. ~~Schéma BDD + migrations (members, gates, permissions, schedules)~~
3. ~~Auth (inscription initiale, login/JWT, refresh tokens)~~
4. ~~Frontend auth (setup, login, dashboard protégé)~~
5. ~~CRUD membres (backend + frontend)~~
6. ~~CRUD portails + gate tokens~~
7. **Permissions (gate_memberships, scoped RBAC 3 niveaux)** ← prochain
8. Plannings horaires
9. Client MQTT + modes d'auth
10. Actions sur portails (open/close) via MQTT/HTTP
11. Statut temps réel (SSE)
12. Codes PIN / mots de passe d'accès
13. API tokens
14. SSO (OIDC)
15. Audit log

---

## Session 5 — 2026-03-14

### Ce qui a été fait
- **Frontend refactorisé** : HeadlessUI v2 + Tailwind CSS v4 (CSS-first), Vite 7, sans shadcn/ui
- **Theme toggle** : Listbox HeadlessUI v2 (Light/Dark/Système) avec icônes Heroicons
- **AppHeader** : composant partagé avec nav dynamique (liens admin only), logout, theme toggle
- **CRUD membres — backend** :
  - `middleware/admin.go` : `RequireAdmin` middleware
  - `middleware/auth.go` : ajout `GetClaimsFromContext(context.Context)` pour les handlers
  - `repository/members_role.sql.go` : `CountMembersByRole` (équivalent sqlc)
  - `service/member.go` : `MemberService` (List, Get, Create, Update, Delete) avec gardes métier
  - `handler/member.go` : 5 routes Huma (`GET/POST /members`, `GET/PATCH/DELETE /members/{id}`)
  - `main.go` : wiring `MemberService` + `MemberHandler` + `authMW`
- **CRUD membres — frontend** :
  - `lib/types.ts` : types partagés `Member`, `MembersPage`
  - `components/ui/select.tsx` : composant Select stylé
  - `pages/members.tsx` : liste paginée + modale create/edit + suppression inline avec confirmation
  - `router.tsx` : route `/members`
  - i18n FR + EN mis à jour (nav, rôles, membres, actions, pagination)

### Décisions techniques prises
- Guard "dernier admin" : `CountMembersByRole` pour compter les admins avant suppression
- Guard "auto-suppression" : le caller ne peut pas supprimer son propre compte
- Unicité username : détection erreur pgconn code 23505 → `ErrUsernameExists`
- Icônes d'action (edit/delete) visibles au survol uniquement (`group-hover:opacity-100`)
- Suppression : confirmation inline dans la ligne (pas de modal) — UX propre

### Prochain bloc prévu
- **CRUD portails** : backend (gates + gate_actions) + frontend (page portails)

### Ce qui reste à faire (grandes étapes)
1. ~~Init projet + Docker Compose + health check~~
2. ~~Schéma BDD + migrations (members, gates, permissions, schedules)~~
3. ~~Auth (inscription initiale, login/JWT, refresh tokens)~~
4. ~~Frontend auth (setup, login, dashboard protégé)~~
5. ~~CRUD membres (backend + frontend)~~
6. CRUD portails + gate tokens
7. Permissions + plannings horaires
8. Client MQTT + modes d'auth
9. Actions sur portails (open/close) via MQTT/HTTP
10. Statut temps réel (SSE)
11. Codes PIN / mots de passe d'accès
12. Domaines personnalisés
13. API tokens
14. SSO (OIDC)
15. Audit log

---

## Session 4 — 2026-03-14

### Ce qui a été fait
- **Frontend scaffolding** : Vite + React 19 + TypeScript
- **Tailwind CSS v4** + **shadcn/ui** (composants Button, Input, Label, Card)
- **next-themes** : thème dark/light/system — dark sobre gris/noir, pas de bleu
- **TanStack Router** : routes `/setup`, `/login`, `/` (dashboard protégé)
- **API client** (`lib/api.ts`) : setup, login, refresh, logout, health
- **Auth store** (`lib/auth.ts`) : gestion token en mémoire, refresh automatique, `useSyncExternalStore`
- **3 pages** :
  - `/setup` — création du premier admin (formulaire avec validation Zod)
  - `/login` — connexion (formulaire avec validation Zod)
  - `/` — dashboard protégé (affiche username, rôle, bouton logout)
- **Theme toggle** : composant pour switcher light → dark → system
- **Vite proxy** : `/setup`, `/auth`, `/health` proxiés vers `localhost:8888`
- Build production OK (npx vite build)
- TypeScript check OK (npx tsc --noEmit)

### Décisions techniques prises
- Access token stocké en mémoire JS (pas localStorage) pour sécurité
- Refresh token stocké en localStorage + envoyé dans le body (le cookie HttpOnly est un bonus côté serveur)
- Route protégée avec `beforeLoad` qui tente un refresh si pas authentifié
- Palette dark sobre : `#0a0a0a` background, `#141414` cards, `#262626` borders — aucun bleu

### Comment tester
1. `cd server && go run ./cmd/server` (backend sur :8888)
2. `cd web && npm run dev` (frontend sur :5173, proxy vers :8888)
3. Ouvrir http://localhost:5173 → redirigé vers `/login`
4. Cliquer "Set up GATIE" → créer l'admin → redirigé vers dashboard
5. Tester logout → login → thème toggle

---

## Session 3 — 2026-03-14

### Ce qui a été fait
- Migration `006_create_refresh_tokens.sql` : table refresh_tokens (token hashé, expiration, lien member)
- Queries sqlc pour refresh_tokens (CRUD + suppression expirés)
- Package `auth` : génération/validation JWT (HS256), hashing bcrypt pour mots de passe, hashing SHA-256 pour refresh tokens
- Service `auth` : logique métier pour setup initial, login, refresh (avec rotation), logout
- Middleware `auth` : extraction Bearer token, validation JWT, injection des claims dans le contexte
- Handlers HTTP : `POST /setup`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`
- Wiring complet dans main.go avec option `JWTSecret` (auto-généré si absent)
- 17 tests unitaires (JWT, bcrypt, middleware, handlers) — tous passent
- Downgrade des dépendances Go pour compatibilité avec Go 1.24.7

### Décisions techniques prises
- Access token JWT (15min) en mémoire côté client, refresh token (7j) en cookie HttpOnly/Secure/SameSite=Strict
- Refresh token rotation : à chaque refresh, l'ancien token est supprimé et un nouveau est créé
- Refresh token hashé en SHA-256 en BDD (pas bcrypt car on n'a pas besoin de résistance au brute-force, le token est aléatoire 256 bits)
- JWT secret auto-généré au démarrage si non configuré (avec warning log)
- Setup initial vérifie `COUNT(members) = 0` avant de créer le premier admin
- Login retourne un message d'erreur générique "invalid credentials" (pas de fuite d'info username/password)

---

## Session 2 — 2026-03-14

### Ce qui a été fait
- Migration `002_create_gates.sql` : table gates (nom, token hashé, TTL statut)
- Migration `003_create_gate_actions.sql` : actions configurables par portail (MQTT/HTTP/NONE)
- Migration `004_create_permissions.sql` : permissions granulaires (open, close, view_status, manage) par membre/portail
- Migration `005_create_schedules.sql` : plannings (JSONB pour expressions logiques combinables) + table de liaison member_gate_schedules
- Queries sqlc pour toutes les nouvelles tables (CRUD + requêtes spécifiques)
- Génération du code Go via sqlc

### Décisions techniques prises
- gate_actions.config en JSONB : flexibilité pour stocker la config MQTT ou HTTP selon le transport
- schedules.expression en JSONB : arbre logique (AND/OR/NOT) pour combiner les règles temporelles
- Contrainte CHECK : un planning PERSONAL doit avoir un owner_id
- gate_token_hash : le jeton brut n'est jamais stocké (comme un mot de passe)

---

## Session 1 — 2026-03-14

### Ce qui a été fait
- Expression de besoin analysée
- Stack technique choisi et validé (Go/Huma + React + PostgreSQL/Valkey/Mosquitto)
- Deux modes MQTT définis (Dynamic Security + broker-agnostic)
- CLAUDE.md créé avec conventions, référence Huma, et protocole de développement
- Protocole anti-vibe-coding mis en place

### Décisions techniques prises
- Backend en Go (pas TypeScript) avec Huma v2 + Chi router
- GATIE est client MQTT, pas broker (broker = infra via Docker)
- Gate token contient les infos gate, vérifié en BDD, deux modes d'auth MQTT
- sqlc pour l'accès DB (SQL pur, type-safe)
- humacli pour la CLI et config par env vars

### Prochain bloc prévu
- **CRUD membres** : endpoints backend + pages frontend pour gérer les membres

### Ce qui reste à faire (grandes étapes)
1. ~~Init projet + Docker Compose + health check~~
2. ~~Schéma BDD + migrations (members, gates, permissions, schedules)~~
3. ~~Auth (inscription initiale, login/JWT, refresh tokens)~~
4. ~~Frontend auth (setup, login, dashboard protégé)~~
5. CRUD membres (backend + frontend)
5. CRUD portails + gate tokens
6. Permissions + plannings horaires
7. Client MQTT + modes d'auth
8. Actions sur portails (open/close) via MQTT/HTTP
9. Statut temps réel (SSE)
10. Codes PIN / mots de passe d'accès
11. Domaines personnalisés
12. API tokens
13. SSO (OIDC)
14. Frontend React
15. Audit log

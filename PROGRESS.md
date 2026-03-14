# GATIE — Avancement

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
- **Auth** : inscription initiale, login/JWT, refresh tokens

### Ce qui reste à faire (grandes étapes)
1. ~~Init projet + Docker Compose + health check~~
2. ~~Schéma BDD + migrations (members, gates, permissions, schedules)~~
3. Auth (inscription initiale, login/JWT, refresh tokens)
4. CRUD membres
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

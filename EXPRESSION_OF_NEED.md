# GATIE — Expression de Besoin Fonctionnel

## 1. Vision du produit

GATIE est une application de contrôle de portails/barrières IoT, conçue pour le self-hosting (homelab). Elle permet de gérer à distance l'ouverture, la fermeture et la supervision d'accès physiques (portails, barrières, portes), avec un contrôle granulaire des permissions, des plages horaires, et un accès invité par code PIN.

**Une installation = une instance.** Pas de multi-tenant, pas de workspaces. L'application est simple à déployer et à administrer.

---

## 2. Modèle économique

### 2.1 Self-host gratuit

- L'application est **gratuite et open-source** en self-hosting
- Toutes les fonctionnalités sont disponibles sans restriction
- L'utilisateur déploie et gère son instance sur son propre serveur
- La communauté open-source assure crédibilité et adoption

### 2.2 Cloud managé (payant)

- Une offre **GATIE Cloud** hébergée est proposée en abonnement
- Chaque client cloud obtient une **instance dédiée et isolée** (instance-per-tenant)
- Le code applicatif est **strictement identique** entre self-host et cloud
- Le cloud vend le **confort**, pas des fonctionnalités supplémentaires :
  - Déploiement instantané, zéro configuration
  - Sauvegardes automatiques
  - Mises à jour gérées
  - Certificats TLS automatiques
  - Support technique
- Le multi-tenant n'existe que dans la couche d'orchestration cloud, **jamais dans l'application**

### 2.3 Portail de gestion cloud (service séparé)

Ce service distinct (hors périmètre de l'application GATIE elle-même) gère :

- Inscription et authentification des clients cloud
- Facturation et abonnements
- Provisioning automatique d'instances
- Gestion des domaines personnalisés côté cloud
- Monitoring et santé des instances

---

## 3. Concepts clés

### 3.1 Membres et rôles

L'application fonctionne avec un unique niveau d'identité : les **membres**. Chaque membre possède un rôle :

- **ADMIN** : accès total à la gestion (portails, membres, plannings, paramètres)
- **MEMBER** : accès restreint aux portails sur lesquels des permissions lui ont été attribuées

### 3.2 Portails (Gates)

Un portail représente un équipement physique (portail, barrière, porte connectée) contrôlable à distance. Chaque portail possède :

- Un **nom**
- Un **statut** en temps réel (en ligne, ouvert, fermé, hors ligne, non réactif, indisponible…)
- Des **actions configurables** : ouverture, fermeture, remontée de statut
- Des **métadonnées en direct** (température, niveau batterie, signal, etc.)
- Un **jeton d'authentification** pour l'appareil physique

### 3.3 Permissions

Les droits d'accès sont attribués **par membre et par portail** avec les permissions suivantes :

- **Déclencher l'ouverture** d'un portail
- **Déclencher la fermeture** d'un portail
- **Consulter le statut** et les données en direct d'un portail
- **Gérer** un portail (configuration, codes PIN, domaines, permissions des autres membres)

### 3.4 Plannings horaires (Schedules)

Les plannings permettent de restreindre temporellement un accès. Ils sont composés d'expressions combinables :

- **Plage horaire** : jours de la semaine + heure de début/fin (ex. : lundi–vendredi 08h–18h)
- **Plage de jours de la semaine** : jours consécutifs avec wrap-around (ex. : samedi–dimanche)
- **Plage de dates** : dates calendaires (ex. : 01/06/2026–31/08/2026)
- **Plage de jours du mois** : jours récurrents (ex. : du 1er au 15 de chaque mois)
- **Plage de mois** : mois récurrents (ex. : juin–septembre)

Les expressions sont combinables avec des opérateurs logiques **ET**, **OU**, **NON** pour créer des règles complexes (ex. : "lundi–vendredi 08h–18h ET pas en août").

Les plannings peuvent être attachés à :
- Un **membre sur un portail** : le membre ne peut agir que pendant les plages autorisées
- Un **code PIN** : le code n'est valide que pendant les plages autorisées

Il existe deux types de plannings :
- **Plannings d'administration** : créés par les admins, utilisables sur n'importe quel membre ou code
- **Plannings personnels** : créés par un membre pour ses propres besoins

---

## 4. Fonctionnalités détaillées

### 4.1 Setup initial

Au premier lancement, l'application détecte qu'aucun membre n'existe et propose un écran de configuration initiale :

- Création du premier compte administrateur (nom d'utilisateur + mot de passe)
- Connexion automatique après création
- Cette étape n'est disponible qu'une seule fois

### 4.2 Authentification

#### 4.2.1 Connexion par mot de passe

- Un membre se connecte avec son **nom d'utilisateur** et son **mot de passe**
- Le système délivre un jeton d'accès (courte durée) et un jeton de rafraîchissement (longue durée)
- Le rafraîchissement est automatique et transparent pour l'utilisateur
- Déconnexion : révocation du jeton de rafraîchissement

#### 4.2.2 Connexion SSO (Single Sign-On)

- Les fournisseurs SSO (OIDC) peuvent être configurés **via l'interface d'administration** ou via **variables d'environnement**
- Si un fournisseur est configuré via variable d'environnement, sa configuration **prend la priorité** sur celle de la base de données pour l'exécution courante, et les paramètres correspondants sont **verrouillés** dans l'interface (non modifiables via le front)
- L'écran de connexion affiche les boutons SSO disponibles
- Flux : redirection vers le fournisseur → callback → échange de code → connexion
- **Auto-provisioning** : si activé, un membre est automatiquement créé lors de sa première connexion SSO
- Mapping de rôle : le rôle du membre peut être déduit d'un claim du token SSO

#### 4.2.3 Jetons API (API Tokens)

Chaque membre peut créer des **jetons API** pour un accès programmatique :

- Libellé descriptif
- Date d'expiration optionnelle
- Restriction optionnelle des permissions : le jeton n'aura accès qu'aux portails et permissions sélectionnés
- Restriction temporelle optionnelle : le jeton est soumis à un planning
- Le jeton brut n'est affiché qu'une seule fois à la création

Les administrateurs peuvent aussi créer et révoquer des jetons pour les autres membres.

#### 4.2.4 Changement de mot de passe

- Un membre peut changer son propre mot de passe (ancien + nouveau requis)
- Un administrateur peut réinitialiser le mot de passe d'un membre (sans connaître l'ancien)

### 4.3 Gestion des membres

*Réservé aux administrateurs.*

- **Créer** un membre : nom d'utilisateur, nom d'affichage optionnel, mot de passe, rôle (membre ou admin)
- **Lister** les membres avec pagination
- **Consulter** les détails d'un membre
- **Modifier** un membre : nom d'affichage, nom d'utilisateur, rôle
- **Supprimer** un membre (suppression définitive)

#### 4.3.1 Configuration d'authentification par membre

Chaque membre peut hériter de la configuration par défaut ou avoir une surcharge individuelle pour :

- Activation/désactivation de la connexion par mot de passe
- Activation/désactivation du SSO
- Activation/désactivation des jetons API

Trois états possibles : **hériter** (utilise la valeur par défaut de l'instance), **activé**, **désactivé**.

### 4.4 Gestion des portails

#### 4.4.1 CRUD des portails

*Création et suppression réservées aux administrateurs. Modification réservée aux administrateurs et gestionnaires du portail.*

- **Créer** un portail : nom, configuration des actions
- **Lister** les portails :
  - Un administrateur voit tous les portails
  - Un membre ne voit que les portails sur lesquels il a au moins une permission
- **Consulter** les détails d'un portail : statut en direct, métadonnées, codes d'accès, domaines
- **Modifier** la configuration d'un portail
- **Supprimer** un portail (suppression définitive)

#### 4.4.2 Actions sur un portail

- **Ouvrir** un portail : envoie la commande d'ouverture à l'appareil
- **Fermer** un portail : envoie la commande de fermeture (si configuré)
- Les actions respectent les permissions et les plannings horaires du membre

#### 4.4.3 Configuration des actions

Chaque portail possède trois actions configurables indépendamment : **ouverture**, **fermeture**, **remontée de statut**. Chaque action peut être de type :

- **MQTT** : publication d'un message sur un topic du broker
- **HTTP** : appel à un webhook externe (URL, méthode, headers, body)
- **Aucune** : action désactivée

#### 4.4.4 Statut et données en direct

- L'appareil pousse son statut vers le serveur (via MQTT ou HTTP entrant)
- Le statut est interprété via un **mapping configurable** (correspondance entre les champs reçus et les statuts affichés)
- Des **règles de statut** permettent de surcharger le statut affiché selon les métadonnées (ex. : si `battery < 20` → statut "low_battery")
- Des **transitions de statut** permettent de programmer un changement automatique après un délai (ex. : si statut "open" depuis 30s → passer à "closed")
- Le système détecte automatiquement les appareils **non réactifs** si aucune donnée n'est reçue dans un délai configurable (TTL)
- Les **statuts personnalisés** peuvent être définis en plus des statuts par défaut (open, closed, unavailable)

#### 4.4.5 Configuration des métadonnées

L'administrateur peut configurer quelles métadonnées sont affichées en direct, avec :

- Clé (chemin dans les données reçues, ex. : `lora.snr`)
- Libellé affiché (ex. : "Signal LoRa")
- Unité (ex. : "dBm")

#### 4.4.6 Jeton d'appareil (Gate Token)

Chaque portail possède un jeton secret utilisé par l'appareil physique pour s'authentifier auprès du serveur :

- Généré automatiquement à la création du portail (affiché une seule fois)
- **Rotation** possible : génère un nouveau jeton et invalide l'ancien
- Consultable par les administrateurs

### 4.5 Codes d'accès (PIN / Mot de passe)

*Gestion réservée aux gestionnaires du portail.*

Les codes d'accès permettent un accès public à un portail sans compte membre.

#### 4.5.1 Types de codes

- **Code PIN** : suite de chiffres (minimum 4), saisie via pavé numérique
- **Mot de passe** : chaîne de texte libre

#### 4.5.2 Création et gestion

- **Créer** un code : type (PIN/mot de passe), libellé, métadonnées de contrôle
- **Lister** les codes d'un portail
- **Modifier** un code : libellé, métadonnées
- **Supprimer** un code
- **Attacher un planning** : le code n'est valide que pendant les plages horaires définies
- **Détacher un planning**

#### 4.5.3 Métadonnées de contrôle d'un code

- **Date d'expiration** : le code devient invalide après cette date
- **Nombre d'utilisations maximum** : le code se désactive après N utilisations
- **Durée de session** : durée pendant laquelle la session ouverte par le code reste active (obligatoire — tout accès par code crée une session)
- **Permissions de session** : quelles actions sont autorisées pendant la session (ouvrir, fermer, consulter le statut)

#### 4.5.4 Accès public par code (mode session)

- L'utilisateur saisit un code PIN ou mot de passe
- Si valide, une **session temporaire** est créée avec les permissions et la durée définis dans le code
- L'utilisateur peut ensuite agir sur le portail pendant la durée de la session
- La session expire automatiquement

#### 4.5.5 PIN autogénéré (rotatif)

Un code PIN peut être configuré en mode **autogénéré** : le serveur génère automatiquement un nouveau code selon une fréquence définie par l'administrateur.

- **Fréquence de rotation** : configurable (ex. : toutes les heures, tous les jours, toutes les semaines)
- Le code courant est affiché dans l'interface d'administration (lisible, jamais hashé avant rotation)
- À chaque rotation, l'ancien code est invalidé et un nouveau est généré
- La rotation peut être déclenchée manuellement en plus du cycle automatique
- Compatible avec les plannings et les métadonnées de contrôle (expiration, nombre d'utilisations)

#### 4.5.6 Sécurité des codes

- **Limitation de tentatives** : maximum de tentatives par IP et par portail sur une fenêtre de temps (anti brute-force)
- **Limitation globale** : maximum de tentatives par IP tous portails confondus
- **Temps de réponse constant** : protection contre les attaques par timing

### 4.6 Permissions et politiques d'accès

#### 4.6.1 Attribution des permissions

- Un administrateur ou gestionnaire de portail peut **accorder** ou **révoquer** des permissions à un membre sur un portail
- Permissions disponibles : ouverture, fermeture, consultation du statut, gestion
- Les permissions sont indépendantes les unes des autres

#### 4.6.2 Restrictions temporelles par membre

- Un planning peut être attaché à un couple **membre + portail**
- Le membre ne pourra agir sur ce portail que pendant les plages autorisées
- La restriction s'applique en plus des permissions (il faut la permission ET être dans la plage horaire)

#### 4.6.3 Permissions par défaut

- L'administrateur peut configurer des **permissions par défaut** qui s'appliquent à tous les nouveaux membres
- Configuration via la page des paramètres

### 4.7 Domaines personnalisés

Chaque portail peut avoir un ou plusieurs **domaines personnalisés** pointant vers sa page d'accès public.

#### 4.7.1 Gestion des domaines

- **Ajouter** un domaine personnalisé à un portail
- **Vérifier** le domaine par challenge DNS : l'administrateur doit créer un enregistrement TXT `_gatie.{domain}` avec le jeton fourni
- **Supprimer** un domaine
- **Lister** les domaines d'un portail avec leur état de vérification

#### 4.7.2 Fonctionnement

- Un domaine vérifié redirige automatiquement vers la page d'accès du portail associé
- Le certificat TLS est provisionné automatiquement (via le proxy)
- La résolution du domaine vers le portail est publique (pas d'authentification requise)

### 4.8 Temps réel (Server-Sent Events)

L'application diffuse les événements en temps réel :

- **Changements de statut** des portails : mis à jour instantanément dans l'interface
- **Métadonnées en direct** : température, signal, batterie… actualisés en continu
- Les événements transitent de l'appareil vers le serveur, puis sont redistribués à tous les clients connectés
- Reconnexion automatique en cas de perte de connexion

### 4.9 Paramètres de l'instance

*Réservé aux administrateurs.*

- **Configuration d'authentification par défaut des membres** :
  - Activer/désactiver la connexion par mot de passe
  - Activer/désactiver le SSO
  - Activer/désactiver les jetons API
  - Nombre maximum de jetons API par membre
  - Durée de session par défaut
- **Fournisseurs SSO** : gestion des fournisseurs OIDC (ajout, modification, suppression). Les fournisseurs configurés via variables d'environnement s'affichent avec un indicateur de verrouillage et ne peuvent pas être modifiés via l'interface.
- **Permissions par défaut** : grille de permissions par portail applicables aux nouveaux membres

### 4.10 Health check

- Le système expose un point de contrôle de santé vérifiant la connectivité à la base de données et au cache

---

## 5. Interfaces utilisateur

L'interface doit couvrir l'ensemble des fonctionnalités décrites dans ce document, avec une UX soignée et adaptée au contexte (administration, accès public). La conception précise des écrans, la structure de navigation et les choix de composants sont laissés à la discrétion de l'implémenteur, en visant une expérience claire, cohérente et accessible.

### 5.1 Périmètre fonctionnel attendu

**Espace d'administration (membres authentifiés)** :
- Setup initial (premier administrateur)
- Connexion par mot de passe et/ou SSO
- Tableau de bord des portails avec statuts en temps réel
- Détail d'un portail : actions, métadonnées live, codes d'accès, domaines personnalisés, configuration
- Gestion des membres : création, édition, permissions, plannings, surcharges d'authentification
- Gestion des plannings : éditeur d'expressions logiques (ET/OU/NON) sur les différents types de règles
- Paramètres de l'instance : authentification, fournisseurs SSO, permissions par défaut
- Gestion des jetons API personnels (depuis le profil utilisateur)

**Espace public (accès invité)** :
- Page d'accès à un portail (via domaine personnalisé ou lien direct)
- Saisie de code PIN (pavé numérique) et/ou mot de passe
- Connexion membre dans le contexte d'un portail
- Interface de session active (actions selon permissions, données en direct)

### 5.2 Exigences transversales

- Design responsive : mobile-first, fonctionnel sur tablette et desktop
- Mode clair / mode sombre
- Internationalisation (i18n)
- Retours visuels clairs sur toutes les opérations (chargement, succès, erreur)
- Feedback de copie dans le presse-papiers

---

## 6. Règles métier transversales

### 6.1 Sécurité

- Les mots de passe respectent des exigences de complexité configurable (longueur minimale, majuscule, minuscule, chiffre)
- Les jetons API sont hashés en base et ne sont affichés qu'une seule fois
- Les codes PIN sont hashés (bcrypt) et jamais exposés en clair
- Les endpoints sensibles sont protégés par limitation de débit (rate limiting)
- Les réponses aux tentatives de code sont à durée constante (anti-timing)
- Les états anti-CSRF sont à usage unique avec expiration courte

### 6.2 Contrôle d'accès

- Toute action sur un portail vérifie : (1) la permission du membre, (2) le planning horaire si applicable
- Les administrateurs ont un accès implicite à tous les portails
- Les membres ne voient que les portails sur lesquels ils ont des permissions

### 6.3 Temps réel

- Les statuts de portail sont propagés en temps réel depuis l'appareil jusqu'à l'interface
- Un appareil qui ne communique plus est automatiquement marqué comme "non réactif" après un délai configurable
- Les transitions de statut automatiques se déclenchent après un délai défini

### 6.4 Gestion des sessions

- Les sessions membres ont une durée configurable avec rafraîchissement automatique
- Les sessions PIN ont une durée définie par le code utilisé
- La déconnexion révoque immédiatement la session côté serveur

### 6.5 Jetons API et contrôle fin

- Un jeton API peut avoir un planning global (restriction temporelle unique appliquée à tous les portails accessibles via ce jeton)

### 6.6 Polling de statut (webhook)

- Pour les portails configurés en mode webhook HTTP, le serveur interroge périodiquement l'appareil (polling) pour récupérer son statut
- Le polling supporte les tentatives multiples avec délai configurable en cas d'échec

### 6.7 Audit

- Le système prévoit un journal d'audit enregistrant les actions sensibles (ouvertures, tentatives de connexion, modifications de configuration) avec l'identifiant du membre, l'adresse IP, le portail concerné et les métadonnées associées

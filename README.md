# Notes API — Serveur

API REST de gestion de notes organisées par espaces, développée en Go.

Ce dépôt contient **le serveur backend**. Le client web se trouve dans un
dépôt séparé : [test-technique-ynov-client](https://github.com/KxroTM/test-technique-ynov-client).

## Présentation

L'application permet à un utilisateur authentifié d'organiser ses notes dans
des espaces thématiques (« Devoirs », « Jobs », « Personnel »…).

Le modèle de données suit la hiérarchie suivante :

```
Utilisateur  1 ──── n  Espace  1 ──── n  Note
```

Une note appartient obligatoirement à un espace, et un espace appartient
obligatoirement à un utilisateur. Chaque utilisateur n'a accès qu'à ses
propres données.

## Stack technique

| Composant       | Choix                        | Raison |
|-----------------|------------------------------|--------|
| Langage         | Go 1.25                      | Imposé par le sujet |
| Routeur HTTP    | Gin                          | Routing et middlewares concis, large adoption |
| Base de données | PostgreSQL 16                | Base relationnelle, contraintes d'intégrité natives |
| Accès aux données | `database/sql` + pgx        | SQL écrit à la main, contrôle total sur les requêtes |
| Mots de passe   | bcrypt                       | Fonction de hachage lente conçue pour les mots de passe |
| Authentification| JWT (HS256)                  | Imposé par le sujet, API sans état |

## Prérequis

- **Go 1.25** ou supérieur
- **Docker** et **Docker Compose** (pour la base de données)

## Installation

### 1. Cloner le dépôt

```bash
git clone https://github.com/KxroTM/test-technique-ynov.git
cd test-technique-ynov
```

### 2. Configurer l'environnement

```bash
cp .env.example .env
```

Le fichier `.env.example` contient des valeurs fonctionnelles pour un
environnement de développement local. Aucune modification n'est nécessaire
pour démarrer.

### 3. Démarrer la base de données

```bash
docker compose up -d
```

Le schéma et les comptes de démonstration sont créés automatiquement au
premier démarrage du conteneur, via les fichiers SQL du dossier
`migrations/`.

### 4. Lancer le serveur

```bash
go run ./cmd/api
```

Le serveur écoute par défaut sur `http://localhost:8080`.

Vérification rapide :

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## Comptes de démonstration

Deux comptes sont créés automatiquement avec les données de démonstration :

| Email               | Mot de passe  | Contenu |
|---------------------|---------------|---------|
| `alice@example.com` | `password123` | 2 espaces (« Devoirs », « Jobs »), 3 notes |
| `bob@example.com`   | `password123` | 1 espace (« Personnel »), 1 note |

Le second compte n'est pas là pour faire nombre : il permet de vérifier
concrètement le cloisonnement des données. Connecté en tant qu'Alice, aucune
donnée de Bob n'est accessible, y compris en ciblant directement les
identifiants de ses ressources.

## API

Toutes les routes sont préfixées par `/api`. Les réponses sont en JSON.

### Routes publiques

| Méthode | Route            | Description |
|---------|------------------|-------------|
| `POST`  | `/auth/register` | Création d'un compte |
| `POST`  | `/auth/login`    | Connexion, retourne un jeton JWT |

### Routes protégées

Elles exigent l'en-tête `Authorization: Bearer <jeton>`.

| Méthode  | Route              | Description |
|----------|--------------------|-------------|
| `GET`    | `/me`              | Profil de l'utilisateur connecté |
| `GET`    | `/spaces`          | Liste des espaces de l'utilisateur |
| `POST`   | `/spaces`          | Création d'un espace |
| `GET`    | `/spaces/:spaceID` | Détail d'un espace |
| `PUT`    | `/spaces/:spaceID` | Modification d'un espace |
| `DELETE` | `/spaces/:spaceID` | Suppression d'un espace et de ses notes |
| `GET`    | `/spaces/:spaceID/notes` | Espace et ses notes (FT3) |
| `POST`   | `/spaces/:spaceID/notes` | Ajout d'une note dans cet espace (FT4) |
| `GET`    | `/notes/:noteID`   | Détail d'une note |
| `PUT`    | `/notes/:noteID`   | Modification d'une note (FT5) |
| `DELETE` | `/notes/:noteID`   | Suppression d'une note (FT6) |

### Modèle de données des notes

Une note comporte un titre, un contenu et un état parmi trois valeurs :

| Valeur        | Libellé affiché |
|---------------|-----------------|
| `todo`        | Non fait |
| `in_progress` | En cours |
| `done`        | Terminé |

L'état est facultatif à la création (`todo` par défaut) mais **obligatoire**
à la modification : un `PUT` remplace l'intégralité de la note, et omettre
l'état reviendrait à le réinitialiser silencieusement.

La note est rattachée à l'espace indiqué dans l'URL, jamais à un espace
transmis dans le corps de la requête. Le rattachement est ainsi porté par la
route elle-même, et non par une donnée que le client pourrait choisir
librement.

### Format des erreurs

Toutes les erreurs partagent la même structure :

```json
{ "error": "ressource introuvable" }
```

Les erreurs de validation détaillent en plus chaque champ fautif :

```json
{
  "error": "données invalides",
  "fields": {
    "email": "adresse email invalide",
    "password": "ce champ doit contenir au moins 8 caractères"
  }
}
```

### Codes de statut utilisés

| Code  | Signification |
|-------|---------------|
| `200` | Succès |
| `201` | Ressource créée |
| `204` | Suppression effectuée, pas de contenu à retourner |
| `400` | Requête malformée ou données invalides |
| `401` | Jeton absent, invalide ou expiré ; identifiants incorrects |
| `404` | Ressource inexistante **ou appartenant à un autre utilisateur** |
| `409` | Conflit : email ou nom d'espace déjà utilisé |
| `500` | Erreur interne (journalisée côté serveur, non détaillée au client) |

Le choix du `404` plutôt que du `403` pour une ressource appartenant à
quelqu'un d'autre est volontaire. Répondre `403` confirmerait que la ressource
existe, ce qui permettrait à un utilisateur de cartographier les données des
autres en énumérant les identifiants.

### Exemple d'utilisation

```bash
# Connexion et récupération du jeton
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password123"}' \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# Liste des espaces
curl http://localhost:8080/api/spaces -H "Authorization: Bearer $TOKEN"

# Création d'un espace
curl -X POST http://localhost:8080/api/spaces \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Projets","description":"Idées personnelles"}'
```

## Tests

Le projet contient deux natures de tests, exécutables séparément.

**Tests unitaires** — sans dépendance externe, ils couvrent le hachage des mots
de passe, la génération et la vérification des jetons JWT, et la validation des
états de note :

```bash
go test ./...
```

**Tests d'intégration** — ils s'exécutent contre une véritable base PostgreSQL,
car c'est le comportement du SQL lui-même que l'on veut vérifier : filtrage par
utilisateur dans les clauses `WHERE`, contraintes d'unicité, suppressions en
cascade. Les simuler avec un faux repository ne prouverait rien de ce qui
compte ici.

```bash
docker compose up -d
TEST_DATABASE_URL="postgres://notes_user:notes_password@localhost:5432/notes_db?sslmode=disable" \
  go test ./internal/repository -v
```

Sans la variable `TEST_DATABASE_URL`, ces tests sont **ignorés** et non en
échec : `go test ./...` reste donc exécutable sans base disponible.

Chaque test d'intégration crée ses propres utilisateurs jetables et les
supprime à la fin. Les données de démonstration ne sont jamais modifiées, et
les tests peuvent être relancés indéfiniment.

## Commandes utiles

Un `Makefile` regroupe les commandes courantes (`make help` pour la liste) :

| Commande            | Effet |
|---------------------|-------|
| `make db-up`        | Démarre PostgreSQL |
| `make db-down`      | Arrête PostgreSQL |
| `make db-reset`     | Recrée la base à zéro |
| `make run`          | Lance le serveur |
| `make test`         | Lance les tests |
| `make test-coverage`| Lance les tests avec le taux de couverture |
| `make test-integration` | Lance les tests d'intégration (base requise) |

Si `make` n'est pas disponible (Windows sans outils GNU), les commandes
équivalentes sont indiquées dans le tableau ci-dessus et utilisables
directement.

## Documentation technique

Le document `docs/documentation-technique.pdf` présente les choix techniques,
l'architecture, la modélisation, les partis pris d'implémentation et les
limites de la solution.

Il est écrit en HTML (`docs/documentation-technique.html`) et converti en PDF,
afin que la source reste versionnable et comparable d'une version à l'autre.
Pour le régénérer après modification :

```bash
chrome --headless --no-pdf-header-footer \
  --print-to-pdf=docs/documentation-technique.pdf \
  docs/documentation-technique.html
```

## Structure du projet

```
cmd/api/              Point d'entrée : assemblage des briques et démarrage
internal/
  apperrors/          Erreurs métier, traduites en statuts HTTP au bord
  config/             Chargement de la configuration depuis l'environnement
  database/           Ouverture et réglage de la connexion PostgreSQL
  models/             Entités métier (Utilisateur, Espace, Note)
  repository/         Accès aux données, requêtes SQL
  service/            Logique métier et règles d'accès
  handlers/           Handlers HTTP, décodage et validation des requêtes
  middleware/         Authentification, traduction des erreurs
migrations/           Schéma SQL et données de démonstration
```

Le découpage suit une dépendance à sens unique :

```
handlers → service → repository → base de données
```

Chaque couche ne connaît que la suivante. Les handlers ne contiennent aucune
règle métier, et le `repository` ne connaît pas le protocole HTTP.

## Configuration

| Variable       | Obligatoire | Défaut | Description |
|----------------|-------------|--------|-------------|
| `PORT`         | non         | `8080` | Port d'écoute du serveur |
| `DATABASE_URL` | **oui**     | —      | Chaîne de connexion PostgreSQL |
| `JWT_SECRET`   | **oui**     | —      | Clé de signature des jetons JWT |

Les deux variables obligatoires n'ont volontairement pas de valeur par
défaut : le serveur refuse de démarrer si elles sont absentes, plutôt que de
tourner avec une configuration incomplète ou un secret prévisible.

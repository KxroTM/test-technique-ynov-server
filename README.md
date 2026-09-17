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
| Langage         | Go 1.21                      | Imposé par le sujet |
| Routeur HTTP    | Gin                          | Routing et middlewares concis, large adoption |
| Base de données | PostgreSQL 16                | Base relationnelle, contraintes d'intégrité natives |
| Accès aux données | `database/sql` + pgx        | SQL écrit à la main, contrôle total sur les requêtes |
| Mots de passe   | bcrypt                       | Fonction de hachage lente conçue pour les mots de passe |
| Authentification| JWT (HS256)                  | Imposé par le sujet, API sans état |

## Prérequis

- **Go 1.21** ou supérieur
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

Si `make` n'est pas disponible (Windows sans outils GNU), les commandes
équivalentes sont indiquées dans le tableau ci-dessus et utilisables
directement.

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

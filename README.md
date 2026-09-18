# Notes API : serveur

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
| Langage         | Go 1.25                      | Binaire statique sans dépendance système, concurrence native |
| Routeur HTTP    | Gin                          | Routing et middlewares concis, large adoption |
| Base de données | PostgreSQL 16                | Base relationnelle, contraintes d'intégrité natives |
| Accès aux données | `database/sql` + pgx        | SQL écrit à la main, contrôle total sur les requêtes |
| Mots de passe   | bcrypt                       | Fonction de hachage lente conçue pour les mots de passe |
| Authentification| JWT (HS256)                  | API sans état : aucune session à conserver côté serveur |

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
go run .
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
| `POST`  | `/auth/google`   | Connexion via un compte Google (facultatif) |

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
| `GET`    | `/spaces/:spaceID/notes` | Espace et ses notes |
| `POST`   | `/spaces/:spaceID/notes` | Ajout d'une note dans cet espace |
| `GET`    | `/notes/:noteID`   | Détail d'une note |
| `PUT`    | `/notes/:noteID`   | Modification d'une note |
| `DELETE` | `/notes/:noteID`   | Suppression d'une note |

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

## Connexion Google

En plus de l'email et du mot de passe, l'API accepte une connexion par compte
Google. **Cette fonctionnalité est facultative** : sans identifiants OAuth
configurés, l'endpoint répond `503` et le client masque le bouton. Toute
l'application reste utilisable.

### Configuration

Créer un identifiant OAuth de type « Application Web » sur
[console.cloud.google.com](https://console.cloud.google.com), avec
`http://localhost:3000/auth/google/callback` en URI de redirection autorisée,
puis renseigner dans `.env` :

```
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
```

### Répartition des rôles

C'est **l'API qui possède l'authentification**. Le client redirige l'utilisateur
vers Google avec le `client_id`, qui est public, mais seul le serveur détient le
`client_secret`, échange le code contre un jeton d'identité et émet le JWT de
l'application. Le secret ne transite jamais par le navigateur, et il n'existe
qu'un seul endroit qui fabrique des jetons.

L'échange est fait avec la bibliothèque standard, sans dépendance
supplémentaire. La signature du jeton d'identité n'est pas revérifiée : il est
reçu directement de Google sur un canal HTTPS authentifié, cas que la
documentation Google dispense explicitement de vérification.

### Règles appliquées

| Situation | Comportement |
|-----------|--------------|
| Compte Google déjà connu | Connexion |
| Email inconnu | Création du compte, sans mot de passe |
| Email déjà utilisé par un compte à mot de passe | Les deux méthodes sont liées sur le même compte |
| Email non vérifié chez Google | **Refus** |

Le dernier point est une garde de sécurité : sans elle, créer un compte Google
portant l'adresse d'un tiers suffirait à prendre le contrôle de son compte.

L'identifiant conservé est le champ `sub` du jeton, stable dans le temps, et non
l'email qui peut changer chez Google.

## Vérifications

Les contrôles suivants ont été menés en interrogeant le serveur en
fonctionnement, et en tentant de violer directement les contraintes en SQL :

| Contrôle | Résultat attendu |
|----------|------------------|
| Connexion avec un mot de passe erroné ou un email inconnu | `401`, message identique dans les deux cas |
| Jeton absent, falsifié, expiré, ou forgé en `alg=none` | `401` |
| Lecture, modification ou suppression d'une ressource d'autrui | `404` |
| Création d'une note dans l'espace d'autrui | `404`, et aucune ligne insérée |
| Email ou nom d'espace déjà utilisé | `409` |
| Identifiant d'URL non numérique ou négatif | `400` |
| État de note hors des trois valeurs autorisées | `400` |
| Suppression d'un espace contenant des notes | `204`, notes supprimées en cascade |

Les invariants du schéma ont été éprouvés directement en SQL, en tentant de les
violer : état de note invalide, espace en doublon, note sans espace. Les trois
tentatives sont rejetées par la base.

## Commandes utiles

| Commande | Effet |
|----------|-------|
| `docker compose up -d` | Démarre PostgreSQL |
| `docker compose down` | Arrête PostgreSQL en conservant les données |
| `docker compose down -v && docker compose up -d` | Recrée la base à zéro et rejoue les migrations |
| `go run .` | Lance le serveur |

## Documentation technique

Le document `docs/documentation-technique.pdf` présente les choix techniques,
l'architecture, la modélisation, les partis pris d'implémentation et les
vérifications menées. Il couvre l'ensemble de la solution, serveur et client, et
il est identique dans les deux dépôts.

## Structure du projet

```
main.go               Point d'entrée : assemblage des briques et démarrage
internal/
  apperrors/          Erreurs métier, traduites en statuts HTTP au bord
  config/             Chargement de la configuration depuis l'environnement
  database/           Ouverture et réglage de la connexion PostgreSQL
  models/             Entités métier (Utilisateur, Espace, Note)
  repository/         Accès aux données, requêtes SQL
  service/            Logique métier et règles d'accès
  handlers/           Handlers HTTP, décodage et validation des requêtes
  middleware/         Authentification, traduction des erreurs
migrations/           Schéma SQL, données de démonstration, colonnes Google
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
| `DATABASE_URL` | **oui**     | aucun  | Chaîne de connexion PostgreSQL |
| `JWT_SECRET`   | **oui**     | aucun  | Clé de signature des jetons JWT |
| `GOOGLE_CLIENT_ID` | non | aucun | Identifiant OAuth Google, active la connexion Google |
| `GOOGLE_CLIENT_SECRET` | non | aucun | Secret OAuth Google, jamais versionné |

Les deux variables obligatoires n'ont volontairement pas de valeur par
défaut : le serveur refuse de démarrer si elles sont absentes, plutôt que de
tourner avec une configuration incomplète ou un secret prévisible.

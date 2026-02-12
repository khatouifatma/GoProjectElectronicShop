# 🏪 Electronic Shop Management API

Backend Go pour un système de gestion multi-boutiques d'électronique avec isolation multi-tenant, gestion des rôles, API publique et redirection WhatsApp dynamique.

## 🚀 Stack Technique

| Composant | Technologie |
|-----------|-------------|
| Langage | Go 1.22 |
| Framework HTTP | Gin v1.10 |
| ORM | GORM v2 |
| Base de données | PostgreSQL 15 |
| Authentification | JWT (HS256) |
| Hashage passwords | bcrypt (cost=12) |

## 📁 Architecture

```
electronic-shop/
├── cmd/
│   └── main.go              # Point d'entrée, routes
├── config/
│   └── config.go            # DB connection, env
├── internal/
│   ├── handlers/            # Contrôleurs HTTP
│   │   ├── auth.go          # Register, Login
│   │   ├── shop.go          # Gestion shop
│   │   ├── product.go       # CRUD produits
│   │   ├── transaction.go   # CRUD transactions
│   │   ├── user.go          # Gestion utilisateurs
│   │   ├── report.go        # Dashboard
│   │   └── public.go        # Routes publiques + WhatsApp
│   ├── middleware/
│   │   └── auth.go          # JWT + CheckRole
│   ├── models/
│   │   └── models.go        # Shop, User, Product, Transaction
│   └── dto/
│       └── dto.go           # Request/Response structs
├── .env.example
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## ⚡ Installation rapide

### Option 1 – Docker (recommandé)

```bash
# Cloner le projet
git clone <repo-url>
cd electronic-shop

# Copier les variables d'environnement
cp .env.example .env

# Lancer avec Docker Compose
docker-compose up --build
```

L'API sera disponible sur `http://localhost:8080`

### Option 2 – Installation manuelle

**Prérequis :** Go 1.22+, PostgreSQL 15+

```bash
# 1. Cloner et installer les dépendances
git clone <repo-url>
cd electronic-shop
go mod download

# 2. Créer la base de données PostgreSQL
psql -U postgres -c "CREATE DATABASE electronic_shop;"

# 3. Configurer les variables d'environnement
cp .env.example .env
# Modifier .env avec vos paramètres
#psql postgres -c "CREATE USER postgres WITH PASSWORD 'postgres' SUPERUSER;"
#verifier la creation du suer psql postgres -c "\du"

# 4. Lancer le serveur (migration automatique au démarrage)
go run cmd/main.go
```

## 🔑 Variables d'environnement

| Variable | Description | Défaut |
|----------|-------------|--------|
| `PORT` | Port du serveur | `8080` |
| `DB_HOST` | Hôte PostgreSQL | `localhost` |
| `DB_USER` | Utilisateur PostgreSQL | `postgres` |
| `DB_PASSWORD` | Mot de passe PostgreSQL | `postgres` |
| `DB_NAME` | Nom de la base de données | `electronic_shop` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `JWT_SECRET` | Clé secrète JWT | ⚠️ **Changer en production** |

## 🌐 Routes API

### 🔓 Auth (publique)
| Méthode | Route | Description |
|---------|-------|-------------|
| POST | `/auth/register` | Créer un compte + shop |
| POST | `/auth/login` | Se connecter → JWT |

### 🌍 Public (sans authentification)
| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/public/:shopID/products` | Liste des produits publics |
| GET | `/public/:shopID/products/:productID/whatsapp` | Lien WhatsApp dynamique |

### 🔒 Privé (JWT requis)

**Produits**
| Méthode | Route | Rôle requis |
|---------|-------|-------------|
| GET | `/api/products` | Admin, SuperAdmin |
| GET | `/api/products/:id` | Admin, SuperAdmin |
| POST | `/api/products` | Admin, SuperAdmin |
| PUT | `/api/products/:id` | Admin, SuperAdmin |
| DELETE | `/api/products/:id` | Admin, SuperAdmin |

**Transactions**
| Méthode | Route | Rôle requis |
|---------|-------|-------------|
| GET | `/api/transactions` | Admin, SuperAdmin |
| POST | `/api/transactions` | Admin, SuperAdmin |

**Utilisateurs (SuperAdmin seulement)**
| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/api/users` | Liste des utilisateurs du shop |
| POST | `/api/users` | Créer un utilisateur |
| DELETE | `/api/users/:id` | Supprimer un utilisateur |

**Shop (SuperAdmin seulement)**
| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/api/shops` | Infos du shop |
| PUT | `/api/shops/whatsapp` | Modifier le numéro WhatsApp |

**Dashboard (SuperAdmin seulement)**
| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/api/reports/dashboard` | Ventes, dépenses, profit, stock faible |

## 📋 Exemples d'utilisation

### 1. Créer un shop + SuperAdmin

```bash
POST /auth/register
{
  "name": "Ahmed Benali",
  "email": "ahmed@techshop.ma",
  "password": "password123",
  "role": "SuperAdmin",
  "shop_name": "TechShop Casablanca",
  "whatsapp_number": "212600000000"
}
```

### 2. Connexion → récupérer le JWT

```bash
POST /auth/login
{
  "email": "ahmed@techshop.ma",
  "password": "password123"
}
# → { "token": "eyJ...", "user": {...} }
```

### 3. Créer un produit (avec JWT)

```bash
POST /api/products
Authorization: Bearer eyJ...
{
  "name": "iPhone 15 Pro",
  "description": "Smartphone Apple dernière génération",
  "category": "Smartphones",
  "purchase_price": 8500,
  "selling_price": 11999,
  "stock": 10,
  "image_url": "https://example.com/iphone15.jpg"
}
```

### 4. Page publique d'un shop

```bash
GET /public/SHOP-UUID/products
# Retourne les produits SANS PurchasePrice
# Chaque produit inclut whatsapp_link
```

### 5. Lien WhatsApp dynamique

```bash
GET /public/SHOP-UUID/products/PRODUCT-UUID/whatsapp
# Retourne:
{
  "whatsapp_link": "https://wa.me/212600000000?text=Bonjour%20je%20veux%20plus%20d%27information%20sur%20iPhone%2015%20Pro"
}
```

### 6. Dashboard SuperAdmin

```bash
GET /api/reports/dashboard
Authorization: Bearer eyJ...
# Retourne:
{
  "total_sales": 45000,
  "total_expenses": 12000,
  "net_profit": 33000,
  "low_stock_products": [...],
  "total_products": 25,
  "total_transactions": 142
}
```

### 7. Créer une transaction de vente

```bash
POST /api/transactions
Authorization: Bearer eyJ...
{
  "type": "Sale",
  "product_id": "PRODUCT-UUID",
  "quantity": 2,
  "amount": 23998
}
# Le stock est automatiquement décrémenté (vérifié pour ne pas aller < 0)
```

## 🔐 Rôles et permissions

| Action | SuperAdmin | Admin | Guest (public) |
|--------|-----------|-------|----------------|
| Voir PurchasePrice | ✅ | ❌ | ❌ |
| Voir profit/dashboard | ✅ | ❌ | ❌ |
| Modifier WhatsApp | ✅ | ❌ | ❌ |
| Gérer utilisateurs | ✅ | ❌ | ❌ |
| CRUD produits | ✅ | ✅ | ❌ |
| CRUD transactions | ✅ | ✅ | ❌ |
| Voir produits publics | ✅ | ✅ | ✅ |

## 🏢 Isolation Multi-tenant

**Principe fondamental :** Le `shopID` est **toujours** extrait du JWT, jamais de l'URL.

- Chaque utilisateur appartient à un shop
- Toutes les requêtes privées filtrent automatiquement par le `shopID` du token
- Impossible d'accéder aux données d'un autre shop, même en modifiant l'URL

## 📊 Modèle de données (ERD simplifié)

```
Shop (1) ──── (N) User
Shop (1) ──── (N) Product
Shop (1) ──── (N) Transaction
Product (1) ── (N) Transaction
```

## 🧪 Tests de sécurité

Pour tester l'isolation multi-tenant :
```bash
# 1. Créer Shop A avec SuperAdmin A
# 2. Créer Shop B avec SuperAdmin B
# 3. Login avec SuperAdmin A → token A
# 4. Essayer GET /api/products avec token A sur des produits de Shop B
# → 404 Not Found (isolation correcte)
```

## 📦 Types de transactions

| Type | Description |
|------|-------------|
| `Sale` | Vente d'un produit (décrémente le stock) |
| `Expense` | Dépense opérationnelle |
| `Withdrawal` | Retrait de fonds |

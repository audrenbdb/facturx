# Factur-X Web Interface

Interface web pour générer des factures Factur-X conformes.

## Prérequis

- [Bun](https://bun.sh/) (gestionnaire de paquets)
- [Go](https://golang.org/) 1.21+ (pour le serveur API)

## Démarrage rapide

### 1. Installer les dépendances

```bash
cd web
bun install
```

### 2. Lancer le serveur API

Dans un terminal :

```bash
cd web/server
go run main.go
```

Le serveur démarre sur `http://localhost:8080`

### 3. Lancer le frontend

Dans un autre terminal :

```bash
cd web
bun run dev
```

Le frontend démarre sur `http://localhost:5173`

## Commandes disponibles

| Commande | Description |
|----------|-------------|
| `bun run dev` | Démarre le serveur de développement Vite |
| `bun run build` | Compile pour la production dans `dist/` |
| `bun run preview` | Prévisualise la version compilée |

## Architecture

```
web/
├── src/
│   ├── components/
│   │   ├── Hero.jsx         # Page d'accueil et features
│   │   ├── InvoiceForm.jsx  # Formulaire de création
│   │   └── Footer.jsx       # Pied de page
│   ├── App.jsx              # Composant racine
│   ├── main.jsx             # Point d'entrée
│   └── index.css            # Styles Tailwind
├── server/
│   └── main.go              # API Go pour générer les PDFs
├── public/
│   └── favicon.svg
├── package.json
├── tailwind.config.js
└── vite.config.js
```

## API

### POST /api/generate

Génère une facture Factur-X PDF.

**Corps de la requête :**

```json
{
  "number": "FAC-2026-001",
  "date": "2026-01-15",
  "seller": {
    "name": "Ma Société SARL",
    "siret": "10900000000009",
    "vatNumber": "FR10900000000",
    "street": "123 Rue de Paris",
    "postalCode": "75001",
    "city": "Paris",
    "email": "contact@masociete.fr"
  },
  "buyer": {
    "name": "Client SA",
    "siret": "11700000000002",
    "street": "456 Avenue des Champs",
    "postalCode": "69001",
    "city": "Lyon",
    "email": "achat@client.fr"
  },
  "lines": [
    {
      "description": "Prestation de conseil",
      "quantity": 10,
      "unitPrice": 150,
      "vatRegime": 0
    }
  ],
  "paymentTerms": {
    "dueDate": "2026-02-15",
    "iban": "FR7630001007941234567890185",
    "bic": "BDFEFRPP",
    "note": "Paiement par virement bancaire"
  }
}
```

**Codes de régime TVA :**

| Code | Régime |
|------|--------|
| 0 | Standard 20% |
| 1 | Réduit 10% |
| 2 | Super-réduit 5,5% |
| 3 | Minimal 2,1% |
| 4 | Auto-entrepreneur (franchise) |
| 5 | Exonéré santé |

**Réponse :** Fichier PDF binaire

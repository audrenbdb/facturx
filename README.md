# Factur-X Go

Librairie Go pour générer des factures électroniques conformes au standard **Factur-X 1.0** (profil BASIC).

```go
go get github.com/audrenbdb/facturx
```

## Caractéristiques

- **Zéro dépendance** : pur Go, aucune librairie externe
- **PDF/A-3** : génération native octet par octet
- **XML CII embarqué** : Cross-Industry Invoice conforme EN 16931
- **Validation SIRET** : algorithme de Luhn intégré
- **Régimes TVA français** : standard, réduit, auto-entrepreneur, exonérations

## Démo

Une interface web de démonstration est disponible sur **[facturx.deiz.fr](https://facturx.deiz.fr)** pour tester la génération de factures.

## Validation

Les factures générées ont été validées par le [service officiel FNFE-MPE](https://services.fnfe-mpe.org) :

| Validation | Statut |
|------------|--------|
| Métadonnées XMD | ✅ Conforme |
| XML contre XSD | ✅ Conforme |
| Règles Schematron | ✅ Conforme |
| PDF/A-3 | ✅ Conforme |

## Utilisation

```go
package main

import (
    "os"
    "github.com/audrenbdb/facturx"
)

func main() {
    req := facturx.InvoiceRequest{
        Number: "FAC-2026-001",
        Date:   "20260131", // Format YYYYMMDD
        Seller: facturx.Contact{
            Name:        "Mon Entreprise SARL",
            Address:     "42 rue de Paris",
            ZipCode:     "75001",
            City:        "Paris",
            CountryCode: "FR",
            Siret:       "12345678901234",
            VatNumber:   "FR12345678901",
        },
        Buyer: facturx.Contact{
            Name:        "Client SA",
            Address:     "15 avenue des Champs",
            ZipCode:     "69001",
            City:        "Lyon",
            CountryCode: "FR",
            Siret:       "98765432101234",
        },
        Lines: []facturx.InvoiceLine{
            {
                Description: "Prestation de conseil",
                Quantity:    5,
                UnitPrice:   500.00,
            },
            {
                Description: "Formation (1 jour)",
                Quantity:    1,
                UnitPrice:   1200.00,
            },
        },
        Regime: facturx.VatStandard(20.0),
    }

    pdf, err := facturx.Generate(req)
    if err != nil {
        panic(err)
    }

    os.WriteFile("facture.pdf", pdf, 0644)
}
```

## Régimes de TVA

```go
// TVA standard (taux personnalisable)
facturx.VatStandard(20.0)  // 20%
facturx.VatStandard(10.0)  // 10%
facturx.VatStandard(5.5)   // 5.5%

// Auto-entrepreneur (franchise en base de TVA)
facturx.VatFranchiseAuto()
// → "TVA non applicable, art. 293 B du CGI"

// Exonération activités de santé
facturx.VatExemptHealth()
// → "Exonération de TVA, art. 261-4-1° du CGI"
```

## Options

```go
req := facturx.InvoiceRequest{
    // ...

    // Ajouter ", EI" après le nom du vendeur
    AddEISuffix: true,

    // Mentions personnalisées (paiement, IBAN, etc.)
    CustomMentions: "Paiement à 30 jours par virement.\nIBAN: FR76 1234 5678 9012",
}
```

## Philosophie

Cette librairie a un objectif précis : **générer des factures conformes, simplement et rapidement**.

Elle ne permet pas de :
- Personnaliser le style ou la mise en page
- Ajouter un logo
- Modifier les couleurs ou polices

Si vous avez besoin de factures personnalisées, cette librairie n'est pas faite pour vous.
Si vous avez besoin de factures conformes Factur-X sans vous prendre la tête, vous êtes au bon endroit.

## Conformité technique

- **PDF/A-3b** : archivage long terme, profil ICC sRGB embarqué
- **Factur-X 1.0 BASIC** : profil suffisant pour la majorité des entreprises françaises
- **EN 16931** : norme européenne de facturation électronique
- **Cross-Industry Invoice (CII)** : syntaxe UN/CEFACT D16B

## Avertissement

**Cette librairie est fournie "en l'état", sans garantie d'aucune sorte.**

L'auteur ne saurait être tenu responsable de tout dommage direct ou indirect résultant de l'utilisation de cette librairie, y compris mais sans s'y limiter : rejets de factures, non-conformité réglementaire, pertes financières ou litiges commerciaux.

**Il appartient à l'utilisateur de :**
- Vérifier la conformité des factures générées auprès des organismes compétents
- S'assurer que les factures respectent la réglementation en vigueur
- Valider les documents avant tout usage en production

Les validations mentionnées (FNFE-MPE) ont été effectuées à une date donnée et ne constituent pas une garantie de conformité permanente. La réglementation évolue ; il est de la responsabilité de l'utilisateur de s'assurer de la conformité continue de ses factures.

## Licence

[MIT License](LICENSE)

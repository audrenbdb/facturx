import { useState, useMemo, useCallback } from 'react'

const VAT_REGIMES = [
  { id: 'standard', label: 'Standard 20%', rate: 20, code: 0 },
  { id: 'reduced', label: 'Réduit 10%', rate: 10, code: 1 },
  { id: 'super_reduced', label: 'Réduit 5,5%', rate: 5.5, code: 2 },
  { id: 'minimal', label: 'Minimal 2,1%', rate: 2.1, code: 3 },
  { id: 'franchise', label: 'Auto-entrepreneur (exonéré)', rate: 0, code: 4 },
  { id: 'health', label: 'Exonéré santé', rate: 0, code: 5 },
  { id: 'education', label: 'Exonéré formation', rate: 0, code: 6 },
]

const VAT_EXEMPTIONS = [
  { id: 'none', label: 'TVA applicable', description: 'Régime normal avec TVA' },
  { id: 'franchise', label: 'Auto-entrepreneur', description: 'TVA non applicable, art. 293 B du CGI' },
  { id: 'health', label: 'Exonération santé', description: 'Art. 261-4-1° du CGI' },
  { id: 'education', label: 'Exonération formation', description: 'Services éducatifs exonérés' },
]

const initialLine = {
  description: '',
  quantity: 1,
  unitPrice: 0,
  vatRegime: 'standard',
}

const initialSeller = {
  name: '',
  siret: '',
  vatNumber: '',
  street: '',
  postalCode: '',
  city: '',
  email: '',
}

const initialBuyer = {
  name: '',
  siret: '',
  street: '',
  postalCode: '',
  city: '',
  email: '',
}

function formatSiret(value) {
  const digits = value.replace(/\D/g, '').slice(0, 14)
  return digits.replace(/(\d{3})(?=\d)/g, '$1 ').trim()
}

function formatCurrency(amount) {
  return new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency: 'EUR',
  }).format(amount)
}

function SectionHeader({ number, title, subtitle }) {
  return (
    <div className="flex items-start gap-3 sm:gap-4 mb-4 sm:mb-6">
      <div className="w-8 h-8 sm:w-10 sm:h-10 rounded-full bg-gradient-to-br from-tricolore-bleu to-marine-700 flex items-center justify-center text-white font-display font-bold text-sm sm:text-lg shrink-0">
        {number}
      </div>
      <div>
        <h2 className="section-title">{title}</h2>
        {subtitle && <p className="text-marine-500 text-xs sm:text-sm mt-1">{subtitle}</p>}
      </div>
    </div>
  )
}

function InputField({ label, error, ...props }) {
  return (
    <div>
      <label className="input-label">{label}</label>
      <input className={`input-field ${error ? 'border-tricolore-rouge focus:ring-tricolore-rouge/20' : ''}`} {...props} />
      {error && <p className="text-tricolore-rouge text-xs mt-1">{error}</p>}
    </div>
  )
}

function InvoiceLine({ line, index, onChange, onRemove, canRemove, isExempt }) {
  const vatRegime = VAT_REGIMES.find(v => v.id === line.vatRegime)
  const lineTotal = line.quantity * line.unitPrice

  return (
    <div className="p-4 bg-marine-50/50 rounded-xl border border-marine-100 animate-fade-in space-y-3">
      {/* Description - full width */}
      <div>
        <label className="input-label">Description</label>
        <input
          type="text"
          className="input-field"
          placeholder="Prestation de service..."
          value={line.description}
          onChange={e => onChange(index, 'description', e.target.value)}
        />
      </div>

      {/* Quantity, Price, TVA, Total - responsive grid */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div>
          <label className="input-label">Quantité</label>
          <input
            type="number"
            className="input-field text-center font-mono"
            min="0.01"
            step="0.01"
            value={line.quantity}
            onChange={e => onChange(index, 'quantity', parseFloat(e.target.value) || 0)}
          />
        </div>
        <div>
          <label className="input-label">Prix unitaire</label>
          <input
            type="number"
            className="input-field text-right font-mono"
            min="0"
            step="0.01"
            value={line.unitPrice}
            onChange={e => onChange(index, 'unitPrice', parseFloat(e.target.value) || 0)}
          />
        </div>
        {!isExempt && (
          <div>
            <label className="input-label">TVA</label>
            <select
              className="input-field"
              value={line.vatRegime}
              onChange={e => onChange(index, 'vatRegime', e.target.value)}
            >
              {VAT_REGIMES.map(regime => (
                <option key={regime.id} value={regime.id}>{regime.label}</option>
              ))}
            </select>
          </div>
        )}
        <div className={isExempt ? 'col-span-2 sm:col-span-1' : ''}>
          <label className="input-label">Total</label>
          <div className="py-3 font-mono font-semibold text-marine-900 text-right sm:text-left">
            {formatCurrency(lineTotal)}
          </div>
        </div>
      </div>

      {/* Delete button - bottom right on mobile */}
      <div className="flex justify-end pt-1">
        <button
          type="button"
          onClick={() => onRemove(index)}
          disabled={!canRemove}
          className="w-10 h-10 rounded-lg border border-marine-200 text-marine-400 hover:text-tricolore-rouge hover:border-tricolore-rouge disabled:opacity-30 disabled:cursor-not-allowed transition-colors flex items-center justify-center"
          title="Supprimer la ligne"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>
    </div>
  )
}

function TotalsDisplay({ lines }) {
  const totals = useMemo(() => {
    let subtotal = 0
    let totalVat = 0
    const vatBreakdown = {}

    lines.forEach(line => {
      const regime = VAT_REGIMES.find(v => v.id === line.vatRegime)
      const lineTotal = line.quantity * line.unitPrice
      const vatAmount = lineTotal * (regime?.rate || 0) / 100

      subtotal += lineTotal
      totalVat += vatAmount

      if (!vatBreakdown[regime.id]) {
        vatBreakdown[regime.id] = { label: regime.label, rate: regime.rate, base: 0, amount: 0 }
      }
      vatBreakdown[regime.id].base += lineTotal
      vatBreakdown[regime.id].amount += vatAmount
    })

    return {
      subtotal,
      totalVat,
      total: subtotal + totalVat,
      vatBreakdown: Object.values(vatBreakdown).filter(v => v.base > 0),
    }
  }, [lines])

  return (
    <div className="bg-gradient-to-br from-marine-800 to-marine-900 rounded-2xl p-6 text-white">
      <h3 className="font-display text-lg font-semibold mb-4 text-marine-200">Récapitulatif</h3>

      <div className="space-y-3 text-sm">
        <div className="flex justify-between">
          <span className="text-marine-300">Total HT</span>
          <span className="font-mono font-semibold">{formatCurrency(totals.subtotal)}</span>
        </div>

        {totals.vatBreakdown.map(vat => (
          <div key={vat.label} className="flex justify-between text-marine-400">
            <span>TVA {vat.rate}%</span>
            <span className="font-mono">{formatCurrency(vat.amount)}</span>
          </div>
        ))}

        <div className="h-px bg-marine-700 my-3" />

        <div className="flex justify-between items-baseline">
          <span className="text-marine-200 font-medium">Total TTC</span>
          <span className="font-mono text-2xl font-bold text-white">{formatCurrency(totals.total)}</span>
        </div>
      </div>
    </div>
  )
}

// Demo data for quick testing
const demoData = {
  invoiceNumber: `FAC-${new Date().getFullYear()}-${String(Math.floor(Math.random() * 999) + 1).padStart(3, '0')}`,
  seller: {
    name: 'TechConsult SARL',
    siret: '528 250 004 00033', // Valid Luhn checksum
    vatNumber: 'FR32528250004',
    street: '42 Boulevard Haussmann',
    postalCode: '75009',
    city: 'Paris',
    email: 'facturation@techconsult.fr',
  },
  buyer: {
    name: 'Dupont Industries SA',
    siret: '356 000 000 00048', // La Poste - valid Luhn checksum
    street: '15 Rue de la République',
    postalCode: '69002',
    city: 'Lyon',
    email: 'comptabilite@dupont-industries.fr',
  },
  lines: [
    {
      description: 'Développement application web - Phase 1',
      quantity: 5,
      unitPrice: 650,
      vatRegime: 'standard',
    },
    {
      description: 'Consultation technique et architecture',
      quantity: 2,
      unitPrice: 450,
      vatRegime: 'standard',
    },
    {
      description: 'Formation équipe développement (1 journée)',
      quantity: 1,
      unitPrice: 1200,
      vatRegime: 'reduced',
    },
  ],
  iban: 'FR76 3000 6000 0112 3456 7890 189',
  note: 'Paiement à 30 jours par virement bancaire.',
}

export default function InvoiceForm() {
  const [invoiceNumber, setInvoiceNumber] = useState('')
  const [invoiceDate, setInvoiceDate] = useState(new Date().toISOString().split('T')[0])
  const [dueDate, setDueDate] = useState('')
  const [vatExemption, setVatExemption] = useState('none')
  const [seller, setSeller] = useState(initialSeller)
  const [buyer, setBuyer] = useState(initialBuyer)
  const [lines, setLines] = useState([{ ...initialLine }])
  const [iban, setIban] = useState('')
  const [bic, setBic] = useState('')
  const [note, setNote] = useState('')
  const [isGenerating, setIsGenerating] = useState(false)
  const [errors, setErrors] = useState({})

  // Fill form with demo data
  const fillDemoData = useCallback(() => {
    setInvoiceNumber(demoData.invoiceNumber)
    setInvoiceDate(new Date().toISOString().split('T')[0])
    // Set due date to 30 days from now
    const due = new Date()
    due.setDate(due.getDate() + 30)
    setDueDate(due.toISOString().split('T')[0])
    setVatExemption('none')
    setSeller(demoData.seller)
    setBuyer(demoData.buyer)
    setLines(demoData.lines.map(l => ({ ...l })))
    setIban(demoData.iban)
    setBic('')
    setNote(demoData.note)
    setErrors({})
  }, [])

  // When VAT exemption changes, update all lines
  const handleVatExemptionChange = (exemptionId) => {
    setVatExemption(exemptionId)
    if (exemptionId !== 'none') {
      // Apply exemption to all lines
      setLines(prev => prev.map(line => ({ ...line, vatRegime: exemptionId })))
    } else {
      // Reset to standard
      setLines(prev => prev.map(line => ({ ...line, vatRegime: 'standard' })))
    }
  }

  const updateSeller = (field, value) => {
    if (field === 'siret') value = formatSiret(value)
    setSeller(prev => ({ ...prev, [field]: value }))
  }

  const updateBuyer = (field, value) => {
    if (field === 'siret') value = formatSiret(value)
    setBuyer(prev => ({ ...prev, [field]: value }))
  }

  const updateLine = (index, field, value) => {
    setLines(prev => {
      const newLines = [...prev]
      newLines[index] = { ...newLines[index], [field]: value }
      return newLines
    })
  }

  const addLine = () => {
    const newLine = {
      ...initialLine,
      vatRegime: vatExemption !== 'none' ? vatExemption : 'standard'
    }
    setLines(prev => [...prev, newLine])
  }

  const removeLine = (index) => {
    if (lines.length > 1) {
      setLines(prev => prev.filter((_, i) => i !== index))
    }
  }

  const validateForm = () => {
    const newErrors = {}

    if (!invoiceNumber.trim()) newErrors.invoiceNumber = 'Numéro requis'
    if (!invoiceDate) newErrors.invoiceDate = 'Date requise'
    if (!seller.name.trim()) newErrors.sellerName = 'Raison sociale requise'
    if (seller.siret && seller.siret.replace(/\s/g, '').length !== 14) {
      newErrors.sellerSiret = 'SIRET invalide (14 chiffres)'
    }
    if (!buyer.name.trim()) newErrors.buyerName = 'Raison sociale requise'

    const hasValidLine = lines.some(l => l.description.trim() && l.quantity > 0)
    if (!hasValidLine) newErrors.lines = 'Au moins une ligne valide requise'

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e) => {
    e.preventDefault()

    if (!validateForm()) return

    setIsGenerating(true)

    const requestData = {
      number: invoiceNumber,
      date: invoiceDate,
      seller: {
        name: seller.name,
        siret: seller.siret.replace(/\s/g, ''),
        vatNumber: seller.vatNumber,
        street: seller.street,
        postalCode: seller.postalCode,
        city: seller.city,
        email: seller.email,
      },
      buyer: {
        name: buyer.name,
        siret: buyer.siret.replace(/\s/g, ''),
        street: buyer.street,
        postalCode: buyer.postalCode,
        city: buyer.city,
        email: buyer.email,
      },
      lines: lines.filter(l => l.description.trim()).map(l => ({
        description: l.description,
        quantity: l.quantity,
        unitPrice: l.unitPrice,
        vatRegime: VAT_REGIMES.find(v => v.id === l.vatRegime)?.code || 0,
      })),
      paymentTerms: {
        dueDate: dueDate || undefined,
        iban: iban || undefined,
        bic: bic || undefined,
        note: note || undefined,
      },
    }

    try {
      const response = await fetch('/api/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(requestData),
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.message || 'Erreur lors de la génération')
      }

      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `facture-${invoiceNumber}.pdf`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Generation error:', error)
      alert(`Erreur: ${error.message}\n\nNote: Le serveur API doit être lancé sur le port 8080.`)
    } finally {
      setIsGenerating(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="max-w-5xl mx-auto px-4 sm:px-6">
      <div className="space-y-6 sm:space-y-8">

        {/* Demo Data Banner */}
        <div className="bg-gradient-to-r from-amber-50 to-orange-50 border border-amber-200 rounded-2xl p-4 opacity-0 animate-fade-up">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-amber-100 flex items-center justify-center shrink-0">
                <svg className="w-5 h-5 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <div>
                <p className="font-medium text-amber-900">Mode démonstration</p>
                <p className="text-sm text-amber-700">Remplissez avec des données d'exemple</p>
              </div>
            </div>
            <button
              type="button"
              onClick={fillDemoData}
              className="w-full sm:w-auto shrink-0 px-5 py-3 bg-amber-500 hover:bg-amber-600 text-white font-semibold rounded-xl shadow-sm hover:shadow transition-all flex items-center justify-center gap-2"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
              Données démo
            </button>
          </div>
        </div>

        {/* Invoice Info */}
        <section className="card p-4 sm:p-8 opacity-0 animate-fade-up stagger-1">
          <SectionHeader number="1" title="Informations de la facture" />

          <div className="grid sm:grid-cols-[repeat(3,minmax(0,1fr))] gap-4 mb-6">
            <InputField
              label="Numéro de facture"
              type="text"
              placeholder="FAC-2026-001"
              value={invoiceNumber}
              onChange={e => setInvoiceNumber(e.target.value)}
              error={errors.invoiceNumber}
            />
            <InputField
              label="Date d'émission"
              type="date"
              className="w-full min-w-0"
              value={invoiceDate}
              onChange={e => setInvoiceDate(e.target.value)}
              error={errors.invoiceDate}
            />
            <InputField
              label="Date d'échéance"
              type="date"
              className="w-full min-w-0"
              value={dueDate}
              onChange={e => setDueDate(e.target.value)}
            />
          </div>

          {/* VAT Exemption Selector */}
          <div className="p-4 bg-marine-50/50 rounded-xl border border-marine-100">
            <label className="input-label mb-3 flex items-center gap-2">
              <svg className="w-4 h-4 text-marine-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2z" />
              </svg>
              Régime de TVA
            </label>
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-2 sm:gap-3">
              {VAT_EXEMPTIONS.map(exemption => (
                <label
                  key={exemption.id}
                  className={`relative flex flex-col p-3 sm:p-4 rounded-lg border-2 cursor-pointer transition-all ${vatExemption === exemption.id
                    ? 'border-tricolore-bleu bg-tricolore-bleu/5'
                    : 'border-marine-200 hover:border-marine-300 bg-white'
                    }`}
                >
                  <input
                    type="radio"
                    name="vatExemption"
                    value={exemption.id}
                    checked={vatExemption === exemption.id}
                    onChange={() => handleVatExemptionChange(exemption.id)}
                    className="sr-only"
                  />
                  <span className={`font-medium text-xs sm:text-sm ${vatExemption === exemption.id ? 'text-tricolore-bleu' : 'text-marine-900'
                    }`}>
                    {exemption.label}
                  </span>
                  <span className="text-xs text-marine-500 mt-1 hidden sm:block">
                    {exemption.description}
                  </span>
                  {vatExemption === exemption.id && (
                    <div className="absolute top-1.5 right-1.5 sm:top-2 sm:right-2">
                      <svg className="w-4 h-4 sm:w-5 sm:h-5 text-tricolore-bleu" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                      </svg>
                    </div>
                  )}
                </label>
              ))}
            </div>
            {vatExemption !== 'none' && (
              <p className="mt-3 text-sm text-marine-600 flex items-center gap-2">
                <svg className="w-4 h-4 text-amber-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                </svg>
                Ce régime sera appliqué à toutes les lignes de la facture.
              </p>
            )}
          </div>
        </section>

        {/* Seller & Buyer */}
        <div className="grid lg:grid-cols-2 gap-8">
          {/* Seller */}
          <section className="card p-4 sm:p-8 opacity-0 animate-fade-up stagger-2">
            <SectionHeader number="2" title="Vendeur" subtitle="Vos informations" />

            <div className="space-y-4">
              <InputField
                label="Raison sociale *"
                type="text"
                placeholder="Ma Société SARL"
                value={seller.name}
                onChange={e => updateSeller('name', e.target.value)}
                error={errors.sellerName}
              />
              <div className="grid grid-cols-2 gap-4">
                <InputField
                  label="SIRET"
                  type="text"
                  placeholder="123 456 789 01234"
                  value={seller.siret}
                  onChange={e => updateSeller('siret', e.target.value)}
                  error={errors.sellerSiret}
                />
                <InputField
                  label="N° TVA"
                  type="text"
                  placeholder="FR12345678901"
                  value={seller.vatNumber}
                  onChange={e => updateSeller('vatNumber', e.target.value)}
                />
              </div>
              <InputField
                label="Adresse"
                type="text"
                placeholder="123 Rue de Paris"
                value={seller.street}
                onChange={e => updateSeller('street', e.target.value)}
              />
              <div className="grid grid-cols-3 gap-4">
                <InputField
                  label="Code postal"
                  type="text"
                  placeholder="75001"
                  value={seller.postalCode}
                  onChange={e => updateSeller('postalCode', e.target.value)}
                />
                <div className="col-span-2">
                  <InputField
                    label="Ville"
                    type="text"
                    placeholder="Paris"
                    value={seller.city}
                    onChange={e => updateSeller('city', e.target.value)}
                  />
                </div>
              </div>
              <InputField
                label="Email"
                type="email"
                placeholder="contact@masociete.fr"
                value={seller.email}
                onChange={e => updateSeller('email', e.target.value)}
              />
            </div>
          </section>

          {/* Buyer */}
          <section className="card p-4 sm:p-8 opacity-0 animate-fade-up stagger-3">
            <SectionHeader number="3" title="Client" subtitle="Destinataire de la facture" />

            <div className="space-y-4">
              <InputField
                label="Raison sociale *"
                type="text"
                placeholder="Client SA"
                value={buyer.name}
                onChange={e => updateBuyer('name', e.target.value)}
                error={errors.buyerName}
              />
              <InputField
                label="SIRET"
                type="text"
                placeholder="987 654 321 09876"
                value={buyer.siret}
                onChange={e => updateBuyer('siret', e.target.value)}
              />
              <InputField
                label="Adresse"
                type="text"
                placeholder="456 Avenue des Champs"
                value={buyer.street}
                onChange={e => updateBuyer('street', e.target.value)}
              />
              <div className="grid grid-cols-3 gap-4">
                <InputField
                  label="Code postal"
                  type="text"
                  placeholder="69001"
                  value={buyer.postalCode}
                  onChange={e => updateBuyer('postalCode', e.target.value)}
                />
                <div className="col-span-2">
                  <InputField
                    label="Ville"
                    type="text"
                    placeholder="Lyon"
                    value={buyer.city}
                    onChange={e => updateBuyer('city', e.target.value)}
                  />
                </div>
              </div>
              <InputField
                label="Email"
                type="email"
                placeholder="achat@client.fr"
                value={buyer.email}
                onChange={e => updateBuyer('email', e.target.value)}
              />
            </div>
          </section>
        </div>

        {/* Invoice Lines */}
        <section className="card p-4 sm:p-8 opacity-0 animate-fade-up stagger-4">
          <SectionHeader number="4" title="Lignes de facturation" subtitle="Produits ou services facturés" />

          {errors.lines && (
            <div className="mb-4 p-3 bg-tricolore-rouge/10 text-tricolore-rouge text-sm rounded-lg">
              {errors.lines}
            </div>
          )}

          <div className="space-y-3 mb-4">
            {lines.map((line, index) => (
              <InvoiceLine
                key={index}
                line={line}
                index={index}
                onChange={updateLine}
                onRemove={removeLine}
                canRemove={lines.length > 1}
                isExempt={vatExemption !== 'none'}
              />
            ))}
          </div>

          <button
            type="button"
            onClick={addLine}
            className="btn-secondary w-full"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            Ajouter une ligne
          </button>
        </section>

        {/* Payment & Totals */}
        <div className="grid lg:grid-cols-3 gap-8">
          <section className="lg:col-span-2 card p-4 sm:p-8 opacity-0 animate-fade-up stagger-5">
            <SectionHeader number="5" title="Paiement" subtitle="Coordonnées bancaires" />

            <div className="grid sm:grid-cols-2 gap-4">
              <InputField
                label="IBAN"
                type="text"
                placeholder="FR76 3000 1007 9412 3456 7890 185"
                value={iban}
                onChange={e => setIban(e.target.value)}
              />
              <InputField
                label="BIC"
                type="text"
                placeholder="BDFEFRPP"
                value={bic}
                onChange={e => setBic(e.target.value)}
              />
            </div>

            <div className="mt-4">
              <label className="input-label">Note / Conditions</label>
              <textarea
                className="input-field min-h-[80px] resize-y"
                placeholder="Conditions de paiement, mentions légales..."
                value={note}
                onChange={e => setNote(e.target.value)}
              />
            </div>
          </section>

          <div className="opacity-0 animate-fade-up stagger-5">
            <TotalsDisplay lines={lines} />

            <button
              type="submit"
              disabled={isGenerating}
              className="btn-primary w-full mt-4 py-4 text-lg"
            >
              {isGenerating ? (
                <>
                  <svg className="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Génération en cours...
                </>
              ) : (
                <>
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  Télécharger la facture
                </>
              )}
            </button>
          </div>
        </div>

      </div>
    </form>
  )
}

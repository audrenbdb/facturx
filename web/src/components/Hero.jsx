import { useEffect, useState } from 'react'

const TricolorBar = () => (
  <div className="flex h-1 w-full">
    <div className="flex-1 bg-tricolore-bleu" />
    <div className="flex-1 bg-white" />
    <div className="flex-1 bg-tricolore-rouge" />
  </div>
)

const FeatureCard = ({ icon, title, description, delay }) => (
  <div
    className={`opacity-0 animate-fade-up stagger-${delay} bg-white/60 backdrop-blur-sm rounded-xl p-6 border border-marine-100 hover:border-marine-200 hover:shadow-lg transition-all duration-300 group`}
  >
    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-tricolore-bleu to-marine-700 flex items-center justify-center text-white text-xl mb-4 group-hover:scale-110 transition-transform duration-300">
      {icon}
    </div>
    <h3 className="font-display text-lg font-semibold text-marine-900 mb-2">{title}</h3>
    <p className="text-marine-600 text-sm leading-relaxed">{description}</p>
  </div>
)

const InfoCard = ({ icon, title, children }) => (
  <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 border border-marine-100">
    <div className="flex items-start gap-4">
      <div className="w-10 h-10 rounded-lg bg-marine-100 flex items-center justify-center text-marine-600 text-lg flex-shrink-0">
        {icon}
      </div>
      <div>
        <h4 className="font-display font-semibold text-marine-900 mb-2">{title}</h4>
        <div className="text-marine-600 text-sm leading-relaxed space-y-2">{children}</div>
      </div>
    </div>
  </div>
)

export default function Hero({ onStart }) {
  const [loaded, setLoaded] = useState(false)

  useEffect(() => {
    setLoaded(true)
  }, [])

  return (
    <div className="min-h-screen flex flex-col">
      <header className="relative overflow-hidden flex-1 flex flex-col">
        {/* Background decorations */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute -top-1/2 -right-1/4 w-[800px] h-[800px] rounded-full bg-gradient-to-br from-tricolore-bleu/5 to-transparent blur-3xl" />
          <div className="absolute -bottom-1/4 -left-1/4 w-[600px] h-[600px] rounded-full bg-gradient-to-tr from-marine-200/30 to-transparent blur-3xl" />

          {/* Geometric accents */}
          <svg className="absolute top-20 right-10 w-32 h-32 text-marine-200/40 rotate-12" viewBox="0 0 100 100">
            <rect x="10" y="10" width="80" height="80" fill="none" stroke="currentColor" strokeWidth="1" />
            <rect x="25" y="25" width="50" height="50" fill="none" stroke="currentColor" strokeWidth="1" />
          </svg>

          <svg className="absolute bottom-40 left-10 w-24 h-24 text-marine-200/30 -rotate-6" viewBox="0 0 100 100">
            <circle cx="50" cy="50" r="40" fill="none" stroke="currentColor" strokeWidth="1" />
            <circle cx="50" cy="50" r="25" fill="none" stroke="currentColor" strokeWidth="1" />
          </svg>
        </div>

        <TricolorBar />

        {/* Navigation */}
        <nav className="relative z-10 px-6 py-4">
          <div className="max-w-6xl mx-auto flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-tricolore-bleu flex items-center justify-center">
                <span className="font-display text-white font-bold text-lg">Fx</span>
              </div>
              <span className="font-display text-xl font-semibold text-marine-900">Factur-X</span>
            </div>

            <div className="flex items-center gap-6 text-sm font-medium text-marine-600">
              <a href="#features" className="hover:text-tricolore-bleu transition-colors">Fonctionnalités</a>
              <a href="#about" className="hover:text-tricolore-bleu transition-colors">À propos</a>
              <a href="https://github.com/audrenbdb/facturx" target="_blank" rel="noopener" className="hover:text-tricolore-bleu transition-colors flex items-center gap-1">
                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" /></svg>
                GitHub
              </a>
            </div>
          </div>
        </nav>

        {/* Hero content */}
        <div className="relative z-10 flex-1 flex flex-col justify-center px-6 py-20">
          <div className="max-w-6xl mx-auto w-full">
            {/* Badge */}
            <div className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-tricolore-bleu/10 text-tricolore-bleu text-sm font-medium mb-6 ${loaded ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'} transition-all duration-500`}>
              <span className="w-2 h-2 rounded-full bg-tricolore-rouge animate-pulse" />
              Réforme 2026
            </div>

            {/* Title */}
            <h1 className={`font-display text-4xl sm:text-5xl lg:text-6xl font-bold text-marine-900 leading-tight mb-6 ${loaded ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'} transition-all duration-500 delay-100`}>
              Créez vos factures
              <br />
              <span className="bg-gradient-to-r from-tricolore-bleu to-marine-600 bg-clip-text text-transparent">
                conformes Factur-X
              </span>
            </h1>

            {/* Description */}
            <p className={`text-lg sm:text-xl text-marine-600 max-w-2xl mb-8 leading-relaxed ${loaded ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'} transition-all duration-500 delay-200`}>
              Préparez-vous à la réforme de la facturation électronique.
              Générez des factures PDF/A-3 avec XML embarqué, conformes au standard européen.
            </p>

            {/* CTA */}
            <div className={`flex flex-wrap gap-4 ${loaded ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'} transition-all duration-500 delay-300`}>
              <button onClick={onStart} className="btn-primary text-lg px-8 py-4 group">
                Tester l'interface
                <svg className="w-5 h-5 group-hover:translate-x-1 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                </svg>
              </button>
              <a href="https://github.com/audrenbdb/facturx" target="_blank" rel="noopener" className="btn-secondary text-lg px-8 py-4">
                Voir la librairie
              </a>
            </div>
          </div>
        </div>

        {/* Scroll indicator */}
        <div className="absolute bottom-8 left-1/2 -translate-x-1/2 animate-bounce">
          <svg className="w-6 h-6 text-marine-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
          </svg>
        </div>
      </header>

      {/* Features section */}
      <section id="features" className="bg-gradient-to-b from-papier-50 to-white px-6 py-24">
        <div className="max-w-6xl mx-auto">
          <h2 className="font-display text-3xl font-bold text-marine-900 mb-4 text-center">Fonctionnalités</h2>
          <p className="text-marine-600 text-center mb-12 max-w-2xl mx-auto">
            Une librairie conçue pour générer des factures conformes rapidement et simplement.
          </p>

          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
            <FeatureCard
              icon="📄"
              title="PDF/A-3 Hybride"
              description="Factures lisibles par l'humain et la machine. Le XML CII est embarqué directement dans le PDF."
              delay={1}
            />
            <FeatureCard
              icon="✓"
              title="Profil BASIC"
              description="Conforme au profil BASIC de Factur-X, suffisant pour la majorité des entreprises françaises."
              delay={2}
            />
            <FeatureCard
              icon="🔢"
              title="Validation SIRET"
              description="Vérification automatique des numéros SIRET et SIREN avec l'algorithme de Luhn."
              delay={3}
            />
            <FeatureCard
              icon="💶"
              title="TVA Automatique"
              description="Tous les régimes de TVA français : standard, réduit, auto-entrepreneur, exonérations."
              delay={4}
            />
            <FeatureCard
              icon="🇫🇷"
              title="Réforme 2026"
              description="Anticipez l'obligation de facturation électronique qui entre en vigueur en France."
              delay={5}
            />
            <FeatureCard
              icon="🔒"
              title="Données Sécurisées"
              description="Vos données restent sur votre machine. Aucune information n'est stockée sur nos serveurs."
              delay={5}
            />
          </div>
        </div>
      </section>

      {/* About section */}
      <section id="about" className="bg-white px-6 py-24 border-t border-marine-100">
        <div className="max-w-6xl mx-auto">
          <h2 className="font-display text-3xl font-bold text-marine-900 mb-4 text-center">À propos de la librairie</h2>
          <p className="text-marine-600 text-center mb-12 max-w-2xl mx-auto">
            Une solution open source pour générer des factures Factur-X conformes.
          </p>

          <div className="grid md:grid-cols-2 gap-6 mb-12">
            <InfoCard icon="🛠" title="Pure Go, zéro dépendance">
              <p>
                La librairie est écrite entièrement en Go sans aucune dépendance externe.
                Le PDF est généré octet par octet, la police TrueType est parsée manuellement,
                et le XML CII est construit à la main.
              </p>
              <p>
                Résultat : une librairie légère, rapide, et sans surprises de compatibilité.
              </p>
            </InfoCard>

            <InfoCard icon="✅" title="Validation officielle FNFE-MPE">
              <p>
                Les factures générées ont été validées par le{' '}
                <a href="https://services.fnfe-mpe.org" target="_blank" rel="noopener" className="text-tricolore-bleu hover:underline">
                  service de validation FNFE-MPE
                </a>
                {' '}:
              </p>
              <ul className="list-disc list-inside text-sm space-y-1 mt-2">
                <li>Validation XMD (métadonnées)</li>
                <li>Validation XML contre XSD</li>
                <li>Validation Schematron</li>
                <li>Conformité PDF/A-3</li>
              </ul>
            </InfoCard>

            <InfoCard icon="📐" title="Simplicité, pas personnalisation">
              <p>
                Cette librairie ne permet <strong>pas</strong> de personnaliser le style des factures.
                Ce n'est pas son objectif.
              </p>
              <p>
                Le but est de proposer un moyen <strong>simple et rapide</strong> de générer des factures
                conformes au standard Factur-X, sans configuration complexe.
              </p>
            </InfoCard>

            <InfoCard icon="📜" title="Open Source MIT">
              <p>
                Le code source est disponible sous licence MIT sur GitHub.
                Vous pouvez l'utiliser, le modifier et le distribuer librement,
                y compris dans des projets commerciaux.
              </p>
              <a
                href="https://github.com/audrenbdb/facturx"
                target="_blank"
                rel="noopener"
                className="inline-flex items-center gap-2 text-tricolore-bleu hover:underline mt-2"
              >
                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" /></svg>
                Voir sur GitHub
              </a>
            </InfoCard>
          </div>

          {/* Warning box */}
          <div className="bg-amber-50 border border-amber-200 rounded-xl p-6 max-w-3xl mx-auto">
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 rounded-lg bg-amber-100 flex items-center justify-center text-amber-600 text-lg flex-shrink-0">
                ⚠️
              </div>
              <div>
                <h4 className="font-display font-semibold text-amber-900 mb-2">
                  Cette interface web est pour les tests uniquement
                </h4>
                <p className="text-amber-800 text-sm leading-relaxed mb-3">
                  L'interface web est fournie pour tester rapidement la librairie.
                  Elle n'est <strong>pas conçue pour un usage en production</strong>.
                </p>
                <ul className="text-amber-700 text-sm space-y-1">
                  <li>• Pas d'API key disponible</li>
                  <li>• Rate limit très restrictif (10 factures/heure par IP)</li>
                  <li>• Aucune garantie de disponibilité</li>
                </ul>
                <p className="text-amber-800 text-sm mt-3">
                  Pour un usage en production, <strong>intégrez la librairie Go directement</strong> dans votre application.
                </p>
              </div>
            </div>
          </div>

          {/* Code example */}
          <div className="mt-12 max-w-3xl mx-auto">
            <h3 className="font-display text-xl font-semibold text-marine-900 mb-4 text-center">
              Utilisation de la librairie
            </h3>
            <div className="bg-marine-900 rounded-xl p-6 overflow-x-auto">
              <pre className="text-sm text-marine-100 font-mono">
                {`go get github.com/audrenbdb/facturx

// Générer une facture
req := facturx.InvoiceRequest{
    Number: "FAC-2026-001",
    Date:   "20260131",
    Seller: facturx.Contact{
        Name:    "Mon Entreprise",
        Siret:   "12345678901234",
        Address: "1 rue de Paris",
        ZipCode: "75001",
        City:    "Paris",
    },
    // ...
}

pdf, err := facturx.Generate(req)
os.WriteFile("facture.pdf", pdf, 0644)`}
              </pre>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}

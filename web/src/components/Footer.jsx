export default function Footer() {
  return (
    <footer className="mt-auto border-t border-marine-200 bg-white/50 backdrop-blur-sm">
      <div className="max-w-6xl mx-auto px-6 py-8">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-tricolore-bleu flex items-center justify-center">
              <span className="font-display text-white font-bold text-sm">Fx</span>
            </div>
            <span className="text-marine-600 text-sm">
              Générateur de factures Factur-X
            </span>
          </div>

          <div className="flex items-center gap-6 text-sm text-marine-500">
            <a
              href="https://fnfe-mpe.org/factur-x/"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-tricolore-bleu transition-colors"
            >
              Documentation Factur-X
            </a>
            <a
              href="https://github.com/audren/facturx"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-tricolore-bleu transition-colors flex items-center gap-1"
            >
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.87 8.17 6.84 9.5.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34-.46-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.87 1.52 2.34 1.07 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.92 0-1.11.38-2 1.03-2.71-.1-.25-.45-1.29.1-2.64 0 0 .84-.27 2.75 1.02.79-.22 1.65-.33 2.5-.33.85 0 1.71.11 2.5.33 1.91-1.29 2.75-1.02 2.75-1.02.55 1.35.2 2.39.1 2.64.65.71 1.03 1.6 1.03 2.71 0 3.82-2.34 4.66-4.57 4.91.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0012 2z" />
              </svg>
              GitHub
            </a>
          </div>
        </div>

        <div className="mt-6 pt-6 border-t border-marine-100 text-center text-xs text-marine-400">
          <p>
            Conforme au profil BASIC de Factur-X v1.0 · Standard européen EN 16931
          </p>
          <p className="mt-1">
            Vos données restent sur votre appareil. Aucune information n'est stockée.
          </p>
        </div>
      </div>

      {/* Tricolor bar at bottom */}
      <div className="flex h-1 w-full">
        <div className="flex-1 bg-tricolore-bleu" />
        <div className="flex-1 bg-white" />
        <div className="flex-1 bg-tricolore-rouge" />
      </div>
    </footer>
  )
}

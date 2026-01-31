import { useState, useEffect } from 'react'
import Hero from './components/Hero'
import InvoiceForm from './components/InvoiceForm'
import Footer from './components/Footer'

function App() {
  const [currentPage, setCurrentPage] = useState('home')

  // Handle browser back/forward navigation
  useEffect(() => {
    const handlePopState = () => {
      setCurrentPage('home')
    }
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  const goToForm = () => {
    window.history.pushState({ page: 'form' }, '', '/form')
    setCurrentPage('form')
  }

  const goToHome = () => {
    window.history.back()
  }

  return (
    <div className="min-h-screen flex flex-col">
      {currentPage === 'home' && (
        <Hero onStart={goToForm} />
      )}

      {currentPage === 'form' && (
        <>
          <FormHeader onBack={goToHome} />
          <main className="flex-1 pt-4 sm:pt-6 pb-12 sm:pb-20 bg-gradient-to-b from-papier-50 to-white">
            <InvoiceForm />
          </main>
        </>
      )}

      <Footer />
    </div>
  )
}

function FormHeader({ onBack }) {
  return (
    <header className="bg-white border-b border-marine-100 sticky top-0 z-50">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 py-3 sm:py-4 flex items-center justify-between">
        <button
          onClick={onBack}
          className="flex items-center gap-1.5 sm:gap-2 text-marine-600 hover:text-tricolore-bleu transition-colors"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          <span className="font-medium hidden sm:inline">Retour</span>
        </button>

        <div className="flex items-center gap-2 sm:gap-3">
          <div className="w-7 h-7 sm:w-8 sm:h-8 rounded-lg bg-tricolore-bleu flex items-center justify-center">
            <span className="font-display text-white font-bold text-xs sm:text-sm">Fx</span>
          </div>
          <span className="font-display text-base sm:text-lg font-semibold text-marine-900">Nouvelle facture</span>
        </div>

        <div className="w-8 sm:w-24" /> {/* Spacer for centering */}
      </div>
    </header>
  )
}

export default App

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider, MutationCache } from '@tanstack/react-query'
import App from './App'
import './globals.css'
import { ToastProvider, useToast } from './components/toast/ToastContext'
import ToastContainer from './components/toast/ToastContainer'
import DeployEventToasts from './components/toast/DeployEventToasts'
import ErrorBoundary from './components/ErrorBoundary'

function AppWithProviders() {
  const toast = useToast()

  // Query client is recreated when toast handler is available so mutation errors become toasts.
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnWindowFocus: false,
        retry: 1,
        staleTime: 10_000,
      },
    },
    mutationCache: new MutationCache({
      onError: (error) => {
        const err = error as { response?: { data?: { error?: string } }; message?: string }
        const message =
          err?.response?.data?.error ?? err?.message ?? 'An unexpected error occurred'
        toast.push({
          kind: 'error',
          title: 'Action failed',
          message,
          duration: 6000,
        })
      },
    }),
  })

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <DeployEventToasts />
        <App />
      </BrowserRouter>
    </QueryClientProvider>
  )
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary>
      <ToastProvider>
        <AppWithProviders />
        <ToastContainer />
      </ToastProvider>
    </ErrorBoundary>
  </StrictMode>,
)

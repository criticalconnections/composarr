import { useToast } from './ToastContext'
import type { ToastKind } from './ToastContext'

const kindStyles: Record<ToastKind, { border: string; icon: string; accent: string }> = {
  success: { border: 'var(--color-success)', icon: '✓', accent: 'var(--color-success)' },
  error: { border: 'var(--color-danger)', icon: '✕', accent: 'var(--color-danger)' },
  warning: { border: 'var(--color-warning)', icon: '!', accent: 'var(--color-warning)' },
  info: { border: 'var(--color-primary)', icon: 'i', accent: 'var(--color-primary)' },
}

export default function ToastContainer() {
  const { toasts, dismiss } = useToast()

  return (
    <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 max-w-sm pointer-events-none">
      {toasts.map((toast) => {
        const style = kindStyles[toast.kind]
        return (
          <div
            key={toast.id}
            className="pointer-events-auto bg-[var(--color-surface)] rounded-lg border-l-4 shadow-xl p-3 pr-8 relative animate-slide-in-right"
            style={{ borderLeftColor: style.border }}
          >
            <button
              onClick={() => dismiss(toast.id)}
              className="absolute top-2 right-2 text-[var(--color-text-muted)] hover:text-[var(--color-text)] text-xs"
              aria-label="Dismiss"
            >
              ✕
            </button>
            <div className="flex gap-3">
              <span
                className="flex-shrink-0 w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold text-white"
                style={{ backgroundColor: style.accent }}
              >
                {style.icon}
              </span>
              <div className="min-w-0 flex-1">
                <p className="font-medium text-sm text-[var(--color-text)]">{toast.title}</p>
                {toast.message && (
                  <p className="text-xs text-[var(--color-text-muted)] mt-0.5 break-words">
                    {toast.message}
                  </p>
                )}
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}

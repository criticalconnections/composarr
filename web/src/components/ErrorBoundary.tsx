import { Component } from 'react'
import type { ErrorInfo, ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

export default class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('ErrorBoundary caught', error, info)
  }

  handleReset = () => {
    this.setState({ error: null })
  }

  render() {
    if (this.state.error) {
      return (
        <div className="min-h-screen flex items-center justify-center p-6">
          <div className="max-w-lg w-full bg-[var(--color-surface)] rounded-xl border border-[var(--color-danger)] p-6">
            <h1 className="text-lg font-semibold text-[var(--color-danger)] mb-2">
              Something went wrong
            </h1>
            <p className="text-sm text-[var(--color-text-muted)] mb-4">
              An unexpected error crashed the UI. This is most likely a bug.
            </p>

            <pre className="bg-[var(--color-bg)] rounded p-3 text-xs text-[var(--color-text-muted)] overflow-auto max-h-48 mb-4 font-mono">
              {this.state.error.message}
              {this.state.error.stack && '\n\n' + this.state.error.stack}
            </pre>

            <div className="flex gap-3">
              <button
                onClick={this.handleReset}
                className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm hover:bg-[var(--color-primary-hover)]"
              >
                Try again
              </button>
              <button
                onClick={() => window.location.reload()}
                className="px-4 py-2 bg-[var(--color-surface-hover)] text-[var(--color-text)] rounded-lg text-sm hover:bg-[var(--color-border)]"
              >
                Reload page
              </button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}

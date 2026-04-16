import { useDependencyGraph } from '../hooks/use-dependencies'
import DependencyGraphView from '../components/dependencies/DependencyGraph'

export default function DependencyGraphPage() {
  const { data: graph, isLoading } = useDependencyGraph()

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Dependency Graph</h1>
        <p className="text-sm text-[var(--color-text-muted)] mt-1">
          Stack dependencies flow left → right. Dependencies must be running before their dependents can deploy.
        </p>
      </div>

      {isLoading ? (
        <p className="text-[var(--color-text-muted)]">Loading graph...</p>
      ) : !graph?.nodes.length ? (
        <div className="bg-[var(--color-surface)] rounded-lg p-12 text-center border border-[var(--color-border)]">
          <p className="text-[var(--color-text-muted)]">No stacks yet</p>
        </div>
      ) : !graph.edges.length ? (
        <div className="bg-[var(--color-surface)] rounded-lg p-6 border border-[var(--color-border)]">
          <p className="text-[var(--color-text-muted)] text-sm mb-4">
            No dependencies configured. Add dependencies from a stack's detail page to see the graph.
          </p>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
            {graph.nodes.map((n) => (
              <div
                key={n.stack.id}
                className="bg-[var(--color-bg)] rounded px-3 py-2 border border-[var(--color-border)] text-sm"
              >
                {n.stack.name}
              </div>
            ))}
          </div>
        </div>
      ) : (
        <>
          <DependencyGraphView graph={graph} />

          <div className="mt-4 text-xs text-[var(--color-text-muted)] flex gap-4">
            <span className="flex items-center gap-2">
              <svg width="20" height="2">
                <line x1="0" y1="1" x2="20" y2="1" stroke="var(--color-primary)" strokeWidth="2" />
              </svg>
              Hard (required)
            </span>
            <span className="flex items-center gap-2">
              <svg width="20" height="2">
                <line x1="0" y1="1" x2="20" y2="1" stroke="var(--color-text-muted)" strokeWidth="2" strokeDasharray="4 4" />
              </svg>
              Soft (preferred)
            </span>
          </div>
        </>
      )}
    </div>
  )
}

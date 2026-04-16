import { Link } from 'react-router-dom'
import type { DependencyGraph as GraphData, DependencyNode } from '../../types/dependency'

interface Props {
  graph: GraphData
}

interface LayoutedNode extends DependencyNode {
  level: number
  x: number
  y: number
}

const NODE_WIDTH = 180
const NODE_HEIGHT = 60
const COLUMN_GAP = 120
const ROW_GAP = 20
const PADDING = 20

/**
 * A simple SVG-based layered graph visualization.
 * Layers are computed from dependency depth (leaves = depth 0).
 * Within a layer, nodes are stacked vertically.
 */
export default function DependencyGraphView({ graph }: Props) {
  const layouted = layoutGraph(graph)
  const width =
    (maxLevel(layouted) + 1) * (NODE_WIDTH + COLUMN_GAP) - COLUMN_GAP + PADDING * 2
  const height = Math.max(...layouted.map((n) => n.y + NODE_HEIGHT)) + PADDING

  const nodeById = new Map(layouted.map((n) => [n.stack.id, n]))

  return (
    <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] overflow-auto">
      <svg width={Math.max(width, 600)} height={Math.max(height, 300)}>
        {/* Edges */}
        {graph.edges.map((edge) => {
          const from = nodeById.get(edge.dependsOnId)
          const to = nodeById.get(edge.stackId)
          if (!from || !to) return null

          // Edge goes from "depends_on" node (source) to "stack" node (target)
          const x1 = from.x + NODE_WIDTH
          const y1 = from.y + NODE_HEIGHT / 2
          const x2 = to.x
          const y2 = to.y + NODE_HEIGHT / 2

          const midX = (x1 + x2) / 2
          const path = `M ${x1},${y1} C ${midX},${y1} ${midX},${y2} ${x2},${y2}`

          return (
            <g key={edge.id}>
              <path
                d={path}
                stroke={edge.dependencyType === 'hard' ? 'var(--color-primary)' : 'var(--color-text-muted)'}
                strokeWidth={2}
                strokeDasharray={edge.dependencyType === 'soft' ? '4 4' : 'none'}
                fill="none"
                opacity={0.7}
              />
              <polygon
                points={`${x2},${y2} ${x2 - 6},${y2 - 4} ${x2 - 6},${y2 + 4}`}
                fill={edge.dependencyType === 'hard' ? 'var(--color-primary)' : 'var(--color-text-muted)'}
              />
            </g>
          )
        })}

        {/* Nodes */}
        {layouted.map((node) => (
          <g key={node.stack.id}>
            <Link to={`/stacks/${node.stack.id}`}>
              <rect
                x={node.x}
                y={node.y}
                width={NODE_WIDTH}
                height={NODE_HEIGHT}
                rx={8}
                fill="var(--color-bg)"
                stroke={statusStroke(node.stack.status)}
                strokeWidth={2}
                className="hover:opacity-80 transition-opacity cursor-pointer"
              />
              <text
                x={node.x + 12}
                y={node.y + 24}
                fill="var(--color-text)"
                fontSize={13}
                fontWeight={500}
                style={{ pointerEvents: 'none' }}
              >
                {truncate(node.stack.name, 20)}
              </text>
              <text
                x={node.x + 12}
                y={node.y + 44}
                fill={statusStroke(node.stack.status)}
                fontSize={11}
                style={{ pointerEvents: 'none' }}
              >
                {node.stack.status}
              </text>
            </Link>
          </g>
        ))}
      </svg>
    </div>
  )
}

function layoutGraph(graph: GraphData): LayoutedNode[] {
  const levels = new Map<string, number>()

  // BFS-like level assignment: level = longest path from a node with no deps
  const assign = (stackId: string, visited: Set<string>): number => {
    if (visited.has(stackId)) return 0 // cycle guard (shouldn't happen)
    visited.add(stackId)
    if (levels.has(stackId)) return levels.get(stackId)!

    const node = graph.nodes.find((n) => n.stack.id === stackId)
    if (!node || !node.dependencies.length) {
      levels.set(stackId, 0)
      return 0
    }

    let max = 0
    for (const dep of node.dependencies) {
      const lvl = assign(dep.dependsOnId, new Set(visited)) + 1
      if (lvl > max) max = lvl
    }
    levels.set(stackId, max)
    return max
  }

  for (const node of graph.nodes) {
    assign(node.stack.id, new Set())
  }

  // Bucket by level, then layout vertically within each column
  const buckets = new Map<number, DependencyNode[]>()
  for (const node of graph.nodes) {
    const lvl = levels.get(node.stack.id) ?? 0
    if (!buckets.has(lvl)) buckets.set(lvl, [])
    buckets.get(lvl)!.push(node)
  }

  const result: LayoutedNode[] = []
  for (const [level, nodes] of buckets.entries()) {
    nodes.forEach((node, idx) => {
      result.push({
        ...node,
        level,
        x: PADDING + level * (NODE_WIDTH + COLUMN_GAP),
        y: PADDING + idx * (NODE_HEIGHT + ROW_GAP),
      })
    })
  }

  return result
}

function maxLevel(nodes: LayoutedNode[]): number {
  return nodes.reduce((max, n) => Math.max(max, n.level), 0)
}

function statusStroke(status: string): string {
  switch (status) {
    case 'running':
      return 'var(--color-success)'
    case 'stopped':
      return 'var(--color-text-muted)'
    case 'degraded':
      return 'var(--color-warning)'
    default:
      return 'var(--color-border)'
  }
}

function truncate(s: string, n: number): string {
  return s.length > n ? s.slice(0, n - 1) + '…' : s
}

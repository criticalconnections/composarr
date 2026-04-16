import type { Stack } from './stack'

export interface StackDependency {
  id: string
  stackId: string
  dependsOnId: string
  dependencyType: 'hard' | 'soft'
  createdAt: string
}

export interface DependencyNode {
  stack: Stack
  dependencies: StackDependency[]
  dependents: StackDependency[]
}

export interface DependencyGraph {
  nodes: DependencyNode[]
  edges: StackDependency[]
}

export interface CommitInfo {
  hash: string
  shortHash: string
  message: string
  author: string
  email: string
  timestamp: string
}

export interface VersionDetail {
  commit: CommitInfo
  content: string
}

export interface StructuredDiff {
  oldHash: string
  newHash: string
  oldContent: string
  newContent: string
}

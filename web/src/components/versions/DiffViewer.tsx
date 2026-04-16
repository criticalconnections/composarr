import ReactDiffViewer, { DiffMethod } from 'react-diff-viewer-continued'

interface Props {
  oldContent: string
  newContent: string
  oldTitle?: string
  newTitle?: string
  splitView?: boolean
}

export default function DiffViewer({
  oldContent,
  newContent,
  oldTitle = 'Previous',
  newTitle = 'Current',
  splitView = true,
}: Props) {
  return (
    <div className="rounded-lg overflow-hidden border border-[var(--color-border)] text-sm">
      <ReactDiffViewer
        oldValue={oldContent}
        newValue={newContent}
        splitView={splitView}
        compareMethod={DiffMethod.WORDS}
        leftTitle={oldTitle}
        rightTitle={newTitle}
        useDarkTheme
        styles={{
          variables: {
            dark: {
              diffViewerBackground: '#1e293b',
              diffViewerColor: '#f1f5f9',
              addedBackground: 'rgba(34, 197, 94, 0.15)',
              addedColor: '#86efac',
              removedBackground: 'rgba(239, 68, 68, 0.15)',
              removedColor: '#fca5a5',
              wordAddedBackground: 'rgba(34, 197, 94, 0.3)',
              wordRemovedBackground: 'rgba(239, 68, 68, 0.3)',
              addedGutterBackground: 'rgba(34, 197, 94, 0.2)',
              removedGutterBackground: 'rgba(239, 68, 68, 0.2)',
              gutterBackground: '#0f172a',
              gutterBackgroundDark: '#0f172a',
              highlightBackground: '#334155',
              highlightGutterBackground: '#334155',
              codeFoldGutterBackground: '#334155',
              codeFoldBackground: '#334155',
              emptyLineBackground: '#0f172a',
              gutterColor: '#94a3b8',
              addedGutterColor: '#86efac',
              removedGutterColor: '#fca5a5',
              codeFoldContentColor: '#94a3b8',
              diffViewerTitleBackground: '#334155',
              diffViewerTitleColor: '#f1f5f9',
              diffViewerTitleBorderColor: '#475569',
            },
          },
        }}
      />
    </div>
  )
}

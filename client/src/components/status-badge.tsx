import { CircleDot, CircleCheck, CircleX, Clock3, CircleHelp } from 'lucide-react'
import type { PayloadState } from '../lib/types'
import { stateClass, titleCase } from '../lib/format'

export function StatusBadge({ state }: { state: PayloadState }) {
  const icon = state === 'completed'
    ? <CircleCheck size={12} />
    : state === 'failed'
      ? <CircleX size={12} />
      : state === 'queued'
        ? <Clock3 size={12} />
        : state === 'uncertain'
          ? <CircleHelp size={12} />
          : state === 'cancelled'
            ? <CircleX size={12} />
            : <CircleDot size={12} />

  return <span className={stateClass(state)}>{icon}{titleCase(state)}</span>
}

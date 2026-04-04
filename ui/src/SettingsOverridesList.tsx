import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { X, ArrowRight } from 'lucide-react'
import type { Engineer, Override } from './types'
import { computeOverrideReplacements, formatOverrideRange, formatSegmentRange, initials } from './utils'

type SettingsOverridesListProps = {
  engineers: Engineer[]
  overrides: Override[]
  setOverrides: (overrides: Override[]) => void
}

function SettingsOverridesList({ engineers, overrides, setOverrides }: SettingsOverridesListProps) {
  function removeOverride(id: string) {
    setOverrides(overrides.filter(o => o.id !== id))
  }

  return (
    <div className="space-y-1">
      {overrides.map(override => {
        const engineer = engineers.find(e => e.id === override.engineerId)
        if (!engineer) return null
        const baseOverrides = overrides.filter(o => o.id !== override.id)
        const replacements = computeOverrideReplacements(engineers, baseOverrides, override.start, override.end)
        return (
          <div key={override.id} className="rounded-xl bg-muted/40 overflow-hidden">
            <div className="flex items-center gap-3 px-3 py-2.5">
              <div className="flex-1 min-w-0 flex items-center gap-2 text-sm">
                <span className="text-muted-foreground shrink-0 tabular-nums">
                  {formatOverrideRange(override.start, override.end)}
                </span>
                <ArrowRight className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                <Avatar className="h-6 w-6 shrink-0">
                  <AvatarImage src={engineer.avatarUrl} />
                  <AvatarFallback className={`text-[10px] font-semibold ${engineer.color} ${engineer.textColor}`}>
                    {initials(engineer.name)}
                  </AvatarFallback>
                </Avatar>
                <span className="font-medium truncate">{engineer.name}</span>
              </div>
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => removeOverride(override.id)}
                className="shrink-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                aria-label="Remove override"
              >
                <X />
              </Button>
            </div>
            {replacements.length > 0 && (
              <div className="px-3 pb-2.5 flex flex-wrap items-center gap-x-3 gap-y-1">
                <span className="text-xs text-muted-foreground">Replaces:</span>
                {replacements.map((seg, i) => (
                  <span key={i} className="flex items-center gap-1 text-xs text-muted-foreground">
                    <Avatar className="h-4 w-4 shrink-0">
                      <AvatarImage src={seg.engineer.avatarUrl} />
                      <AvatarFallback className={`text-[8px] font-semibold ${seg.engineer.color} ${seg.engineer.textColor}`}>
                        {initials(seg.engineer.name)}
                      </AvatarFallback>
                    </Avatar>
                    <span className="font-medium text-foreground">{seg.engineer.name}</span>
                    {replacements.length > 1 && (
                      <span>({formatSegmentRange(seg.start, seg.end)})</span>
                    )}
                  </span>
                ))}
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}

export default SettingsOverridesList
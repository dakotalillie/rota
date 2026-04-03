import { useRef } from "react"
import { GripVertical, X } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "./components/ui/card"
import type { Engineer } from "./types"
import { Avatar, AvatarFallback, AvatarImage } from "./components/ui/avatar"
import { Button } from "./components/ui/button"
import { initials } from "./utils"

type SettingsRotationOrderProps = {
  engineers: Engineer[]
  setEngineers: (engineers: Engineer[]) => void
  removeEngineer: (id: string) => void
}

function SettingsRotationOrder({ engineers, setEngineers, removeEngineer }: SettingsRotationOrderProps) {
    const dragIndexRef = useRef<number | null>(null)

    function handleDragStart(index: number) {
        dragIndexRef.current = index
    }

    function handleDragOver(e: React.DragEvent, index: number) {
        e.preventDefault()
        const from = dragIndexRef.current
        if (from === null || from === index) return
        const next = [...engineers]
        const [item] = next.splice(from, 1)
        next.splice(index, 0, item)
        dragIndexRef.current = index
        setEngineers(next)
    }

    function handleDragEnd() {
        dragIndexRef.current = null
    }

    return (
      <Card className="shadow-sm border-border bg-card">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold">Rotation order</CardTitle>
        </CardHeader>
        <CardContent className="space-y-1">
          {engineers.length > 0 ? engineers.map((engineer, index) => (
            <div
              key={engineer.id}
              draggable
              onDragStart={() => handleDragStart(index)}
              onDragOver={(e) => handleDragOver(e, index)}
              onDragEnd={handleDragEnd}
              className="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-colors bg-muted/40 hover:bg-muted/60 cursor-grab active:cursor-grabbing select-none"
            >
              <GripVertical className="h-4 w-4 text-muted-foreground shrink-0" />
              <Avatar className="h-8 w-8 shrink-0">
                <AvatarImage src={engineer.avatarUrl} />
                <AvatarFallback className={`text-xs font-semibold ${engineer.color} ${engineer.textColor}`}>
                  {initials(engineer.name)}
                </AvatarFallback>
              </Avatar>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{engineer.name}</p>
                {engineer.email && (
                  <p className="text-xs text-muted-foreground truncate">{engineer.email}</p>
                )}
              </div>
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => removeEngineer(engineer.id)}
                className="shrink-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                aria-label={`Remove ${engineer.name}`}
              >
                <X />
              </Button>
            </div>
          )) : (
            <p className="text-sm text-muted-foreground px-1 py-1">No engineers yet.</p>
          )}
        </CardContent>
      </Card>
    )
}

export default SettingsRotationOrder
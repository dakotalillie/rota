import { ArrowLeft } from "lucide-react"
import { Button } from "@/components/ui/button"

type SettingsHeaderProps = {
  onNavigateHome: () => void
}

function SettingsHeader({ onNavigateHome }: SettingsHeaderProps) {
    return (
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon-sm" onClick={onNavigateHome} aria-label="Back to home">
          <ArrowLeft />
        </Button>
        <h1 className="text-xl font-bold tracking-tight">Settings</h1>
      </div>
    )
}

export default SettingsHeader
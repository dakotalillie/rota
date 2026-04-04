import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Engineer, Override } from './types'
import SettingsOverridesList from './SettingsOverridesList'
import SettingsOverridesForm from './SettingsOverridesForm'

type SettingsOverridesProps = {
  engineers: Engineer[]
  overrides: Override[]
  setOverrides: (overrides: Override[]) => void
}

function SettingsOverrides({ engineers, overrides, setOverrides }: SettingsOverridesProps) {
  return (
    <Card className="shadow-sm border-border bg-card">
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-semibold">Schedule overrides</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {overrides.length > 0 && <SettingsOverridesList engineers={engineers} overrides={overrides} setOverrides={setOverrides} />}
        {engineers.length === 0 ? (
          <p className="text-sm text-muted-foreground">Add engineers to the rotation before creating overrides.</p>
        ) : <SettingsOverridesForm engineers={engineers} overrides={overrides} setOverrides={setOverrides} />}
      </CardContent>
    </Card>
  )
}

export default SettingsOverrides
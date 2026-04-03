import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Plus, X, Webhook } from 'lucide-react'
import type { Engineer, Override, WebhookEntry } from './types'
import { inputClass } from './utils'
import SettingsRotationOrder from './SettingsRotationOrder'
import SettingsAddPerson from './SettingsAddPerson'
import SettingsHeader from './SettingsHeader'
import SettingsOverrides from './SettingsOverrides'

function SettingsPage({ engineers, setEngineers, overrides, setOverrides, webhooks, setWebhooks, onNavigateHome }: {
  engineers: Engineer[]
  setEngineers: (engineers: Engineer[]) => void
  overrides: Override[]
  setOverrides: (overrides: Override[]) => void
  webhooks: WebhookEntry[]
  setWebhooks: (webhooks: WebhookEntry[]) => void
  onNavigateHome: () => void
}) {
  const [webhookUrl, setWebhookUrl] = useState('')
  const [webhookLabel, setWebhookLabel] = useState('')

  const webhookUrlValid = (() => {
    try { new URL(webhookUrl); return true } catch { return false }
  })()

  function addWebhook() {
    if (!webhookUrlValid) return
    setWebhooks([...webhooks, { id: crypto.randomUUID(), url: webhookUrl.trim(), label: webhookLabel.trim() }])
    setWebhookUrl('')
    setWebhookLabel('')
  }

  function removeWebhook(id: string) {
    setWebhooks(webhooks.filter(w => w.id !== id))
  }

  function handleWebhookKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') addWebhook()
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      <SettingsHeader onNavigateHome={onNavigateHome} />
      <SettingsRotationOrder
        engineers={engineers}
        setEngineers={setEngineers}
        overrides={overrides}
        setOverrides={setOverrides}
      />
      <SettingsAddPerson engineers={engineers} setEngineers={setEngineers} />
      <SettingsOverrides engineers={engineers} overrides={overrides} setOverrides={setOverrides} />
      {/* Webhooks */}
      <Card className="shadow-sm border-border bg-card">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold">Webhooks</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Existing webhooks */}
          {webhooks.length > 0 && (
            <div className="space-y-1">
              {webhooks.map(wh => (
                <div key={wh.id} className="flex items-center gap-3 px-3 py-2.5 rounded-xl bg-muted/40">
                  <Webhook className="h-4 w-4 text-muted-foreground shrink-0" />
                  <div className="flex-1 min-w-0">
                    {wh.label && <p className="text-sm font-medium truncate">{wh.label}</p>}
                    <p className="text-xs text-muted-foreground truncate">{wh.url}</p>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => removeWebhook(wh.id)}
                    className="shrink-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                    aria-label="Remove webhook"
                  >
                    <X />
                  </Button>
                </div>
              ))}
            </div>
          )}

          {/* Add webhook form */}
          <div className="space-y-2">
            <input
              type="text"
              placeholder="Label (optional)"
              value={webhookLabel}
              onChange={e => setWebhookLabel(e.target.value)}
              onKeyDown={handleWebhookKeyDown}
              className={inputClass}
            />
            <input
              type="url"
              placeholder="https://example.com/webhook"
              value={webhookUrl}
              onChange={e => setWebhookUrl(e.target.value)}
              onKeyDown={handleWebhookKeyDown}
              className={inputClass}
            />
            <Button onClick={addWebhook} disabled={!webhookUrlValid} size="sm" className="gap-1.5">
              <Plus />
              Add webhook
            </Button>
          </div>
        </CardContent>
      </Card>

    </div>
  )
}

export default SettingsPage

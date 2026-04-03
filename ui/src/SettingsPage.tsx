import { useState, useRef } from 'react'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ArrowLeft, Plus, X, ArrowRight, ChevronDown, Webhook } from 'lucide-react'
import type { Engineer, Override, TimeSegment, WebhookEntry } from './types'
import { buildTimeline, formatOverrideRange, formatSegmentRange, initials, inputClass } from './utils'
import SettingsRotationOrder from './SettingsRotationOrder'
import SettingsAddPerson from './SettingsAddPerson'

/**
 * Given a prospective override window [previewStart, previewEnd), compute which
 * segments from the baseline schedule (built without that override) would be
 * displaced. Returns those segments clipped to the override window.
 */
function computeOverrideReplacements(
  engineers: Engineer[],
  baseOverrides: Override[],
  previewStart: string,
  previewEnd: string,
): TimeSegment[] {
  if (engineers.length === 0 || !previewStart || !previewEnd) return []
  const start = new Date(previewStart)
  const end = new Date(previewEnd)
  if (isNaN(start.getTime()) || isNaN(end.getTime()) || end <= start) return []

  const timeline = buildTimeline(engineers, baseOverrides, 8)
  return timeline
    .filter(seg => seg.start < end && seg.end > start)
    .map(seg => ({
      ...seg,
      start: seg.start < start ? new Date(start) : seg.start,
      end: seg.end > end ? new Date(end) : seg.end,
    }))
}

// --- Display helpers ---
const selectClass = 'w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 transition-shadow'


function SettingsPage({ engineers, setEngineers, overrides, setOverrides, webhooks, setWebhooks, onNavigateHome }: {
  engineers: Engineer[]
  setEngineers: (engineers: Engineer[]) => void
  overrides: Override[]
  setOverrides: (overrides: Override[]) => void
  webhooks: WebhookEntry[]
  setWebhooks: (webhooks: WebhookEntry[]) => void
  onNavigateHome: () => void
}) {
  // Rotation order state
  const dragIndexRef = useRef<number | null>(null)

  // Override form state
  const [overrideStart, setOverrideStart] = useState('')
  const [overrideEnd, setOverrideEnd] = useState('')
  const [overrideEngineerId, setOverrideEngineerId] = useState('')

  const validEngineerId = engineers.find(e => e.id === overrideEngineerId)
    ? overrideEngineerId
    : ''

  const overrideValid = overrideStart && overrideEnd && new Date(overrideEnd) > new Date(overrideStart) && validEngineerId
  const formReplacements = overrideValid ? computeOverrideReplacements(engineers, overrides, overrideStart, overrideEnd) : []
  const overrideSelfAssign = formReplacements.some(seg => seg.engineer.id === validEngineerId)

  // Rotation order handlers
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

  function removeEngineer(id: string) {
    setEngineers(engineers.filter(e => e.id !== id))
    setOverrides(overrides.filter(o => o.engineerId !== id))
  }

  // Webhook form state
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

  // Override handlers
  function addOverride() {
    if (!overrideValid) return
    setOverrides([
      ...overrides,
      { id: crypto.randomUUID(), start: overrideStart, end: overrideEnd, engineerId: validEngineerId },
    ])
    setOverrideStart('')
    setOverrideEnd('')
    setOverrideEngineerId('')
  }

  function removeOverride(id: string) {
    setOverrides(overrides.filter(o => o.id !== id))
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon-sm" onClick={onNavigateHome} aria-label="Back to home">
          <ArrowLeft />
        </Button>
        <h1 className="text-xl font-bold tracking-tight">Settings</h1>
      </div>

      <SettingsRotationOrder
        engineers={engineers}
        handleDragStart={handleDragStart}
        handleDragOver={handleDragOver}
        handleDragEnd={handleDragEnd}
        removeEngineer={removeEngineer}
      />

      <SettingsAddPerson engineers={engineers} setEngineers={setEngineers} />

      {/* Schedule overrides */}
      <Card className="shadow-sm border-border bg-card">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold">Schedule overrides</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Existing overrides */}
          {overrides.length > 0 && (
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
          )}

          {/* Add override form */}
          {engineers.length === 0 ? (
            <p className="text-sm text-muted-foreground">Add engineers to the rotation before creating overrides.</p>
          ) : (
            <div className="space-y-3">
              <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-2">
                <input
                  type="datetime-local"
                  value={overrideStart}
                  onChange={e => setOverrideStart(e.target.value)}
                  className={inputClass}
                />
                <ArrowRight className="h-4 w-4 text-muted-foreground shrink-0" />
                <input
                  type="datetime-local"
                  value={overrideEnd}
                  onChange={e => setOverrideEnd(e.target.value)}
                  className={inputClass}
                />
              </div>
              <div className="relative">
                <select
                  value={validEngineerId}
                  onChange={e => setOverrideEngineerId(e.target.value)}
                  className={selectClass + ' appearance-none pr-8'}
                >
                  <option value="" disabled>Select person</option>
                  {engineers.map(e => (
                    <option key={e.id} value={e.id}>{e.name}</option>
                  ))}
                </select>
                <ChevronDown className="absolute right-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
              </div>
              {overrideValid && formReplacements.length > 0 && (
                <div className={`rounded-lg border px-3 py-2.5 space-y-1.5 ${overrideSelfAssign ? 'border-destructive/50 bg-destructive/5' : 'border-border bg-muted/30'}`}>
                  <p className={`text-xs font-medium ${overrideSelfAssign ? 'text-destructive' : 'text-muted-foreground'}`}>Replaces</p>
                  <div className="space-y-1">
                    {formReplacements.map((seg, i) => {
                      const isSelf = seg.engineer.id === validEngineerId
                      return (
                        <div key={i} className="flex items-center gap-2 text-sm">
                          <Avatar className="h-5 w-5 shrink-0">
                            <AvatarImage src={seg.engineer.avatarUrl} />
                            <AvatarFallback className={`text-[9px] font-semibold ${seg.engineer.color} ${seg.engineer.textColor}`}>
                              {initials(seg.engineer.name)}
                            </AvatarFallback>
                          </Avatar>
                          <span className={`font-medium ${isSelf ? 'text-destructive' : ''}`}>{seg.engineer.name}</span>
                          <span className="text-muted-foreground text-xs">{formatSegmentRange(seg.start, seg.end)}</span>
                          {isSelf && <span className="text-xs text-destructive">already on call</span>}
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}
              <Button onClick={addOverride} disabled={!overrideValid || overrideSelfAssign} className="gap-1.5">
                <Plus />
                Add override
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

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

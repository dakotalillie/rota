import { useState, useRef } from 'react'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Phone, Clock, CalendarDays, Settings2, ArrowLeft, GripVertical, Plus, X, ArrowRight, ChevronDown, Webhook } from 'lucide-react'

// --- Types ---

type Engineer = {
  id: string
  name: string
  email: string
  avatarUrl?: string
  color: string        // tailwind bg class for avatar
  lightColor: string   // tailwind bg class for row highlight (light mode)
  darkColor: string    // tailwind bg class for row highlight (dark mode)
  textColor: string    // tailwind text class for avatar text
}

type Override = {
  id: string
  start: string      // datetime-local string: "YYYY-MM-DDThh:mm"
  end: string        // datetime-local string: "YYYY-MM-DDThh:mm"
  engineerId: string
}

type WebhookEntry = {
  id: string
  url: string
  label: string
}

type TimeSegment = {
  start: Date
  end: Date
  engineer: Engineer
  isOverride: boolean
}

// --- Color palette ---

const COLOR_PALETTE = [
  { color: 'bg-violet-500', lightColor: 'bg-violet-50',  darkColor: 'dark:bg-violet-950/50',  textColor: 'text-white' },
  { color: 'bg-sky-500',    lightColor: 'bg-sky-50',     darkColor: 'dark:bg-sky-950/50',     textColor: 'text-white' },
  { color: 'bg-emerald-500',lightColor: 'bg-emerald-50', darkColor: 'dark:bg-emerald-950/50', textColor: 'text-white' },
  { color: 'bg-orange-400', lightColor: 'bg-orange-50',  darkColor: 'dark:bg-orange-950/50',  textColor: 'text-white' },
  { color: 'bg-rose-500',   lightColor: 'bg-rose-50',    darkColor: 'dark:bg-rose-950/50',    textColor: 'text-white' },
  { color: 'bg-teal-500',   lightColor: 'bg-teal-50',    darkColor: 'dark:bg-teal-950/50',    textColor: 'text-white' },
  { color: 'bg-amber-500',  lightColor: 'bg-amber-50',   darkColor: 'dark:bg-amber-950/50',   textColor: 'text-white' },
  { color: 'bg-pink-500',   lightColor: 'bg-pink-50',    darkColor: 'dark:bg-pink-950/50',    textColor: 'text-white' },
]

// --- Prototype data ---

const INITIAL_ENGINEERS: Engineer[] = [
  { id: '1', name: 'Alex Rivera',   email: 'alex@example.com',   color: 'bg-violet-500',  lightColor: 'bg-violet-50',  darkColor: 'dark:bg-violet-950/50',  textColor: 'text-white' },
  { id: '2', name: 'Jordan Kim',    email: 'jordan@example.com', color: 'bg-sky-500',     lightColor: 'bg-sky-50',     darkColor: 'dark:bg-sky-950/50',     textColor: 'text-white' },
  { id: '3', name: 'Sam Patel',     email: 'sam@example.com',    color: 'bg-emerald-500', lightColor: 'bg-emerald-50', darkColor: 'dark:bg-emerald-950/50', textColor: 'text-white' },
  { id: '4', name: 'Casey Morgan',  email: 'casey@example.com',  color: 'bg-orange-400',  lightColor: 'bg-orange-50',  darkColor: 'dark:bg-orange-950/50',  textColor: 'text-white' },
  { id: '5', name: 'Taylor Brooks', email: 'taylor@example.com', color: 'bg-rose-500',    lightColor: 'bg-rose-50',    darkColor: 'dark:bg-rose-950/50',    textColor: 'text-white' },
]

// --- Schedule helpers ---

function mondayOf(date: Date): Date {
  const d = new Date(date)
  const day = d.getDay()
  const diff = (day === 0 ? -6 : 1 - day)
  d.setDate(d.getDate() + diff)
  d.setHours(0, 0, 0, 0)
  return d
}

function toDateTimeLocal(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

/**
 * Build a timeline of non-overlapping segments by:
 * 1. Generating base week segments from the rotation
 * 2. Collecting all time boundaries (week starts + override starts/ends)
 * 3. For each sub-interval, checking whether an override covers it; if not, using the rotation engineer
 * 4. Merging adjacent segments with the same engineer
 */
function buildTimeline(engineers: Engineer[], overrides: Override[], weeksCount: number): TimeSegment[] {
  if (engineers.length === 0) return []

  const start = mondayOf(new Date())
  const endTime = new Date(start)
  endTime.setDate(endTime.getDate() + weeksCount * 7)

  // Collect all relevant time boundaries within the window
  const boundarySet = new Set<number>()
  boundarySet.add(start.getTime())
  boundarySet.add(endTime.getTime())

  // Add a boundary for each week
  const d = new Date(start)
  for (let i = 1; i < weeksCount; i++) {
    d.setDate(d.getDate() + 7)
    boundarySet.add(d.getTime())
  }

  // Add override boundaries, clamped to the window
  for (const ov of overrides) {
    const ovStart = new Date(ov.start).getTime()
    const ovEnd = new Date(ov.end).getTime()
    if (ovEnd <= start.getTime() || ovStart >= endTime.getTime()) continue
    boundarySet.add(Math.max(ovStart, start.getTime()))
    boundarySet.add(Math.min(ovEnd, endTime.getTime()))
  }

  const boundaries = [...boundarySet].sort((a, b) => a - b)

  // Build raw segments
  const raw: TimeSegment[] = []
  for (let i = 0; i < boundaries.length - 1; i++) {
    const segStart = new Date(boundaries[i])
    const segEnd = new Date(boundaries[i + 1])
    const midMs = (boundaries[i] + boundaries[i + 1]) / 2

    // Find the most-recently-added override that covers this segment (last one wins)
    let overrideEngineer: Engineer | undefined
    for (let j = overrides.length - 1; j >= 0; j--) {
      const ov = overrides[j]
      const ovStart = new Date(ov.start).getTime()
      const ovEnd = new Date(ov.end).getTime()
      if (ovStart <= midMs && ovEnd > midMs) {
        overrideEngineer = engineers.find(e => e.id === ov.engineerId)
        break
      }
    }

    if (overrideEngineer) {
      raw.push({ start: segStart, end: segEnd, engineer: overrideEngineer, isOverride: true })
    } else {
      // Which week index does this fall in?
      const weekIndex = Math.floor((boundaries[i] - start.getTime()) / (7 * 24 * 60 * 60 * 1000))
      const engineer = engineers[weekIndex % engineers.length]
      raw.push({ start: segStart, end: segEnd, engineer, isOverride: false })
    }
  }

  // Merge adjacent segments with the same engineer and override status
  const merged: TimeSegment[] = []
  for (const seg of raw) {
    const last = merged[merged.length - 1]
    if (last && last.engineer.id === seg.engineer.id && last.isOverride === seg.isOverride) {
      last.end = seg.end
    } else {
      merged.push({ ...seg })
    }
  }

  return merged
}

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

function initials(name: string) {
  return name.split(' ').map(p => p[0]).join('').toUpperCase()
}

const fmtDate = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric' })
const fmtDateTime = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' })

function isMidnight(date: Date) {
  return date.getHours() === 0 && date.getMinutes() === 0 && date.getSeconds() === 0
}

/**
 * Format a half-open segment [start, end).
 * If both boundaries are at midnight, show as "Apr 1 – Apr 7" (end is exclusive, so display end-1 day).
 * If either has a time component, show full datetimes.
 */
function formatSegmentRange(start: Date, end: Date): string {
  if (isMidnight(start) && isMidnight(end)) {
    const displayEnd = new Date(end)
    displayEnd.setDate(displayEnd.getDate() - 1)
    return `${fmtDateTime.format(start)} – ${fmtDateTime.format(displayEnd)}`
  }
  return `${fmtDateTime.format(start)} – ${fmtDateTime.format(end)}`
}

/** Format an override's stored datetime-local strings for display. */
function formatOverrideRange(start: string, end: string): string {
  return `${fmtDateTime.format(new Date(start))} – ${fmtDateTime.format(new Date(end))}`
}

function isActiveNow(seg: TimeSegment): boolean {
  const now = new Date()
  return now >= seg.start && now < seg.end
}

const selectClass = 'w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 transition-shadow'
const inputClass = 'w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 transition-shadow'

// --- Home page components ---

function OnCallHero({ segment }: { segment: TimeSegment }) {
  return (
    <div className="relative overflow-hidden rounded-2xl bg-linear-to-br from-violet-500 via-indigo-500 to-sky-500 dark:from-violet-700 dark:via-indigo-700 dark:to-sky-700 p-px shadow-lg shadow-indigo-200 dark:shadow-indigo-950">
      <div className="relative rounded-[calc(1rem-1px)] bg-linear-to-br from-violet-500 via-indigo-500 to-sky-500 dark:from-violet-700 dark:via-indigo-700 dark:to-sky-700 px-6 py-5">
        {/* Decorative blobs */}
        <div className="absolute -top-6 -right-6 h-32 w-32 rounded-full bg-white/10 blur-2xl" />
        <div className="absolute -bottom-8 -left-4 h-24 w-24 rounded-full bg-white/10 blur-2xl" />

        <div className="relative">
          <div className="flex items-center gap-2 mb-4">
            <div className="h-2 w-2 rounded-full bg-green-300 animate-pulse shadow-sm shadow-green-400" />
            <span className="text-sm font-semibold text-white/80 uppercase tracking-widest">
              Currently On Call
            </span>
          </div>

          <div className="flex items-center gap-5">
            <Avatar className="ring-4 ring-white/30 shadow-xl" style={{ height: '4.5rem', width: '4.5rem' }}>
              <AvatarImage src={segment.engineer.avatarUrl} />
              <AvatarFallback className="bg-white/20 text-white font-bold text-xl backdrop-blur-sm">
                {initials(segment.engineer.name)}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col gap-1.5">
              <h2 className="text-2xl font-bold text-white">{segment.engineer.name}</h2>
              <div className="flex items-center gap-1.5 text-sm text-white/70">
                <Phone className="h-3.5 w-3.5" />
                <span>{segment.engineer.email}</span>
              </div>
              <div className="flex items-center gap-1.5 text-sm text-white/70">
                <Clock className="h-3.5 w-3.5" />
                <span>{formatSegmentRange(segment.start, segment.end)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function ScheduleRow({ segment, index }: { segment: TimeSegment; index: number }) {
  const isActive = isActiveNow(segment)
  const { engineer } = segment

  return (
    <div
      className={`flex items-center gap-4 px-4 py-3 rounded-xl transition-colors ${
        isActive
          ? `${engineer.lightColor} ${engineer.darkColor} ring-1 ring-inset ring-current/10`
          : index % 2 === 0
          ? 'bg-muted/40'
          : ''
      }`}
    >
      {/* Time range */}
      <div className="w-64 shrink-0 flex items-center gap-2">
        <p className={`text-sm font-medium whitespace-nowrap ${isActive ? 'text-foreground' : 'text-muted-foreground'}`}>
          {formatSegmentRange(segment.start, segment.end)}
        </p>
        {segment.isOverride && (
          <Badge className="text-xs px-1.5 py-0 bg-amber-100 text-amber-700 border-amber-200 hover:bg-amber-100 dark:bg-amber-900/50 dark:text-amber-400 dark:border-amber-800 dark:hover:bg-amber-900/50">
            Override
          </Badge>
        )}
      </div>

      <Separator orientation="vertical" className="h-8" />

      {/* Engineer */}
      <div className="flex items-center gap-3 flex-1 min-w-0">
        <Avatar className="h-8 w-8 shrink-0">
          <AvatarImage src={engineer.avatarUrl} />
          <AvatarFallback className={`text-xs font-semibold ${engineer.color} ${engineer.textColor}`}>
            {initials(engineer.name)}
          </AvatarFallback>
        </Avatar>
        <span className="text-sm font-medium truncate">{engineer.name}</span>
      </div>
    </div>
  )
}

function HomePage({ timeline, onNavigateEdit }: {
  timeline: TimeSegment[]
  onNavigateEdit: () => void
}) {
  const now = new Date()
  const activeSegment = timeline.find(s => now >= s.start && now < s.end) ?? timeline[0]

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold tracking-tight">Rota</h1>
        <Button variant="outline" size="sm" onClick={onNavigateEdit} className="gap-1.5">
          <Settings2 />
          Settings
        </Button>
      </div>

      {/* Hero */}
      {activeSegment ? (
        <OnCallHero segment={activeSegment} />
      ) : (
        <div className="rounded-2xl border border-dashed border-border p-10 text-center text-sm text-muted-foreground">
          No engineers in the rotation yet.{' '}
          <button onClick={onNavigateEdit} className="underline underline-offset-4 hover:text-foreground transition-colors">
            Add some.
          </button>
        </div>
      )}

      {/* Upcoming schedule */}
      <Card className="shadow-sm border-border bg-card">
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base font-semibold">
            <CalendarDays className="h-4 w-4 text-indigo-500" />
            Upcoming Rotation
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-1">
          {timeline.length > 0 ? (
            timeline.map((seg, i) => (
              <ScheduleRow key={i} segment={seg} index={i} />
            ))
          ) : (
            <p className="text-sm text-muted-foreground px-4 py-2">No engineers in the rotation yet.</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

// --- Settings page ---


function EditPage({ engineers, setEngineers, overrides, setOverrides, webhooks, setWebhooks, onNavigateHome }: {
  engineers: Engineer[]
  setEngineers: (engineers: Engineer[]) => void
  overrides: Override[]
  setOverrides: (overrides: Override[]) => void
  webhooks: WebhookEntry[]
  setWebhooks: (webhooks: WebhookEntry[]) => void
  onNavigateHome: () => void
}) {
  // Rotation order state
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
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

  function addEngineer() {
    const trimmedName = name.trim()
    if (!trimmedName) return
    const palette = COLOR_PALETTE[engineers.length % COLOR_PALETTE.length]
    const newEngineer = { id: crypto.randomUUID(), name: trimmedName, email: email.trim(), ...palette }
    setEngineers([...engineers, newEngineer])
    setName('')
    setEmail('')
  }

  function handleAddEngineerKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') addEngineer()
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

      {/* Rotation order */}
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

      {/* Add person */}
      <Card className="shadow-sm border-border bg-card">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold">Add person</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <input
            type="text"
            placeholder="Name"
            value={name}
            onChange={e => setName(e.target.value)}
            onKeyDown={handleAddEngineerKeyDown}
            className={inputClass}
          />
          <input
            type="email"
            placeholder="Email (optional)"
            value={email}
            onChange={e => setEmail(e.target.value)}
            onKeyDown={handleAddEngineerKeyDown}
            className={inputClass}
          />
          <Button onClick={addEngineer} disabled={!name.trim()} size="sm" className="gap-1.5">
            <Plus />
            Add person
          </Button>
        </CardContent>
      </Card>

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

// --- Root ---

export default function App() {
  const [engineers, setEngineers] = useState<Engineer[]>(INITIAL_ENGINEERS)
  const [overrides, setOverrides] = useState<Override[]>([])
  const [webhooks, setWebhooks] = useState<WebhookEntry[]>([])
  const [page, setPage] = useState<'home' | 'edit'>('home')

  const timeline = buildTimeline(engineers, overrides, 8)

  return (
    <div className="min-h-screen bg-background">
      {page === 'home' ? (
        <HomePage
          timeline={timeline}
          onNavigateEdit={() => setPage('edit')}
        />
      ) : (
        <EditPage
          engineers={engineers}
          setEngineers={setEngineers}
          overrides={overrides}
          setOverrides={setOverrides}
          webhooks={webhooks}
          setWebhooks={setWebhooks}
          onNavigateHome={() => setPage('home')}
        />
      )}
    </div>
  )
}

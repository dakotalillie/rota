import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Phone, Clock, CalendarDays } from 'lucide-react'

// --- Prototype data ---

type Engineer = {
  name: string
  email: string
  avatarUrl?: string
  color: string        // tailwind bg class for avatar
  lightColor: string   // tailwind bg class for row highlight
  textColor: string    // tailwind text class for avatar text
}

type ScheduleEntry = {
  weekStart: Date
  weekEnd: Date
  engineer: Engineer
}

const engineers: Engineer[] = [
  { name: 'Alex Rivera',   email: 'alex@example.com',   color: 'bg-violet-500',  lightColor: 'bg-violet-50',  textColor: 'text-white' },
  { name: 'Jordan Kim',    email: 'jordan@example.com', color: 'bg-sky-500',     lightColor: 'bg-sky-50',     textColor: 'text-white' },
  { name: 'Sam Patel',     email: 'sam@example.com',    color: 'bg-emerald-500', lightColor: 'bg-emerald-50', textColor: 'text-white' },
  { name: 'Casey Morgan',  email: 'casey@example.com',  color: 'bg-orange-400',  lightColor: 'bg-orange-50',  textColor: 'text-white' },
  { name: 'Taylor Brooks', email: 'taylor@example.com', color: 'bg-rose-500',    lightColor: 'bg-rose-50',    textColor: 'text-white' },
]

function mondayOf(date: Date): Date {
  const d = new Date(date)
  const day = d.getDay()
  const diff = (day === 0 ? -6 : 1 - day)
  d.setDate(d.getDate() + diff)
  d.setHours(0, 0, 0, 0)
  return d
}

function sundayOf(monday: Date): Date {
  const d = new Date(monday)
  d.setDate(d.getDate() + 6)
  return d
}

function buildSchedule(weeksCount: number): ScheduleEntry[] {
  const schedule: ScheduleEntry[] = []
  let monday = mondayOf(new Date())
  for (let i = 0; i < weeksCount; i++) {
    schedule.push({
      weekStart: new Date(monday),
      weekEnd: sundayOf(monday),
      engineer: engineers[i % engineers.length],
    })
    monday = new Date(monday)
    monday.setDate(monday.getDate() + 7)
  }
  return schedule
}

const schedule = buildSchedule(8)
const currentEntry = schedule[0]

// --- Helpers ---

function initials(name: string) {
  return name.split(' ').map(p => p[0]).join('').toUpperCase()
}

const fmt = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric' })

function formatRange(start: Date, end: Date) {
  return `${fmt.format(start)} – ${fmt.format(end)}`
}

function isCurrentWeek(entry: ScheduleEntry) {
  const now = new Date()
  return now >= entry.weekStart && now <= entry.weekEnd
}

// --- Components ---

function OnCallHero({ entry }: { entry: ScheduleEntry }) {
  return (
    <div className="relative overflow-hidden rounded-2xl bg-linear-to-br from-violet-500 via-indigo-500 to-sky-500 p-px shadow-lg shadow-indigo-200">
      <div className="relative rounded-[calc(1rem-1px)] bg-linear-to-br from-violet-500 via-indigo-500 to-sky-500 px-6 py-5">
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
            <Avatar className="h-18 w-18 ring-4 ring-white/30 shadow-xl" style={{ height: '4.5rem', width: '4.5rem' }}>
              <AvatarImage src={entry.engineer.avatarUrl} />
              <AvatarFallback className="bg-white/20 text-white font-bold text-xl backdrop-blur-sm">
                {initials(entry.engineer.name)}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col gap-1.5">
              <h2 className="text-2xl font-bold text-white">{entry.engineer.name}</h2>
              <div className="flex items-center gap-1.5 text-sm text-white/70">
                <Phone className="h-3.5 w-3.5" />
                <span>{entry.engineer.email}</span>
              </div>
              <div className="flex items-center gap-1.5 text-sm text-white/70">
                <Clock className="h-3.5 w-3.5" />
                <span>Week of {formatRange(entry.weekStart, entry.weekEnd)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function ScheduleRow({ entry, index }: { entry: ScheduleEntry; index: number }) {
  const isCurrent = isCurrentWeek(entry)
  const { engineer } = entry

  return (
    <div
      className={`flex items-center gap-4 px-4 py-3 rounded-xl transition-colors ${
        isCurrent
          ? `${engineer.lightColor} ring-1 ring-inset ring-current/10`
          : index % 2 === 0
          ? 'bg-muted/40'
          : ''
      }`}
    >
      {/* Week label */}
      <div className="w-36 shrink-0">
        <p className={`text-sm font-medium ${isCurrent ? 'text-foreground' : 'text-muted-foreground'}`}>
          {formatRange(entry.weekStart, entry.weekEnd)}
        </p>
        {isCurrent && (
          <Badge className="mt-0.5 text-xs px-1.5 py-0 bg-green-100 text-green-700 border-green-200 hover:bg-green-100">
            This week
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

export default function App() {
  return (
    <div className="min-h-screen bg-linear-to-br from-slate-50 via-indigo-50/40 to-violet-50/30">
      <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">

        {/* Hero */}
        <OnCallHero entry={currentEntry} />

        {/* Upcoming schedule */}
        <Card className="shadow-sm border-slate-200">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base font-semibold">
              <CalendarDays className="h-4 w-4 text-indigo-500" />
              Upcoming Rotation
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-1">
            {schedule.map((entry, i) => (
              <ScheduleRow key={i} entry={entry} index={i} />
            ))}
          </CardContent>
        </Card>

      </div>
    </div>
  )
}

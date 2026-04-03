import { CalendarDays } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { TimeSegment } from "./types"
import HomePageScheduleRow from "./HomePageScheduleRow"

type HomePageScheduleProps = {
    timeline: TimeSegment[]
}

function HomePageSchedule({ timeline }: HomePageScheduleProps) {
    return (
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
              <HomePageScheduleRow key={i} segment={seg} index={i} />
            ))
          ) : (
            <p className="text-sm text-muted-foreground px-4 py-2">No engineers in the rotation yet.</p>
          )}
        </CardContent>
      </Card>
    )
}

export default HomePageSchedule
import type { TimeSegment } from "@/types"
import HomePageHeader from "./HomePageHeader"
import HomePageHeroEmpty from "./HomePageHeroEmpty"
import HomePageHero from "./HomePageHero"
import HomePageSchedule from "./HomePageSchedule"

type HomePageProps = {
  timeline: TimeSegment[]
  onNavigateEdit: () => void
}

function HomePage({ timeline, onNavigateEdit }: HomePageProps) {
  const now = new Date()
  const activeSegment = timeline.find(s => now >= s.start && now < s.end) ?? timeline[0]

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      <HomePageHeader onNavigateEdit={onNavigateEdit} />

      {activeSegment ? (
        <HomePageHero segment={activeSegment} />
      ) : <HomePageHeroEmpty onNavigateEdit={onNavigateEdit} />}

      <HomePageSchedule timeline={timeline} />
    </div>
  )
}

export default HomePage
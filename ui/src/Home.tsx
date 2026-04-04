import type { TimeSegment } from "@/types"
import HomeHeader from "./HomeHeader"
import HomeHeroEmpty from "./HomeHeroEmpty"
import HomeHero from "./HomeHero"
import HomeSchedule from "./HomeSchedule"

type HomeProps = {
  timeline: TimeSegment[]
  onNavigateEdit: () => void
}

function Home({ timeline, onNavigateEdit }: HomeProps) {
  const now = new Date()
  const activeSegment = timeline.find(s => now >= s.start && now < s.end) ?? timeline[0]

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      <HomeHeader onNavigateEdit={onNavigateEdit} />

      {activeSegment ? (
        <HomeHero segment={activeSegment} />
      ) : <HomeHeroEmpty onNavigateEdit={onNavigateEdit} />}

      <HomeSchedule timeline={timeline} />
    </div>
  )
}

export default Home
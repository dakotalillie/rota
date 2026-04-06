import { useAppState } from "./AppStateContext";
import HomeHeader from "./HomeHeader";
import HomeHero from "./HomeHero";
import HomeHeroEmpty from "./HomeHeroEmpty";
import HomeSchedule from "./HomeSchedule";
import { buildTimeline } from "./utils";

function Home() {
  const { engineers, overrides } = useAppState();
  const timeline = buildTimeline(engineers, overrides, 8);
  const now = new Date();
  const activeSegment =
    timeline.find((s) => now >= s.start && now < s.end) ?? timeline[0];

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      <HomeHeader />

      {activeSegment ? <HomeHero segment={activeSegment} /> : <HomeHeroEmpty />}

      <HomeSchedule timeline={timeline} />
    </div>
  );
}

export default Home;

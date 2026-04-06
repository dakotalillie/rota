import { Link, useParams } from "@tanstack/react-router";
import { Settings2 } from "lucide-react";
import { useEffect, useState } from "react";

import { useBreadcrumbs } from "./BreadcrumbContext";
import { Button } from "./Button";
import HomeHero from "./HomeHero";
import HomeHeroEmpty from "./HomeHeroEmpty";
import HomeSchedule from "./HomeSchedule";
import PageHeader from "./PageHeader";
import type { Engineer, TimeSegment } from "./types";

const COLORS: Pick<
  Engineer,
  "color" | "textColor" | "lightColor" | "darkColor"
>[] = [
  {
    color: "bg-indigo-500",
    textColor: "text-white",
    lightColor: "bg-indigo-50",
    darkColor: "dark:bg-indigo-950",
  },
  {
    color: "bg-emerald-500",
    textColor: "text-white",
    lightColor: "bg-emerald-50",
    darkColor: "dark:bg-emerald-950",
  },
  {
    color: "bg-rose-500",
    textColor: "text-white",
    lightColor: "bg-rose-50",
    darkColor: "dark:bg-rose-950",
  },
  {
    color: "bg-amber-500",
    textColor: "text-white",
    lightColor: "bg-amber-50",
    darkColor: "dark:bg-amber-950",
  },
  {
    color: "bg-sky-500",
    textColor: "text-white",
    lightColor: "bg-sky-50",
    darkColor: "dark:bg-sky-950",
  },
  {
    color: "bg-violet-500",
    textColor: "text-white",
    lightColor: "bg-violet-50",
    darkColor: "dark:bg-violet-950",
  },
];

interface ApiMember {
  type: "members";
  id: string;
  relationships: { user: { data: { type: "users"; id: string } } };
}

interface ApiUser {
  type: "users";
  id: string;
  attributes: { name: string; email: string };
}

interface ApiScheduleBlock {
  type: "scheduleBlocks";
  id: string;
  attributes: { start: string; end: string };
  relationships: { member: { data: { type: "members"; id: string } } };
}

interface GetRotationResponse {
  data?: { attributes: { name: string } };
  errors?: { detail?: string }[];
}

interface GetScheduleResponse {
  data: ApiScheduleBlock[];
  included?: (ApiMember | ApiUser)[];
  errors?: { detail?: string }[];
}

function buildTimelineFromSchedule(
  data: ApiScheduleBlock[],
  included: (ApiMember | ApiUser)[] | undefined,
): TimeSegment[] {
  const userMap = new Map<string, ApiUser>();
  const memberUserMap = new Map<string, string>();

  for (const item of included ?? []) {
    if (item.type === "users") userMap.set(item.id, item);
    if (item.type === "members")
      memberUserMap.set(item.id, item.relationships.user.data.id);
  }

  const userColorIndex = new Map<string, number>();
  const engineers = new Map<string, Engineer>();

  for (const block of data) {
    const memberId = block.relationships.member.data.id;
    const userId = memberUserMap.get(memberId);
    if (!userId) continue;
    if (!engineers.has(userId)) {
      const user = userMap.get(userId);
      if (!user) continue;
      const idx = userColorIndex.size;
      userColorIndex.set(userId, idx);
      engineers.set(userId, {
        id: userId,
        name: user.attributes.name,
        email: user.attributes.email,
        ...COLORS[idx % COLORS.length],
      });
    }
  }

  return data.map((block) => {
    const memberId = block.relationships.member.data.id;
    const userId = memberUserMap.get(memberId) ?? "";
    const engineer = engineers.get(userId) ?? {
      id: userId,
      name: "Unknown",
      email: "",
      ...COLORS[0],
    };
    return {
      start: new Date(block.attributes.start),
      end: new Date(block.attributes.end),
      engineer,
      isOverride: false,
    };
  });
}

function RotationDetail() {
  const { rotationId } = useParams({ from: "/rotations/$rotationId" });

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [rotationName, setRotationName] = useState<string | null>(null);
  const [timeline, setTimeline] = useState<TimeSegment[]>([]);

  useEffect(() => {
    Promise.all([
      fetch(`/api/rotations/${rotationId}`).then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json() as Promise<GetRotationResponse>;
      }),
      fetch(`/api/rotations/${rotationId}/schedule`).then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json() as Promise<GetScheduleResponse>;
      }),
    ])
      .then(([rotationRes, scheduleRes]) => {
        if (rotationRes.data) {
          setRotationName(rotationRes.data.attributes.name);
        }
        setTimeline(
          buildTimelineFromSchedule(scheduleRes.data, scheduleRes.included),
        );
        setLoading(false);
      })
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : "Unknown error");
        setLoading(false);
      });
  }, [rotationId]);

  useBreadcrumbs([
    { label: "Rotations", to: "/rotations" },
    { label: rotationName ?? "…" },
  ]);

  const now = new Date();
  const activeSegment =
    timeline.find((s) => now >= s.start && now < s.end) ?? timeline[0];

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      <PageHeader
        title={rotationName ?? "Loading…"}
        actions={
          <Link to="/rotations/$rotationId/settings" params={{ rotationId }}>
            <Button variant="outline" size="sm" className="gap-1.5">
              <Settings2 />
              Settings
            </Button>
          </Link>
        }
      />

      {loading && <p className="text-sm text-neutral-500">Loading…</p>}

      {error && (
        <p className="text-sm text-red-500">Failed to load rotation: {error}</p>
      )}

      {!loading && !error && (
        <>
          {activeSegment ? (
            <HomeHero segment={activeSegment} />
          ) : (
            <HomeHeroEmpty />
          )}
          <HomeSchedule timeline={timeline} />
        </>
      )}
    </div>
  );
}

export default RotationDetail;

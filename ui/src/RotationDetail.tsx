import { Link, useParams } from "@tanstack/react-router";
import { Settings2 } from "lucide-react";
import { useEffect, useState } from "react";

import { useBreadcrumbs } from "./BreadcrumbContext";
import { Button } from "./Button";
import HomeHero from "./OnCallHero";
import OnCallHeroEmpty from "./OnCallHeroEmpty";
import PageHeader from "./PageHeader";
import Schedule from "./Schedule";
import type { TimeSegment } from "./types";
import {
  type ApiMember,
  type ApiScheduleBlock,
  type ApiUser,
  buildTimelineFromSchedule,
} from "./utils";

interface GetRotationResponse {
  data?: {
    attributes: { name: string };
    relationships?: {
      currentMember?: { data: { type: "members"; id: string } | null };
    };
  };
  errors?: { detail?: string }[];
}

interface GetScheduleResponse {
  data: ApiScheduleBlock[];
  included?: (ApiMember | ApiUser)[];
  errors?: { detail?: string }[];
}

function RotationDetail() {
  const { rotationId } = useParams({ from: "/rotations/$rotationId" });

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [rotationName, setRotationName] = useState<string | null>(null);
  const [timeline, setTimeline] = useState<TimeSegment[]>([]);
  const [currentMemberId, setCurrentMemberId] = useState<string | null>(null);

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
          setCurrentMemberId(
            rotationRes.data.relationships?.currentMember?.data?.id ?? null,
          );
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

  const activeSegment = currentMemberId
    ? (timeline.find((s) => s.member.id === currentMemberId) ?? timeline[0])
    : timeline[0];

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
            <OnCallHeroEmpty />
          )}
          <Schedule timeline={timeline} />
        </>
      )}
    </div>
  );
}

export default RotationDetail;

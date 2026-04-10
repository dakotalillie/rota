import { useParams } from "@tanstack/react-router";
import { useEffect, useState } from "react";

import { useAppState } from "./AppStateContext";
import { useBreadcrumbs } from "./BreadcrumbContext";
import { Card, CardContent, CardHeader, CardTitle } from "./Card";
import { colorsForName } from "./colorPalette";
import Members from "./Members";
import Overrides from "./Overrides";
import PageHeader from "./PageHeader";
import type { Member, Override } from "./types";

type Cadence = {
  weekly?: { day: string; time: string; timeZone: string };
};

type ApiMember = {
  type: "members";
  id: string;
  attributes: { position: number; color: string };
  relationships: { user: { data: { id: string } } };
};

type ApiUser = {
  type: "users";
  id: string;
  attributes: { name: string; email: string };
};

type ApiOverride = {
  type: "overrides";
  id: string;
  attributes: { start: string; end: string };
  relationships: { member: { data: { id: string } } };
};

type GetRotationResponse = {
  data?: {
    attributes: { name: string; cadence: Cadence };
    relationships: {
      members: { data: { id: string }[] };
      overrides: { data: { id: string }[] };
      scheduledMember?: { data: { id: string } | null };
    };
  };
  included?: (ApiMember | ApiUser | ApiOverride)[];
  errors?: { detail?: string }[];
};

type IncludedMaps = {
  memberMap: Map<string, ApiMember>;
  userMap: Map<string, ApiUser>;
  overrideMap: Map<string, ApiOverride>;
};

function buildIncludedMaps(
  included: (ApiMember | ApiUser | ApiOverride)[],
): IncludedMaps {
  const memberMap = new Map<string, ApiMember>();
  const userMap = new Map<string, ApiUser>();
  const overrideMap = new Map<string, ApiOverride>();

  for (const item of included) {
    if (item.type === "members") memberMap.set(item.id, item);
    else if (item.type === "users") userMap.set(item.id, item);
    else if (item.type === "overrides") overrideMap.set(item.id, item);
  }

  return { memberMap, userMap, overrideMap };
}

function loadMembers(
  memberRefs: { id: string }[],
  { memberMap, userMap }: IncludedMaps,
): Member[] {
  const sortedRefs = [...memberRefs].sort((a, b) => {
    const positionA = memberMap.get(a.id)?.attributes.position ?? 0;
    const positionB = memberMap.get(b.id)?.attributes.position ?? 0;
    return positionA - positionB;
  });

  return sortedRefs.flatMap((ref) => {
    const member = memberMap.get(ref.id);
    if (!member) return [];

    const userId = member.relationships.user.data.id;
    const user = userMap.get(userId);
    if (!user) return [];

    return [
      {
        id: ref.id,
        userId,
        name: user.attributes.name,
        email: user.attributes.email,
        ...colorsForName(member.attributes.color),
      },
    ];
  });
}

function cadenceSummary(cadence: Cadence): string {
  const { weekly } = cadence;
  if (!weekly) return "Unknown cadence";

  const { day, time, timeZone } = weekly;
  const [hourStr, minuteStr] = time.split(":");
  const hour = Number(hourStr);
  const minute = Number(minuteStr);

  const refDate = new Date(2024, 0, 1, hour, minute);
  const timeText = new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  }).format(refDate);

  const tzParts = new Intl.DateTimeFormat("en-US", {
    timeZone,
    timeZoneName: "shortGeneric",
  }).formatToParts(new Date());
  const tzText =
    tzParts.find((p) => p.type === "timeZoneName")?.value ?? timeZone;

  return `Rotates weekly on ${day}s at ${timeText} ${tzText}`;
}

function toDateTimeLocalValue(dateTime: string): string {
  const date = new Date(dateTime);
  if (Number.isNaN(date.getTime())) return dateTime;

  const pad = (value: number) => String(value).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate(),
  )}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function loadOverrides(
  overrideRefs: { id: string }[],
  { memberMap, overrideMap }: IncludedMaps,
): Override[] {
  return overrideRefs.flatMap((ref) => {
    const override = overrideMap.get(ref.id);
    if (!override) return [];

    const memberId = override.relationships.member.data.id;
    if (!memberMap.has(memberId)) return [];

    return [
      {
        id: override.id,
        start: toDateTimeLocalValue(override.attributes.start),
        end: toDateTimeLocalValue(override.attributes.end),
        memberId,
      },
    ];
  });
}

function Settings() {
  const { rotationId } = useParams({ from: "/rotations/$rotationId/settings" });
  const [rotationName, setRotationName] = useState<string | null>(null);
  const [cadence, setCadence] = useState<Cadence | null>(null);

  useBreadcrumbs([
    { label: "Rotations", to: "/rotations" },
    {
      label: rotationName ?? "…",
      to: "/rotations/$rotationId",
      params: { rotationId },
    },
    { label: "Settings" },
  ]);

  const { members, setMembers, overrides, setOverrides, setScheduledMemberId } =
    useAppState();

  useEffect(() => {
    let cancelled = false;

    function clearState() {
      if (cancelled) return;
      setRotationName(null);
      setCadence(null);
      setMembers([]);
      setOverrides([]);
      setScheduledMemberId(null);
    }

    void (async () => {
      try {
        const res = await fetch(`/api/rotations/${rotationId}`);
        const body = (await res.json()) as GetRotationResponse;
        if (!res.ok || !body.data) {
          clearState();
          return;
        }

        if (cancelled) return;

        const includedMaps = buildIncludedMaps(body.included ?? []);
        const loadedMembers = loadMembers(
          body.data.relationships.members.data,
          includedMaps,
        );
        const loadedOverrides = loadOverrides(
          body.data.relationships.overrides.data,
          includedMaps,
        );
        const scheduledMemberId =
          body.data.relationships.scheduledMember?.data?.id ?? null;

        setRotationName(body.data.attributes.name);
        setCadence(body.data.attributes.cadence);
        setMembers(loadedMembers);
        setOverrides(loadedOverrides);
        setScheduledMemberId(scheduledMemberId);
      } catch {
        clearState();
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [rotationId, setMembers, setOverrides, setScheduledMemberId]);

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      <PageHeader title="Settings" />
      <Card className="shadow-sm border-border bg-card">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold">Cadence</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            {cadence ? cadenceSummary(cadence) : "…"}
          </p>
        </CardContent>
      </Card>
      <Members
        members={members}
        setMembers={setMembers}
        overrides={overrides}
        setOverrides={setOverrides}
      />
      <Overrides
        members={members}
        overrides={overrides}
        setOverrides={setOverrides}
      />
    </div>
  );
}

export default Settings;

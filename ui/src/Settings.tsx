import { useParams } from "@tanstack/react-router";
import { useEffect, useState } from "react";

import { useAppState } from "./AppStateContext";
import { useBreadcrumbs } from "./BreadcrumbContext";
import { colorsForName } from "./colorPalette";
import Members from "./Members";
import Overrides from "./Overrides";
import PageHeader from "./PageHeader";
import type { Member, Override } from "./types";

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
    attributes: { name: string };
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

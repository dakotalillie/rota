import { useParams } from "@tanstack/react-router";
import { useEffect, useState } from "react";

import { useAppState } from "./AppStateContext";
import { useBreadcrumbs } from "./BreadcrumbContext";
import PageHeader from "./PageHeader";
import SettingsOverrides from "./SettingsOverrides";
import SettingsRotationOrder from "./SettingsRotationOrder";
import type { Member } from "./types";

type ApiMember = {
  type: "members";
  id: string;
  attributes: { order: number };
  relationships: { user: { data: { id: string } } };
};

type ApiUser = {
  type: "users";
  id: string;
  attributes: { name: string; email: string };
};

type GetRotationResponse = {
  data?: {
    attributes: { name: string };
    relationships: {
      members: { data: { id: string }[] };
    };
  };
  included?: (ApiMember | ApiUser)[];
  errors?: { detail?: string }[];
};

const COLOR_PALETTE = [
  {
    color: "bg-violet-500",
    lightColor: "bg-violet-50",
    darkColor: "dark:bg-violet-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-sky-500",
    lightColor: "bg-sky-50",
    darkColor: "dark:bg-sky-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-emerald-500",
    lightColor: "bg-emerald-50",
    darkColor: "dark:bg-emerald-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-orange-400",
    lightColor: "bg-orange-50",
    darkColor: "dark:bg-orange-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-rose-500",
    lightColor: "bg-rose-50",
    darkColor: "dark:bg-rose-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-teal-500",
    lightColor: "bg-teal-50",
    darkColor: "dark:bg-teal-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-amber-500",
    lightColor: "bg-amber-50",
    darkColor: "dark:bg-amber-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-pink-500",
    lightColor: "bg-pink-50",
    darkColor: "dark:bg-pink-950/50",
    textColor: "text-white",
  },
];

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

  const { members, setMembers, overrides, setOverrides } = useAppState();

  useEffect(() => {
    void (async () => {
      const res = await fetch(`/api/rotations/${rotationId}`);
      const body = (await res.json()) as GetRotationResponse;
      if (!res.ok || !body.data) return;

      setRotationName(body.data.attributes.name);

      const memberRefs = body.data.relationships.members.data;
      const included = body.included ?? [];

      const memberMap = new Map<string, ApiMember>();
      const userMap = new Map<string, ApiUser>();
      for (const item of included) {
        if (item.type === "members") memberMap.set(item.id, item);
        else if (item.type === "users") userMap.set(item.id, item);
      }

      const sortedRefs = [...memberRefs].sort((a, b) => {
        const orderA = memberMap.get(a.id)?.attributes.order ?? 0;
        const orderB = memberMap.get(b.id)?.attributes.order ?? 0;
        return orderA - orderB;
      });

      const loadedMembers: Member[] = sortedRefs.flatMap((ref, i) => {
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
            ...COLOR_PALETTE[i % COLOR_PALETTE.length],
          },
        ];
      });

      setMembers(loadedMembers);
    })();
  }, [rotationId, setMembers, setRotationName]);

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      <PageHeader title="Settings" />
      <SettingsRotationOrder
        members={members}
        setMembers={setMembers}
        overrides={overrides}
        setOverrides={setOverrides}
      />
      <SettingsOverrides
        members={members}
        overrides={overrides}
        setOverrides={setOverrides}
      />
    </div>
  );
}

export default Settings;

import { ArrowRight, ChevronDown, Plus } from "lucide-react";
import { useState } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "./Avatar";
import { Button } from "./Button";
import { Input } from "./Input";
import type { Member, Override } from "./types";
import {
  computeOverrideReplacements,
  formatDateTimeRange,
  initials,
} from "./utils";

function todayAt9am(): string {
  const d = new Date();
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T09:00`;
}

// --- Display helpers ---
const selectClass =
  "w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 transition-shadow";

type SettingsOverridesFormProps = {
  members: Member[];
  overrides: Override[];
  setOverrides: (overrides: Override[]) => void;
};

function SettingsOverridesForm({
  members,
  overrides,
  setOverrides,
}: SettingsOverridesFormProps) {
  const [overrideStart, setOverrideStart] = useState("");
  const [overrideEnd, setOverrideEnd] = useState("");
  const [overrideMemberId, setOverrideMemberId] = useState("");

  const validMemberId = members.find((m) => m.id === overrideMemberId)
    ? overrideMemberId
    : "";

  const overrideValid =
    overrideStart &&
    overrideEnd &&
    new Date(overrideEnd) > new Date(overrideStart) &&
    validMemberId;
  const formReplacements = overrideValid
    ? computeOverrideReplacements(
        members,
        overrides,
        overrideStart,
        overrideEnd,
      )
    : [];
  const overrideSelfAssign = formReplacements.some(
    (seg) => seg.member.id === validMemberId,
  );

  function addOverride() {
    if (!overrideValid) return;
    setOverrides([
      ...overrides,
      {
        id: crypto.randomUUID(),
        start: overrideStart,
        end: overrideEnd,
        memberId: validMemberId,
      },
    ]);
    setOverrideStart("");
    setOverrideEnd("");
    setOverrideMemberId("");
  }

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-2">
        <Input
          type="datetime-local"
          value={overrideStart}
          onFocus={() => {
            if (!overrideStart) setOverrideStart(todayAt9am());
          }}
          onChange={(e) => setOverrideStart(e.target.value)}
        />
        <ArrowRight className="h-4 w-4 text-muted-foreground shrink-0" />
        <Input
          type="datetime-local"
          value={overrideEnd}
          onFocus={() => {
            if (!overrideEnd) setOverrideEnd(todayAt9am());
          }}
          onChange={(e) => setOverrideEnd(e.target.value)}
        />
      </div>
      <div className="relative">
        <select
          value={validMemberId}
          onChange={(e) => setOverrideMemberId(e.target.value)}
          className={selectClass + " appearance-none pr-8"}
        >
          <option value="" disabled>
            Select person
          </option>
          {members.map((m) => (
            <option key={m.id} value={m.id}>
              {m.name}
            </option>
          ))}
        </select>
        <ChevronDown className="absolute right-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
      </div>
      {overrideValid && formReplacements.length > 0 && (
        <div
          className={`rounded-lg border px-3 py-2.5 space-y-1.5 ${overrideSelfAssign ? "border-destructive/50 bg-destructive/5" : "border-border bg-muted/30"}`}
        >
          <p
            className={`text-xs font-medium ${overrideSelfAssign ? "text-destructive" : "text-muted-foreground"}`}
          >
            Replaces
          </p>
          <div className="space-y-1">
            {formReplacements.map((seg, i) => {
              const isSelf = seg.member.id === validMemberId;
              return (
                <div key={i} className="flex items-center gap-2 text-sm">
                  <Avatar className="h-5 w-5 shrink-0">
                    <AvatarImage src={seg.member.avatarUrl} />
                    <AvatarFallback
                      className={`text-[9px] font-semibold ${seg.member.color} ${seg.member.textColor}`}
                    >
                      {initials(seg.member.name)}
                    </AvatarFallback>
                  </Avatar>
                  <span
                    className={`font-medium ${isSelf ? "text-destructive" : ""}`}
                  >
                    {seg.member.name}
                  </span>
                  <span className="text-muted-foreground text-xs">
                    {formatDateTimeRange(seg.start, seg.end)}
                  </span>
                  {isSelf && (
                    <span className="text-xs text-destructive">
                      already on call
                    </span>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}
      <Button
        onClick={addOverride}
        disabled={!overrideValid || overrideSelfAssign}
        className="gap-1.5"
      >
        <Plus />
        Add override
      </Button>
    </div>
  );
}

export default SettingsOverridesForm;

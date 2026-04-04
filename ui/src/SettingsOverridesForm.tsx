import { useState } from "react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Plus, ArrowRight, ChevronDown } from "lucide-react";
import type { Engineer, Override } from "./types";
import {
  computeOverrideReplacements,
  formatSegmentRange,
  initials,
  inputClass,
} from "./utils";

// --- Display helpers ---
const selectClass =
  "w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 transition-shadow";

type SettingsOverridesFormProps = {
  engineers: Engineer[];
  overrides: Override[];
  setOverrides: (overrides: Override[]) => void;
};

function SettingsOverridesForm({
  engineers,
  overrides,
  setOverrides,
}: SettingsOverridesFormProps) {
  const [overrideStart, setOverrideStart] = useState("");
  const [overrideEnd, setOverrideEnd] = useState("");
  const [overrideEngineerId, setOverrideEngineerId] = useState("");

  const validEngineerId = engineers.find((e) => e.id === overrideEngineerId)
    ? overrideEngineerId
    : "";

  const overrideValid =
    overrideStart &&
    overrideEnd &&
    new Date(overrideEnd) > new Date(overrideStart) &&
    validEngineerId;
  const formReplacements = overrideValid
    ? computeOverrideReplacements(
        engineers,
        overrides,
        overrideStart,
        overrideEnd,
      )
    : [];
  const overrideSelfAssign = formReplacements.some(
    (seg) => seg.engineer.id === validEngineerId,
  );

  function addOverride() {
    if (!overrideValid) return;
    setOverrides([
      ...overrides,
      {
        id: crypto.randomUUID(),
        start: overrideStart,
        end: overrideEnd,
        engineerId: validEngineerId,
      },
    ]);
    setOverrideStart("");
    setOverrideEnd("");
    setOverrideEngineerId("");
  }

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-2">
        <input
          type="datetime-local"
          value={overrideStart}
          onChange={(e) => setOverrideStart(e.target.value)}
          className={inputClass}
        />
        <ArrowRight className="h-4 w-4 text-muted-foreground shrink-0" />
        <input
          type="datetime-local"
          value={overrideEnd}
          onChange={(e) => setOverrideEnd(e.target.value)}
          className={inputClass}
        />
      </div>
      <div className="relative">
        <select
          value={validEngineerId}
          onChange={(e) => setOverrideEngineerId(e.target.value)}
          className={selectClass + " appearance-none pr-8"}
        >
          <option value="" disabled>
            Select person
          </option>
          {engineers.map((e) => (
            <option key={e.id} value={e.id}>
              {e.name}
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
              const isSelf = seg.engineer.id === validEngineerId;
              return (
                <div key={i} className="flex items-center gap-2 text-sm">
                  <Avatar className="h-5 w-5 shrink-0">
                    <AvatarImage src={seg.engineer.avatarUrl} />
                    <AvatarFallback
                      className={`text-[9px] font-semibold ${seg.engineer.color} ${seg.engineer.textColor}`}
                    >
                      {initials(seg.engineer.name)}
                    </AvatarFallback>
                  </Avatar>
                  <span
                    className={`font-medium ${isSelf ? "text-destructive" : ""}`}
                  >
                    {seg.engineer.name}
                  </span>
                  <span className="text-muted-foreground text-xs">
                    {formatSegmentRange(seg.start, seg.end)}
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

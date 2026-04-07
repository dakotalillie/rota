import { Dialog } from "@base-ui/react/dialog";
import { Tooltip } from "@base-ui/react/tooltip";
import { useParams } from "@tanstack/react-router";
import { ArrowRight, ChevronDown, Plus, X } from "lucide-react";
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

type CreateOverrideResponse = {
  data?: { id: string };
  errors?: { detail?: string }[];
};

function todayAt9am(): string {
  const d = new Date();
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T09:00`;
}

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
  const { rotationId } = useParams({ strict: false });
  const [open, setOpen] = useState(false);
  const [overrideStart, setOverrideStart] = useState("");
  const [overrideEnd, setOverrideEnd] = useState("");
  const [overrideMemberId, setOverrideMemberId] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function handleOpenChange(newOpen: boolean) {
    setOpen(newOpen);
    if (newOpen) {
      setOverrideStart("");
      setOverrideEnd("");
      setOverrideMemberId("");
      setError(null);
    }
  }

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

  async function addOverride() {
    if (!overrideValid) return;
    setSubmitting(true);
    setError(null);
    try {
      const res = await fetch(`/api/rotations/${rotationId}/overrides`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          data: {
            attributes: {
              start: new Date(overrideStart).toISOString(),
              end: new Date(overrideEnd).toISOString(),
            },
            relationships: {
              member: { data: { type: "members", id: validMemberId } },
            },
          },
        }),
      });
      const body = (await res.json()) as CreateOverrideResponse;
      if (!res.ok) {
        setError(body.errors?.[0]?.detail ?? `HTTP ${res.status}`);
        return;
      }
      setOverrides([
        ...overrides,
        {
          id: body.data!.id,
          start: overrideStart,
          end: overrideEnd,
          memberId: validMemberId,
        },
      ]);
      setOpen(false);
    } catch {
      setError("An unexpected error occurred");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={handleOpenChange}>
      <Tooltip.Root>
        <Tooltip.Trigger
          render={
            <span
              className={members.length === 0 ? "cursor-not-allowed" : ""}
            />
          }
        >
          <Dialog.Trigger
            render={
              <Button
                variant="outline"
                size="sm"
                className={`gap-1.5${members.length === 0 ? " pointer-events-none" : ""}`}
                disabled={members.length === 0}
                tabIndex={members.length === 0 ? -1 : undefined}
              >
                <Plus />
                Add override
              </Button>
            }
          />
        </Tooltip.Trigger>
        {members.length === 0 && (
          <Tooltip.Portal>
            <Tooltip.Positioner>
              <Tooltip.Popup className="rounded-md bg-popover border border-border px-2.5 py-1.5 text-xs text-popover-foreground shadow-md">
                Add members to the rotation first
              </Tooltip.Popup>
            </Tooltip.Positioner>
          </Tooltip.Portal>
        )}
      </Tooltip.Root>
      <Dialog.Portal>
        <Dialog.Backdrop className="fixed inset-0 bg-black/40 dark:bg-black/60 animate-in fade-in" />
        <Dialog.Viewport className="fixed inset-0 flex items-center justify-center p-4">
          <Dialog.Popup className="bg-card w-full max-w-lg rounded-xl border border-border shadow-lg animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between px-5 pt-5 pb-4">
              <Dialog.Title className="text-base font-semibold">
                Add override
              </Dialog.Title>
              <Dialog.Close
                render={
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label="Close"
                    className="text-muted-foreground hover:text-foreground"
                  >
                    <X />
                  </Button>
                }
              />
            </div>
            <div className="space-y-3 px-5 pb-5">
              <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-2">
                <Input
                  type="datetime-local"
                  value={overrideStart}
                  className="min-w-0"
                  onFocus={() => {
                    if (!overrideStart) setOverrideStart(todayAt9am());
                  }}
                  onChange={(e) => setOverrideStart(e.target.value)}
                />
                <ArrowRight className="h-4 w-4 text-muted-foreground shrink-0" />
                <Input
                  type="datetime-local"
                  value={overrideEnd}
                  className="min-w-0"
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
                        <div
                          key={i}
                          className="flex items-center gap-2 text-sm"
                        >
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
                onClick={() => void addOverride()}
                disabled={!overrideValid || overrideSelfAssign || submitting}
                size="sm"
                className="w-full gap-1.5"
              >
                <Plus />
                Add override
              </Button>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
          </Dialog.Popup>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export default SettingsOverridesForm;

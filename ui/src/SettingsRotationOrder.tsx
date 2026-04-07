import { useParams } from "@tanstack/react-router";
import { GripVertical, X } from "lucide-react";
import { useRef } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "./Avatar";
import { Button } from "./Button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "./Card";
import SettingsAddPerson from "./SettingsAddPerson";
import type { Member, Override } from "./types";
import { initials } from "./utils";

type SettingsRotationOrderProps = {
  members: Member[];
  setMembers: (members: Member[]) => void;
  overrides: Override[];
  setOverrides: (overrides: Override[]) => void;
};

function SettingsRotationOrder({
  members,
  setMembers,
  overrides,
  setOverrides,
}: SettingsRotationOrderProps) {
  const { rotationId } = useParams({ strict: false });
  const dragIndexRef = useRef<number | null>(null);
  const didReorderRef = useRef(false);

  function handleDragStart(index: number) {
    dragIndexRef.current = index;
  }

  function handleDragOver(e: React.DragEvent, index: number) {
    e.preventDefault();
    const from = dragIndexRef.current;
    if (from === null || from === index) return;
    const next = [...members];
    const [item] = next.splice(from, 1);
    next.splice(index, 0, item);
    dragIndexRef.current = index;
    didReorderRef.current = true;
    setMembers(next);
  }

  function handleDragEnd() {
    dragIndexRef.current = null;
    if (didReorderRef.current) {
      didReorderRef.current = false;
      void fetch(`/api/rotations/${rotationId}/members`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          data: members.map((m) => ({ type: "members", id: m.id })),
        }),
      });
    }
  }

  function removeMember(id: string) {
    setMembers(members.filter((m) => m.id !== id));
    setOverrides(overrides.filter((o) => o.memberId !== id));
  }

  return (
    <Card className="shadow-sm border-border bg-card">
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-semibold">Members</CardTitle>
        <CardAction>
          <SettingsAddPerson members={members} setMembers={setMembers} />
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-1">
        {members.length > 0 ? (
          members.map((member, index) => (
            <div
              key={member.id}
              draggable
              onDragStart={() => handleDragStart(index)}
              onDragOver={(e) => handleDragOver(e, index)}
              onDragEnd={handleDragEnd}
              className="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-colors bg-muted/40 hover:bg-muted/60 cursor-grab active:cursor-grabbing select-none"
            >
              <GripVertical className="h-4 w-4 text-muted-foreground shrink-0" />
              <Avatar className="h-8 w-8 shrink-0">
                <AvatarImage src={member.avatarUrl} />
                <AvatarFallback
                  className={`text-xs font-semibold ${member.color} ${member.textColor}`}
                >
                  {initials(member.name)}
                </AvatarFallback>
              </Avatar>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{member.name}</p>
                {member.email && (
                  <p className="text-xs text-muted-foreground truncate">
                    {member.email}
                  </p>
                )}
              </div>
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => removeMember(member.id)}
                className="shrink-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                aria-label={`Remove ${member.name}`}
              >
                <X />
              </Button>
            </div>
          ))
        ) : (
          <p className="text-sm text-muted-foreground px-1 py-1">
            No members yet.
          </p>
        )}
      </CardContent>
    </Card>
  );
}

export default SettingsRotationOrder;

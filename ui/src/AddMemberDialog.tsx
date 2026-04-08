import { Dialog } from "@base-ui/react/dialog";
import { useParams } from "@tanstack/react-router";
import { UserPlus, X } from "lucide-react";
import { useState } from "react";

import { Button } from "./Button";
import { colorsForName } from "./colorPalette";
import { Input } from "./Input";
import type { Member } from "./types";

type AddMemberDialogProps = {
  members: Member[];
  setMembers: (members: Member[]) => void;
};

type CreateMemberResponse = {
  data: {
    id: string;
    attributes: { color: string };
  };
  included: {
    id: string;
    attributes: { name: string; email: string };
  }[];
  errors?: { detail?: string }[];
};

function AddMemberDialog({ members, setMembers }: AddMemberDialogProps) {
  const { rotationId } = useParams({ strict: false });
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function handleOpenChange(newOpen: boolean) {
    setOpen(newOpen);
    if (newOpen) {
      setName("");
      setEmail("");
      setError(null);
    }
  }

  async function addMember() {
    const trimmedName = name.trim();
    const trimmedEmail = email.trim();
    if (!trimmedName || !trimmedEmail) return;
    setSubmitting(true);
    setError(null);
    try {
      const res = await fetch(`/api/rotations/${rotationId}/members`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          data: { attributes: { name: trimmedName, email: trimmedEmail } },
        }),
      });
      const body = (await res.json()) as CreateMemberResponse;
      if (!res.ok) {
        setError(body.errors?.[0]?.detail ?? `HTTP ${res.status}`);
        return;
      }
      const user = body.included[0];
      const newMember: Member = {
        id: body.data.id,
        userId: user.id,
        name: user.attributes.name,
        email: user.attributes.email,
        ...colorsForName(body.data.attributes.color),
      };
      setMembers([...members, newMember]);
      setOpen(false);
    } catch {
      setError("An unexpected error occurred");
    } finally {
      setSubmitting(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") void addMember();
  }

  return (
    <Dialog.Root open={open} onOpenChange={handleOpenChange}>
      <Dialog.Trigger
        render={
          <Button variant="outline" size="sm" className="gap-1.5">
            <UserPlus />
            Add member
          </Button>
        }
      />
      <Dialog.Portal>
        <Dialog.Backdrop className="fixed inset-0 bg-black/40 dark:bg-black/60 animate-in fade-in" />
        <Dialog.Viewport className="fixed inset-0 flex items-center justify-center p-4">
          <Dialog.Popup className="bg-card w-full max-w-sm rounded-xl border border-border shadow-lg animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between px-5 pt-5 pb-4">
              <Dialog.Title className="text-base font-semibold">
                Add person
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
            <div className="space-y-2 px-5 pb-5">
              <Input
                type="text"
                placeholder="Name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={handleKeyDown}
                autoFocus
              />
              <Input
                type="email"
                placeholder="Email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onKeyDown={handleKeyDown}
              />
              <Button
                onClick={() => void addMember()}
                disabled={!name.trim() || !email.trim() || submitting}
                size="sm"
                className="w-full gap-1.5"
              >
                <UserPlus />
                Add person
              </Button>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
          </Dialog.Popup>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export default AddMemberDialog;

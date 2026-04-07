import { Dialog } from "@base-ui/react/dialog";
import { useParams } from "@tanstack/react-router";
import { UserPlus, X } from "lucide-react";
import { useState } from "react";

import { Button } from "./Button";
import { Input } from "./Input";
import type { Engineer } from "./types";

type SettingsAddPersonProps = {
  engineers: Engineer[];
  setEngineers: (engineers: Engineer[]) => void;
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

type CreateMemberResponse = {
  data: {
    id: string;
  };
  included: {
    id: string;
    attributes: { name: string; email: string };
  }[];
  errors?: { detail?: string }[];
};

function SettingsAddPerson({
  engineers,
  setEngineers,
}: SettingsAddPersonProps) {
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

  async function addEngineer() {
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
      const palette = COLOR_PALETTE[engineers.length % COLOR_PALETTE.length];
      const newEngineer: Engineer = {
        id: user.id,
        name: user.attributes.name,
        email: user.attributes.email,
        ...palette,
      };
      setEngineers([...engineers, newEngineer]);
      setOpen(false);
    } catch {
      setError("An unexpected error occurred");
    } finally {
      setSubmitting(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") void addEngineer();
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
                onClick={() => void addEngineer()}
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

export default SettingsAddPerson;

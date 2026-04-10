import { Dialog } from "@base-ui/react/dialog";
import { Link } from "@tanstack/react-router";
import { Plus } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "./Button";
import { Input } from "./Input";
import PageHeader from "./PageHeader";

interface ApiRotation {
  type: "rotations";
  id: string;
  attributes: {
    name: string;
  };
  relationships: {
    currentMember: { data: { type: "members"; id: string } | null };
  };
}

interface ApiMember {
  type: "members";
  id: string;
  relationships: { user: { data: { type: "users"; id: string } } };
}

interface ApiUser {
  type: "users";
  id: string;
  attributes: { name: string; email: string };
}

interface ListRotationsResponse {
  data: ApiRotation[];
  included?: (ApiMember | ApiUser)[];
}

function currentMemberName(
  rotation: ApiRotation,
  included: (ApiMember | ApiUser)[] | undefined,
): string | null {
  const memberData = rotation.relationships.currentMember.data;
  if (!memberData || !included) return null;

  const member = included.find(
    (r): r is ApiMember => r.type === "members" && r.id === memberData.id,
  );
  if (!member) return null;

  const userId = member.relationships.user.data.id;
  const user = included.find(
    (r): r is ApiUser => r.type === "users" && r.id === userId,
  );
  return user?.attributes.name ?? null;
}

function RotationsList() {
  const [response, setResponse] = useState<ListRotationsResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [name, setName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  useEffect(() => {
    fetch("/api/rotations")
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json() as Promise<ListRotationsResponse>;
      })
      .then(setResponse)
      .catch((err: unknown) =>
        setError(err instanceof Error ? err.message : "Unknown error"),
      );
  }, []);

  function handleOpenChange(open: boolean) {
    setDialogOpen(open);
    if (!open) {
      setName("");
      setSubmitError(null);
    }
  }

  async function handleCreate() {
    const trimmed = name.trim();
    if (!trimmed) return;
    setSubmitting(true);
    setSubmitError(null);
    try {
      const res = await fetch("/api/rotations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ data: { attributes: { name: trimmed } } }),
      });
      if (!res.ok) {
        const body = (await res.json()) as { errors?: { detail?: string }[] };
        const detail = body.errors?.[0]?.detail ?? `HTTP ${res.status}`;
        setSubmitError(detail);
        return;
      }
      const body = (await res.json()) as { data: ApiRotation };
      setResponse((prev) => ({
        data: [...(prev?.data ?? []), body.data],
        included: prev?.included,
      }));
      setDialogOpen(false);
      setName("");
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setSubmitting(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") void handleCreate();
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      <Dialog.Root open={dialogOpen} onOpenChange={handleOpenChange}>
        <PageHeader
          title="Rotations"
          actions={
            <Dialog.Trigger render={<Button size="sm" className="gap-1.5" />}>
              <Plus />
              Create Rotation
            </Dialog.Trigger>
          }
        />
        <Dialog.Portal>
          <Dialog.Backdrop className="fixed inset-0 bg-black/40 dark:bg-black/60" />
          <Dialog.Popup className="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-sm rounded-xl border border-border bg-background p-6 shadow-lg space-y-4 outline-none">
            <Dialog.Title className="text-base font-semibold">
              Create Rotation
            </Dialog.Title>
            <Input
              type="text"
              placeholder="Rotation name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              onKeyDown={handleKeyDown}
              autoFocus
            />
            {submitError && (
              <p className="text-sm text-red-500">{submitError}</p>
            )}
            <div className="flex justify-end gap-2">
              <Dialog.Close render={<Button variant="outline" size="sm" />}>
                Cancel
              </Dialog.Close>
              <Button
                size="sm"
                disabled={!name.trim() || submitting}
                onClick={() => void handleCreate()}
              >
                Create
              </Button>
            </div>
          </Dialog.Popup>
        </Dialog.Portal>
      </Dialog.Root>

      {error && (
        <p className="text-sm text-red-500">
          Failed to load rotations: {error}
        </p>
      )}

      {!error && !response && (
        <p className="text-sm text-neutral-500">Loading…</p>
      )}

      {response && (
        <div className="space-y-2">
          {response.data.length === 0 && (
            <p className="text-sm text-neutral-500">No rotations yet.</p>
          )}
          {response.data.map((rotation) => {
            const memberName = currentMemberName(rotation, response.included);
            return (
              <Link
                key={rotation.id}
                to="/rotations/$rotationId"
                params={{ rotationId: rotation.id }}
                className="block rounded-xl border border-neutral-200 dark:border-neutral-800 px-4 py-3 hover:bg-neutral-50 dark:hover:bg-neutral-900 transition-colors"
              >
                <div className="flex items-center justify-between gap-4">
                  <div className="min-w-0">
                    <p className="font-medium truncate">
                      {rotation.attributes.name}
                    </p>
                  </div>
                  <div className="shrink-0 text-sm text-neutral-500">
                    {memberName ? (
                      <span>
                        <span className="text-neutral-400 mr-1">On call:</span>
                        {memberName}
                      </span>
                    ) : (
                      <span className="text-neutral-400">
                        No current member
                      </span>
                    )}
                  </div>
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}

export default RotationsList;

import { Dialog } from "@base-ui/react/dialog";
import { Link } from "@tanstack/react-router";
import { Plus, Trash2 } from "lucide-react";
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

  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);
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

  function handleDeleteOpenChange(open: boolean) {
    if (!open) {
      setDeletingId(null);
      setDeleteError(null);
    }
  }

  async function handleDeleteConfirm() {
    if (!deletingId) return;
    setDeleting(true);
    setDeleteError(null);
    try {
      const res = await fetch(`/api/rotations/${deletingId}`, {
        method: "DELETE",
      });
      if (!res.ok) {
        const body = (await res.json()) as { errors?: { detail?: string }[] };
        const detail = body.errors?.[0]?.detail ?? `HTTP ${res.status}`;
        setDeleteError(detail);
        return;
      }
      setResponse((prev) =>
        prev
          ? { ...prev, data: prev.data.filter((r) => r.id !== deletingId) }
          : prev,
      );
      setDeletingId(null);
    } catch (err) {
      setDeleteError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setDeleting(false);
    }
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

      <Dialog.Root
        open={deletingId !== null}
        onOpenChange={handleDeleteOpenChange}
      >
        <Dialog.Portal>
          <Dialog.Backdrop className="fixed inset-0 bg-black/40 dark:bg-black/60" />
          <Dialog.Popup className="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-sm rounded-xl border border-border bg-background p-6 shadow-lg space-y-4 outline-none">
            <Dialog.Title className="text-base font-semibold">
              Delete rotation?
            </Dialog.Title>
            <p className="text-sm text-neutral-500">
              This will permanently delete{" "}
              <span className="font-medium text-foreground">
                {
                  response?.data.find((r) => r.id === deletingId)?.attributes
                    .name
                }
              </span>
              .
            </p>
            {deleteError && (
              <p className="text-sm text-red-500">{deleteError}</p>
            )}
            <div className="flex justify-end gap-2">
              <Dialog.Close render={<Button variant="outline" size="sm" />}>
                Cancel
              </Dialog.Close>
              <Button
                variant="destructive"
                size="sm"
                disabled={deleting}
                onClick={() => void handleDeleteConfirm()}
              >
                Delete
              </Button>
            </div>
          </Dialog.Popup>
        </Dialog.Portal>
      </Dialog.Root>

      {response && (
        <div className="space-y-2">
          {response.data.length === 0 && (
            <p className="text-sm text-neutral-500">No rotations yet.</p>
          )}
          {response.data.map((rotation) => {
            const memberName = currentMemberName(rotation, response.included);
            return (
              <div
                key={rotation.id}
                className="flex items-center rounded-xl border border-neutral-200 dark:border-neutral-800 hover:bg-neutral-50 dark:hover:bg-neutral-900 transition-colors"
              >
                <Link
                  to="/rotations/$rotationId"
                  params={{ rotationId: rotation.id }}
                  className="flex-1 min-w-0 px-4 py-3"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <p className="font-medium truncate">
                      {rotation.attributes.name}
                    </p>
                    <span className="shrink-0 text-sm text-neutral-500">
                      {memberName ? (
                        <>
                          <span className="text-neutral-400 mr-1">
                            On call:
                          </span>
                          {memberName}
                        </>
                      ) : (
                        <span className="text-neutral-400">
                          No current member
                        </span>
                      )}
                    </span>
                  </div>
                </Link>
                <div className="px-2">
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={`Delete ${rotation.attributes.name}`}
                    onClick={() => setDeletingId(rotation.id)}
                  >
                    <Trash2 />
                  </Button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

export default RotationsList;

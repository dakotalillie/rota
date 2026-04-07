import { Link, useParams } from "@tanstack/react-router";

function OnCallHeroEmpty() {
  const { rotationId } = useParams({ strict: false });

  return (
    <div className="rounded-2xl border border-dashed border-border p-10 text-center text-sm text-muted-foreground">
      This rotation doesn't have any members yet.{" "}
      <Link
        to="/rotations/$rotationId/settings"
        params={{ rotationId: rotationId! }}
        className="underline underline-offset-4 hover:text-foreground transition-colors"
      >
        Add some.
      </Link>
    </div>
  );
}

export default OnCallHeroEmpty;

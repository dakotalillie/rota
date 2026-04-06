import { Link } from "@tanstack/react-router";

function HomeHeroEmpty() {
  return (
    <div className="rounded-2xl border border-dashed border-border p-10 text-center text-sm text-muted-foreground">
      No engineers in the rotation yet.{" "}
      <Link
        to="/settings"
        className="underline underline-offset-4 hover:text-foreground transition-colors"
      >
        Add some.
      </Link>
    </div>
  );
}

export default HomeHeroEmpty;

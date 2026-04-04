type HomeHeroEmptyProps = {
  onNavigateEdit: () => void;
};

function HomeHeroEmpty({ onNavigateEdit }: HomeHeroEmptyProps) {
  return (
    <div className="rounded-2xl border border-dashed border-border p-10 text-center text-sm text-muted-foreground">
      No engineers in the rotation yet.{" "}
      <button
        onClick={onNavigateEdit}
        className="underline underline-offset-4 hover:text-foreground transition-colors"
      >
        Add some.
      </button>
    </div>
  );
}

export default HomeHeroEmpty;

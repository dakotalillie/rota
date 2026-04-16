import AddOverrideDialog from "./AddOverrideDialog";
import { Card, CardContent, CardHeader, CardTitle } from "./Card";
import OverridesList from "./OverridesList";
import type { Member, Override, TimeSegment } from "./types";

type OverridesProps = {
  members: Member[];
  overrides: Override[];
  setOverrides: (overrides: Override[]) => void;
  schedule: TimeSegment[];
};

function Overrides({
  members,
  overrides,
  setOverrides,
  schedule,
}: OverridesProps) {
  return (
    <Card className="shadow-sm border-border bg-card">
      <CardHeader className="flex flex-row items-center justify-between pb-3">
        <CardTitle className="text-base font-semibold">Overrides</CardTitle>
        <AddOverrideDialog
          members={members}
          overrides={overrides}
          setOverrides={setOverrides}
          schedule={schedule}
        />
      </CardHeader>
      <CardContent>
        {overrides.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No overrides scheduled.
          </p>
        ) : (
          <OverridesList
            members={members}
            overrides={overrides}
            setOverrides={setOverrides}
            schedule={schedule}
          />
        )}
      </CardContent>
    </Card>
  );
}

export default Overrides;

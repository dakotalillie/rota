import { Card, CardContent, CardHeader, CardTitle } from "./Card";
import SettingsOverridesForm from "./SettingsOverridesForm";
import SettingsOverridesList from "./SettingsOverridesList";
import type { Member, Override } from "./types";

type SettingsOverridesProps = {
  members: Member[];
  overrides: Override[];
  setOverrides: (overrides: Override[]) => void;
};

function SettingsOverrides({
  members,
  overrides,
  setOverrides,
}: SettingsOverridesProps) {
  return (
    <Card className="shadow-sm border-border bg-card">
      <CardHeader className="flex flex-row items-center justify-between pb-3">
        <CardTitle className="text-base font-semibold">Overrides</CardTitle>
        <SettingsOverridesForm
          members={members}
          overrides={overrides}
          setOverrides={setOverrides}
        />
      </CardHeader>
      <CardContent>
        {overrides.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No overrides scheduled.
          </p>
        ) : (
          <SettingsOverridesList
            members={members}
            overrides={overrides}
            setOverrides={setOverrides}
          />
        )}
      </CardContent>
    </Card>
  );
}

export default SettingsOverrides;

import { Card, CardContent, CardHeader, CardTitle } from "./Card";
import SettingsOverridesForm from "./SettingsOverridesForm";
import SettingsOverridesList from "./SettingsOverridesList";
import type { Engineer, Override } from "./types";

type SettingsOverridesProps = {
  engineers: Engineer[];
  overrides: Override[];
  setOverrides: (overrides: Override[]) => void;
};

function SettingsOverrides({
  engineers,
  overrides,
  setOverrides,
}: SettingsOverridesProps) {
  return (
    <Card className="shadow-sm border-border bg-card">
      <CardHeader className="flex flex-row items-center justify-between pb-3">
        <CardTitle className="text-base font-semibold">Overrides</CardTitle>
        <SettingsOverridesForm
          engineers={engineers}
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
            engineers={engineers}
            overrides={overrides}
            setOverrides={setOverrides}
          />
        )}
      </CardContent>
    </Card>
  );
}

export default SettingsOverrides;

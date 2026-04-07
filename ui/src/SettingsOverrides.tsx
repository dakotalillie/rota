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
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-semibold">Overrides</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {overrides.length > 0 && (
          <SettingsOverridesList
            members={members}
            overrides={overrides}
            setOverrides={setOverrides}
          />
        )}
        {members.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Add engineers to the rotation before creating overrides.
          </p>
        ) : (
          <SettingsOverridesForm
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

import { useParams } from "@tanstack/react-router";
import { useEffect, useState } from "react";

import { useAppState } from "./AppStateContext";
import { useBreadcrumbs } from "./BreadcrumbContext";
import PageHeader from "./PageHeader";
import SettingsAddPerson from "./SettingsAddPerson";
import SettingsOverrides from "./SettingsOverrides";
import SettingsRotationOrder from "./SettingsRotationOrder";
import SettingsWebhooks from "./SettingsWebhooks";

function Settings() {
  const { rotationId } = useParams({ from: "/rotations/$rotationId/settings" });
  const [rotationName, setRotationName] = useState<string | null>(null);

  useEffect(() => {
    fetch(`/api/rotations/${rotationId}`)
      .then((res) => (res.ok ? res.json() : Promise.reject()))
      .then((body: { data?: { attributes: { name: string } } }) => {
        if (body.data) setRotationName(body.data.attributes.name);
      })
      .catch(() => {});
  }, [rotationId]);

  useBreadcrumbs([
    { label: "Rotations", to: "/rotations" },
    {
      label: rotationName ?? "…",
      to: "/rotations/$rotationId",
      params: { rotationId },
    },
    { label: "Settings" },
  ]);

  const {
    engineers,
    setEngineers,
    overrides,
    setOverrides,
    webhooks,
    setWebhooks,
  } = useAppState();

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      <PageHeader title="Settings" />
      <SettingsRotationOrder
        engineers={engineers}
        setEngineers={setEngineers}
        overrides={overrides}
        setOverrides={setOverrides}
      />
      <SettingsAddPerson engineers={engineers} setEngineers={setEngineers} />
      <SettingsOverrides
        engineers={engineers}
        overrides={overrides}
        setOverrides={setOverrides}
      />
      <SettingsWebhooks webhooks={webhooks} setWebhooks={setWebhooks} />
    </div>
  );
}

export default Settings;

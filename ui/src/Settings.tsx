import type { Engineer, Override, WebhookEntry } from "./types";
import SettingsRotationOrder from "./SettingsRotationOrder";
import SettingsAddPerson from "./SettingsAddPerson";
import SettingsHeader from "./SettingsHeader";
import SettingsOverrides from "./SettingsOverrides";
import SettingsWebhooks from "./SettingsWebhooks";

type SettingsProps = {
  engineers: Engineer[];
  setEngineers: (engineers: Engineer[]) => void;
  overrides: Override[];
  setOverrides: (overrides: Override[]) => void;
  webhooks: WebhookEntry[];
  setWebhooks: (webhooks: WebhookEntry[]) => void;
  onNavigateHome: () => void;
};

function Settings({
  engineers,
  setEngineers,
  overrides,
  setOverrides,
  webhooks,
  setWebhooks,
  onNavigateHome,
}: SettingsProps) {
  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-8">
      <SettingsHeader onNavigateHome={onNavigateHome} />
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

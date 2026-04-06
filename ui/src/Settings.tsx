import { useAppState } from "./AppStateContext";
import SettingsAddPerson from "./SettingsAddPerson";
import SettingsHeader from "./SettingsHeader";
import SettingsOverrides from "./SettingsOverrides";
import SettingsRotationOrder from "./SettingsRotationOrder";
import SettingsWebhooks from "./SettingsWebhooks";

function Settings() {
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
      <SettingsHeader />
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

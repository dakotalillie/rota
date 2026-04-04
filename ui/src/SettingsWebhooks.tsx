import { Card, CardContent, CardHeader, CardTitle } from "./Card";
import SettingsWebhooksForm from "./SettingsWebhooksForm";
import SettingsWebhooksList from "./SettingsWebhooksList";
import type { WebhookEntry } from "./types";

type SettingsWebhooksProps = {
  webhooks: WebhookEntry[];
  setWebhooks: (webhooks: WebhookEntry[]) => void;
};

function SettingsWebhooks({ webhooks, setWebhooks }: SettingsWebhooksProps) {
  return (
    <Card className="shadow-sm border-border bg-card">
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-semibold">Webhooks</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <SettingsWebhooksList webhooks={webhooks} setWebhooks={setWebhooks} />
        <SettingsWebhooksForm webhooks={webhooks} setWebhooks={setWebhooks} />
      </CardContent>
    </Card>
  );
}

export default SettingsWebhooks;

import { Button } from "@/components/ui/button";
import { X, Webhook } from "lucide-react";
import type { WebhookEntry } from "./types";

type SettingsWebhooksListProps = {
  webhooks: WebhookEntry[];
  setWebhooks: (webhooks: WebhookEntry[]) => void;
};

function SettingsWebhooksList({
  webhooks,
  setWebhooks,
}: SettingsWebhooksListProps) {
  function removeWebhook(id: string) {
    setWebhooks(webhooks.filter((w) => w.id !== id));
  }

  return (
    webhooks.length > 0 && (
      <div className="space-y-1">
        {webhooks.map((wh) => (
          <div
            key={wh.id}
            className="flex items-center gap-3 px-3 py-2.5 rounded-xl bg-muted/40"
          >
            <Webhook className="h-4 w-4 text-muted-foreground shrink-0" />
            <div className="flex-1 min-w-0">
              {wh.label && (
                <p className="text-sm font-medium truncate">{wh.label}</p>
              )}
              <p className="text-xs text-muted-foreground truncate">{wh.url}</p>
            </div>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => removeWebhook(wh.id)}
              className="shrink-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
              aria-label="Remove webhook"
            >
              <X />
            </Button>
          </div>
        ))}
      </div>
    )
  );
}

export default SettingsWebhooksList;

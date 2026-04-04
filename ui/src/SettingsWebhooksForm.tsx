import { Plus } from "lucide-react";
import { useState } from "react";

import { Button } from "./Button";
import type { WebhookEntry } from "./types";
import { inputClass } from "./utils";

type SettingsWebhooksFormProps = {
  webhooks: WebhookEntry[];
  setWebhooks: (webhooks: WebhookEntry[]) => void;
};

function SettingsWebhooksForm({
  webhooks,
  setWebhooks,
}: SettingsWebhooksFormProps) {
  const [webhookUrl, setWebhookUrl] = useState("");
  const [webhookLabel, setWebhookLabel] = useState("");

  const webhookUrlValid = (() => {
    try {
      new URL(webhookUrl);
      return true;
    } catch {
      return false;
    }
  })();

  function addWebhook() {
    if (!webhookUrlValid) return;
    setWebhooks([
      ...webhooks,
      {
        id: crypto.randomUUID(),
        url: webhookUrl.trim(),
        label: webhookLabel.trim(),
      },
    ]);
    setWebhookUrl("");
    setWebhookLabel("");
  }

  function handleWebhookKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") addWebhook();
  }

  return (
    <div className="space-y-2">
      <input
        type="text"
        placeholder="Label (optional)"
        value={webhookLabel}
        onChange={(e) => setWebhookLabel(e.target.value)}
        onKeyDown={handleWebhookKeyDown}
        className={inputClass}
      />
      <input
        type="url"
        placeholder="https://example.com/webhook"
        value={webhookUrl}
        onChange={(e) => setWebhookUrl(e.target.value)}
        onKeyDown={handleWebhookKeyDown}
        className={inputClass}
      />
      <Button
        onClick={addWebhook}
        disabled={!webhookUrlValid}
        size="sm"
        className="gap-1.5"
      >
        <Plus />
        Add webhook
      </Button>
    </div>
  );
}

export default SettingsWebhooksForm;

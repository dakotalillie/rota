/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useState } from "react";

import type { Engineer, Override, WebhookEntry } from "./types";

type AppState = {
  engineers: Engineer[];
  setEngineers: (engineers: Engineer[]) => void;
  overrides: Override[];
  setOverrides: (overrides: Override[]) => void;
  webhooks: WebhookEntry[];
  setWebhooks: (webhooks: WebhookEntry[]) => void;
};

const AppStateContext = createContext<AppState | null>(null);

export function AppStateProvider({ children }: { children: React.ReactNode }) {
  const [engineers, setEngineers] = useState<Engineer[]>([]);
  const [overrides, setOverrides] = useState<Override[]>([]);
  const [webhooks, setWebhooks] = useState<WebhookEntry[]>([]);

  return (
    <AppStateContext.Provider
      value={{
        engineers,
        setEngineers,
        overrides,
        setOverrides,
        webhooks,
        setWebhooks,
      }}
    >
      {children}
    </AppStateContext.Provider>
  );
}

export function useAppState(): AppState {
  const ctx = useContext(AppStateContext);
  if (!ctx) throw new Error("useAppState must be used within AppStateProvider");
  return ctx;
}

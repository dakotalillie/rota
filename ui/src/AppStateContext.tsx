/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useState } from "react";

import type { Engineer, Override, WebhookEntry } from "./types";

const INITIAL_ENGINEERS: Engineer[] = [
  {
    id: "1",
    name: "Alex Rivera",
    email: "alex@example.com",
    color: "bg-violet-500",
    lightColor: "bg-violet-50",
    darkColor: "dark:bg-violet-950/50",
    textColor: "text-white",
  },
  {
    id: "2",
    name: "Jordan Kim",
    email: "jordan@example.com",
    color: "bg-sky-500",
    lightColor: "bg-sky-50",
    darkColor: "dark:bg-sky-950/50",
    textColor: "text-white",
  },
  {
    id: "3",
    name: "Sam Patel",
    email: "sam@example.com",
    color: "bg-emerald-500",
    lightColor: "bg-emerald-50",
    darkColor: "dark:bg-emerald-950/50",
    textColor: "text-white",
  },
  {
    id: "4",
    name: "Casey Morgan",
    email: "casey@example.com",
    color: "bg-orange-400",
    lightColor: "bg-orange-50",
    darkColor: "dark:bg-orange-950/50",
    textColor: "text-white",
  },
  {
    id: "5",
    name: "Taylor Brooks",
    email: "taylor@example.com",
    color: "bg-rose-500",
    lightColor: "bg-rose-50",
    darkColor: "dark:bg-rose-950/50",
    textColor: "text-white",
  },
];

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
  const [engineers, setEngineers] = useState<Engineer[]>(INITIAL_ENGINEERS);
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

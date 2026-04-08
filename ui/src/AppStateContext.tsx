/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useState } from "react";

import type { Member, Override } from "./types";

type AppState = {
  members: Member[];
  setMembers: (members: Member[]) => void;
  overrides: Override[];
  setOverrides: (overrides: Override[]) => void;
  scheduledMemberId: string | null;
  setScheduledMemberId: (scheduledMemberId: string | null) => void;
};

const AppStateContext = createContext<AppState | null>(null);

export function AppStateProvider({ children }: { children: React.ReactNode }) {
  const [members, setMembers] = useState<Member[]>([]);
  const [overrides, setOverrides] = useState<Override[]>([]);
  const [scheduledMemberId, setScheduledMemberId] = useState<string | null>(
    null,
  );

  return (
    <AppStateContext.Provider
      value={{
        members,
        setMembers,
        overrides,
        setOverrides,
        scheduledMemberId,
        setScheduledMemberId,
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

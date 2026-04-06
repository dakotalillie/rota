import { createContext, useContext, useEffect } from "react";

export type BreadcrumbItem = {
  label: string;
  to?: string;
  params?: Record<string, string>;
};

type ContextValue = {
  setBreadcrumbs: (items: BreadcrumbItem[]) => void;
};

export const BreadcrumbContext = createContext<ContextValue>({
  setBreadcrumbs: () => {},
});

export function useBreadcrumbs(items: BreadcrumbItem[]) {
  const { setBreadcrumbs } = useContext(BreadcrumbContext);
  const key = JSON.stringify(items);
  useEffect(() => {
    setBreadcrumbs(JSON.parse(key) as BreadcrumbItem[]);
    return () => setBreadcrumbs([]);
  }, [key, setBreadcrumbs]);
}

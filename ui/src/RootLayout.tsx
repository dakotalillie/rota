import { Outlet } from "@tanstack/react-router";
import { Link } from "@tanstack/react-router";
import { useState } from "react";

import { BreadcrumbContext, type BreadcrumbItem } from "./BreadcrumbContext";
import Navbar from "./Navbar";

function BreadcrumbStrip({ breadcrumbs }: { breadcrumbs: BreadcrumbItem[] }) {
  return (
    <div className="bg-background">
      <nav className="max-w-2xl mx-auto px-4 py-2 flex items-center gap-1.5 text-sm text-muted-foreground">
        {breadcrumbs.map((crumb, i) => (
          <span key={i} className="flex items-center gap-1.5">
            {i > 0 && <span aria-hidden="true">›</span>}
            {crumb.to ? (
              <Link
                to={crumb.to}
                params={crumb.params as never}
                className="hover:text-foreground transition-colors"
              >
                {crumb.label}
              </Link>
            ) : (
              <span>{crumb.label}</span>
            )}
          </span>
        ))}
      </nav>
    </div>
  );
}

export default function RootLayout() {
  const [breadcrumbs, setBreadcrumbs] = useState<BreadcrumbItem[]>([]);

  return (
    <BreadcrumbContext.Provider value={{ setBreadcrumbs }}>
      <Navbar />
      {breadcrumbs.length > 0 && <BreadcrumbStrip breadcrumbs={breadcrumbs} />}
      <Outlet />
    </BreadcrumbContext.Provider>
  );
}

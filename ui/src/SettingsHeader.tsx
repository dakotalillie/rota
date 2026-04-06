import { Link } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";

import { Button } from "./Button";

function SettingsHeader() {
  return (
    <div className="flex items-center gap-3">
      <Link to="/">
        <Button variant="ghost" size="icon-sm" aria-label="Back to home">
          <ArrowLeft />
        </Button>
      </Link>
      <h1 className="text-xl font-bold tracking-tight">Settings</h1>
    </div>
  );
}

export default SettingsHeader;

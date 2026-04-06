import { Link, useParams } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";

import { Button } from "./Button";

function SettingsHeader() {
  const { rotationId } = useParams({ strict: false });

  return (
    <div className="flex items-center gap-3">
      <Link to="/rotations/$rotationId" params={{ rotationId: rotationId! }}>
        <Button variant="ghost" size="icon-sm" aria-label="Back to rotation">
          <ArrowLeft />
        </Button>
      </Link>
      <h1 className="text-xl font-bold tracking-tight">Settings</h1>
    </div>
  );
}

export default SettingsHeader;

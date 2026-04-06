import { Link, useParams } from "@tanstack/react-router";
import { Settings2 } from "lucide-react";

import { Button } from "./Button";

function HomeHeader() {
  const { rotationId } = useParams({ strict: false });

  return (
    <div className="flex items-center justify-between">
      <Link to="/" className="text-xl font-bold tracking-tight">
        Rota
      </Link>
      <Link
        to="/rotations/$rotationId/settings"
        params={{ rotationId: rotationId! }}
      >
        <Button variant="outline" size="sm" className="gap-1.5">
          <Settings2 />
          Settings
        </Button>
      </Link>
    </div>
  );
}

export default HomeHeader;

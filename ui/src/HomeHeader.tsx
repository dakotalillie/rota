import { Link } from "@tanstack/react-router";
import { Settings2 } from "lucide-react";

import { Button } from "./Button";

function HomeHeader() {
  return (
    <div className="flex items-center justify-between">
      <h1 className="text-xl font-bold tracking-tight">Rota</h1>
      <Link to="/settings">
        <Button variant="outline" size="sm" className="gap-1.5">
          <Settings2 />
          Settings
        </Button>
      </Link>
    </div>
  );
}

export default HomeHeader;

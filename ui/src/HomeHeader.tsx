import { Settings2 } from "lucide-react";
import { Button } from "@/components/ui/button";

type HomeHeaderProps = {
  onNavigateEdit: () => void;
};

function HomeHeader({ onNavigateEdit }: HomeHeaderProps) {
  return (
    <div className="flex items-center justify-between">
      <h1 className="text-xl font-bold tracking-tight">Rota</h1>
      <Button
        variant="outline"
        size="sm"
        onClick={onNavigateEdit}
        className="gap-1.5"
      >
        <Settings2 />
        Settings
      </Button>
    </div>
  );
}

export default HomeHeader;

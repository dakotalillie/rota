import { Link } from "@tanstack/react-router";

export default function Navbar() {
  return (
    <header className="border-b border-border bg-background">
      <div className="max-w-2xl mx-auto px-4 h-14 flex items-center justify-between">
        <Link
          to="/rotations"
          className="text-lg font-bold tracking-tight hover:opacity-70 transition-opacity"
        >
          Rota
        </Link>
        <nav className="flex items-center gap-4">
          <Link
            to="/rotations"
            className="text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            Rotations
          </Link>
        </nav>
      </div>
    </header>
  );
}

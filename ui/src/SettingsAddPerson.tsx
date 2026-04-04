import { Plus } from "lucide-react";
import { useState } from "react";

import { Button } from "./Button";
import { Card, CardContent, CardHeader, CardTitle } from "./Card";
import { Input } from "./Input";
import type { Engineer } from "./types";

type SettingsAddPersonProps = {
  engineers: Engineer[];
  setEngineers: (engineers: Engineer[]) => void;
};

const COLOR_PALETTE = [
  {
    color: "bg-violet-500",
    lightColor: "bg-violet-50",
    darkColor: "dark:bg-violet-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-sky-500",
    lightColor: "bg-sky-50",
    darkColor: "dark:bg-sky-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-emerald-500",
    lightColor: "bg-emerald-50",
    darkColor: "dark:bg-emerald-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-orange-400",
    lightColor: "bg-orange-50",
    darkColor: "dark:bg-orange-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-rose-500",
    lightColor: "bg-rose-50",
    darkColor: "dark:bg-rose-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-teal-500",
    lightColor: "bg-teal-50",
    darkColor: "dark:bg-teal-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-amber-500",
    lightColor: "bg-amber-50",
    darkColor: "dark:bg-amber-950/50",
    textColor: "text-white",
  },
  {
    color: "bg-pink-500",
    lightColor: "bg-pink-50",
    darkColor: "dark:bg-pink-950/50",
    textColor: "text-white",
  },
];

function SettingsAddPerson({
  engineers,
  setEngineers,
}: SettingsAddPersonProps) {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");

  function addEngineer() {
    const trimmedName = name.trim();
    if (!trimmedName) return;
    const palette = COLOR_PALETTE[engineers.length % COLOR_PALETTE.length];
    const newEngineer: Engineer = {
      id: crypto.randomUUID(),
      name: trimmedName,
      email: email.trim(),
      ...palette,
    };
    setEngineers([...engineers, newEngineer]);
    setName("");
    setEmail("");
  }

  function handleAddEngineerKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") addEngineer();
  }

  return (
    <Card className="shadow-sm border-border bg-card">
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-semibold">Add person</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        <Input
          type="text"
          placeholder="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={handleAddEngineerKeyDown}
        />
        <Input
          type="email"
          placeholder="Email (optional)"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          onKeyDown={handleAddEngineerKeyDown}
        />
        <Button
          onClick={addEngineer}
          disabled={!name.trim()}
          size="sm"
          className="gap-1.5"
        >
          <Plus />
          Add person
        </Button>
      </CardContent>
    </Card>
  );
}

export default SettingsAddPerson;

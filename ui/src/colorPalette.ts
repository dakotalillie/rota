type MemberColors = {
  color: string;
  lightColor: string;
  darkColor: string;
  textColor: string;
};

const COLOR_MAP: Record<string, MemberColors> = {
  violet: {
    color: "bg-violet-500",
    lightColor: "bg-violet-50",
    darkColor: "dark:bg-violet-950/50",
    textColor: "text-white",
  },
  sky: {
    color: "bg-sky-500",
    lightColor: "bg-sky-50",
    darkColor: "dark:bg-sky-950/50",
    textColor: "text-white",
  },
  emerald: {
    color: "bg-emerald-500",
    lightColor: "bg-emerald-50",
    darkColor: "dark:bg-emerald-950/50",
    textColor: "text-white",
  },
  orange: {
    color: "bg-orange-400",
    lightColor: "bg-orange-50",
    darkColor: "dark:bg-orange-950/50",
    textColor: "text-white",
  },
  rose: {
    color: "bg-rose-500",
    lightColor: "bg-rose-50",
    darkColor: "dark:bg-rose-950/50",
    textColor: "text-white",
  },
  teal: {
    color: "bg-teal-500",
    lightColor: "bg-teal-50",
    darkColor: "dark:bg-teal-950/50",
    textColor: "text-white",
  },
  amber: {
    color: "bg-amber-500",
    lightColor: "bg-amber-50",
    darkColor: "dark:bg-amber-950/50",
    textColor: "text-white",
  },
  pink: {
    color: "bg-pink-500",
    lightColor: "bg-pink-50",
    darkColor: "dark:bg-pink-950/50",
    textColor: "text-white",
  },
};

export function colorsForName(name: string): MemberColors {
  return COLOR_MAP[name] ?? COLOR_MAP.violet;
}

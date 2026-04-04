import { Avatar, AvatarFallback, AvatarImage } from "./components/ui/avatar";
import { Badge } from "./components/ui/badge";
import { Separator } from "./components/ui/separator";
import type { TimeSegment } from "./types";
import { formatSegmentRange, initials } from "./utils";

type HomeScheduleRowProps = {
  segment: TimeSegment;
  index: number;
};

function isActiveNow(seg: TimeSegment): boolean {
  const now = new Date();
  return now >= seg.start && now < seg.end;
}

function HomeScheduleRow({ segment, index }: HomeScheduleRowProps) {
  const isActive = isActiveNow(segment);
  const { engineer } = segment;

  return (
    <div
      className={`flex items-center gap-4 px-4 py-3 rounded-xl transition-colors ${
        isActive
          ? `${engineer.lightColor} ${engineer.darkColor} ring-1 ring-inset ring-current/10`
          : index % 2 === 0
            ? "bg-muted/40"
            : ""
      }`}
    >
      {/* Time range */}
      <div className="w-64 shrink-0">
        <p
          className={`text-sm font-medium whitespace-nowrap ${isActive ? "text-foreground" : "text-muted-foreground"}`}
        >
          {formatSegmentRange(segment.start, segment.end)}
        </p>
      </div>

      <Separator orientation="vertical" className="h-8" />

      {/* Engineer */}
      <div className="flex items-center gap-3 flex-1 min-w-0">
        <Avatar className="h-8 w-8 shrink-0">
          <AvatarImage src={engineer.avatarUrl} />
          <AvatarFallback
            className={`text-xs font-semibold ${engineer.color} ${engineer.textColor}`}
          >
            {initials(engineer.name)}
          </AvatarFallback>
        </Avatar>
        <span className="text-sm font-medium truncate flex-1">
          {engineer.name}
        </span>
        {segment.isOverride && (
          <Badge className="text-xs px-1.5 py-0 shrink-0 bg-amber-100 text-amber-700 border-amber-200 hover:bg-amber-100 dark:bg-amber-900/50 dark:text-amber-400 dark:border-amber-800 dark:hover:bg-amber-900/50">
            Override
          </Badge>
        )}
      </div>
    </div>
  );
}

export default HomeScheduleRow;

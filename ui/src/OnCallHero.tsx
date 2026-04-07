import { Clock, Mail } from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "./Avatar";
import type { TimeSegment } from "./types";
import { formatDateTimeRange, initials } from "./utils";

type OnCallHeroProps = {
  segment: TimeSegment;
};

function OnCallHero({ segment }: OnCallHeroProps) {
  return (
    <div className="relative overflow-hidden rounded-2xl bg-linear-to-br from-violet-500 via-indigo-500 to-sky-500 dark:from-violet-700 dark:via-indigo-700 dark:to-sky-700 p-px shadow-lg shadow-indigo-200 dark:shadow-indigo-950">
      <div className="relative rounded-[calc(1rem-1px)] bg-linear-to-br from-violet-500 via-indigo-500 to-sky-500 dark:from-violet-700 dark:via-indigo-700 dark:to-sky-700 px-6 py-5">
        <div className="absolute -top-6 -right-6 h-32 w-32 rounded-full bg-white/10 blur-2xl" />
        <div className="absolute -bottom-8 -left-4 h-24 w-24 rounded-full bg-white/10 blur-2xl" />
        <div className="relative">
          <div className="flex items-center gap-2 mb-4">
            <div className="h-2 w-2 rounded-full bg-green-300 animate-pulse shadow-sm shadow-green-400" />
            <span className="text-sm font-semibold text-white/80 uppercase tracking-widest">
              Currently On Call
            </span>
          </div>
          <div className="flex items-center gap-5">
            <Avatar
              className="ring-4 ring-white/30 shadow-xl"
              style={{ height: "4.5rem", width: "4.5rem" }}
            >
              <AvatarImage src={segment.member.avatarUrl} />
              <AvatarFallback className="bg-white/20 text-white font-bold text-xl backdrop-blur-sm">
                {initials(segment.member.name)}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col gap-1.5">
              <h2 className="text-2xl font-bold text-white">
                {segment.member.name}
              </h2>
              <div className="flex items-center gap-1.5 text-sm text-white/70">
                <Mail className="h-3.5 w-3.5" />
                <span>{segment.member.email}</span>
              </div>
              <div className="flex items-center gap-1.5 text-sm text-white/70">
                <Clock className="h-3.5 w-3.5" />
                <span>{formatDateTimeRange(segment.start, segment.end)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default OnCallHero;

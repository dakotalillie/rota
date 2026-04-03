import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function initials(name: string) {
  return name.split(' ').map(p => p[0]).join('').toUpperCase()
}


const fmtDateTime = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' })

function isMidnight(date: Date) {
  return date.getHours() === 0 && date.getMinutes() === 0 && date.getSeconds() === 0
}

/**
 * Format a half-open segment [start, end).
 * If both boundaries are at midnight, show as "Apr 1 – Apr 7" (end is exclusive, so display end-1 day).
 * If either has a time component, show full datetimes.
 */
export function formatSegmentRange(start: Date, end: Date): string {
  if (isMidnight(start) && isMidnight(end)) {
    const displayEnd = new Date(end)
    displayEnd.setDate(displayEnd.getDate() - 1)
    return `${fmtDateTime.format(start)} – ${fmtDateTime.format(displayEnd)}`
  }
  return `${fmtDateTime.format(start)} – ${fmtDateTime.format(end)}`
}

/** Format an override's stored datetime-local strings for display. */
export function formatOverrideRange(start: string, end: string): string {
  return `${fmtDateTime.format(new Date(start))} – ${fmtDateTime.format(new Date(end))}`
}

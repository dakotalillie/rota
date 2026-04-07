export type Member = {
  id: string; // member ID (matches server-side Member.id)
  userId: string; // user ID
  name: string;
  email: string;
  avatarUrl?: string;
  color: string; // tailwind bg class for avatar
  lightColor: string; // tailwind bg class for row highlight (light mode)
  darkColor: string; // tailwind bg class for row highlight (dark mode)
  textColor: string; // tailwind text class for avatar text
};

export type Override = {
  id: string;
  start: string; // datetime-local string: "YYYY-MM-DDThh:mm"
  end: string; // datetime-local string: "YYYY-MM-DDThh:mm"
  memberId: string;
};

export type TimeSegment = {
  start: Date;
  end: Date;
  member: Member;
  isOverride: boolean;
};

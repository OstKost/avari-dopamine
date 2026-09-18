"use client";

import { Flame } from "lucide-react";

interface StreakBadgeProps {
  days: number;
  className?: string;
}

export function StreakBadge({ days, className = "" }: StreakBadgeProps) {
  if (days <= 0) return null;

  return (
    <div
      className={`inline-flex items-center gap-1.5 rounded-full bg-gradient-to-r from-amber-500/15 via-orange-500/15 to-rose-500/15 px-3 py-1 text-xs font-black text-amber-600 dark:text-amber-400 border border-amber-500/30 shadow-sm transition-all hover:scale-105 ${className}`}
      title={`Вы заказываете ${days} дн. подряд!`}
    >
      <Flame className="h-4 w-4 text-orange-500 fill-orange-500 animate-pulse" />
      <span>{days} {getDaysWord(days)} подряд</span>
    </div>
  );
}

function getDaysWord(num: number): string {
  const n = Math.abs(num) % 100;
  const n1 = n % 10;
  if (n > 10 && n < 20) return "дней";
  if (n1 > 1 && n1 < 5) return "дня";
  if (n1 === 1) return "день";
  return "дней";
}

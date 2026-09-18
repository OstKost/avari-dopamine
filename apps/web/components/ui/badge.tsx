import * as React from "react";
import { cn } from "@/lib/utils";

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "gold" | "teal" | "secondary" | "success" | "warning" | "destructive" | "outline" | "dopamine";
}

export function Badge({ className, variant = "default", ...props }: BadgeProps) {
  const variants = {
    default: "bg-[#122234] text-[#F4F1E8] border border-[#1E3A50]",
    gold: "bg-amber-400/15 text-amber-300 border border-amber-400/40 shadow-sm shadow-amber-500/10 font-bold",
    teal: "bg-teal-400/15 text-teal-300 border border-teal-400/40 shadow-sm shadow-teal-500/10 font-bold",
    secondary: "bg-[#122234]/80 text-[#9FB3C4] border border-[#1E3A50]",
    success: "bg-emerald-500/15 text-emerald-300 border border-emerald-500/40 font-semibold",
    warning: "bg-amber-500/15 text-amber-300 border border-amber-500/40 font-semibold",
    destructive: "bg-red-500/15 text-red-300 border border-red-500/40 font-semibold",
    outline: "border border-[#1E3A50] text-[#9FB3C4] bg-transparent",
    dopamine: "bg-amber-400/20 text-amber-300 border border-amber-400/50 font-bold shadow-glow-amber",
  };

  return (
    <div
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors select-none",
        variants[variant],
        className
      )}
      {...props}
    />
  );
}

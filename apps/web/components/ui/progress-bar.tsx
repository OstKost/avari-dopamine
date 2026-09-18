import { cn } from "@/lib/utils";

export interface ProgressBarProps {
  progress: number; // 0 to 100
  className?: string;
  variant?: "rose" | "gradient";
}

export function ProgressBar({ progress, className, variant = "gradient" }: ProgressBarProps) {
  const clampedProgress = Math.min(100, Math.max(0, progress));

  return (
    <div className={cn("h-2 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-800", className)}>
      <div
        className={cn(
          "h-full transition-all duration-500 ease-out rounded-full",
          variant === "gradient" ? "bg-gradient-to-r from-rose-500 via-pink-500 to-purple-600" : "bg-rose-500"
        )}
        style={{ width: `${clampedProgress}%` }}
      />
    </div>
  );
}

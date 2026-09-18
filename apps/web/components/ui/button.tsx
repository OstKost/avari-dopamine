import * as React from "react";
import { cn } from "@/lib/utils";

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "gold" | "reward" | "teal" | "secondary" | "outline" | "ghost" | "destructive" | "glow";
  size?: "sm" | "md" | "lg" | "icon" | "pill";
  isLoading?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "md", isLoading, children, disabled, ...props }, ref) => {
    const baseStyles =
      "inline-flex items-center justify-center font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-amber-400 disabled:pointer-events-none disabled:opacity-50 select-none active:scale-[0.98]";

    const variants = {
      default:
        "bg-gradient-to-r from-[#F2B84B] via-[#FFD37A] to-[#F2B84B] text-[#050B14] font-bold shadow-md shadow-amber-500/20 hover:brightness-110 hover:shadow-glow-amber",
      gold:
        "bg-gradient-to-r from-[#F2B84B] via-[#FFD37A] to-[#F2B84B] text-[#050B14] font-extrabold shadow-md shadow-amber-500/25 hover:brightness-110 hover:shadow-glow-amber",
      reward:
        "bg-gradient-to-r from-[#FFD37A] via-[#F2B84B] to-[#C6912F] text-[#050B14] font-black shadow-lg shadow-amber-500/30 hover:brightness-110 hover:shadow-glow-amber-lg",
      teal:
        "bg-gradient-to-r from-[#54ACBF] to-[#0D9488] text-[#050B14] font-bold shadow-md shadow-teal-500/20 hover:brightness-110 hover:shadow-glow-teal",
      secondary:
        "bg-[#122234] text-[#F4F1E8] border border-[#1E3A50] hover:bg-[#1E3A50] hover:text-white",
      outline:
        "border border-[#1E3A50] bg-transparent hover:bg-[#122234] text-[#F4F1E8] hover:border-amber-400/50",
      ghost:
        "hover:bg-[#122234] text-[#9FB3C4] hover:text-[#F4F1E8]",
      destructive:
        "bg-red-500/90 text-white hover:bg-red-600 shadow-sm shadow-red-500/20",
      glow:
        "bg-gradient-to-r from-[#F2B84B] via-[#FFD37A] to-[#F2B84B] text-[#050B14] font-bold shadow-lg shadow-amber-500/30 hover:shadow-glow-amber-lg hover:scale-[1.02]",
    };

    const sizes = {
      sm: "h-8 px-3 text-xs rounded-lg gap-1.5",
      md: "h-10 px-4 py-2 text-sm rounded-xl gap-2",
      lg: "h-12 px-6 text-base rounded-2xl gap-2.5",
      pill: "h-9 px-4 text-xs font-bold rounded-full gap-1.5",
      icon: "h-10 w-10 p-0 rounded-xl",
    };

    return (
      <button
        ref={ref}
        disabled={disabled || isLoading}
        className={cn(baseStyles, variants[variant], sizes[size], className)}
        {...props}
      >
        {isLoading && (
          <svg
            className="animate-spin -ml-1 mr-2 h-4 w-4 text-current"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        )}
        {children}
      </button>
    );
  }
);

Button.displayName = "Button";

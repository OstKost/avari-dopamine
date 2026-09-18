"use client";

import { useEffect, useState } from "react";
import { Sparkles, Trophy, CheckCircle2 } from "lucide-react";

interface CelebrationProps {
  show: boolean;
  orderNumber?: string;
}

export function CelebrationConfetti({ show, orderNumber }: CelebrationProps) {
  const [particles, setParticles] = useState<Array<{ id: number; x: number; y: number; color: string; size: number; delay: number }>>([]);
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);

  useEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
    setPrefersReducedMotion(mediaQuery.matches);

    const listener = (e: MediaQueryListEvent) => setPrefersReducedMotion(e.matches);
    mediaQuery.addEventListener("change", listener);
    return () => mediaQuery.removeEventListener("change", listener);
  }, []);

  useEffect(() => {
    if (show && !prefersReducedMotion) {
      const colors = ["#f43f5e", "#8b5cf6", "#ec4899", "#10b981", "#3b82f6", "#f59e0b"];
      const newParticles = Array.from({ length: 40 }).map((_, i) => ({
        id: i,
        x: Math.random() * 100,
        y: Math.random() * 100,
        color: colors[Math.floor(Math.random() * colors.length)],
        size: Math.random() * 8 + 4,
        delay: Math.random() * 0.5,
      }));
      setParticles(newParticles);
    }
  }, [show, prefersReducedMotion]);

  if (!show) return null;

  return (
    <div className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-emerald-500/10 via-rose-500/10 to-purple-500/10 border border-emerald-500/20 p-6 sm:p-8 text-center space-y-4 shadow-lg animate-fadeIn">
      {!prefersReducedMotion && (
        <div className="pointer-events-none absolute inset-0 overflow-hidden">
          {particles.map((p) => (
            <div
              key={p.id}
              className="absolute rounded-full opacity-80 animate-ping"
              style={{
                left: `${p.x}%`,
                top: `${p.y}%`,
                width: `${p.size}px`,
                height: `${p.size}px`,
                backgroundColor: p.color,
                animationDuration: "1.5s",
                animationDelay: `${p.delay}s`,
              }}
            />
          ))}
        </div>
      )}

      <div className="relative z-10 flex flex-col items-center space-y-2">
        <div className="flex h-16 w-16 items-center justify-center rounded-3xl bg-emerald-500 text-white shadow-xl shadow-emerald-500/30 animate-bounce">
          <CheckCircle2 className="h-8 w-8" />
        </div>
        <div className="space-y-1">
          <div className="flex items-center justify-center gap-2">
            <Trophy className="h-5 w-5 text-amber-500" />
            <h3 className="text-2xl font-black text-zinc-900 dark:text-zinc-50 tracking-tight">
              Заказ доставлен в ПВЗ!
            </h3>
            <Sparkles className="h-5 w-5 text-rose-500" />
          </div>
          <p className="text-sm text-zinc-500 max-w-md mx-auto">
            {orderNumber ? `Заказ #${orderNumber.slice(0, 8)} готов к выдаче.` : "Ваш заказ готов к выдаче."} Заберите его в удобное время!
          </p>
        </div>
      </div>
    </div>
  );
}

"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { Sparkles, Trophy } from "lucide-react";

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
      const colors = ["#F2B84B", "#FFD37A", "#54ACBF", "#A7EBF2", "#5FD98A"];
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
    <div className="relative overflow-hidden rounded-[28px] bg-gradient-to-br from-[#122234] via-[#0B1622] to-[#122234] border border-amber-500/30 p-6 sm:p-8 text-center space-y-4 shadow-2xl shadow-amber-500/10 animate-fadeIn">
      {/* Background glow */}
      <div className="pointer-events-none absolute -top-16 left-1/2 -translate-x-1/2 w-64 h-64 bg-amber-500/15 rounded-full blur-3xl" />
      <div className="pointer-events-none absolute -bottom-16 left-1/2 -translate-x-1/2 w-64 h-64 bg-teal-500/15 rounded-full blur-3xl" />

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

      <div className="relative z-10 flex flex-col items-center space-y-3">
        <div className="flex h-16 w-16 items-center justify-center rounded-3xl bg-gradient-to-tr from-[#1E3A50] to-[#0E1B29] border border-amber-400 text-amber-400 shadow-xl shadow-amber-500/30 animate-bounce">
          <Image
            src="/logo-minimal-star.png"
            alt="Success Star"
            width={36}
            height={36}
            className="object-contain"
          />
        </div>
        <div className="space-y-1">
          <div className="flex items-center justify-center gap-2">
            <Trophy className="h-5 w-5 text-amber-400" />
            <h3 className="text-2xl font-black text-[#F4F1E8] tracking-tight">
              Заказ доставлен в ПВЗ!
            </h3>
            <Sparkles className="h-5 w-5 text-teal-400" />
          </div>
          <p className="text-sm text-[#9FB3C4] max-w-md mx-auto">
            {orderNumber ? `Заказ #${orderNumber.slice(0, 8)} готов к выдаче.` : "Ваш заказ готов к выдаче."} Заберите его в удобное время!
          </p>
        </div>
      </div>
    </div>
  );
}

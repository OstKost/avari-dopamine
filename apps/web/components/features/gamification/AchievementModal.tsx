"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { Sparkles, X, Trophy } from "lucide-react";
import { Button } from "@/components/ui/button";

export interface Achievement {
  id: string;
  title: string;
  description: string;
  xp_reward: number;
  icon?: string;
  unlocked_at?: string;
  is_unlocked: boolean;
  progress?: number;
  max_progress?: number;
}

interface AchievementModalProps {
  achievement: Achievement | null;
  onClose: () => void;
  onClaim?: (achievement: Achievement) => void;
}

export function AchievementModal({ achievement, onClose, onClaim }: AchievementModalProps) {
  const [particles, setParticles] = useState<Array<{ id: number; x: number; y: number; color: string; size: number; delay: number }>>([]);

  useEffect(() => {
    if (achievement) {
      const colors = ["#F2B84B", "#FFD37A", "#54ACBF", "#A7EBF2", "#5FD98A"];
      const newParticles = Array.from({ length: 30 }).map((_, i) => ({
        id: i,
        x: Math.random() * 100,
        y: Math.random() * 100,
        color: colors[Math.floor(Math.random() * colors.length)],
        size: Math.random() * 6 + 3,
        delay: Math.random() * 0.4,
      }));
      setParticles(newParticles);
    }
  }, [achievement]);

  if (!achievement) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-fadeIn">
      <div className="relative w-full max-w-sm rounded-[28px] bg-gradient-to-b from-[#122234] to-[#0B1622] border border-[#1E3A50] p-8 text-center shadow-2xl shadow-amber-500/10 overflow-hidden animate-scaleUp">
        {/* Ambient glow backgrounds */}
        <div className="pointer-events-none absolute -top-24 left-1/2 -translate-x-1/2 w-64 h-64 bg-amber-500/20 rounded-full blur-3xl" />
        <div className="pointer-events-none absolute -bottom-24 left-1/2 -translate-x-1/2 w-64 h-64 bg-teal-500/15 rounded-full blur-3xl" />

        {/* Floating Sparks */}
        <div className="pointer-events-none absolute inset-0 overflow-hidden">
          {particles.map((p) => (
            <div
              key={p.id}
              className="absolute rounded-full opacity-75 animate-ping"
              style={{
                left: `${p.x}%`,
                top: `${p.y}%`,
                width: `${p.size}px`,
                height: `${p.size}px`,
                backgroundColor: p.color,
                animationDuration: "1.8s",
                animationDelay: `${p.delay}s`,
              }}
            />
          ))}
        </div>

        {/* Close Button */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-2 rounded-xl text-[#5E7488] hover:text-[#F4F1E8] hover:bg-[#1E3A50]/50 transition-colors z-10"
        >
          <X className="h-5 w-5" />
        </button>

        {/* Radiant Center Emblem matching screen-03 */}
        <div className="relative z-10 flex flex-col items-center space-y-5">
          <div className="relative flex items-center justify-center">
            {/* Pulsing ring */}
            <div className="absolute inset-0 rounded-full bg-gradient-to-tr from-amber-500 to-amber-300 blur-xl opacity-40 animate-pulse" />
            
            <div className="relative flex h-24 w-24 items-center justify-center rounded-3xl bg-gradient-to-tr from-[#1E3A50] to-[#0E1B29] border-2 border-amber-400/60 p-4 shadow-xl shadow-amber-500/20">
              <Image
                src="/logo-minimal-star.png"
                alt="Achievement Medallion"
                width={56}
                height={56}
                className="object-contain drop-shadow-[0_0_12px_rgba(242,184,75,0.8)]"
              />
            </div>
          </div>

          <div className="space-y-2">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-amber-500/15 border border-amber-500/30 text-amber-400 text-xs font-bold uppercase tracking-wider">
              <Sparkles className="h-3.5 w-3.5" />
              <span>{achievement.is_unlocked ? "Достижение открыто!" : "Новое достижение!"}</span>
            </div>

            <h3 className="text-2xl font-black text-[#F4F1E8] tracking-tight">
              «{achievement.title}»
            </h3>

            <p className="text-sm text-[#9FB3C4] max-w-xs mx-auto leading-relaxed">
              {achievement.description}
            </p>
          </div>

          {/* XP Reward Badge */}
          <div className="flex items-center justify-center gap-2 px-5 py-2.5 rounded-2xl bg-[#0E1B29] border border-[#1E3A50] shadow-inner">
            <Trophy className="h-5 w-5 text-amber-400" />
            <span className="text-lg font-black text-amber-400 tracking-wide">
              +{achievement.xp_reward} XP
            </span>
          </div>

          {/* Claim / Close Button */}
          <Button
            variant="reward"
            size="lg"
            className="w-full text-base font-bold py-3.5 rounded-2xl shadow-lg shadow-amber-500/25"
            onClick={() => {
              if (onClaim) onClaim(achievement);
              onClose();
            }}
          >
            <span>Забрать награду</span>
            <Sparkles className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </div>
    </div>
  );
}

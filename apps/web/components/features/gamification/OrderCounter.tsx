"use client";

import { PackageCheck } from "lucide-react";

interface OrderCounterProps {
  count: number;
  className?: string;
}

export function OrderCounter({ count, className = "" }: OrderCounterProps) {
  if (count <= 0) return null;

  return (
    <div
      className={`inline-flex items-center gap-1.5 rounded-full bg-zinc-100 dark:bg-zinc-800 px-3 py-1 text-xs font-bold text-zinc-700 dark:text-zinc-300 border border-zinc-200 dark:border-zinc-700 ${className}`}
    >
      <PackageCheck className="h-3.5 w-3.5 text-rose-500" />
      <span>{count} {getOrdersWord(count)}</span>
    </div>
  );
}

function getOrdersWord(num: number): string {
  const n = Math.abs(num) % 100;
  const n1 = n % 10;
  if (n > 10 && n < 20) return "заказов";
  if (n1 > 1 && n1 < 5) return "заказа";
  if (n1 === 1) return "заказ";
  return "заказов";
}

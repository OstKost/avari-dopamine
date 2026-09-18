"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, Tag, ShoppingBag, User } from "lucide-react";
import { cn } from "@/lib/utils";

export function BottomNav() {
  const pathname = usePathname();

  const navItems = [
    {
      href: "/",
      label: "Главная",
      icon: Home,
      isActive: pathname === "/",
    },
    {
      href: "/catalog",
      label: "Каталог",
      icon: Tag,
      isActive: pathname.startsWith("/catalog") || pathname.startsWith("/product"),
    },
    {
      href: "/cart",
      label: "Корзина",
      icon: ShoppingBag,
      isActive: pathname === "/cart",
      badge: "1",
    },
    {
      href: "/profile",
      label: "Профиль",
      icon: User,
      isActive: pathname === "/profile" || pathname.startsWith("/orders"),
    },
  ];

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 md:hidden bg-[#050B14]/90 backdrop-blur-2xl border-t border-[#1E3A50]/90 px-4 py-2 shadow-2xl">
      <div className="flex items-center justify-around max-w-md mx-auto">
        {navItems.map((item) => {
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "relative flex flex-col items-center justify-center py-1 px-3 rounded-2xl transition-all duration-300",
                item.isActive
                  ? "text-amber-400 font-bold"
                  : "text-[#9FB3C4] hover:text-[#F4F1E8]"
              )}
            >
              <div className="relative">
                <Icon
                  className={cn(
                    "h-5 w-5 transition-transform duration-200",
                    item.isActive && "scale-110 drop-shadow-[0_0_8px_rgba(242,184,75,0.6)]"
                  )}
                />
                {item.badge && (
                  <span className="absolute -top-1.5 -right-2.5 flex h-4 w-4 items-center justify-center rounded-full bg-gradient-to-r from-amber-400 to-amber-500 text-[10px] font-black text-[#050B14] shadow-sm">
                    {item.badge}
                  </span>
                )}
              </div>
              <span className="text-[11px] mt-1 tracking-tight">{item.label}</span>
              {item.isActive && (
                <div className="absolute -bottom-1 h-1 w-6 rounded-full bg-gradient-to-r from-amber-400 to-amber-300 shadow-glow-amber" />
              )}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}

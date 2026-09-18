"use client";

import { useEffect, useState } from "react";
import { MapPin, User, Sparkles, ChevronRight, Bike } from "lucide-react";
import { CourierInfo } from "@/hooks/useOrderStatus";

interface DeliveryProgressProps {
  status: string;
  courier?: CourierInfo;
  estimatedCompletionAt?: string;
  isLive?: boolean;
  pickupPointName?: string;
}

export function DeliveryProgress({
  status,
  courier,
  estimatedCompletionAt,
  pickupPointName,
}: DeliveryProgressProps) {
  const [timeLeft, setTimeLeft] = useState<string>("03:42");

  useEffect(() => {
    if (!estimatedCompletionAt || status === "delivered") {
      if (status === "delivered") setTimeLeft("00:00");
      return;
    }

    const updateTimer = () => {
      const now = new Date().getTime();
      const target = new Date(estimatedCompletionAt).getTime();
      const diff = Math.max(0, Math.floor((target - now) / 1000));
      const mins = String(Math.floor(diff / 60)).padStart(2, "0");
      const secs = String(diff % 60).padStart(2, "0");
      setTimeLeft(`${mins}:${secs}`);
    };

    updateTimer();
    const interval = setInterval(updateTimer, 1000);
    return () => clearInterval(interval);
  }, [estimatedCompletionAt, status]);

  const stages = [
    { id: "assembled", label: "Собран", active: ["paid", "assembling", "courier_assigned", "in_transit", "delivery_delayed", "delivered"].includes(status) },
    { id: "courier", label: "Курьер назначен", active: ["courier_assigned", "in_transit", "delivery_delayed", "delivered"].includes(status) },
    { id: "transit", label: "В пути", active: ["in_transit", "delivery_delayed", "delivered"].includes(status), current: ["in_transit", "delivery_delayed"].includes(status) },
    { id: "delivered", label: "Доставлен", active: status === "delivered", current: status === "delivered" },
  ];

  const getStatusHeading = () => {
    switch (status) {
      case "in_transit":
        return `В пути — ${timeLeft}`;
      case "delivery_delayed":
        return `Задерживается — ${timeLeft}`;
      case "courier_assigned":
        return `Курьер спешит — ${timeLeft}`;
      case "assembling":
        return "Собираем заказ...";
      case "paid":
        return "Оплачен — готовим к отправке";
      case "delivered":
        return "Доставлен в ПВЗ! 🎉";
      default:
        return "Обработка заказа";
    }
  };

  return (
    <div className="space-y-4">
      {/* Big Glowing Live Delivery Card matching screen-02 */}
      <div className="relative rounded-3xl overflow-hidden border border-teal-400/50 bg-gradient-to-b from-[#0B2533] via-[#091D28] to-[#08151E] p-6 sm:p-8 shadow-glow-teal-lg text-center">
        {/* Animated Scooter Illustration */}
        <div className="relative mx-auto mb-6 flex items-center justify-center h-28 w-44">
          <div className="absolute inset-0 bg-teal-400/20 blur-2xl rounded-full" />
          <div className="relative z-10 flex flex-col items-center">
            <div className="relative flex items-center justify-center h-20 w-20 rounded-full bg-[#050B14] border-2 border-teal-400 shadow-[0_0_24px_rgba(84,172,191,0.8)] animate-pulse">
              <Bike className="h-10 w-10 text-teal-300 drop-shadow-[0_0_8px_rgba(84,172,191,0.9)]" />
            </div>
            {/* Speed trails */}
            <div className="flex gap-1 mt-2">
              <div className="h-0.5 w-6 rounded-full bg-teal-400/80 animate-pulse" />
              <div className="h-0.5 w-10 rounded-full bg-teal-300 shadow-glow-teal" />
              <div className="h-0.5 w-4 rounded-full bg-teal-400/60" />
            </div>
          </div>
        </div>

        {/* Big Status Heading */}
        <h2 className="text-2xl sm:text-3xl font-black text-[#F4F1E8] tracking-tight mb-6">
          {getStatusHeading()}
        </h2>

        {/* Stepper Timeline matching screen-02 */}
        <div className="relative max-w-lg mx-auto mb-8 px-2">
          {/* Timeline track line */}
          <div className="absolute top-3 left-6 right-6 h-0.5 bg-[#1E3A50] -z-0">
            <div
              className="h-full bg-gradient-to-r from-amber-400 via-teal-400 to-teal-300 transition-all duration-700 shadow-glow-teal"
              style={{
                width: status === "delivered" ? "100%" : ["in_transit", "delivery_delayed"].includes(status) ? "70%" : ["courier_assigned"].includes(status) ? "40%" : "15%",
              }}
            />
          </div>

          <div className="relative z-10 flex justify-between">
            {stages.map((st) => (
              <div key={st.id} className="flex flex-col items-center space-y-2 max-w-[80px]">
                <div
                  className={`h-6 w-6 rounded-full flex items-center justify-center border-2 transition-all duration-500 ${
                    st.current
                      ? "border-teal-300 bg-[#050B14] shadow-[0_0_16px_rgba(84,172,191,0.9)] scale-110"
                      : st.active
                      ? "border-amber-400 bg-amber-400/20 text-amber-300 shadow-glow-amber"
                      : "border-[#1E3A50] bg-[#0B1622] text-[#5E7488]"
                  }`}
                >
                  {st.current ? (
                    <div className="h-2 w-2 rounded-full bg-teal-300 animate-ping" />
                  ) : st.active ? (
                    <div className="h-2 w-2 rounded-full bg-amber-400" />
                  ) : null}
                </div>
                <span
                  className={`text-[11px] text-center font-bold tracking-tight leading-tight ${
                    st.current
                      ? "text-teal-300"
                      : st.active
                      ? "text-amber-300"
                      : "text-[#5E7488]"
                  }`}
                >
                  {st.label}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* Reward XP Badge */}
        <div className="inline-flex items-center gap-1.5 px-5 py-2 rounded-full bg-amber-400/15 border border-amber-400/50 shadow-glow-amber text-amber-300 text-xs sm:text-sm font-extrabold">
          <Sparkles className="h-4 w-4 text-amber-300 animate-pulse" />
          <span>+10 XP за этот заказ</span>
        </div>
      </div>

      {/* Info Cards below matching screen-02 */}
      <div className="space-y-3">
        {/* Pickup Point Card */}
        <div className="flex items-center justify-between p-4 px-5 rounded-2xl bg-[#0B1622]/90 border border-[#1E3A50] hover:border-teal-400/40 transition-all">
          <div className="flex items-center gap-3.5">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-amber-400/15 text-amber-300 border border-amber-400/30">
              <MapPin className="h-5 w-5" />
            </div>
            <div>
              <span className="text-[11px] font-semibold text-[#9FB3C4] block">Пункт выдачи</span>
              <span className="text-sm font-bold text-[#F4F1E8]">
                {pickupPointName || "ул. Малая Садовая, 12"}
              </span>
            </div>
          </div>
          <ChevronRight className="h-5 w-5 text-[#5E7488]" />
        </div>

        {/* Courier Card */}
        {courier && (
          <div className="flex items-center justify-between p-4 px-5 rounded-2xl bg-[#0B1622]/90 border border-[#1E3A50] hover:border-teal-400/40 transition-all">
            <div className="flex items-center gap-3.5">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-teal-400/15 text-teal-300 border border-teal-400/30">
                <User className="h-5 w-5" />
              </div>
              <div>
                <span className="text-[11px] font-semibold text-[#9FB3C4] block">Курьер</span>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-bold text-[#F4F1E8]">
                    {courier.name}
                  </span>
                  <span className="text-xs font-bold text-amber-300 bg-amber-400/10 px-1.5 py-0.2 rounded-md border border-amber-400/30">
                    ★ {courier.rating.toFixed(2)}
                  </span>
                </div>
              </div>
            </div>
            <ChevronRight className="h-5 w-5 text-[#5E7488]" />
          </div>
        )}
      </div>
    </div>
  );
}

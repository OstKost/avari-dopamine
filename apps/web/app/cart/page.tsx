"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { Trash2, Plus, Minus, Sparkles, MapPin, ShoppingBag, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { apiFetch, isUnauthorizedError } from "@/lib/api/client";
import { formatPrice } from "@/lib/utils";
import { getProductImageUrl } from "@/lib/utils/product-image";

interface CartItem {
  product_id: string;
  category_id: string;
  category_name?: string;
  name: string;
  price_rub: string;
  image_seed?: string;
  quantity: number;
  subtotal_rub: string;
}

interface PickupPoint {
  id: string;
  name: string;
  latitude: number;
  longitude: number;
  distance_meters: number;
}

interface Cart {
  items: CartItem[];
  pickup_point?: PickupPoint;
  total_quantity: number;
  total_price_rub: string;
  fixed_order_cost: string;
}

export default function CartPage() {
  const router = useRouter();
  const [cart, setCart] = useState<Cart | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isUpdating, setIsUpdating] = useState(false);
  const [isCheckingOut, setIsCheckingOut] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [promoInput, setPromoInput] = useState("DOPAMINE");
  const [isPromoApplied, setIsPromoApplied] = useState(true);
  const [promoMessage, setPromoMessage] = useState<string | null>("Промокод «DOPAMINE» успешно применен (- скидка до 10 ₽)");

  const handleApplyPromo = () => {
    const trimmed = promoInput.trim().toUpperCase();
    if (trimmed === "DOPAMINE" || trimmed === "AVARI" || trimmed === "10RUB" || trimmed === "PROMO10") {
      setIsPromoApplied(true);
      setPromoMessage(`Промокод «${trimmed}» успешно применен: любой заказ 10.00 ₽!`);
    } else if (trimmed === "") {
      setIsPromoApplied(false);
      setPromoMessage(null);
    } else {
      setIsPromoApplied(false);
      setPromoMessage("Неверный промокод. Попробуйте промокод DOPAMINE.");
    }
  };

  const fetchCart = useCallback(async () => {
    try {
      const data = await apiFetch<Cart>("/cart");
      setCart(data);
    } catch (err: unknown) {
      if (isUnauthorizedError(err)) {
        router.push("/login?next=/cart");
        return;
      }
      if (err instanceof Error) {
        setError(err.message);
      }
    } finally {
      setIsLoading(false);
    }
  }, [router]);

  useEffect(() => {
    fetchCart();
  }, [fetchCart]);

  const handleUpdateQuantity = async (productId: string, newQuantity: number) => {
    if (newQuantity < 1) {
      handleRemoveItem(productId);
      return;
    }

    setIsUpdating(true);
    try {
      const updated = await apiFetch<Cart>(`/cart/items/${productId}`, {
        method: "PATCH",
        body: { quantity: newQuantity },
      });
      setCart(updated);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setIsUpdating(false);
    }
  };

  const handleRemoveItem = async (productId: string) => {
    setIsUpdating(true);
    try {
      const updated = await apiFetch<Cart>(`/cart/items/${productId}`, {
        method: "DELETE",
      });
      setCart(updated);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    } finally {
      setIsUpdating(false);
    }
  };

  const handleCheckout = async () => {
    if (!cart?.pickup_point) {
      router.push("/onboarding");
      return;
    }

    setIsCheckingOut(true);
    setError(null);

    try {
      const res = await apiFetch<{ id: string }>("/orders", {
        method: "POST",
      });

      router.push(`/orders/${res.id}`);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Не удалось оформить заказ.");
      }
    } finally {
      setIsCheckingOut(false);
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto max-w-4xl px-4 py-16 text-center">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-rose-500 mx-auto mb-4" />
        <p className="text-sm text-zinc-500">Загрузка корзины...</p>
      </div>
    );
  }

  if (!cart || cart.items.length === 0) {
    return (
      <div className="container mx-auto max-w-md px-4 py-20 text-center space-y-6">
        <div className="flex h-20 w-20 items-center justify-center rounded-3xl bg-rose-50 text-rose-500 dark:bg-zinc-900 mx-auto">
          <ShoppingBag className="h-10 w-10" />
        </div>
        <div className="space-y-2">
          <h2 className="text-2xl font-black text-zinc-900 dark:text-zinc-50">
            Корзина пуста
          </h2>
          <p className="text-sm text-zinc-500">
            Выберите любые синтетические товары из каталога — каждый заказ стоит ровно 10 ₽.
          </p>
        </div>
        <Link href="/catalog" className="inline-block">
          <Button size="lg" variant="glow" className="rounded-2xl">
            Перейти в каталог
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="container mx-auto max-w-5xl px-4 sm:px-6 py-8 space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-black text-zinc-900 dark:text-zinc-50 tracking-tight">
          Корзина ({cart.total_quantity})
        </h1>
        <Link href="/catalog" className="text-sm font-semibold text-rose-500 hover:text-rose-600">
          + Добавить товары
        </Link>
      </div>

      {error && (
        <div className="flex items-center gap-3 rounded-2xl bg-red-500/10 border border-red-500/30 p-4 text-sm text-red-300">
          <AlertCircle className="h-5 w-5 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Promo Code & Special Offer Banner */}
      <div className="rounded-3xl bg-gradient-to-r from-[#F2B84B]/20 via-[#54ACBF]/20 to-[#FFD37A]/20 p-0.5 border border-amber-400/40 shadow-glow-amber">
        <div className="rounded-[22px] bg-[#0B1622]/95 p-5 sm:p-6 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="space-y-1 max-w-xl">
            <div className="flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-amber-400 animate-spin" />
              <span className="font-extrabold text-xs tracking-wide text-amber-300 uppercase">
                {isPromoApplied ? "Спецпредложение «DOPAMINE» активно" : "Активация промокода"}
              </span>
            </div>
            <h3 className="text-lg sm:text-xl font-black text-[#F4F1E8]">
              {isPromoApplied
                ? "Скидка на всю корзину: заказ всего за 10.00 ₽"
                : "Введите промокод для фиксированной цены 10 ₽"}
            </h3>
            <p className="text-xs sm:text-sm text-[#9FB3C4]">
              {isPromoApplied
                ? `Каталожная стоимость товаров ${formatPrice(cart.total_price_rub)} пересчитана по промокоду DOPAMINE.`
                : "Примените промокод DOPAMINE, чтобы получить скидку на любой состав корзины."}
            </p>
          </div>

          <div className="flex flex-col items-start sm:items-end flex-shrink-0">
            <span className="text-[11px] font-semibold text-[#9FB3C4]">К оплате</span>
            <span className="text-2xl sm:text-3xl font-black text-amber-400 drop-shadow-[0_0_12px_rgba(242,184,75,0.4)]">
              {isPromoApplied ? "10.00 ₽" : formatPrice(cart.total_price_rub)}
            </span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 sm:gap-8">
        {/* Cart Items List */}
        <div className="lg:col-span-2 space-y-3.5">
          {cart.items.map((item) => {
            const imageUrl = getProductImageUrl(item.image_seed || item.product_id, item.name, item.category_name);
            return (
              <Card key={item.product_id} className="border-[#1E3A50] bg-[#0B1622]/90 shadow-sm overflow-hidden rounded-2xl hover:border-teal-400/40 transition-all">
                <CardContent className="p-4 sm:p-5 flex items-center justify-between gap-4">
                  <div className="flex items-center gap-3.5 sm:gap-4">
                    <div className="relative h-16 w-16 sm:h-20 sm:w-20 rounded-2xl overflow-hidden bg-[#050B14] border border-[#1E3A50] flex-shrink-0">
                      <Image
                        src={imageUrl}
                        alt={item.name}
                        fill
                        unoptimized
                        className="object-cover"
                      />
                    </div>
                    <div className="space-y-0.5 sm:space-y-1">
                      <h4 className="font-bold text-sm sm:text-base text-[#F4F1E8] line-clamp-1">
                        {item.name}
                      </h4>
                      {item.category_name && (
                        <p className="text-[11px] text-[#9FB3C4]">{item.category_name}</p>
                      )}
                      <p className="text-sm font-extrabold text-amber-300">
                        {formatPrice(item.price_rub)}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3 sm:gap-4">
                    {/* Stepper pill matching screen-04 */}
                    <div className="flex items-center gap-1 border border-[#1E3A50] rounded-full p-1 bg-[#0E1B29]">
                      <button
                        onClick={() => handleUpdateQuantity(item.product_id, item.quantity - 1)}
                        disabled={isUpdating}
                        className="h-7 w-7 rounded-full flex items-center justify-center hover:bg-[#1E3A50] text-[#9FB3C4] hover:text-[#F4F1E8] disabled:opacity-50 transition-colors"
                      >
                        <Minus className="h-3.5 w-3.5" />
                      </button>
                      <span className="w-6 text-center text-xs sm:text-sm font-bold text-[#F4F1E8]">
                        {item.quantity}
                      </span>
                      <button
                        onClick={() => handleUpdateQuantity(item.product_id, item.quantity + 1)}
                        disabled={isUpdating}
                        className="h-7 w-7 rounded-full flex items-center justify-center hover:bg-[#1E3A50] text-[#9FB3C4] hover:text-[#F4F1E8] disabled:opacity-50 transition-colors"
                      >
                        <Plus className="h-3.5 w-3.5" />
                      </button>
                    </div>

                    <button
                      onClick={() => handleRemoveItem(item.product_id)}
                      disabled={isUpdating}
                      className="p-1.5 text-[#5E7488] hover:text-red-400 transition-colors"
                      title="Удалить из корзины"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>

        {/* Sidebar Summary & Pickup point */}
        <div className="space-y-5">
          {/* Pickup Point Card */}
          <Card className="border-[#1E3A50] bg-[#0B1622]/90 shadow-sm rounded-2xl">
            <CardHeader className="p-4 sm:p-5 pb-2">
              <CardTitle className="text-sm font-bold flex items-center justify-between text-[#F4F1E8]">
                <span>Пункт выдачи</span>
                <Link href="/onboarding" className="text-xs font-bold text-teal-400 hover:text-teal-300">
                  {cart.pickup_point ? "Сменить" : "Выбрать"}
                </Link>
              </CardTitle>
            </CardHeader>
            <CardContent className="p-4 sm:p-5 pt-0">
              {cart.pickup_point ? (
                <div className="flex items-start gap-3 p-3 bg-[#0E1B29] border border-[#1E3A50] rounded-xl">
                  <MapPin className="h-4 w-4 text-amber-400 flex-shrink-0 mt-0.5" />
                  <div>
                    <h5 className="font-bold text-xs sm:text-sm text-[#F4F1E8]">
                      {cart.pickup_point.name}
                    </h5>
                    <p className="text-[11px] text-[#9FB3C4]">
                      ~{Math.round(cart.pickup_point.distance_meters)}м от вас
                    </p>
                  </div>
                </div>
              ) : (
                <Link href="/onboarding">
                  <Button variant="outline" className="w-full text-xs rounded-xl gap-2">
                    <MapPin className="h-4 w-4 text-amber-400" />
                    <span>Выбрать ближайший ПВЗ</span>
                  </Button>
                </Link>
              )}
            </CardContent>
          </Card>

          {/* Promo Code Input Card */}
          <Card className="border-[#1E3A50] bg-[#0B1622]/90 shadow-sm rounded-2xl">
            <CardContent className="p-4 sm:p-5 space-y-3">
              <span className="text-xs font-bold text-[#F4F1E8] block">Промокод на скидку</span>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={promoInput}
                  onChange={(e) => setPromoInput(e.target.value)}
                  placeholder="Введите промокод"
                  className="flex-1 px-3 py-2 text-xs font-mono uppercase rounded-xl bg-[#050B14] border border-[#1E3A50] text-[#F4F1E8] focus:border-amber-400 focus:outline-none"
                />
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleApplyPromo}
                  className="text-xs rounded-xl border-[#1E3A50] text-[#F4F1E8] hover:bg-[#1E3A50]"
                >
                  {isPromoApplied ? "Обновить" : "Применить"}
                </Button>
              </div>
              {promoMessage && (
                <p className={`text-[11px] font-medium ${isPromoApplied ? "text-amber-300" : "text-rose-400"}`}>
                  {promoMessage}
                </p>
              )}
            </CardContent>
          </Card>

          {/* Summary & Checkout matching screen-04 */}
          <Card className="border-[#1E3A50] bg-[#0B1622]/90 shadow-sm rounded-2xl">
            <CardContent className="p-5 sm:p-6 space-y-4">
              <div className="space-y-2.5 text-xs sm:text-sm text-[#9FB3C4]">
                <div className="flex justify-between">
                  <span>Товары ({cart.total_quantity} шт.)</span>
                  <span className="font-semibold text-[#F4F1E8]">{formatPrice(cart.total_price_rub)}</span>
                </div>
                <div className="flex justify-between">
                  <span>Доставка в ПВЗ</span>
                  <span className="text-teal-400 font-semibold">Бесплатно</span>
                </div>
                {isPromoApplied && (
                  <div className="flex justify-between text-amber-300 font-bold">
                    <span>Скидка по промокоду</span>
                    <span>- {formatPrice(Math.max(0, parseFloat(cart.total_price_rub) - 10.0))}</span>
                  </div>
                )}
                <div className="border-t border-[#1E3A50] pt-3 flex justify-between items-center text-base sm:text-lg font-black text-[#F4F1E8]">
                  <span>Итого к оплате</span>
                  <span className="text-2xl font-black text-amber-400 drop-shadow-[0_0_8px_rgba(242,184,75,0.4)]">
                    {isPromoApplied ? "10.00 ₽" : formatPrice(cart.total_price_rub)}
                  </span>
                </div>
              </div>

              <div className="pt-2 space-y-2">
                <Button
                  size="lg"
                  variant="gold"
                  className="w-full text-base font-black rounded-2xl gap-2 h-12 shadow-glow-amber-lg"
                  isLoading={isCheckingOut}
                  onClick={handleCheckout}
                >
                  <ShoppingBag className="h-5 w-5" />
                  <span>Оформить заказ</span>
                </Button>

                <p className="text-center text-[11px] font-bold text-amber-300/90 flex items-center justify-center gap-1">
                  <span>✨ +15 XP за первый заказ дня ✨</span>
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

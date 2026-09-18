"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { Trash2, Plus, Minus, Sparkles, MapPin, ArrowRight, ShoppingBag, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { apiFetch } from "@/lib/api/client";
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

  const fetchCart = useCallback(async () => {
    try {
      const data = await apiFetch<Cart>("/cart");
      setCart(data);
    } catch (err: unknown) {
      if (err instanceof Error) {
        if (err.message.includes("401")) {
          router.push("/login?next=/cart");
          return;
        }
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
        <div className="flex items-center gap-3 rounded-2xl bg-red-50 p-4 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-300">
          <AlertCircle className="h-5 w-5 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* INV-01 Prominent Invariant Banner */}
      <div className="rounded-3xl bg-gradient-to-r from-rose-500 via-pink-500 to-purple-600 p-1 text-white shadow-xl shadow-rose-500/20">
        <div className="rounded-[22px] bg-white dark:bg-zinc-950 p-6 sm:p-8 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Sparkles className="h-5 w-5 text-rose-500 animate-spin" />
              <span className="font-bold text-sm tracking-wide text-rose-500 uppercase">
                Инвариант INV-01
              </span>
            </div>
            <h3 className="text-2xl font-black text-zinc-900 dark:text-zinc-50">
              Сумма заказа: строго 10.00 ₽
            </h3>
            <p className="text-xs sm:text-sm text-zinc-500">
              Каталожная стоимость товаров ({formatPrice(cart.total_price_rub)}) носит справочный характер. Итоговый платёж всегда фиксирован.
            </p>
          </div>

          <div className="flex flex-col items-end flex-shrink-0">
            <span className="text-xs text-zinc-400">Итого к оплате</span>
            <span className="text-3xl font-black text-rose-500">10.00 ₽</span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Cart Items List */}
        <div className="lg:col-span-2 space-y-4">
          {cart.items.map((item) => {
            const imageUrl = getProductImageUrl(item.image_seed || item.product_id, item.name, item.category_name);
            return (
              <Card key={item.product_id} className="border-zinc-200/80 dark:border-zinc-800 shadow-sm overflow-hidden">
                <CardContent className="p-4 sm:p-6 flex items-center justify-between gap-4">
                  <div className="flex items-center gap-4">
                    <div className="relative h-20 w-20 rounded-2xl overflow-hidden bg-zinc-100 dark:bg-zinc-800 flex-shrink-0">
                      <Image
                        src={imageUrl}
                        alt={item.name}
                        fill
                        unoptimized
                        className="object-cover"
                      />
                    </div>
                    <div className="space-y-1">
                      <h4 className="font-bold text-base text-zinc-900 dark:text-zinc-100 line-clamp-1">
                        {item.name}
                      </h4>
                      {item.category_name && (
                        <p className="text-xs text-zinc-400">{item.category_name}</p>
                      )}
                      <p className="text-sm font-semibold text-zinc-600 dark:text-zinc-300">
                        {formatPrice(item.price_rub)}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-1.5 border border-zinc-200 dark:border-zinc-700 rounded-xl p-1 bg-zinc-50/50 dark:bg-zinc-900/50">
                      <button
                        onClick={() => handleUpdateQuantity(item.product_id, item.quantity - 1)}
                        disabled={isUpdating}
                        className="p-1 rounded-lg hover:bg-zinc-200 dark:hover:bg-zinc-800 text-zinc-600 dark:text-zinc-300 disabled:opacity-50"
                      >
                        <Minus className="h-4 w-4" />
                      </button>
                      <span className="w-6 text-center text-sm font-bold">
                        {item.quantity}
                      </span>
                      <button
                        onClick={() => handleUpdateQuantity(item.product_id, item.quantity + 1)}
                        disabled={isUpdating}
                        className="p-1 rounded-lg hover:bg-zinc-200 dark:hover:bg-zinc-800 text-zinc-600 dark:text-zinc-300 disabled:opacity-50"
                      >
                        <Plus className="h-4 w-4" />
                      </button>
                    </div>

                    <button
                      onClick={() => handleRemoveItem(item.product_id)}
                      disabled={isUpdating}
                      className="p-2 text-zinc-400 hover:text-red-500 transition-colors"
                    >
                      <Trash2 className="h-5 w-5" />
                    </button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>

        {/* Sidebar Summary & Pickup point */}
        <div className="space-y-6">
          {/* Pickup Point Card */}
          <Card className="border-zinc-200/80 dark:border-zinc-800 shadow-sm">
            <CardHeader className="p-5 pb-3">
              <CardTitle className="text-base flex items-center justify-between">
                <span>Пункт выдачи</span>
                <Link href="/onboarding" className="text-xs font-semibold text-rose-500 hover:text-rose-600">
                  {cart.pickup_point ? "Сменить" : "Выбрать"}
                </Link>
              </CardTitle>
            </CardHeader>
            <CardContent className="p-5 pt-0">
              {cart.pickup_point ? (
                <div className="flex items-start gap-3 p-3 bg-zinc-50 dark:bg-zinc-900 rounded-2xl">
                  <MapPin className="h-5 w-5 text-rose-500 flex-shrink-0 mt-0.5" />
                  <div>
                    <h5 className="font-bold text-sm text-zinc-900 dark:text-zinc-100">
                      {cart.pickup_point.name}
                    </h5>
                    <p className="text-xs text-zinc-500">
                      Дистанция ~{Math.round(cart.pickup_point.distance_meters)}м
                    </p>
                  </div>
                </div>
              ) : (
                <Link href="/onboarding">
                  <Button variant="outline" className="w-full text-xs rounded-xl gap-2">
                    <MapPin className="h-4 w-4 text-rose-500" />
                    <span>Выбрать ближайший ПВЗ</span>
                  </Button>
                </Link>
              )}
            </CardContent>
          </Card>

          {/* Summary & Checkout */}
          <Card className="border-zinc-200/80 dark:border-zinc-800 shadow-sm">
            <CardContent className="p-6 space-y-4">
              <div className="space-y-2 text-sm text-zinc-600 dark:text-zinc-400">
                <div className="flex justify-between">
                  <span>Товары ({cart.total_quantity} шт.)</span>
                  <span>{formatPrice(cart.total_price_rub)}</span>
                </div>
                <div className="flex justify-between">
                  <span>Доставка в ПВЗ</span>
                  <span className="text-emerald-500 font-medium">Бесплатно</span>
                </div>
                <div className="flex justify-between text-rose-500 font-semibold">
                  <span>Dopamine скидка</span>
                  <span>- {formatPrice(parseFloat(cart.total_price_rub) - 10.0 > 0 ? parseFloat(cart.total_price_rub) - 10.0 : 0)}</span>
                </div>
                <div className="border-t border-zinc-200 dark:border-zinc-800 pt-3 flex justify-between text-lg font-black text-zinc-900 dark:text-zinc-100">
                  <span>К оплате</span>
                  <span className="text-rose-500">10.00 ₽</span>
                </div>
              </div>

              <Button
                size="lg"
                variant="glow"
                className="w-full text-base rounded-2xl gap-2"
                isLoading={isCheckingOut}
                onClick={handleCheckout}
              >
                <span>Оформить заказ за 10 ₽</span>
                <ArrowRight className="h-5 w-5" />
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

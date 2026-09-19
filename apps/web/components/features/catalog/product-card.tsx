"use client";

import { useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { apiFetch, isUnauthorizedError } from "@/lib/api/client";
import { getProductImageUrl } from "@/lib/utils/product-image";
import { formatPrice } from "@/lib/utils";

export interface ProductCardProps {
  id: string;
  name: string;
  categoryName?: string;
  priceRub: string;
  imageSeed: string;
  description?: string;
}

export function ProductCard({
  id,
  name,
  categoryName,
  priceRub,
  imageSeed,
}: ProductCardProps) {
  const [isAdding, setIsAdding] = useState(false);
  const [isAdded, setIsAdded] = useState(false);

  const handleAddToCart = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    setIsAdding(true);
    try {
      await apiFetch("/cart/items", {
        method: "POST",
        body: {
          product_id: id,
          quantity: 1,
        },
      });

      setIsAdded(true);
      setTimeout(() => setIsAdded(false), 2000);
    } catch (err) {
      console.error("Failed to add item to cart:", err);
      // If unauthorized, redirect to login
      if (isUnauthorizedError(err)) {
        const nextUrl = typeof window !== "undefined" ? (window.location.pathname + window.location.search) : "/catalog";
        window.location.href = `/login?next=${encodeURIComponent(nextUrl || "/catalog")}`;
      }
    } finally {
      setIsAdding(false);
    }
  };

  const imageUrl = getProductImageUrl(imageSeed || id, name, categoryName);

  return (
    <Link href={`/product/${id}`} className="group block h-full">
      <Card className="h-full overflow-hidden border-[#1E3A50] bg-[#0B1622]/90 hover:border-amber-400/60 hover:shadow-glow-amber transition-all duration-300 flex flex-col justify-between rounded-2xl">
        <div>
          <div className="relative aspect-[4/3] w-full overflow-hidden bg-[#050B14]">
            <Image
              src={imageUrl}
              alt={name}
              fill
              unoptimized
              sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 25vw"
              className="object-cover transition-transform duration-500 group-hover:scale-105"
            />
            {categoryName && (
              <div className="absolute left-2.5 top-2.5">
                <Badge variant="teal" className="backdrop-blur-md text-[10px] font-bold px-2 py-0.5">
                  {categoryName}
                </Badge>
              </div>
            )}
          </div>

          <CardContent className="p-3.5 space-y-1">
            <h3 className="font-bold text-sm sm:text-base text-[#F4F1E8] line-clamp-1 group-hover:text-amber-300 transition-colors">
              {name}
            </h3>
            <div className="text-base sm:text-lg font-black text-[#F4F1E8] tracking-tight">
              {formatPrice(priceRub)}
            </div>
          </CardContent>
        </div>

        <div className="p-3.5 pt-0 mt-auto">
          <Button
            size="sm"
            variant={isAdded ? "secondary" : "gold"}
            className="w-full gap-1.5 rounded-xl font-bold py-2 text-xs sm:text-sm"
            isLoading={isAdding}
            onClick={handleAddToCart}
          >
            {isAdded ? (
              <>
                <Check className="h-4 w-4 text-emerald-400" />
                <span>Добавлено</span>
              </>
            ) : (
              <>
                <span className="text-base leading-none font-black">+</span>
                <span>В корзину</span>
              </>
            )}
          </Button>
        </div>
      </Card>
    </Link>
  );
}

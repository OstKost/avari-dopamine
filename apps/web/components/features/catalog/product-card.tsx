"use client";

import { useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { Check, ShoppingBag } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { apiFetch } from "@/lib/api/client";
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
  description,
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
      if (err instanceof Error && err.message.includes("401")) {
        window.location.href = "/login?next=/catalog";
      }
    } finally {
      setIsAdding(false);
    }
  };

  const imageUrl = getProductImageUrl(imageSeed || id, name, categoryName);

  return (
    <Link href={`/product/${id}`} className="group block h-full">
      <Card className="h-full overflow-hidden border-zinc-200/80 bg-white hover:shadow-xl hover:border-rose-200 dark:border-zinc-800 dark:bg-zinc-900 transition-all duration-300 flex flex-col justify-between">
        <div>
          <div className="relative aspect-[4/3] w-full overflow-hidden bg-zinc-100 dark:bg-zinc-800">
            <Image
              src={imageUrl}
              alt={name}
              fill
              unoptimized
              sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 25vw"
              className="object-cover transition-transform duration-500 group-hover:scale-105"
            />
            {categoryName && (
              <div className="absolute left-3 top-3">
                <Badge variant="secondary" className="backdrop-blur-md bg-white/90 dark:bg-zinc-900/90 text-xs font-medium">
                  {categoryName}
                </Badge>
              </div>
            )}
          </div>

          <CardContent className="p-4 space-y-2">
            <h3 className="font-bold text-base text-zinc-900 dark:text-zinc-100 line-clamp-1 group-hover:text-rose-500 transition-colors">
              {name}
            </h3>
            {description && (
              <p className="text-xs text-zinc-500 dark:text-zinc-400 line-clamp-2">
                {description}
              </p>
            )}
          </CardContent>
        </div>

        <div className="p-4 pt-0 flex items-center justify-between mt-auto">
          <div>
            <span className="text-xs text-zinc-400 block -mb-0.5">В корзине за</span>
            <span className="text-lg font-black text-zinc-900 dark:text-zinc-50">
              {formatPrice(priceRub)}
            </span>
          </div>

          <Button
            size="sm"
            variant={isAdded ? "secondary" : "default"}
            className="gap-1.5 rounded-xl transition-all"
            isLoading={isAdding}
            onClick={handleAddToCart}
          >
            {isAdded ? (
              <>
                <Check className="h-4 w-4 text-emerald-600" />
                <span>В корзине</span>
              </>
            ) : (
              <>
                <ShoppingBag className="h-3.5 w-3.5" />
                <span>Добавить</span>
              </>
            )}
          </Button>
        </div>
      </Card>
    </Link>
  );
}

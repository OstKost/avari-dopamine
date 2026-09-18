import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowLeft, Sparkles, ShieldCheck, Zap, Package } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { apiFetch } from "@/lib/api/client";
import { formatPrice } from "@/lib/utils";
import { getProductImageUrl } from "@/lib/utils/product-image";

export const dynamic = "force-dynamic";

interface Product {
  id: string;
  category_id: string;
  category_name?: string;
  name: string;
  description?: string;
  price_rub: string;
  image_seed: string;
  created_at: string;
}

export default async function ProductDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let product: Product | null = null;
  try {
    const res = await apiFetch<{ products: Product[] }>("/catalog/products", {
      cache: "no-store",
    });
    product = res.products?.find((p) => p.id === id) || null;
  } catch {
    // handled below
  }

  if (!product) {
    notFound();
  }

  const imageUrl = getProductImageUrl(product.image_seed || id, product.name, product.category_name);

  return (
    <div className="container mx-auto max-w-5xl px-4 sm:px-6 py-8 space-y-8">
      <Link
        href="/catalog"
        className="inline-flex items-center gap-2 text-sm text-[#9FB3C4] hover:text-[#F4F1E8] transition-colors"
      >
        <ArrowLeft className="h-4 w-4" />
        <span>Назад в каталог</span>
      </Link>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-10">
        {/* Product Image */}
        <div className="relative aspect-square w-full rounded-3xl overflow-hidden bg-[#0B1622] border border-[#1E3A50] shadow-xl">
          <Image
            src={imageUrl}
            alt={product.name}
            fill
            priority
            unoptimized
            sizes="(max-width: 768px) 100vw, 50vw"
            className="object-cover"
          />
          {product.category_name && (
            <div className="absolute top-4 left-4">
              <Badge variant="secondary" className="backdrop-blur-md bg-[#0B1622]/90 border border-[#1E3A50] text-[#F4F1E8] text-sm font-semibold">
                {product.category_name}
              </Badge>
            </div>
          )}
        </div>

        {/* Product Info */}
        <div className="flex flex-col justify-between space-y-6">
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <span className="px-3 py-1 rounded-full text-xs font-bold bg-amber-400/15 border border-amber-400/30 text-amber-400 flex items-center gap-1.5">
                <Sparkles className="h-3.5 w-3.5" />
                <span>Инвариант INV-01: Любой заказ ровно 10 ₽</span>
              </span>
            </div>

            <h1 className="text-3xl sm:text-4xl font-black text-[#F4F1E8] tracking-tight">
              {product.name}
            </h1>

            <p className="text-base text-[#9FB3C4] leading-relaxed">
              {product.description || "Высококачественный синтетический товар из каталога Avari Dopamine. Гарантия моментального удовольствия."}
            </p>

            <div className="pt-4 border-t border-[#1E3A50]">
              <span className="text-xs text-[#5E7488] block mb-1">Сумма в каталоге</span>
              <div className="flex items-baseline gap-3">
                <span className="text-3xl font-black text-[#F4F1E8]">
                  {formatPrice(product.price_rub)}
                </span>
                <span className="text-xs text-amber-400 font-semibold">
                  (при заказе действует промокод на 10 ₽)
                </span>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <Link href="/catalog" className="block">
              <Button size="lg" variant="reward" className="w-full gap-2 text-base rounded-2xl shadow-lg shadow-amber-500/20">
                <Package className="h-5 w-5" />
                <span>Добавить и перейти к покупкам</span>
              </Button>
            </Link>

            <div className="grid grid-cols-2 gap-3 pt-2">
              <Card className="border-[#1E3A50] bg-[#0B1622]/60">
                <CardContent className="p-3.5 flex items-center gap-3">
                  <ShieldCheck className="h-5 w-5 text-teal-400 flex-shrink-0" />
                  <span className="text-xs text-[#9FB3C4] font-medium">
                    Синтетический оригинал
                  </span>
                </CardContent>
              </Card>

              <Card className="border-[#1E3A50] bg-[#0B1622]/60">
                <CardContent className="p-3.5 flex items-center gap-3">
                  <Zap className="h-5 w-5 text-amber-400 flex-shrink-0" />
                  <span className="text-xs text-[#9FB3C4] font-medium">
                    ПВЗ в радиусе 100-500м
                  </span>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

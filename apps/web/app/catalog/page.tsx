import Link from "next/link";
import { ShoppingBag, MapPin, ChevronRight } from "lucide-react";
import { SearchBar } from "@/components/features/catalog/search-bar";
import { ProductCard } from "@/components/features/catalog/product-card";
import { Badge } from "@/components/ui/badge";
import { apiFetch } from "@/lib/api/client";

export const dynamic = "force-dynamic";

interface Category {
  id: string;
  name: string;
  slug: string;
  description: string;
}

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

interface CategoriesResponse {
  categories: Category[];
}

interface ProductsResponse {
  products: Product[];
  total: number;
  limit: number;
  offset: number;
}

export default async function CatalogPage({
  searchParams,
}: {
  searchParams: Promise<{ category_id?: string; q?: string }>;
}) {
  const { category_id, q } = await searchParams;

  let categories: Category[] = [];
  let products: Product[] = [];
  let total = 0;

  try {
    const [catRes, prodRes] = await Promise.all([
      apiFetch<CategoriesResponse>("/catalog/categories", { cache: "no-store" }),
      apiFetch<ProductsResponse>("/catalog/products", {
        params: {
          query: {
            category_id,
            q,
            limit: 50,
          },
        },
        cache: "no-store",
      }),
    ]);

    categories = catRes.categories || [];
    products = prodRes.products || [];
    total = prodRes.total || 0;
  } catch (err) {
    console.error("Error fetching catalog data:", err);
  }

  return (
    <div className="container mx-auto max-w-7xl px-4 sm:px-6 py-6 sm:py-8 space-y-6">
      {/* Pickup Point Selection Pill */}
      <Link href="/onboarding" className="block group">
        <div className="flex items-center justify-between p-3.5 px-4 rounded-2xl bg-[#0B1622]/90 border border-[#1E3A50] group-hover:border-teal-400/50 group-hover:shadow-glow-teal transition-all">
          <div className="flex items-center gap-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-amber-400/15 text-amber-400 border border-amber-400/30">
              <MapPin className="h-4 w-4" />
            </div>
            <div>
              <span className="text-[11px] font-semibold text-[#9FB3C4] block">Пункт выдачи</span>
              <span className="text-sm font-bold text-[#F4F1E8] group-hover:text-amber-300 transition-colors">
                ул. Малая Садовая, 12 (240м)
              </span>
            </div>
          </div>
          <div className="flex items-center gap-1 text-xs font-bold text-teal-400">
            <span>Сменить</span>
            <ChevronRight className="h-4 w-4" />
          </div>
        </div>
      </Link>

      {/* Top Banner & Search */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4">
        <div className="space-y-1">
          <h1 className="text-2xl sm:text-3xl font-black text-[#F4F1E8] tracking-tight">
            Каталог дофамина
          </h1>
          <p className="text-xs sm:text-sm text-[#9FB3C4]">
            {total} товаров · Любой заказ за 10 ₽ по промокоду
          </p>
        </div>

        <SearchBar />
      </div>

      {/* Category Pills matching screen-01 */}
      <div className="flex items-center gap-2.5 overflow-x-auto pb-2 scrollbar-none">
        <Link href="/catalog">
          <div
            className={`px-5 py-2 text-xs sm:text-sm font-bold rounded-full cursor-pointer transition-all ${
              !category_id
                ? "bg-amber-400/20 text-amber-300 border border-amber-400 shadow-glow-amber"
                : "bg-[#0B1622] text-[#9FB3C4] border border-[#1E3A50] hover:border-teal-400/50 hover:text-[#F4F1E8]"
            }`}
          >
            Все
          </div>
        </Link>
        {categories.map((cat) => {
          const isActive = category_id === cat.id;
          return (
            <Link key={cat.id} href={`/catalog?category_id=${cat.id}${q ? `&q=${encodeURIComponent(q)}` : ""}`}>
              <div
                className={`px-5 py-2 text-xs sm:text-sm font-bold rounded-full cursor-pointer whitespace-nowrap transition-all ${
                  isActive
                    ? "bg-amber-400/20 text-amber-300 border border-amber-400 shadow-glow-amber"
                    : "bg-[#0B1622] text-[#9FB3C4] border border-[#1E3A50] hover:border-teal-400/50 hover:text-[#F4F1E8]"
                }`}
              >
                {cat.name}
              </div>
            </Link>
          );
        })}
      </div>

      {/* Product Grid (2 cols mobile, 3 tablet, 4 desktop) */}
      {products.length === 0 ? (
        <div className="flex min-h-[40vh] flex-col items-center justify-center text-center p-8 border border-dashed border-[#1E3A50] rounded-3xl bg-[#0B1622]/40">
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-[#122234] text-amber-400 mb-4 border border-[#1E3A50]">
            <ShoppingBag className="h-8 w-8" />
          </div>
          <h3 className="text-lg font-bold text-[#F4F1E8]">
            Ничего не найдено
          </h3>
          <p className="text-sm text-[#9FB3C4] max-w-sm mt-1">
            Попробуйте изменить поисковый запрос или выбрать другую категорию.
          </p>
          <Link href="/catalog" className="mt-4">
            <Badge variant="gold" className="px-4 py-1.5 cursor-pointer">
              Сбросить фильтры
            </Badge>
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3.5 sm:gap-6">
          {products.map((product) => (
            <ProductCard
              key={product.id}
              id={product.id}
              name={product.name}
              categoryName={product.category_name}
              description={product.description}
              priceRub={product.price_rub}
              imageSeed={product.image_seed}
            />
          ))}
        </div>
      )}
    </div>
  );
}

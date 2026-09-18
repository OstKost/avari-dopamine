import Link from "next/link";
import { Sparkles, ShoppingBag } from "lucide-react";
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
    <div className="container mx-auto max-w-7xl px-4 sm:px-6 py-8 space-y-8">
      {/* Top Banner & Search */}
      <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 pb-6 border-b border-zinc-200 dark:border-zinc-800">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Badge variant="dopamine" className="gap-1">
              <Sparkles className="h-3.5 w-3.5 text-rose-500" />
              10 ₽ фиксированная цена
            </Badge>
          </div>
          <h1 className="text-3xl font-black text-zinc-900 dark:text-zinc-50 tracking-tight">
            Каталог товаров
          </h1>
          <p className="text-sm text-zinc-500">
            Найдено {total} синтетических товаров
          </p>
        </div>

        <SearchBar />
      </div>

      {/* Category Pills */}
      <div className="flex items-center gap-2 overflow-x-auto pb-2 scrollbar-none">
        <Link href="/catalog">
          <Badge
            variant={!category_id ? "default" : "secondary"}
            className="px-4 py-2 text-sm rounded-xl cursor-pointer hover:bg-zinc-800 dark:hover:bg-zinc-200 transition-colors"
          >
            Все категории
          </Badge>
        </Link>
        {categories.map((cat) => {
          const isActive = category_id === cat.id;
          return (
            <Link key={cat.id} href={`/catalog?category_id=${cat.id}${q ? `&q=${encodeURIComponent(q)}` : ""}`}>
              <Badge
                variant={isActive ? "default" : "secondary"}
                className="px-4 py-2 text-sm rounded-xl cursor-pointer whitespace-nowrap hover:bg-zinc-200 dark:hover:bg-zinc-700 transition-colors"
              >
                {cat.name}
              </Badge>
            </Link>
          );
        })}
      </div>

      {/* Product Grid */}
      {products.length === 0 ? (
        <div className="flex min-h-[40vh] flex-col items-center justify-center text-center p-8 border border-dashed border-zinc-300 dark:border-zinc-800 rounded-3xl">
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-400 dark:bg-zinc-800 mb-4">
            <ShoppingBag className="h-8 w-8" />
          </div>
          <h3 className="text-lg font-bold text-zinc-900 dark:text-zinc-100">
            Ничего не найдено
          </h3>
          <p className="text-sm text-zinc-500 max-w-sm mt-1">
            Попробуйте изменить поисковый запрос или выбрать другую категорию.
          </p>
          <Link href="/catalog" className="mt-4">
            <Badge variant="outline" className="px-4 py-1.5 cursor-pointer">
              Сбросить фильтры
            </Badge>
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
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

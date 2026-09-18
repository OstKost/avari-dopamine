"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useTransition, useState, useEffect } from "react";
import { Search, X } from "lucide-react";
import { Input } from "@/components/ui/input";

export function SearchBar() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const initialQuery = searchParams.get("q") || "";

  const [query, setQuery] = useState(initialQuery);
  const [, startTransition] = useTransition();

  useEffect(() => {
    setQuery(initialQuery);
  }, [initialQuery]);

  const handleSearch = (term: string) => {
    setQuery(term);

    startTransition(() => {
      const params = new URLSearchParams(searchParams.toString());
      if (term.trim()) {
        params.set("q", term.trim());
      } else {
        params.delete("q");
      }
      params.set("offset", "0");
      router.replace(`${pathname}?${params.toString()}`);
    });
  };

  return (
    <div className="relative w-full max-w-md">
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
      <Input
        type="text"
        placeholder="Поиск дофаминовых товаров..."
        value={query}
        onChange={(e) => handleSearch(e.target.value)}
        className="pl-9 pr-8 bg-white/80 dark:bg-zinc-900/80 backdrop-blur-sm"
      />
      {query && (
        <button
          onClick={() => handleSearch("")}
          className="absolute right-2.5 top-1/2 -translate-y-1/2 p-1 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-200"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      )}
    </div>
  );
}

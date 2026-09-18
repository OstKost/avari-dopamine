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
      <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-teal-400" />
      <Input
        type="text"
        placeholder="Найти дофамин..."
        value={query}
        onChange={(e) => handleSearch(e.target.value)}
        className="pl-10 pr-9 bg-[#0E1B29]/90 border-[#1E3A50] focus-visible:border-teal-400 text-[#F4F1E8] placeholder:text-[#5E7488] rounded-2xl shadow-sm"
      />
      {query && (
        <button
          onClick={() => handleSearch("")}
          className="absolute right-3 top-1/2 -translate-y-1/2 p-1 text-[#5E7488] hover:text-[#F4F1E8]"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      )}
    </div>
  );
}

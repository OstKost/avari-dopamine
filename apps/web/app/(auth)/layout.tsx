import Link from "next/link";
import { Sparkles } from "lucide-react";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-[calc(100vh-8rem)] flex-col items-center justify-center px-4 py-12 sm:px-6 lg:px-8">
      <div className="w-full max-w-md space-y-8">
        <div className="flex flex-col items-center text-center">
          <Link href="/" className="flex items-center gap-2 group mb-2">
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-tr from-rose-500 to-purple-600 text-white shadow-lg shadow-rose-500/20 group-hover:scale-105 transition-transform">
              <Sparkles className="h-6 w-6 animate-pulse" />
            </div>
          </Link>
          <h2 className="text-2xl font-black tracking-tight bg-gradient-to-r from-rose-600 via-pink-600 to-purple-600 bg-clip-text text-transparent">
            Dopamine Market
          </h2>
        </div>

        {children}
      </div>
    </div>
  );
}

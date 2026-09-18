import Link from "next/link";
import Image from "next/image";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-[calc(100vh-12rem)] flex-col items-center justify-center px-4 py-12 sm:px-6 lg:px-8">
      <div className="w-full max-w-md space-y-8">
        <div className="flex flex-col items-center text-center">
          <Link href="/" className="flex items-center gap-3 group mb-3">
            <div className="relative flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-tr from-[#1E3A50] to-[#0E1B29] border border-[#1E3A50] shadow-xl group-hover:scale-105 group-hover:border-amber-400/50 transition-all p-2.5">
              <Image
                src="/logo-detailed.png"
                alt="Avari Dopamine Logo"
                width={40}
                height={40}
                className="object-contain"
              />
            </div>
          </Link>
          <h2 className="text-2xl font-black tracking-tight text-[#F4F1E8]">
            Avari Dopamine
          </h2>
          <p className="text-xs text-[#9FB3C4] mt-1">
            Синтетический маркетплейс мгновенного удовольствия
          </p>
        </div>

        {children}
      </div>
    </div>
  );
}

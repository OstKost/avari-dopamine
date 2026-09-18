import { Spinner } from "@/components/ui/spinner";

export default function Loading() {
  return (
    <div className="container mx-auto flex min-h-[50vh] items-center justify-center">
      <div className="flex flex-col items-center gap-4">
        <Spinner size="lg" />
        <p className="text-sm font-medium text-zinc-500 animate-pulse">Загрузка...</p>
      </div>
    </div>
  );
}

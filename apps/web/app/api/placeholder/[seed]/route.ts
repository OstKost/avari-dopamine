import { NextRequest, NextResponse } from "next/server";
import { getProductImageUrl } from "@/lib/utils/product-image";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ seed: string }> }
) {
  const { seed } = await params;
  const searchParams = request.nextUrl.searchParams;
  const name = searchParams.get("name") || undefined;
  const category = searchParams.get("category") || undefined;

  const dataUri = getProductImageUrl(seed, name, category);
  const svgContent = decodeURIComponent(dataUri.replace("data:image/svg+xml;charset=utf-8,", ""));

  return new NextResponse(svgContent, {
    status: 200,
    headers: {
      "Content-Type": "image/svg+xml",
      "Cache-Control": "public, max-age=31536000, immutable",
    },
  });
}

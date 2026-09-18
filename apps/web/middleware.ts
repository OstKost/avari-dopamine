import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { parseJwt } from "./lib/auth/auth";

const PROTECTED_ROUTES = ["/cart", "/checkout", "/orders"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  const isProtected = PROTECTED_ROUTES.some((route) => pathname.startsWith(route));

  if (isProtected) {
    const token = request.cookies.get("access_token")?.value;
    const payload = token ? parseJwt(token) : null;

    if (!payload) {
      const loginUrl = new URL("/login", request.url);
      loginUrl.searchParams.set("next", pathname);
      return NextResponse.redirect(loginUrl);
    }
  }

  // Redirect authenticated users from /login or /register to home
  if (pathname === "/login" || pathname === "/register") {
    const token = request.cookies.get("access_token")?.value;
    const payload = token ? parseJwt(token) : null;

    if (payload) {
      return NextResponse.redirect(new URL("/", request.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public assets
     */
    "/((?!api|_next/static|_next/image|favicon.ico|.*\\..*).*)",
  ],
};

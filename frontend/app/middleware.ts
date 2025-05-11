import { NextResponse } from "next/server"
import type { NextRequest } from "next/server"

export function middleware(request: NextRequest) {
  const token = request.cookies.get("token")?.value
  
  if (!token) {
    const isProtectedRoute = 
      request.nextUrl.pathname === "/" || 
      request.nextUrl.pathname === "/history" ||
      request.nextUrl.pathname.startsWith("/api/")
      
    if (isProtectedRoute && request.nextUrl.pathname !== "/auth") {
      return NextResponse.redirect(new URL("/auth", request.url))
    }
  }
  
  if (token && request.nextUrl.pathname === "/auth") {
    return NextResponse.redirect(new URL("/", request.url))
  }
  
  return NextResponse.next()
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|public).*)",
  ],
}
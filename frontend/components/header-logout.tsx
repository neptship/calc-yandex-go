"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { LogOut, User } from "lucide-react"
import { useAuth } from "@/contexts/auth-context"

export default function Header() {
  const pathname = usePathname()
  const { user, logout } = useAuth()
  
  if (pathname === "/auth" || pathname === "/auth/register") {
    return null
  }

  return (
    <header className="fixed top-0 left-0 right-0 bg-black/80 backdrop-blur-sm border-b border-white/10 z-10">
      <div className="container mx-auto px-4 flex justify-between items-center h-16">
        <Link href="/" className="text-white font-bold text-xl">
          neptship.go
        </Link>
        
        <div className="flex items-center gap-6">
          <Link 
            href="/history" 
            className={`text-sm ${pathname === "/history" ? "text-white" : "text-white/70 hover:text-white"} transition-colors`}
          >
            История
          </Link>
          
          <div className="relative group">
            <button className="flex items-center gap-2 text-sm text-white/70 hover:text-white transition-colors">
              <User size={16} />
              <span>{user?.username}</span>
            </button>
            
            <div className="absolute right-0 top-full mt-2 w-48 py-1 bg-black border border-white/10 rounded-md shadow-lg hidden group-hover:block">
              <button
                onClick={logout}
                className="w-full px-4 py-2 text-left text-sm text-white/70 hover:text-white hover:bg-white/5 transition-colors flex items-center gap-2"
              >
                <LogOut size={14} />
                <span>Выйти</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </header>
  )
}
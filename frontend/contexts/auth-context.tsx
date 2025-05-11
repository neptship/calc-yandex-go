"use client"

import { createContext, useContext, useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import type { ReactNode } from "react"

type User = {
  id: number
  username: string
}

type AuthContextType = {
  user: User | null
  token: string | null
  login: (token: string, userData: User) => void
  logout: () => void
  isLoading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()

  useEffect(() => {
    try {
      const storedToken = localStorage.getItem("token")
      const storedUser = localStorage.getItem("user")
      
      if (storedToken && storedUser) {
        setToken(storedToken)
        setUser(JSON.parse(storedUser))
      }
    } catch (error) {
      localStorage.removeItem("token")
      localStorage.removeItem("user")
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!isLoading) {
      const path = window.location.pathname
      
      if (token && (path === "/auth" || path === "/auth/register")) {
        router.push("/")
      }
      
      if (!token && path !== "/auth" && path !== "/auth/register") {
        router.push("/auth/register")
      }
    }
  }, [isLoading, token, router])

  const login = (newToken: string, userData: User) => {
    setToken(newToken)
    setUser(userData)
    
    localStorage.setItem("token", newToken)
    localStorage.setItem("user", JSON.stringify(userData))
  }

  const logout = () => {
    setToken(null)
    setUser(null)
    
    localStorage.removeItem("token")
    localStorage.removeItem("user")
    
    router.push("/auth")
  }

  return (
    <AuthContext.Provider value={{ user, token, login, logout, isLoading }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
}
"use client"

import Calculator from "@/components/calculator"
import { useAuth } from "@/contexts/auth-context"
import { useRouter } from "next/navigation"
import { useEffect } from "react"

export default function Home() {
  const { token, isLoading } = useAuth()
  const router = useRouter()
  
  useEffect(() => {
    if (!isLoading && !token) {
      router.push("/auth")
    }
  }, [isLoading, token, router])
  
  if (isLoading || !token) {
    return (
      <main className="min-h-screen bg-black flex flex-col items-center justify-center p-4">
        <div className="text-white">Loading...</div>
      </main>
    )
  }

  return (
    <main className="min-h-screen bg-black flex flex-col items-center justify-center p-4">
      <Calculator />
    </main>
  )
}
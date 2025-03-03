import Calculator from "@/components/calculator"
import Link from "next/link"
import { Suspense } from "react"

export default function Home() {
  return (
    <main className="min-h-screen bg-black flex items-center justify-center p-4">
    <Suspense fallback={<div>Loading...</div>}>
      <Calculator />
    </Suspense>
    </main>
  )
}


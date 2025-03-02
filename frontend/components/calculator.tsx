"use client"

import type React from "react"

import { useState, useEffect } from "react"
import { useSearchParams } from "next/navigation"
import { ArrowRight } from "lucide-react"

export default function Calculator() {
  const searchParams = useSearchParams()
  const [expression, setExpression] = useState("")
  const [result, setResult] = useState<string | null>(null)

  useEffect(() => {
    const expressionParam = searchParams.get("expression")
    const resultParam = searchParams.get("result")

    if (expressionParam) {
      setExpression(expressionParam)
    }

    if (resultParam) {
      setResult(resultParam)
    }
  }, [searchParams])

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    if (/^[0-9+\-*/().%\s]*$/.test(value)) {
      setExpression(value)
    }
  }

  return (
    <div className="w-full max-w-md space-y-4 animate-fade-in">
      <div className="relative">
        <input
          type="text"
          value={expression}
          onChange={handleInputChange}
          placeholder="Введите выражение..."
          className="w-full bg-black border border-white/20 rounded-xl px-4 py-3 text-white placeholder:text-white/50 focus:outline-none focus:border-white/40 transition-colors"
        />
        {result && <div className="absolute right-4 top-1/2 -translate-y-1/2 text-white/70">= {result}</div>}
      </div>
      <form action="/api/calculate" method="POST" className="flex gap-2">
        <button
          type="submit"
          className="flex items-center gap-2 px-4 py-2 bg-white/10 hover:bg-white/15 text-white rounded-xl transition-colors"
        >
          Вычислить
          <ArrowRight className="w-4 h-4" />
        </button>
      </form>
    </div>
  )
}


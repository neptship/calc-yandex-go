"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Eye, EyeOff, Lock, User, Loader2 } from "lucide-react"
import Link from "next/link"
import { useAuth } from "@/contexts/auth-context"

export default function RegisterPage() {
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const router = useRouter()
  const { login } = useAuth()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError("")

    try {
      console.log("Sending registration request with:", { login: username, password });
      
      const response = await fetch("http://localhost:8080/api/v1/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ 
          login: username,
          password: password 
        }),
      })

      const data = await response.json()
      console.log("Registration response:", data);

      if (!response.ok) {
        throw new Error(data.error || "Ошибка регистрации")
      }

      if (data.token) {
        const userId = data.user?.id || data.id || 0;
        const userLogin = data.user?.login || data.login || username;
        
        login(data.token, {
          id: userId,
          username: userLogin
        })

        router.push("/")
      } else {
        setError("Регистрация успешна. Пожалуйста, войдите в систему.")
        setTimeout(() => {
          router.push("/auth")
        }, 2000)
      }
    } catch (error) {
      console.error("Registration error:", error);
      setError(error instanceof Error ? error.message : "Ошибка регистрации")
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <main className="min-h-screen bg-black flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-md animate-fade-in">
        <div className="mb-8 text-center">
          <h1 className="text-2xl font-bold text-white mb-2">Регистрация</h1>
          <p className="text-white/50">Создайте аккаунт для доступа к калькулятору</p>
        </div>

        {error && (
          <div className={`mb-4 p-3 rounded-lg text-sm ${
            error.includes("успешна") 
              ? "bg-green-500/10 border border-green-500/20 text-green-400"
              : "bg-red-500/10 border border-red-500/20 text-red-400"
          }`}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="username" className="block text-sm text-white/70">
              Имя пользователя
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-white/40">
                <User size={18} />
              </div>
              <input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full bg-black border border-white/20 rounded-xl pl-10 pr-4 py-3 text-white placeholder:text-white/30 focus:outline-none focus:border-white/40 transition-colors"
                placeholder="Введите имя пользователя"
                required
              />
            </div>
          </div>

          <div className="space-y-2">
            <label htmlFor="password" className="block text-sm text-white/70">
              Пароль
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-white/40">
                <Lock size={18} />
              </div>
              <input
                id="password"
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full bg-black border border-white/20 rounded-xl pl-10 pr-10 py-3 text-white placeholder:text-white/30 focus:outline-none focus:border-white/40 transition-colors"
                placeholder="Введите пароль"
                required
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute inset-y-0 right-0 flex items-center pr-3 text-white/40 hover:text-white/70 transition-colors"
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </div>

          <div className="pt-2">
            <button
              type="submit"
              disabled={isLoading}
              className="w-full bg-white/10 hover:bg-white/15 text-white rounded-xl py-3 transition-colors disabled:opacity-50"
            >
              {isLoading ? (
                <div className="flex items-center justify-center">
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Регистрация...
                </div>
              ) : (
                "Зарегистрироваться"
              )}
            </button>
          </div>
        </form>

        <div className="mt-6 text-center">
          <p className="text-white/50 text-sm">
            Уже есть аккаунт? <Link href="/auth" className="text-white hover:underline">Войти</Link>
          </p>
        </div>
      </div>
    </main>
  )
}
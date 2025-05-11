"use client";

import { useState, useEffect, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { ArrowRight, Loader2, LogOut, History } from "lucide-react";
import Link from "next/link";
import { useAuth } from "@/contexts/auth-context";

export default function Calculator() {
  const searchParams = useSearchParams();
  const [expression, setExpression] = useState("");
  const [result, setResult] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [expressionId, setExpressionId] = useState<number | null>(null);
  const [history, setHistory] = useState<any[]>([]);
  const pollInterval = useRef<NodeJS.Timeout | null>(null);
  const { token, logout, user } = useAuth();

  useEffect(() => {
    const savedHistory = localStorage.getItem("calculatorHistory");
    if (savedHistory) {
      try {
        setHistory(JSON.parse(savedHistory));
      } catch (e) {
        console.error("Failed to parse history:", e);
      }
    }
    const expressionParam = searchParams.get("expression");
    const resultParam = searchParams.get("result");

    if (expressionParam) {
      setExpression(expressionParam);
    }

    if (resultParam) {
      setResult(resultParam);
    }
  }, [searchParams]);

  const saveToHistory = (historyItem: any) => {
    try {
      const skipErrors = ["Expression cannot be empty", "= Expression cannot be empty"];
      
      if (skipErrors.includes(historyItem.result)) {
        console.log(`Пропускаем сохранение ошибки: ${historyItem.result}`);
        return;
      }
      
      const storageKey = user?.username ? `calculationsHistory_${user.username}` : "calculationsHistory";
      console.log(`Сохраняем историю для пользователя ${user?.username} с ключом ${storageKey}`);
      
      const savedHistory = JSON.parse(localStorage.getItem(storageKey) || "[]");
      const updatedHistory = [historyItem, ...savedHistory].slice(0, 50);
      localStorage.setItem(storageKey, JSON.stringify(updatedHistory));
      
      setHistory(updatedHistory);
    } catch (e) {
      console.error("Failed to save to history:", e);
    }
  };

  useEffect(() => {
    if (expressionId !== null && loading) {
      const pollResult = async () => {
        try {
          if (!token) {
            throw new Error("Not authenticated");
          }

          const response = await fetch(`http://localhost:8080/api/v1/expressions/${expressionId}`, {
            headers: {
              "Authorization": `Bearer ${token}`
            }
          });

          if (response.status === 401) {
            logout();
            return;
          }
          
          if (!response.ok) {
            throw new Error("Ошибка получения результата");
          }
          
          const data = await response.json();
          const fetchedExpr = data.expression;
          
          if (fetchedExpr.status === "completed" || fetchedExpr.status === "failed") {
            if (pollInterval.current) {
              clearInterval(pollInterval.current);
              pollInterval.current = null;
            }
            
            if (fetchedExpr.status === "completed" && fetchedExpr.result !== null) {
              setResult(fetchedExpr.result.toString());
              
              const updatedExpression = { ...fetchedExpr };
              delete updatedExpression.expression;
              
              const dataWithoutDuplication = {
                ...data,
                expression: updatedExpression
              };
              
              saveToHistory({
                expression: expression,
                result: fetchedExpr.result.toString(),
                timestamp: new Date().toISOString(),
                fullData: dataWithoutDuplication
              });
            } else {
              setResult("Ошибка вычисления");
  
              const { expression: _, ...dataWithoutExpression } = data;
              
              const skipErrorsInHistory = ["Expression cannot be empty", "= Expression cannot be empty"];
              if (!skipErrorsInHistory.includes("Ошибка вычисления")) {
                saveToHistory({
                  expression: expression,
                  result: "Ошибка вычисления",
                  timestamp: new Date().toISOString(),
                  fullData: {
                    expression: {
                      id: fetchedExpr.id,
                      status: fetchedExpr.status
                    }
                  }
                });
              }
            }
            
            setLoading(false);
          }
        } catch (error) {
          console.error("Ошибка при опросе результата:", error);
          
          if (pollInterval.current) {
            clearInterval(pollInterval.current);
            pollInterval.current = null;
          }
          
          setResult("Ошибка запроса");
          setLoading(false);
        }
      };

      pollResult();
      
      pollInterval.current = setInterval(pollResult, 1000);

      return () => {
        if (pollInterval.current) {
          clearInterval(pollInterval.current);
          pollInterval.current = null;
        }
      };
    }
  }, [expressionId, loading, token, logout, expression]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    if (/^[0-9+\-*/().%\s]*$/.test(value)) {
      setExpression(value);
    }
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    
    if (!token) {
      return;
    }
    
    setLoading(true);
    setResult(null);
    setExpressionId(null);
    
    try {
      const res = await fetch("http://localhost:8080/api/v1/calculate", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify({ expression }),
      });

      if (res.status === 401) {
        logout();
        return;
      }
      
      const data = await res.json();
      
      if (!res.ok || data.error) {
        let errorMsg = data.error;
        
        const skipErrorsInHistory = ["Expression cannot be empty", "= Expression cannot be empty"];
        
        if (data.error === "Invalid expression" || data.error === "invalid expression") {
          errorMsg = "Недопустимое выражение";
          setResult(errorMsg);
        } else if (!data.error) {
          errorMsg = "Неизвестная ошибка";
          
          if (!skipErrorsInHistory.includes(errorMsg)) {
            saveToHistory({
              expression: expression,
              result: errorMsg,
              timestamp: new Date().toISOString(),
            });
          }
        } else {
          if (!skipErrorsInHistory.includes(errorMsg)) {
            saveToHistory({
              expression: expression,
              result: errorMsg,
              timestamp: new Date().toISOString(),
            });
          }
        }
        
        setLoading(false);
        return;
      }
      
      if (data.status === "completed" && data.result !== undefined) {
        setResult(data.result.toString());
        setLoading(false);
        
        const { expression: _, ...dataWithoutExpression } = data;
        
        saveToHistory({
          expression: expression,
          result: data.result.toString(),
          timestamp: new Date().toISOString(),
          fullData: dataWithoutExpression
        });
        
        return;
      }
      
      setExpressionId(data.id);
    } catch (error) {
      console.error("Error during calculation:", error);
      setResult("Ошибка запроса");
      setLoading(false);
    }
  };

  return (
    <div className="w-full max-w-md space-y-4 animate-fade-in relative">
      <div className="absolute -top-12 right-0 flex items-center gap-4">
        <Link href="/history" className="text-white/70 hover:text-white transition-colors flex items-center gap-1">
          <History size={16} />
          <span>История</span>
        </Link>
        <button 
          onClick={logout} 
          className="text-white/70 hover:text-white transition-colors flex items-center gap-1"
        >
          <LogOut size={16} />
          <span>Выйти</span>
        </button>
      </div>
      <div className="relative">
        <input
          type="text"
          value={expression}
          onChange={handleInputChange}
          placeholder="Введите выражение..."
          className="w-full bg-black border border-white/20 rounded-xl px-4 py-3 text-white placeholder:text-white/50 focus:outline-none focus:border-white/40 transition-colors"
        />
        {result && (
          <div className="absolute right-4 top-1/2 -translate-y-1/2 text-white/70">
            {result === "Недопустимое выражение" || result === "Неизвестная ошибка" || result === "Ошибка запроса" || result === "Ошибка вычисления" ? 
              result : 
              `= ${result}`
            }
          </div>
        )}
      </div>
      <form onSubmit={handleSubmit} className="flex gap-2">
        <button
          type="submit"
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 bg-white/10 hover:bg-white/15 text-white rounded-xl transition-colors disabled:opacity-50"
        >
          {loading ? (
            <>
              <Loader2 className="w-4 h-4 animate-spin" />
              <span>Вычисляется...</span>
            </>
          ) : (
            <>
              <span>Вычислить</span>
              <ArrowRight className="w-4 h-4" />
            </>
          )}
        </button>
      </form>
    </div>
  );
}
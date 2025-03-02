"use client";

import { useState, useEffect, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { ArrowRight, Loader2 } from "lucide-react";

export default function Calculator() {
  const searchParams = useSearchParams();
  const [expression, setExpression] = useState("");
  const [result, setResult] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [expressionId, setExpressionId] = useState<number | null>(null);
  const pollInterval = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    const expressionParam = searchParams.get("expression");
    const resultParam = searchParams.get("result");

    if (expressionParam) {
      setExpression(expressionParam);
    }

    if (resultParam) {
      setResult(resultParam);
    }
  }, [searchParams]);

  useEffect(() => {
    if (expressionId !== null && loading) {
      const pollResult = async () => {
        try {
          const response = await fetch(`http://localhost:8080/api/v1/expressions/${expressionId}`);
          if (!response.ok) {
            throw new Error("Ошибка получения результата");
          }
          
          const data = await response.json();
          const expression = data.expression;
          
          if (expression.status === "completed" || expression.status === "failed") {
            if (pollInterval.current) {
              clearInterval(pollInterval.current);
              pollInterval.current = null;
            }
            
            if (expression.status === "completed" && expression.result !== null) {
              setResult(expression.result.toString());
            } else {
              setResult("Ошибка вычисления");
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
  }, [expressionId, loading]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    if (/^[0-9+\-*/().%\s]*$/.test(value)) {
      setExpression(value);
    }
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);
    setResult(null);
    setExpressionId(null);
    
    try {
      const res = await fetch("http://localhost:8080/api/v1/calculate", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ expression }),
      });
      
      console.log("Статус ответа:", res.status);
      
      const data = await res.json();
      console.log("Данные ответа:", data);
      
      if (!res.ok || data.error) {
        // Проверяем текст ошибки и подменяем английский вариант на русский
        if (data.error === "Invalid expression" || data.error === "invalid expression") {
          setResult("Недопустимое выражение");
        } else {
          setResult(data.error || "Неизвестная ошибка");
        }
        setLoading(false);
        return;
      }
      
      if (data.status === "completed" && data.result !== undefined) {
        setResult(data.result);
        setLoading(false);
        return;
      }
      
      setExpressionId(data.id);
    } catch (error) {
      console.error("Ошибка запроса:", error);
      setResult("Ошибка запроса");
      setLoading(false);
    }
  };

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
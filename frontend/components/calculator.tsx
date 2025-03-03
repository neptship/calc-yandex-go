"use client";

import { useState, useEffect, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { ArrowRight, Loader2 } from "lucide-react";
import Link from "next/link";

export default function Calculator() {
  const searchParams = useSearchParams();
  const [expression, setExpression] = useState("");
  const [result, setResult] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [expressionId, setExpressionId] = useState<number | null>(null);
  const [history, setHistory] = useState<string[]>([])
  const pollInterval = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    const savedHistory = localStorage.getItem("calculatorHistory")
    if (savedHistory) {
      setHistory(JSON.parse(savedHistory))
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
              
              // Удаляем expression из data перед сохранением в историю
              const updatedExpression = { ...data.expression };
              delete updatedExpression.expression;
              
              const dataWithoutDuplication = {
                ...data,
                expression: updatedExpression
              };
              
              saveToHistory({
                expression: expression.expression,
                result: expression.result.toString(),
                timestamp: new Date().toISOString(),
                fullData: dataWithoutDuplication
              });
            } else {
              setResult("Ошибка вычисления");
              
              // Удаляем expression из data перед сохранением в историю
              const { expression: _, ...dataWithoutExpression } = data;
              
              saveToHistory({
                expression: expression.expression,
                result: "Ошибка вычисления",
                timestamp: new Date().toISOString(),
                fullData: dataWithoutExpression
              });
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

  const saveToHistory = (historyItem: any) => {
    const savedHistory = JSON.parse(localStorage.getItem("calculationsHistory") || "[]");
    const updatedHistory = [historyItem, ...savedHistory].slice(0, 50);
    localStorage.setItem("calculationsHistory", JSON.stringify(updatedHistory));
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
      
      const data = await res.json();
      
      if (!res.ok || data.error) {
        let errorMsg = data.error;
        if (data.error === "Invalid expression" || data.error === "invalid expression") {
          errorMsg = "Недопустимое выражение";
        } else if (!data.error) {
          errorMsg = "Неизвестная ошибка";
        }
        
        setResult(errorMsg);
        
        const { expression: _, ...dataWithoutExpression } = data;
        
        saveToHistory({
          expression: expression,
          result: errorMsg,
          timestamp: new Date().toISOString(),
          fullData: dataWithoutExpression
        });
        
        setLoading(false);
        return;
      }
      
      if (data.status === "completed" && data.result !== undefined) {
        setResult(data.result);
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
      const errorMsg = "Ошибка запроса";
      setResult(errorMsg);
      
      // Сохраняем ошибку запроса в историю
      saveToHistory({
        expression: expression,
        result: errorMsg,
        timestamp: new Date().toISOString(),
        fullData: { error }
      });
      
      setLoading(false);
    }
  };

  return (
    <div className="w-full max-w-md space-y-4 animate-fade-in relative">
      <Link href="/history" className="absolute -top-12 right-0 text-white/70 hover:text-white transition-colors">
        История
      </Link>
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
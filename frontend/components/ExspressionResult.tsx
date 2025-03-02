"use client";

import { useState, useEffect } from "react";

interface ExpressionResultProps {
  expression: string;
}

export default function ExpressionResult({ expression }: ExpressionResultProps) {
  const [result, setResult] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchResult() {
      try {
        const res = await fetch("http://localhost:8080/api/v1/calculate", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ expression }),
        });
        const data = await res.json();
        if (data.error) {
          setResult("Ошибка: " + data.error);
        } else {
          setResult(data.result);
        }
      } catch (error) {
        console.error("Ошибка запроса:", error);
        setResult("Ошибка запроса");
      } finally {
        setLoading(false);
      }
    }

    fetchResult();
  }, [expression]);

  if (loading) return <p>Загрузка...</p>;
  return <p>Результат: {result}</p>;
}
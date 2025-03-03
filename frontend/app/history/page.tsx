"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { ArrowLeft, Trash2 } from "lucide-react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { tomorrow } from "react-syntax-highlighter/dist/esm/styles/prism";

interface HistoryItem {
  expression: string;
  result: string;
  timestamp: string;
  fullData: any;
}

export default function HistoryPage() {
  const [history, setHistory] = useState<HistoryItem[]>([]);
  const [expandedItem, setExpandedItem] = useState<number | null>(null);

  useEffect(() => {
    try {
      const savedHistory = localStorage.getItem("calculationsHistory");
      if (savedHistory) {
        const parsedHistory = JSON.parse(savedHistory);
        
        const cleanHistory = parsedHistory.map((item: any) => ({
          expression: typeof item.expression === 'string' ? item.expression : 
                     JSON.stringify(item.expression),
          result: typeof item.result === 'string' ? item.result : 
                 String(item.result),
          timestamp: item.timestamp || new Date().toISOString(),
          fullData: item.fullData || {}
        }));
        
        setHistory(cleanHistory);
      }
    } catch (error) {
      console.error("Error loading history:", error);
      localStorage.removeItem("calculationsHistory");
      setHistory([]);
    }
  }, []);

  const toggleExpanded = (index: number) => {
    setExpandedItem(expandedItem === index ? null : index);
  };

  const clearHistory = () => {
    localStorage.removeItem("calculationsHistory");
    setHistory([]);
  };

  return (
    <main className="min-h-screen bg-black flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-md space-y-4 animate-fade-in relative text-white">
        <div className="flex justify-between items-center">
          <Link
            href="/"
            className="text-white/70 hover:text-white transition-colors flex items-center"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Назад
          </Link>
          {history.length > 0 && (
            <button 
              onClick={clearHistory}
              className="text-red-400 hover:text-red-300 transition-colors flex items-center text-sm"
            >
              <Trash2 className="w-4 h-4 mr-1" />
              Очистить историю
            </button>
          )}
        </div>
        
        <h1 className="text-2xl font-bold text-white mb-4">История запросов</h1>
        {history.length === 0 && <p className="text-white/50">История вычислений пуста</p>}
        
        {history.map((item, i) => (
          <div key={i} className="bg-white/10 p-3 rounded-md mb-2">
            <div className="flex justify-between items-center">
              <div className="text-sm">
                {/* Safely render expression with proper type checking */}
                <span className="text-white">
                  {item.expression}
                </span>
                {typeof item.result === 'string' && item.result.startsWith("Ошибка") ? (
                  <span className="ml-2 text-red-500">
                    {item.result}
                  </span>
                ) : (
                  <span className="ml-2">= {item.result}</span>
                )}
              </div>
              <button
                className="text-xs text-white/70 hover:text-white transition-colors"
                onClick={() => toggleExpanded(i)}
              >
                {expandedItem === i ? "Скрыть JSON" : "Показать JSON"}
              </button>
            </div>
            {expandedItem === i && (
              <SyntaxHighlighter
                language="json"
                style={tomorrow}
                customStyle={{
                  marginTop: "0.5rem",
                  padding: "0.5rem",
                  borderRadius: "0.375rem",
                  overflow: "auto",
                  fontSize: "0.75rem"
                }}
              >
                {JSON.stringify(item.fullData, null, 2)}
              </SyntaxHighlighter>
            )}
          </div>
        ))}
      </div>
    </main>
  );
}
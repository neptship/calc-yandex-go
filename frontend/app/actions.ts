"use server"

export async function calculateExpression(expression: string): Promise<string> {
  try {
    if (!/^[0-9+\-*/().%\s]+$/.test(expression)) {
      throw new Error("Invalid expression")
    }

    const result = eval(expression)

    if (typeof result === "number") {
      return Number.isInteger(result) ? result.toString() : result.toFixed(4).replace(/\.?0+$/, "")
    }

    return String(result)
  } catch (error) {
    console.error("Calculation error:", error)
    return "Error in calculation"
  }
}


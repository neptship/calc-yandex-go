"use server"

export async function calculateExpression(expression: string): Promise<string> {
  try {
    // In a real application, you might want to use a safer evaluation method
    // or a dedicated math expression parser library

    // For demonstration purposes, we're using a simple approach
    // This is NOT safe for production as it could allow code injection
    // Consider using a library like math.js in a real application

    // Sanitize input (very basic example)
    if (!/^[0-9+\-*/().%\s]+$/.test(expression)) {
      throw new Error("Invalid expression")
    }

    // Calculate the result
    // Note: In a real app, use a proper math expression parser instead of eval
    const result = eval(expression)

    // Format the result
    if (typeof result === "number") {
      // Handle floating point precision
      return Number.isInteger(result) ? result.toString() : result.toFixed(4).replace(/\.?0+$/, "")
    }

    return String(result)
  } catch (error) {
    console.error("Calculation error:", error)
    return "Error in calculation"
  }
}


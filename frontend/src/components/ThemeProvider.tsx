"use client"

import { createContext, useContext, useEffect, useState, useCallback } from "react"

type ThemeContextType = {
  dark: boolean
  toggle: () => void
}

const ThemeContext = createContext<ThemeContextType>({ dark: false, toggle: () => {} })
export const useTheme = () => useContext(ThemeContext)

export default function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [dark, setDark] = useState(false)

  useEffect(() => {
    const stored = localStorage.getItem("theme")
    if (stored === "dark") {
      setDark(true)
      document.documentElement.classList.add("dark")
    }
  }, [])

  const toggle = useCallback(() => {
    setDark((prev) => {
      const next = !prev
      localStorage.setItem("theme", next ? "dark" : "light")
      if (next) document.documentElement.classList.add("dark")
      else document.documentElement.classList.remove("dark")
      return next
    })
  }, [])

  return (
    <ThemeContext.Provider value={{ dark, toggle }}>
      {children}
    </ThemeContext.Provider>
  )
}

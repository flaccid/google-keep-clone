"use client"

import { useState } from "react"
import Link from "next/link"
import { Lightbulb, Menu, RefreshCw, Settings, Moon, Sun } from "lucide-react"
import { useTheme } from "@/components/ThemeProvider"

export default function Header({
  sidebarExpanded,
  onToggleSidebar,
  search,
  onSearchChange,
}: {
  sidebarExpanded: boolean
  onToggleSidebar: () => void
  search: string
  onSearchChange: (v: string) => void
}) {
  const [showSettings, setShowSettings] = useState(false)
  const { dark, toggle } = useTheme()

  return (
    <header className="fixed top-0 left-0 right-0 h-16 bg-white dark:bg-[#202124] border-b border-gray-100 dark:border-[#3c4043] flex items-center z-50">
      {/* Left section: hamburger + logo — matches sidebar width */}
      <div className={`flex items-center gap-2 px-2 sm:px-4 flex-shrink-0 transition-all duration-200 ${
        sidebarExpanded ? "sm:w-72" : "sm:w-[68px]"
      }`}>
        <button
          onClick={onToggleSidebar}
          className="p-2 hover:bg-gray-100 dark:hover:bg-white/10 rounded-full transition-colors"
          aria-label="Toggle sidebar"
        >
          <Menu size={20} className="text-gray-600 dark:text-[#e8eaed]" />
        </button>

        {sidebarExpanded && (
          <Link href="/" className="hidden sm:flex items-center gap-2 hover:opacity-80">
            <Lightbulb size={24} className="text-yellow-500" fill="currentColor" />
            <span className="text-2xl font-medium text-gray-700 dark:text-[#e8eaed]">Keep</span>
          </Link>
        )}
      </div>

      {/* Search bar — aligns with main content area */}
      <div className="flex-1 flex items-center gap-2 sm:gap-4 pr-2 sm:pr-4">
        <div className="relative w-full max-w-2xl">
          <input
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="Search notes..."
            className="w-full px-4 py-2 pl-10 rounded-lg bg-white dark:bg-[#202124] border border-gray-200 dark:border-[#5f6368] shadow-sm text-sm outline-none focus:shadow-md focus:ring-1 focus:ring-gray-300 dark:focus:ring-[#5f6368] transition-all placeholder-gray-400 dark:placeholder-[#9aa0a6] text-gray-900 dark:text-[#e8eaed]"
          />
          <svg
            className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-[#9aa0a6]"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.35-4.35" />
          </svg>
        </div>

        <button
          className="p-2 hover:bg-gray-100 dark:hover:bg-white/10 rounded-full transition-colors hidden sm:block flex-shrink-0"
          title="Refresh"
        >
          <RefreshCw size={20} className="text-gray-600 dark:text-[#e8eaed]" />
        </button>
        <div className="relative flex-shrink-0">
          <button
            onClick={() => setShowSettings(!showSettings)}
            className="p-2 hover:bg-gray-100 dark:hover:bg-white/10 rounded-full transition-colors"
            title="Settings"
          >
            <Settings size={20} className="text-gray-600 dark:text-[#e8eaed]" />
          </button>
          {showSettings && (
            <>
              <div className="fixed inset-0 z-40" onClick={() => setShowSettings(false)} />
              <div className="absolute right-0 top-full mt-1 w-56 bg-white dark:bg-[#2d2e30] rounded-lg shadow-lg border border-gray-200 dark:border-[#5f6368] z-50 py-1">
                <div className="flex items-center justify-between px-4 py-2.5">
                  <span className="text-sm text-gray-700 dark:text-[#e8eaed] flex items-center gap-2">
                    {dark ? <Moon size={16} /> : <Sun size={16} />}
                    Dark theme
                  </span>
                  <button
                    onClick={(e) => { e.stopPropagation(); toggle() }}
                    className={`relative w-10 h-5 rounded-full transition-colors ${
                      dark ? "bg-blue-600" : "bg-gray-300"
                    }`}
                  >
                    <span
                      className={`absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform ${
                        dark ? "translate-x-5" : "translate-x-0"
                      }`}
                    />
                  </button>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </header>
  )
}

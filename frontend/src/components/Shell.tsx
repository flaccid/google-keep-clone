"use client"

import { createContext, useContext, useState, useEffect } from "react"
import Header from "@/components/Header"
import Sidebar from "@/components/Sidebar"

const SearchContext = createContext<{ search: string }>({ search: "" })
export const useSearch = () => useContext(SearchContext)

function useIsMobile() {
  const [mobile, setMobile] = useState(false)
  useEffect(() => {
    const mq = window.matchMedia("(max-width: 639px)")
    setMobile(mq.matches)
    const handler = (e: MediaQueryListEvent) => setMobile(e.matches)
    mq.addEventListener("change", handler)
    return () => mq.removeEventListener("change", handler)
  }, [])
  return mobile
}

export default function Shell({ children }: { children: React.ReactNode }) {
  const isMobile = useIsMobile()
  const [sidebarExpanded, setSidebarExpanded] = useState(false)
  const [sidebarHover, setSidebarHover] = useState(false)
  const [search, setSearch] = useState("")

  // Expand sidebar by default on desktop after hydration
  useEffect(() => {
    if (!isMobile) setSidebarExpanded(true)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  function toggleSidebar() {
    setSidebarExpanded((v) => !v)
  }

  function closeSidebarOnMobile() {
    if (isMobile) setSidebarExpanded(false)
  }

  return (
    <SearchContext.Provider value={{ search }}>
      <Header
        sidebarExpanded={sidebarExpanded}
        onToggleSidebar={toggleSidebar}
        search={search}
        onSearchChange={setSearch}
      />
      <Sidebar
        expanded={sidebarExpanded}
        hover={sidebarHover}
        onHoverChange={isMobile ? () => {} : setSidebarHover}
        isMobile={isMobile}
        onNavigate={closeSidebarOnMobile}
      />
      {/* Backdrop for mobile sidebar */}
      {isMobile && sidebarExpanded && (
        <div
          className="fixed inset-0 z-30 bg-black/50"
          onClick={() => setSidebarExpanded(false)}
        />
      )}
      <main
        className={`pt-16 min-h-screen transition-all duration-200 ${
          isMobile ? "ml-0" : sidebarExpanded ? "ml-72" : "ml-[68px]"
        }`}
      >
        {children}
      </main>
    </SearchContext.Provider>
  )
}

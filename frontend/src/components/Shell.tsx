"use client"

import { createContext, useContext, useState } from "react"
import Header from "@/components/Header"
import Sidebar from "@/components/Sidebar"

const SearchContext = createContext<{ search: string }>({ search: "" })
export const useSearch = () => useContext(SearchContext)

export default function Shell({ children }: { children: React.ReactNode }) {
  const [sidebarExpanded, setSidebarExpanded] = useState(true)
  const [sidebarHover, setSidebarHover] = useState(false)
  const [search, setSearch] = useState("")

  return (
    <SearchContext.Provider value={{ search }}>
      <Header
        sidebarExpanded={sidebarExpanded}
        onToggleSidebar={() => setSidebarExpanded((v) => !v)}
        search={search}
        onSearchChange={setSearch}
      />
      <Sidebar
        expanded={sidebarExpanded}
        hover={sidebarHover}
        onHoverChange={setSidebarHover}
      />
      <main
        className={`pt-16 min-h-screen transition-all duration-200 ${
          sidebarExpanded ? "ml-72" : "ml-[68px]"
        }`}
      >
        {children}
      </main>
    </SearchContext.Provider>
  )
}

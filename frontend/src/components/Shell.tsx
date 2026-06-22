"use client"

import { createContext, useContext, useState } from "react"
import Header from "@/components/Header"
import Sidebar from "@/components/Sidebar"

const SearchContext = createContext<{ search: string }>({ search: "" })
export const useSearch = () => useContext(SearchContext)

export default function Shell({ children }: { children: React.ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [search, setSearch] = useState("")

  return (
    <SearchContext.Provider value={{ search }}>
      <Header
        sidebarOpen={sidebarOpen}
        onToggleSidebar={() => setSidebarOpen((v) => !v)}
        search={search}
        onSearchChange={setSearch}
      />
      <Sidebar open={sidebarOpen} />
      <main
        className={`pt-16 min-h-screen transition-all duration-200 ${
          sidebarOpen ? "ml-72" : "ml-0"
        }`}
      >
        {children}
      </main>
    </SearchContext.Provider>
  )
}

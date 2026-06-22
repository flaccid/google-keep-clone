"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { Lightbulb, Bell, Archive, Trash2, Plus } from "lucide-react"
import { api } from "@/lib/api"
import type { Label } from "@/lib/types"

const navItems = [
  { href: "/", icon: Lightbulb, label: "Notes" },
  { href: "/reminders", icon: Bell, label: "Reminders" },
  { href: "/archive", icon: Archive, label: "Archive" },
  { href: "/trash", icon: Trash2, label: "Trash" },
]

export default function Sidebar({ open }: { open: boolean }) {
  const pathname = usePathname()
  const [labels, setLabels] = useState<Label[]>([])
  const [newLabel, setNewLabel] = useState("")

  useEffect(() => {
    api.labels.list().then(setLabels).catch(() => {})
  }, [])

  async function createLabel() {
    if (!newLabel.trim()) return
    await api.labels.create(newLabel.trim())
    setNewLabel("")
    setLabels(await api.labels.list())
  }

  return (
    <aside
      className={`fixed left-0 top-16 h-[calc(100vh-4rem)] w-72 bg-white dark:bg-[#2d2e30] border-r border-gray-100 dark:border-[#5f6368] z-40 transition-transform duration-200 overflow-y-auto scrollbar-thin ${
        open ? "translate-x-0" : "-translate-x-72"
      }`}
    >
      <nav className="py-2">
        {navItems.map(({ href, icon: Icon, label }) => {
          const active = href === "/" ? pathname === "/" : pathname.startsWith(href)
          return (
            <Link
              key={href}
              href={href}
              className={`flex items-center gap-4 mx-2 px-4 py-2.5 rounded-r-full text-sm transition-colors ${
                active
                  ? "bg-yellow-50 dark:bg-yellow-900/30 text-gray-800 dark:text-[#e8eaed] font-medium"
                  : "text-gray-600 dark:text-[#bdc1c6] hover:bg-gray-100 dark:hover:bg-white/10"
              }`}
            >
              <Icon size={20} strokeWidth={active ? 2 : 1.5} />
              {label}
            </Link>
          )
        })}
      </nav>

      <div className="border-t border-gray-100 dark:border-[#5f6368] mt-2 pt-2">
        <p className="px-6 text-xs font-medium text-gray-400 dark:text-[#9aa0a6] uppercase tracking-wide mb-1">
          Labels
        </p>
        {labels.map((l) => {
          const id = l.name?.replace("labels/", "") || ""
          return (
            <Link
              key={l.name}
              href={`/labels/${id}`}
              className="flex items-center gap-4 mx-2 px-4 py-2 rounded-r-full text-sm text-gray-600 dark:text-[#bdc1c6] hover:bg-gray-100 dark:hover:bg-white/10 transition-colors"
            >
              <span className="w-5 h-5 rounded-full border border-gray-300 dark:border-[#5f6368] flex items-center justify-center text-[10px] text-gray-400 dark:text-[#9aa0a6]">
                {l.displayName?.charAt(0).toUpperCase()}
              </span>
              {l.displayName}
            </Link>
          )
        })}
        <div className="flex items-center gap-2 mx-2 px-4 py-2">
          <input
            value={newLabel}
            onChange={(e) => setNewLabel(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && createLabel()}
            placeholder="Create new label..."
            className="flex-1 text-sm bg-transparent outline-none placeholder-gray-400 dark:placeholder-[#9aa0a6] text-gray-900 dark:text-[#e8eaed]"
          />
          <button onClick={createLabel} className="text-gray-400 dark:text-[#9aa0a6] hover:text-gray-600 dark:hover:text-[#e8eaed]">
            <Plus size={16} />
          </button>
        </div>
      </div>
    </aside>
  )
}

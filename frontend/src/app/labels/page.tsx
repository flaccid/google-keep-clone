"use client"

import { useState, useEffect, useCallback } from "react"
import { api } from "@/lib/api"
import type { Label } from "@/lib/types"
import Link from "next/link"

export default function LabelsPage() {
  const [labels, setLabels] = useState<Label[]>([])
  const [newName, setNewName] = useState("")

  const load = useCallback(async () => {
    try {
      setLabels(await api.labels.list())
    } catch {}
  }, [])

  useEffect(() => { load() }, [load])

  async function create() {
    if (!newName.trim()) return
    await api.labels.create(newName.trim())
    setNewName("")
    load()
  }

  async function remove(id: string) {
    await api.labels.delete(id)
    load()
  }

  return (
    <div className="max-w-3xl mx-auto px-6 py-6">
      <h2 className="text-xs font-medium text-gray-400 dark:text-[#9aa0a6] uppercase tracking-wide mb-3">Labels</h2>

      <div className="flex gap-2 mb-6">
        <input
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && create()}
          placeholder="New label name..."
          className="flex-1 max-w-xs px-3 py-1.5 rounded-lg border border-gray-200 dark:border-[#5f6368] text-sm outline-none focus:border-blue-400 bg-white dark:bg-[#2d2e30] text-gray-900 dark:text-[#e8eaed] placeholder-gray-400 dark:placeholder-[#9aa0a6]"
        />
        <button onClick={create} className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700">
          Add
        </button>
      </div>

      <div className="space-y-1">
        {labels.map((l) => {
          const id = l.name?.replace("labels/", "") || ""
          return (
            <div key={l.name} className="flex items-center justify-between px-3 py-2 rounded-lg hover:bg-gray-50 dark:hover:bg-white/5">
              <Link href={`/labels/${id}`} className="text-sm text-gray-900 dark:text-[#e8eaed] hover:text-blue-600 dark:hover:text-blue-400">
                {l.displayName}
              </Link>
              <button onClick={() => remove(id)} className="text-xs text-red-500 dark:text-red-400 hover:underline">
                Delete
              </button>
            </div>
          )
        })}
        {labels.length === 0 && (
          <p className="text-gray-400 dark:text-[#9aa0a6] text-sm text-center py-8">No labels yet.</p>
        )}
      </div>
    </div>
  )
}

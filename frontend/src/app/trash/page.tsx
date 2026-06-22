"use client"

import { useState, useEffect, useCallback } from "react"
import { api } from "@/lib/api"
import type { Note } from "@/lib/types"

export default function TrashPage() {
  const [notes, setNotes] = useState<Note[]>([])

  const load = useCallback(async () => {
    try {
      const res = await api.notes.list(undefined, undefined, undefined, "trashed=true")
      setNotes(res.notes || [])
    } catch {}
  }, [])

  useEffect(() => { load() }, [load])

  async function restore(n: Note) {
    const id = n.name?.replace("notes/", "") || ""
    await api.notes.restore(id)
    load()
  }

  async function removePermanently(n: Note) {
    const id = n.name?.replace("notes/", "") || ""
    if (confirm("Delete this note permanently?")) {
      await api.notes.delete(id)
      load()
    }
  }

  return (
    <div className="max-w-3xl mx-auto px-6 py-6">
      <h2 className="text-xs font-medium text-gray-400 dark:text-[#9aa0a6] uppercase tracking-wide mb-3">Trash</h2>
      <div className="space-y-2">
        {notes.map((n) => {
          const id = n.name?.replace("notes/", "") || ""
          return (
            <div key={n.name} className="flex items-center justify-between bg-white dark:bg-[#2d2e30] rounded-lg border border-gray-200 dark:border-[#5f6368] p-3">
              <div className="flex items-center gap-3">
                <span className="text-sm font-medium text-gray-900 dark:text-[#e8eaed]">{n.title || "Untitled"}</span>
                <span className="text-xs text-gray-400 dark:text-[#9aa0a6]">{n.trashTime}</span>
              </div>
              <div className="flex gap-2">
                <button onClick={() => restore(n)} className="text-xs text-blue-600 dark:text-blue-400 hover:underline">
                  Restore
                </button>
                <button onClick={() => removePermanently(n)} className="text-xs text-red-500 dark:text-red-400 hover:underline">
                  Delete forever
                </button>
              </div>
            </div>
          )
        })}
        {notes.length === 0 && (
          <p className="text-gray-400 dark:text-[#9aa0a6] text-sm text-center py-12">Trash is empty.</p>
        )}
      </div>
    </div>
  )
}

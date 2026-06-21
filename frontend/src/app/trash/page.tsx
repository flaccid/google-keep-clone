"use client"

import { useState, useEffect, useCallback } from "react"
import { api } from "@/lib/api"
import type { Note } from "@/lib/types"
import Link from "next/link"

export default function TrashPage() {
  const [notes, setNotes] = useState<Note[]>([])

  const load = useCallback(async () => {
    try {
      const res = await api.notes.list()
      setNotes((res.notes || []).filter((n) => n.trashed))
    } catch {}
  }, [])

  useEffect(() => { load() }, [load])

  async function restore(n: Note) {
    const id = n.name?.replace("notes/", "") || ""
    await api.notes.restore(id)
    load()
  }

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <Link href="/" className="text-gray-500 hover:text-gray-700 text-sm">Notes</Link>
        <Link href="/archive" className="text-gray-500 hover:text-gray-700 text-sm">Archive</Link>
        <Link href="/trash" className="text-blue-600 font-medium text-sm">Trash</Link>
        <Link href="/labels" className="text-gray-500 hover:text-gray-700 text-sm">Labels</Link>
      </div>

      <h2 className="text-xs font-medium text-gray-400 uppercase tracking-wide mb-3">Trash</h2>
      <div className="space-y-2">
        {notes.map((n) => (
          <div key={n.name} className="flex items-center justify-between bg-white rounded-lg border border-gray-200 p-3">
            <div>
              <p className="text-sm font-medium">{n.title || "Untitled"}</p>
              <p className="text-xs text-gray-400">{n.trashTime}</p>
            </div>
            <button onClick={() => restore(n)} className="text-xs text-blue-600 hover:underline">
              Restore
            </button>
          </div>
        ))}
        {notes.length === 0 && (
          <p className="text-gray-400 text-sm text-center py-12">Trash is empty.</p>
        )}
      </div>
    </div>
  )
}

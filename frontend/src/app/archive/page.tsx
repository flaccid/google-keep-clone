"use client"

import { useState, useEffect, useCallback } from "react"
import { api } from "@/lib/api"
import type { Note } from "@/lib/types"
import NoteCard from "@/components/NoteCard"
import Link from "next/link"

export default function ArchivePage() {
  const [notes, setNotes] = useState<Note[]>([])

  const load = useCallback(async () => {
    try {
      const res = await api.notes.list()
      setNotes((res.notes || []).filter((n) => n.archived && !n.trashed))
    } catch {}
  }, [])

  useEffect(() => { load() }, [load])

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <Link href="/" className="text-gray-500 hover:text-gray-700 text-sm">Notes</Link>
        <Link href="/archive" className="text-blue-600 font-medium text-sm">Archive</Link>
        <Link href="/trash" className="text-gray-500 hover:text-gray-700 text-sm">Trash</Link>
        <Link href="/labels" className="text-gray-500 hover:text-gray-700 text-sm">Labels</Link>
      </div>

      <h2 className="text-xs font-medium text-gray-400 uppercase tracking-wide mb-3">Archived</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {notes.map((n) => (
          <NoteCard key={n.name} note={n} onUpdate={load} />
        ))}
        {notes.length === 0 && (
          <p className="text-gray-400 text-sm col-span-full text-center py-12">No archived notes.</p>
        )}
      </div>
    </div>
  )
}

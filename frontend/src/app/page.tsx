"use client"

import { useState, useEffect, useCallback, useRef } from "react"
import { api } from "@/lib/api"
import type { Note } from "@/lib/types"
import NoteCard from "@/components/NoteCard"
import NoteEditor from "@/components/NoteEditor"
import Link from "next/link"

export default function HomePage() {
  const [notes, setNotes] = useState<Note[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [search, setSearch] = useState("")
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined)

  const load = useCallback(async (searchTerm?: string) => {
    try {
      const res = await api.notes.list(undefined, undefined, searchTerm)
      setNotes(res.notes || [])
    } catch {}
  }, [])

  useEffect(() => {
    load(search || undefined)
  }, [load])

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      load(search || undefined)
    }, 300)
    return () => { if (debounceRef.current) clearTimeout(debounceRef.current) }
  }, [search, load])

  const active = notes.filter((n) => !n.archived && !n.trashed && !n.pinned)
  const pinned = notes.filter((n) => n.pinned && !n.archived && !n.trashed)

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <input
          placeholder="Search notes..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="flex-1 max-w-md px-4 py-2 rounded-lg border border-gray-200 text-sm outline-none focus:border-blue-400"
        />
        <div className="flex gap-3 text-sm">
          <Link href="/" className="text-blue-600 font-medium">Notes</Link>
          <Link href="/archive" className="text-gray-500 hover:text-gray-700">Archive</Link>
          <Link href="/trash" className="text-gray-500 hover:text-gray-700">Trash</Link>
          <Link href="/labels" className="text-gray-500 hover:text-gray-700">Labels</Link>
        </div>
      </div>

      <button
        onClick={() => setShowCreate(!showCreate)}
        className="mb-6 px-4 py-2 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700"
      >
        {showCreate ? "Cancel" : "New Note"}
      </button>

      {showCreate && (
        <div className="mb-8">
          <NoteEditor onSave={() => { setShowCreate(false); load() }} />
        </div>
      )}

      {pinned.length > 0 && (
        <div className="mb-8">
          <h2 className="text-xs font-medium text-gray-400 uppercase tracking-wide mb-3">Pinned</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {pinned.map((n) => (
              <NoteCard key={n.name} note={n} onUpdate={load} />
            ))}
          </div>
        </div>
      )}

      <h2 className="text-xs font-medium text-gray-400 uppercase tracking-wide mb-3">Notes</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {active.map((n) => (
          <NoteCard key={n.name} note={n} onUpdate={load} />
        ))}
        {active.length === 0 && pinned.length === 0 && (
          <p className="text-gray-400 text-sm col-span-full text-center py-12">
            No notes yet. Click &quot;New Note&quot; to create one.
          </p>
        )}
      </div>
    </div>
  )
}

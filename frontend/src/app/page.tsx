"use client"

import { useState, useEffect, useCallback, useRef } from "react"
import { api } from "@/lib/api"
import type { Note } from "@/lib/types"
import NoteCard from "@/components/NoteCard"
import NoteEditor from "@/components/NoteEditor"
import NoteModal from "@/components/NoteModal"
import { useSearch } from "@/components/Shell"

export default function HomePage() {
  const [notes, setNotes] = useState<Note[]>([])
  const [editorOpen, setEditorOpen] = useState(false)
  const [openNoteId, setOpenNoteId] = useState<string | null>(null)
  const { search } = useSearch()
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
    <div className="max-w-5xl mx-auto px-3 sm:px-6 py-4 sm:py-6">
      <div className="mb-6">
        <NoteEditor
          onSave={() => { setEditorOpen(false); load() }}
          onClose={() => setEditorOpen(false)}
        />
      </div>

      {pinned.length > 0 && (
        <div className="mb-8">
          <h2 className="text-xs font-medium text-gray-400 dark:text-[#9aa0a6] uppercase tracking-wide mb-3">Pinned</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 sm:gap-4">
            {pinned.map((n) => (
              <NoteCard key={n.name} note={n} onUpdate={load} onOpen={setOpenNoteId} />
            ))}
          </div>
        </div>
      )}

      <h2 className="text-xs font-medium text-gray-400 dark:text-[#9aa0a6] uppercase tracking-wide mb-3">
        {pinned.length > 0 ? "Others" : "Notes"}
      </h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 sm:gap-4">
        {active.map((n) => (
          <NoteCard key={n.name} note={n} onUpdate={load} onOpen={setOpenNoteId} />
        ))}
        {active.length === 0 && pinned.length === 0 && (
          <p className="text-gray-400 dark:text-[#9aa0a6] text-sm col-span-full text-center py-12">
            {search
              ? "No matching notes found."
              : "No notes yet. Click \"Take a note...\" to create one."}
          </p>
        )}
      </div>

      {openNoteId && (
        <NoteModal noteId={openNoteId} onClose={() => { setOpenNoteId(null); load() }} />
      )}
    </div>
  )
}

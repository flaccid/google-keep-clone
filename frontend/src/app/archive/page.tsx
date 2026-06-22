"use client"

import { useState, useEffect, useCallback, useRef } from "react"
import { api } from "@/lib/api"
import type { Note } from "@/lib/types"
import NoteCard from "@/components/NoteCard"
import NoteModal from "@/components/NoteModal"
import { useSearch } from "@/components/Shell"

export default function ArchivePage() {
  const [notes, setNotes] = useState<Note[]>([])
  const [openNoteId, setOpenNoteId] = useState<string | null>(null)
  const { search } = useSearch()
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined)

  const load = useCallback(async (searchTerm?: string) => {
    try {
      const res = await api.notes.list(undefined, undefined, searchTerm, "archived=true AND trashed=false")
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

  return (
    <div className="max-w-5xl mx-auto px-3 sm:px-6 py-4 sm:py-6">
      <h2 className="text-xs font-medium text-gray-400 dark:text-[#9aa0a6] uppercase tracking-wide mb-3">Archived</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 sm:gap-4">
        {notes.map((n) => (
          <NoteCard key={n.name} note={n} onUpdate={load} onOpen={setOpenNoteId} />
        ))}
        {notes.length === 0 && (
          <p className="text-gray-400 dark:text-[#9aa0a6] text-sm col-span-full text-center py-12">No archived notes.</p>
        )}
      </div>

      {openNoteId && (
        <NoteModal noteId={openNoteId} onClose={() => { setOpenNoteId(null); load() }} />
      )}
    </div>
  )
}

"use client"

import { useEffect, useState, useCallback, useRef } from "react"
import { useParams } from "next/navigation"
import { api } from "@/lib/api"
import type { Note, Label } from "@/lib/types"
import NoteCard from "@/components/NoteCard"
import NoteModal from "@/components/NoteModal"
import { useSearch } from "@/components/Shell"

export default function LabelNotesPage() {
  const { id } = useParams<{ id: string }>()
  const [notes, setNotes] = useState<Note[]>([])
  const [openNoteId, setOpenNoteId] = useState<string | null>(null)
  const [label, setLabel] = useState<Label | null>(null)
  const { search } = useSearch()
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined)

  const load = useCallback(async (searchTerm?: string) => {
    try {
      const [labels, res] = await Promise.all([api.labels.list(), api.notes.list(undefined, undefined, searchTerm)])
      const found = labels.find((l) => l.name === `labels/${id}`) || null
      setLabel(found)
      setNotes((res.notes || []).filter((n) => found && n.labels?.includes(found.displayName || "")))
    } catch {}
  }, [id])

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
    <div className="max-w-5xl mx-auto px-6 py-6">
      <h2 className="text-xs font-medium text-gray-400 dark:text-[#9aa0a6] uppercase tracking-wide mb-3">Label: {label?.displayName || id}</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {notes.map((n) => (
          <NoteCard key={n.name} note={n} onUpdate={load} onOpen={setOpenNoteId} />
        ))}
        {notes.length === 0 && (
          <p className="text-gray-400 dark:text-[#9aa0a6] text-sm col-span-full text-center py-12">No notes with this label.</p>
        )}
      </div>

      {openNoteId && (
        <NoteModal noteId={openNoteId} onClose={() => { setOpenNoteId(null); load() }} />
      )}
    </div>
  )
}

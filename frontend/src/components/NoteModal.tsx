"use client"

import { useEffect, useState, useCallback } from "react"
import { api } from "@/lib/api"
import type { Note } from "@/lib/types"
import NoteEditor from "./NoteEditor"

export default function NoteModal({
  noteId,
  onClose,
}: {
  noteId: string
  onClose: () => void
}) {
  const [note, setNote] = useState<Note | null>(null)

  const load = useCallback(async () => {
    try {
      setNote(await api.notes.get(noteId))
    } catch {
      onClose()
    }
  }, [noteId, onClose])

  useEffect(() => { load() }, [load])

  if (!note) return null

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-16 bg-black/50">
      <div className="w-full max-w-2xl px-4 py-6">
        <NoteEditor
          note={note}
          onSave={onClose}
          onDelete={onClose}
          onClose={onClose}
        />
      </div>
    </div>
  )
}

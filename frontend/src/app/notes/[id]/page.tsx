"use client"

import { useEffect, useState, useCallback } from "react"
import { useParams, useRouter } from "next/navigation"
import { api } from "@/lib/api"
import type { Note } from "@/lib/types"
import NoteEditor from "@/components/NoteEditor"

export default function NotePage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const [note, setNote] = useState<Note | null>(null)

  const load = useCallback(async () => {
    try {
      setNote(await api.notes.get(id))
    } catch {
      router.push("/")
    }
  }, [id, router])

  useEffect(() => { load() }, [load])

  if (!note) return <p className="text-gray-400 text-sm text-center py-12">Loading...</p>

  return (
    <div className="max-w-2xl mx-auto">
      <button onClick={() => router.push("/")} className="text-sm text-blue-600 mb-4 hover:underline">
        &larr; Back
      </button>
      <NoteEditor
        note={note}
        onSave={() => router.push("/")}
        onDelete={() => router.push("/")}
      />
    </div>
  )
}

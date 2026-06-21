"use client"

import type { Note } from "@/lib/types"
import { api } from "@/lib/api"

const COLORS: Record<string, string> = {
  DEFAULT: "bg-white",
  RED: "bg-red-100",
  ORANGE: "bg-orange-100",
  YELLOW: "bg-yellow-100",
  GREEN: "bg-green-100",
  TEAL: "bg-teal-100",
  BLUE: "bg-blue-100",
  DARK_BLUE: "bg-blue-200",
  PURPLE: "bg-purple-100",
  PINK: "bg-pink-100",
  BROWN: "bg-amber-100",
  GRAY: "bg-gray-100",
}

function noteTitle(note: Note): string {
  return note.title || ""
}

function notePreview(note: Note): string {
  if (note.body?.text?.text) return note.body.text.text
  if (note.body?.list?.listItems?.length) {
    const done = note.body.list.listItems.filter((i) => i.checked).length
    const total = note.body.list.listItems.length
    return `${done}/${total} items`
  }
  return ""
}

function noteId(name?: string): string {
  if (!name) return ""
  return name.replace("notes/", "")
}

export default function NoteCard({
  note,
  onUpdate,
}: {
  note: Note
  onUpdate: () => void
}) {
  const colorClass = COLORS[note.color || "DEFAULT"] || COLORS.DEFAULT
  const id = noteId(note.name)

  async function togglePin(e: React.MouseEvent) {
    e.stopPropagation()
    e.preventDefault()
    if (note.pinned) await api.notes.unpin(id)
    else await api.notes.pin(id)
    onUpdate()
  }

  async function toggleArchive(e: React.MouseEvent) {
    e.stopPropagation()
    e.preventDefault()
    if (note.archived) await api.notes.unarchive(id)
    else await api.notes.archive(id)
    onUpdate()
  }

  async function toggleTrash(e: React.MouseEvent) {
    e.stopPropagation()
    e.preventDefault()
    await api.notes.trash(id)
    onUpdate()
  }

  return (
    <a
      href={`/notes/${id}`}
      className={`${colorClass} rounded-xl border border-gray-200 p-4 shadow-sm hover:shadow-md transition-shadow cursor-pointer block relative group`}
    >
      <div className="flex justify-between items-start">
        <h3 className="font-medium text-sm mb-1">{noteTitle(note) || "Untitled"}</h3>
        {note.pinned && <span className="text-yellow-500 text-xs">pinned</span>}
      </div>
      <p className="text-gray-600 text-xs whitespace-pre-wrap line-clamp-4">
        {notePreview(note)}
      </p>
      {note.labels && note.labels.length > 0 && (
        <div className="flex flex-wrap gap-1 mt-2">
          {note.labels.map((l) => (
            <span key={l} className="text-[10px] bg-gray-200/60 rounded px-1.5 py-0.5">
              {l}
            </span>
          ))}
        </div>
      )}
      <div className="absolute top-1 right-1 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <button onClick={togglePin} className="text-xs p-1 hover:bg-black/10 rounded" title={note.pinned ? "Unpin" : "Pin"}>
          {note.pinned ? "unpin" : "pin"}
        </button>
        <button onClick={toggleArchive} className="text-xs p-1 hover:bg-black/10 rounded" title={note.archived ? "Unarchive" : "Archive"}>
          {note.archived ? "unarchive" : "archive"}
        </button>
        <button onClick={toggleTrash} className="text-xs p-1 hover:bg-black/10 rounded" title="Delete">
          delete
        </button>
      </div>
    </a>
  )
}

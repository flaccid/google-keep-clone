"use client"

import { useState } from "react"
import type { Note } from "@/lib/types"
import { api } from "@/lib/api"

const COLOR_OPTIONS = [
  "DEFAULT", "RED", "ORANGE", "YELLOW", "GREEN", "TEAL",
  "BLUE", "DARK_BLUE", "PURPLE", "PINK", "BROWN", "GRAY",
]

export default function NoteEditor({
  note,
  onSave,
  onDelete,
}: {
  note?: Note
  onSave: () => void
  onDelete?: () => void
}) {
  const [title, setTitle] = useState(note?.title || "")
  const [text, setText] = useState(note?.body?.text?.text || "")
  const [color, setColor] = useState(note?.color || "DEFAULT")
  const [saving, setSaving] = useState(false)

  const id = note?.name?.replace("notes/", "") || ""

  async function handleSave() {
    setSaving(true)
    try {
      if (id) {
        await api.notes.update(id, { title, body: { text: { text } }, color })
      } else {
        await api.notes.create({ title, body: { text: { text } }, color })
      }
      onSave()
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!id || !onDelete) return
    await api.notes.trash(id)
    onDelete()
  }

  const colorBg: Record<string, string> = {
    DEFAULT: "bg-white",
    RED: "bg-red-50",
    ORANGE: "bg-orange-50",
    YELLOW: "bg-yellow-50",
    GREEN: "bg-green-50",
    TEAL: "bg-teal-50",
    BLUE: "bg-blue-50",
    DARK_BLUE: "bg-blue-100",
    PURPLE: "bg-purple-50",
    PINK: "bg-pink-50",
    BROWN: "bg-amber-50",
    GRAY: "bg-gray-50",
  }

  return (
    <div className={`${colorBg[color] || "bg-white"} rounded-xl border border-gray-200 p-6 max-w-2xl mx-auto`}>
      <input
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="Title"
        className="w-full text-lg font-medium bg-transparent border-none outline-none mb-4 placeholder-gray-400"
      />
      <textarea
        value={text}
        onChange={(e) => setText(e.target.value)}
        placeholder="Take a note..."
        rows={10}
        className="w-full bg-transparent border-none outline-none resize-none text-sm placeholder-gray-400"
      />

      <div className="flex items-center justify-between mt-4 pt-4 border-t border-gray-100">
        <div className="flex gap-1">
          {COLOR_OPTIONS.map((c) => (
            <button
              key={c}
              onClick={() => setColor(c)}
              className={`w-5 h-5 rounded-full border ${c === "DEFAULT" ? "bg-white border-gray-300" : colorBg[c]} ${color === c ? "ring-2 ring-blue-500" : ""}`}
              title={c}
            />
          ))}
        </div>

        <div className="flex gap-2">
          {id && onDelete && (
            <button onClick={handleDelete} className="text-xs text-red-500 hover:underline">
              Delete
            </button>
          )}
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-4 py-1.5 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? "Saving..." : "Save"}
          </button>
        </div>
      </div>
    </div>
  )
}

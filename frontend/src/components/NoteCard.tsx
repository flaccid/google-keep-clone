"use client"

import type { Note } from "@/lib/types"
import { api } from "@/lib/api"
import { Pin, Archive, Trash2, Palette, MoreHorizontal } from "lucide-react"
import { useState } from "react"

const THEME_NAMES = ["THEME_SHORE", "THEME_BLOOM", "THEME_PLUM", "THEME_NIGHT", "THEME_BAMBOO", "THEME_CANDY", "THEME_SUNSET", "THEME_OCEAN"]

const BG_CLASSES: Record<string, string> = {
  DEFAULT: "bg-white dark:bg-[#202124]",
  RED: "bg-keep-red dark:bg-[#202124]",
  ORANGE: "bg-keep-orange dark:bg-[#202124]",
  YELLOW: "bg-keep-yellow dark:bg-[#202124]",
  GREEN: "bg-keep-green dark:bg-[#202124]",
  TEAL: "bg-keep-teal dark:bg-[#202124]",
  BLUE: "bg-keep-blue dark:bg-[#202124]",
  CERULEAN: "bg-keep-cerulean dark:bg-[#202124]",
  PURPLE: "bg-keep-purple dark:bg-[#202124]",
  PINK: "bg-keep-pink dark:bg-[#202124]",
  BROWN: "bg-keep-brown dark:bg-[#202124]",
  GRAY: "bg-keep-gray dark:bg-[#202124]",
  THEME_SHORE: "bg-theme-shore dark:bg-theme-shore-dark",
  THEME_BLOOM: "bg-theme-bloom dark:bg-theme-bloom-dark",
  THEME_PLUM: "bg-theme-plum dark:bg-theme-plum-dark",
  THEME_NIGHT: "bg-theme-night dark:bg-theme-night-dark",
  THEME_BAMBOO: "bg-theme-bamboo dark:bg-theme-bamboo-dark",
  THEME_CANDY: "bg-theme-candy dark:bg-theme-candy-dark",
  THEME_SUNSET: "bg-theme-sunset dark:bg-theme-sunset-dark",
  THEME_OCEAN: "bg-theme-ocean dark:bg-theme-ocean-dark",
}

const BORDER_CLASSES: Record<string, string> = {
  DEFAULT: "border-gray-200 dark:border-[#5f6368]",
  RED: "border-red-200 dark:border-[#5f6368]",
  ORANGE: "border-orange-200 dark:border-[#5f6368]",
  YELLOW: "border-yellow-200 dark:border-[#5f6368]",
  GREEN: "border-green-200 dark:border-[#5f6368]",
  TEAL: "border-teal-200 dark:border-[#5f6368]",
  BLUE: "border-blue-200 dark:border-[#5f6368]",
  CERULEAN: "border-sky-200 dark:border-[#5f6368]",
  PURPLE: "border-purple-200 dark:border-[#5f6368]",
  PINK: "border-pink-200 dark:border-[#5f6368]",
  BROWN: "border-amber-200 dark:border-[#5f6368]",
  GRAY: "border-gray-200 dark:border-[#5f6368]",
  THEME_SHORE: "border-amber-300 dark:border-[#5f6368]",
  THEME_BLOOM: "border-amber-300 dark:border-[#5f6368]",
  THEME_PLUM: "border-amber-300 dark:border-[#5f6368]",
  THEME_NIGHT: "border-amber-300 dark:border-[#5f6368]",
  THEME_BAMBOO: "border-amber-300 dark:border-[#5f6368]",
  THEME_CANDY: "border-amber-300 dark:border-[#5f6368]",
  THEME_SUNSET: "border-amber-300 dark:border-[#5f6368]",
  THEME_OCEAN: "border-amber-300 dark:border-[#5f6368]",
}

const COLOR_DOTS: Record<string, string> = {
  DEFAULT: "bg-white dark:bg-[#3c4043] border border-gray-300 dark:border-[#5f6368]",
  RED: "bg-keep-red border border-keep-red-dark",
  YELLOW: "bg-keep-yellow border border-keep-yellow-dark",
  GREEN: "bg-keep-green border border-green-300",
  TEAL: "bg-keep-teal border border-teal-300",
  BLUE: "bg-keep-blue border border-keep-blue-dark",
  CERULEAN: "bg-keep-cerulean border border-blue-300",
  PURPLE: "bg-keep-purple border border-purple-300",
  PINK: "bg-keep-pink border border-pink-300",
  BROWN: "bg-keep-brown border border-amber-300",
  GRAY: "bg-keep-gray border border-gray-300",
  ...Object.fromEntries(THEME_NAMES.map((t) => [t, "bg-white/50 border border-amber-300"])),
}

const COLOR_VALUES = ["DEFAULT", "RED", "ORANGE", "YELLOW", "GREEN", "TEAL", "BLUE", "CERULEAN", "PURPLE", "PINK", "BROWN", "GRAY"]

function noteTitle(note: Note): string {
  return note.title || ""
}

function notePreview(note: Note): { kind: "text"; text: string } | { kind: "list"; items: Array<{ text: string; checked: boolean }> } | null {
  if (note.body?.text?.text) return { kind: "text", text: note.body.text.text }
  if (note.body?.list?.listItems?.length) return { kind: "list", items: note.body.list.listItems.map((li) => ({ text: li.text?.text || "", checked: li.checked || false })) }
  return null
}

function noteId(name?: string): string {
  if (!name) return ""
  return name.replace("notes/", "")
}

export default function NoteCard({
  note,
  onUpdate,
  onOpen,
}: {
  note: Note
  onUpdate: () => void
  onOpen?: (id: string) => void
}) {
  const [showPalette, setShowPalette] = useState(false)
  const noteColor = note.color || "DEFAULT"
  const bgClass = BG_CLASSES[noteColor] || BG_CLASSES.DEFAULT
  const borderClass = BORDER_CLASSES[noteColor] || BORDER_CLASSES.DEFAULT
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

  async function changeColor(e: React.MouseEvent, c: string) {
    e.stopPropagation()
    e.preventDefault()
    await api.notes.update(id, { color: c } as any)
    setShowPalette(false)
    onUpdate()
  }

  const preview = notePreview(note)
  const hasContent = noteTitle(note) || preview

  return (
    <div
      className={`${bgClass} ${borderClass} rounded-lg border shadow-sm hover:shadow-lg dark:hover:shadow-xl dark:hover:shadow-black/30 transition-shadow cursor-pointer relative group`}
      onClick={() => onOpen?.(id)}
    >
      <button
        onClick={togglePin}
        className={`absolute top-1.5 right-1.5 p-1 rounded-full transition-opacity ${
          note.pinned
            ? "opacity-100 text-yellow-600"
            : "opacity-0 group-hover:opacity-100 text-gray-400 dark:text-[#9aa0a6] hover:text-gray-600 dark:hover:text-[#e8eaed]"
        }`}
        title={note.pinned ? "Unpin note" : "Pin note"}
      >
        <Pin size={14} fill={note.pinned ? "currentColor" : "none"} />
      </button>

      <div className="px-4 pt-4 pb-2">
        {noteTitle(note) && (
          <h3 className="text-sm font-medium mb-1 pr-6 text-gray-900 dark:text-[#e8eaed]">{noteTitle(note)}</h3>
        )}
        {preview?.kind === "text" && (
          <p className="text-gray-700 dark:text-[#bdc1c6] text-sm whitespace-pre-wrap leading-relaxed line-clamp-6">
            {preview.text}
          </p>
        )}
        {preview?.kind === "list" && (
          <div className="space-y-0.5">
            {preview.items.filter((i) => !i.checked).slice(0, 5).map((item, i) => (
              <div key={i} className="flex items-center gap-2">
                <span className="flex-shrink-0 w-3.5 h-3.5 rounded-full border-2 border-gray-400" />
                <span className="text-sm truncate text-gray-700 dark:text-[#bdc1c6]">{item.text}</span>
              </div>
            ))}
            {preview.items.some((i) => i.checked) && (
              <div className="flex items-center gap-2 pt-1">
                <hr className="flex-1 border-gray-200 dark:border-[#5f6368]" />
                <span className="text-[11px] text-gray-400 dark:text-[#9aa0a6] whitespace-nowrap">
                  {preview.items.filter((i) => i.checked).length} completed
                </span>
                <hr className="flex-1 border-gray-200 dark:border-[#5f6368]" />
              </div>
            )}
            {preview.items.filter((i) => i.checked).slice(0, 3).map((item, i) => (
              <div key={i} className="flex items-center gap-2">
                <span className="flex-shrink-0 w-3.5 h-3.5 rounded-full border-2 bg-gray-400 border-gray-400 flex items-center justify-center">
                  <span className="text-white text-[8px]">✓</span>
                </span>
                <span className="text-sm truncate text-gray-400 dark:text-[#9aa0a6] line-through">{item.text}</span>
              </div>
            ))}
            {(preview.items.filter((i) => !i.checked).length > 5 || preview.items.filter((i) => i.checked).length > 3) && (
              <span className="text-xs text-gray-400">+ more items</span>
            )}
          </div>
        )}
        {!hasContent && (
          <p className="text-gray-400 dark:text-[#9aa0a6] text-sm">Empty note</p>
        )}
      </div>

      {note.labels && note.labels.length > 0 && (
        <div className="flex flex-wrap gap-1 px-4 pb-1">
          {note.labels.map((l) => (
            <span key={l} className="text-[11px] bg-gray-200/60 dark:bg-white/10 text-gray-600 dark:text-[#bdc1c6] rounded px-1.5 py-0.5">
              {l}
            </span>
          ))}
        </div>
      )}

      <div className="flex items-center gap-0.5 px-1 py-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <button onClick={toggleArchive} className="p-1.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10 text-gray-500 dark:text-[#9aa0a6]" title={note.archived ? "Unarchive" : "Archive"}>
          <Archive size={16} />
        </button>
        <div className="relative">
          <button onClick={(e) => { e.stopPropagation(); e.preventDefault(); setShowPalette(!showPalette) }} className="p-1.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10 text-gray-500 dark:text-[#9aa0a6]" title="Background options">
            <Palette size={16} />
          </button>
          {showPalette && (
            <div className="absolute bottom-full left-0 mb-1 bg-white dark:bg-[#2d2e30] rounded-lg shadow-lg border border-gray-200 dark:border-[#5f6368] z-10 p-2 min-w-[200px]" onClick={(e) => e.stopPropagation()}>
              <p className="text-[11px] font-medium text-gray-400 dark:text-[#9aa0a6] uppercase tracking-wide mb-1.5 px-0.5">Themes</p>
              <div className="grid grid-cols-4 gap-1 mb-2">
                {THEME_NAMES.map((t) => (
                  <button
                    key={t}
                    onClick={(e) => changeColor(e, t)}
                    className={`w-9 h-9 rounded-lg ${t === noteColor ? "ring-2 ring-blue-500" : "ring-1 ring-black/10 dark:ring-white/20"} overflow-hidden`}
                    title={t.replace("THEME_", "").replace(/_/g, " ").toLowerCase().replace(/\b\w/g, (l) => l.toUpperCase())}
                  >
                    <div className={`w-full h-full ${BG_CLASSES[t] || ""}`} />
                  </button>
                ))}
              </div>
              <p className="text-[11px] font-medium text-gray-400 dark:text-[#9aa0a6] uppercase tracking-wide mb-1.5 px-0.5">Colors</p>
              <div className="flex gap-0.5 flex-wrap">
                {COLOR_VALUES.map((c) => (
                  <button
                    key={c}
                    onClick={(e) => changeColor(e, c)}
                    className={`w-5 h-5 rounded-full ${COLOR_DOTS[c]} ${c === noteColor ? "ring-2 ring-blue-500" : ""}`}
                    title={c === "DEFAULT" ? "Default" : c.charAt(0) + c.slice(1).toLowerCase()}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
        <button onClick={toggleTrash} className="p-1.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10 text-gray-500 dark:text-[#9aa0a6]" title="Delete">
          <Trash2 size={16} />
        </button>
        <button className="p-1.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10 text-gray-500 dark:text-[#9aa0a6] ml-auto" title="More">
          <MoreHorizontal size={16} />
        </button>
      </div>
    </div>
  )
}

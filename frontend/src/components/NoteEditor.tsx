"use client"

import { useState, useRef, useEffect } from "react"
import type { Note, ListItem as ListItemType, Label } from "@/lib/types"
import { api } from "@/lib/api"
import { Pin, Palette, Image, ListChecks, Type, X, Check, Tag } from "lucide-react"

const COLOR_OPTIONS = [
  "DEFAULT", "RED", "ORANGE", "YELLOW", "GREEN", "TEAL",
  "BLUE", "DARK_BLUE", "PURPLE", "PINK", "BROWN", "GRAY",
]

const COLOR_VALUES: Record<string, string> = {
  DEFAULT: "bg-white dark:bg-[#3c4043] border border-gray-300 dark:border-[#5f6368]",
  RED: "bg-keep-red border border-keep-red-dark",
  ORANGE: "bg-keep-orange border border-orange-300",
  YELLOW: "bg-keep-yellow border border-keep-yellow-dark",
  GREEN: "bg-keep-green border border-green-300",
  TEAL: "bg-keep-teal border border-teal-300",
  BLUE: "bg-keep-blue border border-keep-blue-dark",
  DARK_BLUE: "bg-keep-dark-blue border border-blue-300",
  PURPLE: "bg-keep-purple border border-purple-300",
  PINK: "bg-keep-pink border border-pink-300",
  BROWN: "bg-keep-brown border border-amber-300",
  GRAY: "bg-keep-gray border border-gray-300",
}

const BG_COLORS: Record<string, string> = {
  DEFAULT: "bg-white dark:bg-[#202124]",
  RED: "bg-keep-red dark:bg-[#202124]",
  ORANGE: "bg-keep-orange dark:bg-[#202124]",
  YELLOW: "bg-keep-yellow dark:bg-[#202124]",
  GREEN: "bg-keep-green dark:bg-[#202124]",
  TEAL: "bg-keep-teal dark:bg-[#202124]",
  BLUE: "bg-keep-blue dark:bg-[#202124]",
  DARK_BLUE: "bg-keep-dark-blue dark:bg-[#202124]",
  PURPLE: "bg-keep-purple dark:bg-[#202124]",
  PINK: "bg-keep-pink dark:bg-[#202124]",
  BROWN: "bg-keep-brown dark:bg-[#202124]",
  GRAY: "bg-keep-gray dark:bg-[#202124]",
}

function initListItems(note?: Note): Array<{ text: string; checked: boolean }> {
  if (!note?.body?.list?.listItems) return []
  return note.body.list.listItems.map((li) => ({
    text: li.text?.text || "",
    checked: li.checked || false,
  }))
}

export default function NoteEditor({
  note,
  onSave,
  onDelete,
  onClose,
}: {
  note?: Note
  onSave: () => void
  onDelete?: () => void
  onClose?: () => void
}) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [expanded, setExpanded] = useState(!!note)
  const [mode, setMode] = useState<"text" | "list">(
    note?.body?.list?.listItems?.length ? "list" : "text"
  )
  const [title, setTitle] = useState(note?.title || "")
  const [text, setText] = useState(note?.body?.text?.text || "")
  const [listItems, setListItems] = useState<Array<{ text: string; checked: boolean }>>(initListItems(note))
  const [color, setColor] = useState(note?.color || "DEFAULT")
  const [pinned, setPinned] = useState(note?.pinned || false)
  const [saving, setSaving] = useState(false)
  const [files, setFiles] = useState<File[]>([])
  const [selectedLabels, setSelectedLabels] = useState<string[]>(note?.labels || [])
  const [showLabelPicker, setShowLabelPicker] = useState(false)
  const [showColors, setShowColors] = useState(false)
  const [availableLabels, setAvailableLabels] = useState<Label[]>([])
  const fileRef = useRef<HTMLInputElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const listEndRef = useRef<HTMLDivElement>(null)
  const labelPickerRef = useRef<HTMLDivElement>(null)
  const colorPickerRef = useRef<HTMLDivElement>(null)

  const id = note?.name?.replace("notes/", "") || ""
  const isNew = !id
  const bg = BG_COLORS[color] || BG_COLORS.DEFAULT

  useEffect(() => {
    if (expanded && inputRef.current) {
      inputRef.current.focus()
    }
    if (expanded) {
      api.labels.list().then((labels) => setAvailableLabels(labels)).catch(() => {})
    }
  }, [expanded])

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (labelPickerRef.current && !labelPickerRef.current.contains(e.target as Node)) {
        setShowLabelPicker(false)
      }
      if (colorPickerRef.current && !colorPickerRef.current.contains(e.target as Node)) {
        setShowColors(false)
      }
    }
    document.addEventListener("mousedown", handleClick)
    return () => document.removeEventListener("mousedown", handleClick)
  }, [])

  const handleCloseRef = useRef(handleClose)
  handleCloseRef.current = handleClose

  useEffect(() => {
    if (!expanded) return
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        handleCloseRef.current()
      }
    }
    document.addEventListener("mousedown", handleClick)
    return () => document.removeEventListener("mousedown", handleClick)
  }, [expanded])

  function buildBody() {
    if (mode === "list") {
      const items = listItems
        .filter((li) => li.text.trim())
        .map((li) => ({
          text: { text: li.text },
          checked: li.checked,
        }))
      return { list: { listItems: items } }
    }
    return { text: { text } }
  }

  async function handleSave() {
    if (saving) return
    setSaving(true)
    let savedId = id
    try {
      const body = buildBody()
      const labels = selectedLabels.length > 0 ? selectedLabels : undefined
      if (savedId) {
        const payload: any = { title, body, color, labels }
        if (pinned !== !!note?.pinned) payload.pinned = pinned
        await api.notes.update(savedId, payload)
      } else {
        const items = mode === "list" ? listItems.filter((li) => li.text.trim()) : []
        if (!title && !text && items.length === 0 && files.length === 0) return
        const created = await api.notes.create({ title, body, color, pinned, labels })
        savedId = created.name?.replace("notes/", "") || ""
      }
      if (files.length > 0 && savedId) {
        for (const f of files) {
          await api.notes.uploadAttachment(savedId, f)
        }
        setFiles([])
      }
      onSave()
    } finally {
      setSaving(false)
    }
  }

  async function handleClose() {
    if (isNew) {
      const hasContent = title || text || listItems.some((li) => li.text.trim()) || files.length > 0
      if (hasContent) await handleSave()
    } else {
      await handleSave()
    }
    setExpanded(false)
    setTitle("")
    setText("")
    setListItems([])
    setMode("text")
    setColor("DEFAULT")
    setPinned(false)
    setFiles([])
    setSelectedLabels([])
    setShowLabelPicker(false)
    setShowColors(false)
    onClose?.()
  }

  async function handleDelete() {
    if (!id || !onDelete) return
    await api.notes.trash(id)
    onDelete()
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Escape") handleClose()
  }

  function addListItem(afterIndex?: number) {
    setListItems((prev) => {
      const next = [...prev]
      const newItem = { text: "", checked: false }
      if (afterIndex !== undefined) {
        next.splice(afterIndex + 1, 0, newItem)
      } else {
        next.push(newItem)
      }
      return next
    })
  }

  function updateListItem(index: number, text: string) {
    setListItems((prev) => {
      const next = [...prev]
      next[index] = { ...next[index], text }
      return next
    })
  }

  function toggleListItem(index: number) {
    setListItems((prev) => {
      const next = [...prev]
      next[index] = { ...next[index], checked: !next[index].checked }
      return next
    })
  }

  function removeListItem(index: number) {
    setListItems((prev) => prev.filter((_, i) => i !== index))
  }

  function handleListItemKeyDown(e: React.KeyboardEvent<HTMLInputElement>, index: number) {
    if (e.key === "Enter") {
      e.preventDefault()
      addListItem(index)
      setTimeout(() => {
        const next = document.querySelector(`[data-li="${index + 1}"]`) as HTMLInputElement
        next?.focus()
      }, 0)
    }
    if (e.key === "Backspace" && listItems[index].text === "" && listItems.length > 1) {
      removeListItem(index)
    }
  }

  function openWithMode(m: "text" | "list") {
    setMode(m)
    setExpanded(true)
  }

  if (!expanded) {
    return (
      <div className="bg-white dark:bg-[#202124] rounded-lg border border-gray-200 dark:border-[#5f6368] shadow-sm max-w-2xl mx-auto">
        <div
          onClick={() => openWithMode("text")}
          className="px-4 py-3 flex items-center justify-between cursor-text hover:shadow-md transition-shadow rounded-t-lg"
        >
          <span className="text-gray-400 dark:text-[#9aa0a6] text-sm">Take a note...</span>
          <div className="flex items-center gap-2">
            <button
              onClick={(e) => { e.stopPropagation(); openWithMode("list") }}
              className="p-1.5 rounded-full hover:bg-gray-100 dark:hover:bg-white/10 text-gray-400 dark:text-[#9aa0a6]"
              title="New list"
            >
              <ListChecks size={18} />
            </button>
            <button
              onClick={(e) => { e.stopPropagation(); setPinned(!pinned) }}
              className="p-1.5 rounded-full hover:bg-gray-100 dark:hover:bg-white/10 text-gray-400 dark:text-[#9aa0a6]"
              title="Pin note"
            >
              <Pin size={16} fill={pinned ? "currentColor" : "none"} className={pinned ? "text-yellow-600" : ""} />
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div
      ref={containerRef}
      className={`${bg} rounded-lg border border-gray-200 dark:border-[#5f6368] shadow-sm max-w-2xl mx-auto`}
      onKeyDown={handleKeyDown}
    >
      <div className="px-4 pt-4 pb-2">
        <input
          ref={inputRef}
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Title"
          className="w-full text-base font-medium bg-transparent border-none outline-none mb-1 placeholder-gray-400 dark:placeholder-[#9aa0a6] text-gray-900 dark:text-[#e8eaed]"
        />

        {mode === "list" ? (
          <div className="space-y-0.5">
            {listItems
              .map((item, origIndex) => ({ item, origIndex }))
              .filter(({ item }) => !item.checked)
              .map(({ item, origIndex }) => (
                <div key={origIndex} className="group flex items-center gap-2 -ml-1">
                  <button
                    onClick={() => toggleListItem(origIndex)}
                    className="flex-shrink-0 w-4 h-4 rounded-full border-2 border-gray-400 dark:border-[#9aa0a6] hover:border-gray-600"
                  />
                  <input
                    data-li={origIndex}
                    value={item.text}
                    onChange={(e) => updateListItem(origIndex, e.target.value)}
                    onKeyDown={(e) => handleListItemKeyDown(e, origIndex)}
                    placeholder={origIndex === listItems.length - 1 ? "List item" : ""}
                    className="flex-1 bg-transparent border-none outline-none text-sm py-1 placeholder-gray-400 dark:placeholder-[#9aa0a6] text-gray-900 dark:text-[#e8eaed]"
                  />
                  <button
                    onClick={() => removeListItem(origIndex)}
                    className="opacity-0 group-hover:opacity-100 p-0.5 rounded hover:bg-black/10 dark:hover:bg-white/10 text-gray-400 dark:text-[#9aa0a6] transition-opacity"
                    title="Delete item"
                  >
                    <X size={14} />
                  </button>
                </div>
              ))}
            {listItems.some((i) => i.checked) && (
              <div className="flex items-center gap-2 pt-1">
                <hr className="flex-1 border-gray-200 dark:border-[#5f6368]" />
                <span className="text-[11px] text-gray-400 dark:text-[#9aa0a6] whitespace-nowrap">
                  {listItems.filter((i) => i.checked).length} completed
                </span>
                <hr className="flex-1 border-gray-200 dark:border-[#5f6368]" />
              </div>
            )}
            {listItems
              .map((item, origIndex) => ({ item, origIndex }))
              .filter(({ item }) => item.checked)
              .map(({ item, origIndex }) => (
                <div key={origIndex} className="group flex items-center gap-2 -ml-1">
                  <button
                    onClick={() => toggleListItem(origIndex)}
                    className={`flex-shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center transition-colors bg-gray-500 border-gray-500 dark:bg-[#9aa0a6] dark:border-[#9aa0a6]`}
                  >
                    <Check size={10} className="text-white" strokeWidth={3} />
                  </button>
                  <input
                    data-li={origIndex}
                    value={item.text}
                    onChange={(e) => updateListItem(origIndex, e.target.value)}
                    onKeyDown={(e) => handleListItemKeyDown(e, origIndex)}
                    placeholder={origIndex === listItems.length - 1 ? "List item" : ""}
                    className="flex-1 bg-transparent border-none outline-none text-sm py-1 placeholder-gray-400 dark:placeholder-[#9aa0a6] text-gray-400 dark:text-[#9aa0a6] line-through"
                  />
                  <button
                    onClick={() => removeListItem(origIndex)}
                    className="opacity-0 group-hover:opacity-100 p-0.5 rounded hover:bg-black/10 dark:hover:bg-white/10 text-gray-400 dark:text-[#9aa0a6] transition-opacity"
                    title="Delete item"
                  >
                    <X size={14} />
                  </button>
                </div>
              ))}
            <div ref={listEndRef} />
            <button
              onClick={() => addListItem()}
              className="text-sm text-gray-400 dark:text-[#9aa0a6] hover:text-gray-600 dark:hover:text-[#e8eaed] py-1"
            >
              + Add item
            </button>
          </div>
        ) : (
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Take a note..."
            rows={isNew ? 4 : 10}
            className="w-full bg-transparent border-none outline-none resize-none text-sm placeholder-gray-400 dark:placeholder-[#9aa0a6] text-gray-900 dark:text-[#e8eaed]"
          />
        )}
      </div>

      <div className="flex items-center justify-between px-2 py-1 border-t border-black/10 dark:border-white/10">
        <div className="flex items-center gap-0.5">
          <button
            onClick={() => setMode(mode === "text" ? "list" : "text")}
            className="p-1.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10 text-gray-500 dark:text-[#9aa0a6]"
            title={mode === "text" ? "Switch to list" : "Switch to text"}
          >
            {mode === "text" ? <ListChecks size={16} /> : <Type size={16} />}
          </button>

          <div className="relative" ref={colorPickerRef}>
            <button
              onClick={() => setShowColors(!showColors)}
              className="p-1.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10 text-gray-500 dark:text-[#9aa0a6]"
              title="Background options"
            >
              <Palette size={16} />
            </button>
            {showColors && (
              <div className="absolute bottom-full left-0 mb-2 flex gap-0.5 p-1.5 bg-white dark:bg-[#2d2e30] rounded-lg shadow-lg border border-gray-200 dark:border-[#5f6368] z-10">
                {COLOR_OPTIONS.map((c) => (
                  <button
                    key={c}
                    onClick={() => setColor(c)}
                    className={`w-5 h-5 rounded-full ${COLOR_VALUES[c]} ${c === color ? "ring-2 ring-blue-500" : ""}`}
                    title={c}
                  />
                ))}
              </div>
            )}
          </div>

          <button
            onClick={() => fileRef.current?.click()}
            className="p-1.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10 text-gray-500 dark:text-[#9aa0a6]"
            title="Attach file"
          >
            <Image size={16} />
          </button>
          <input
            ref={fileRef}
            type="file"
            multiple
            onChange={(e) => {
              if (e.target.files) setFiles(Array.from(e.target.files))
            }}
            className="hidden"
          />
          {files.length > 0 && (
            <span className="text-xs text-gray-400 dark:text-[#9aa0a6] ml-1">{files.length} file(s)</span>
          )}

          <div className="relative" ref={labelPickerRef}>
            <button
              onClick={() => setShowLabelPicker(!showLabelPicker)}
              className="p-1.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10 text-gray-500 dark:text-[#9aa0a6]"
              title="Labels"
            >
              <Tag size={16} />
            </button>
            {showLabelPicker && (
              <div className="absolute bottom-full left-0 mb-1 bg-white dark:bg-[#2d2e30] rounded-lg shadow-lg border border-gray-200 dark:border-[#5f6368] z-10 min-w-[180px] max-h-48 overflow-y-auto py-1">
                {availableLabels.length === 0 && (
                  <p className="px-3 py-2 text-xs text-gray-400 dark:text-[#9aa0a6]">No labels yet</p>
                )}
                {availableLabels.map((label) => {
                  const displayName = label.displayName || ""
                  const checked = selectedLabels.includes(displayName)
                  return (
                    <label
                      key={label.name}
                      className="flex items-center gap-2 px-3 py-1.5 hover:bg-black/5 dark:hover:bg-white/10 cursor-pointer text-sm"
                    >
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={() => {
                          setSelectedLabels((prev) =>
                            checked
                              ? prev.filter((n) => n !== displayName)
                              : [...prev, displayName]
                          )
                        }}
                        className="accent-blue-500"
                      />
                      <span className="text-gray-700 dark:text-[#e8eaed]">{displayName}</span>
                    </label>
                  )
                })}
              </div>
            )}
          </div>

          {selectedLabels.length > 0 && (
            <div className="flex items-center gap-1 ml-1">
              {selectedLabels.slice(0, 3).map((l) => (
                <span key={l} className="text-[10px] bg-black/10 dark:bg-white/10 text-gray-600 dark:text-[#c4c7c5] px-1.5 py-0.5 rounded">
                  {l}
                </span>
              ))}
              {selectedLabels.length > 3 && (
                <span className="text-[10px] text-gray-400 dark:text-[#9aa0a6]">+{selectedLabels.length - 3}</span>
              )}
            </div>
          )}
        </div>

        <div className="flex items-center gap-1">
          {!isNew && onDelete && (
            <button onClick={handleDelete} className="px-3 py-1 text-sm text-gray-500 dark:text-[#9aa0a6] hover:bg-black/10 dark:hover:bg-white/10 rounded">
              Delete
            </button>
          )}
          <button
            onClick={handleClose}
            className="px-3 py-1 text-sm text-gray-500 dark:text-[#9aa0a6] hover:bg-black/10 dark:hover:bg-white/10 rounded"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  )
}

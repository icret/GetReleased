'use client'

import { Select as SelectPrimitive } from '@base-ui/react/select'
import { ChevronDown } from 'lucide-react'

import { cn } from '@/lib/utils'

interface SelectOption {
  value: string
  label: string
}

interface SelectProps {
  value: string
  onValueChange: (value: string) => void
  options: SelectOption[]
  placeholder?: string
  className?: string
  disabled?: boolean
}

export function Select({ value, onValueChange, options, placeholder = '选择...', className, disabled }: SelectProps) {
  const selected = options.find((o) => o.value === value)

  return (
    <SelectPrimitive.Root
      value={value}
      onValueChange={(v) => {
        if (v !== null) onValueChange(v)
      }}
      disabled={disabled}
    >
      <SelectPrimitive.Trigger
        className={cn(
          'flex h-9 w-full items-center justify-between gap-2 rounded-lg border border-border bg-muted/40 px-3 py-1.5',
          'text-sm text-foreground backdrop-blur transition-colors hover:border-primary/50 hover:bg-muted/60',
          'focus-visible:border-primary/50 focus-visible:ring-3 focus-visible:ring-ring/40',
          'disabled:cursor-not-allowed disabled:opacity-50 data-[placeholder]:text-muted-foreground',
          className,
        )}
      >
        <SelectPrimitive.Value className="truncate">{selected ? selected.label : placeholder}</SelectPrimitive.Value>
        <ChevronDown className="size-4 shrink-0 text-muted-foreground" />
      </SelectPrimitive.Trigger>
      <SelectPrimitive.Portal>
        <SelectPrimitive.Positioner className="z-50 min-w-[var(--anchor-width)]">
          <SelectPrimitive.Popup className="glass-strong max-h-64 overflow-y-auto rounded-xl p-1.5 text-popover-foreground">
            <SelectPrimitive.List>
              {options.map((option) => (
                <SelectPrimitive.Item key={option.value} value={option.value} className="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm outline-none transition-colors data-[highlighted]:bg-primary/15 data-[highlighted]:text-primary">
                  <SelectPrimitive.ItemIndicator className="ml-auto" />
                  <SelectPrimitive.ItemText>{option.label}</SelectPrimitive.ItemText>
                </SelectPrimitive.Item>
              ))}
            </SelectPrimitive.List>
          </SelectPrimitive.Popup>
        </SelectPrimitive.Positioner>
      </SelectPrimitive.Portal>
    </SelectPrimitive.Root>
  )
}

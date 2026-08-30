'use client'

import { Tooltip as TooltipPrimitive } from '@base-ui/react/tooltip'
import type { ReactNode } from 'react'

interface TooltipProps {
  children: ReactNode
  content: ReactNode
  className?: string
}

function Tooltip({ children, content, className }: TooltipProps) {
  return (
    <TooltipPrimitive.Provider delay={100} closeDelay={0}>
      <TooltipPrimitive.Root>
        <TooltipPrimitive.Trigger render={<span />} className={className}>
          {children}
        </TooltipPrimitive.Trigger>
        <TooltipPrimitive.Portal>
          <TooltipPrimitive.Positioner side="top" sideOffset={8}>
            <TooltipPrimitive.Popup className="glass-strong max-w-xs rounded-lg px-3 py-2 text-xs text-popover-foreground">{content}</TooltipPrimitive.Popup>
          </TooltipPrimitive.Positioner>
        </TooltipPrimitive.Portal>
      </TooltipPrimitive.Root>
    </TooltipPrimitive.Provider>
  )
}

export { Tooltip }

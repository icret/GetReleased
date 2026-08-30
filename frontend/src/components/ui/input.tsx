'use client'

import { Input as InputPrimitive } from '@base-ui/react/input'
import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@/lib/utils'

const inputVariants = cva(
  'flex h-9 w-full rounded-lg border border-border bg-muted/40 px-3 py-1.5 text-sm text-foreground ' +
    'backdrop-blur placeholder:text-muted-foreground transition-colors file:border-0 file:bg-transparent ' +
    'file:text-sm file:font-medium focus-visible:border-primary/50 focus-visible:bg-muted/60 ' +
    'focus-visible:ring-3 focus-visible:ring-ring/40 disabled:cursor-not-allowed disabled:opacity-50',
)

function Input({ className, ...props }: InputPrimitive.Props & VariantProps<typeof inputVariants>) {
  return <InputPrimitive data-slot="input" className={cn(inputVariants(), className)} {...props} />
}

export { Input, inputVariants }

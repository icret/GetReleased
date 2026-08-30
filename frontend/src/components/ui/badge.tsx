import type { ComponentProps } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@/lib/utils'

const badgeVariants = cva('inline-flex shrink-0 items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium whitespace-nowrap transition-colors [&_svg]:size-3 [&_svg]:shrink-0', {
  variants: {
    variant: {
      default: 'border-primary/20 bg-primary/12 text-primary',
      secondary: 'border-border bg-muted text-muted-foreground',
      outline: 'border-border bg-transparent text-muted-foreground',
      destructive: 'border-destructive/20 bg-destructive/12 text-destructive',
      success: 'border-success/25 bg-success/12 text-success',
      warning: 'border-warning/25 bg-warning/12 text-warning',
    },
  },
  defaultVariants: {
    variant: 'default',
  },
})

function Badge({ className, variant, ...props }: ComponentProps<'span'> & VariantProps<typeof badgeVariants>) {
  return <span data-slot="badge" className={cn(badgeVariants({ variant }), className)} {...props} />
}

export { Badge, badgeVariants }

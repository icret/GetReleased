'use client'

import { ChevronLeft, ChevronRight } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

export interface PaginationProps {
  totalItems: number
  itemsPerPage: number
  currentPage: number
  onPageChange: (page: number) => void
  className?: string
}

function getPageNumbers(totalPages: number, currentPage: number): (number | '...')[] {
  const pages: (number | '...')[] = []
  const delta = 2

  pages.push(1)

  for (let i = Math.max(2, currentPage - delta); i <= Math.min(totalPages - 1, currentPage + delta); i++) {
    pages.push(i)
  }

  if (currentPage + delta < totalPages - 1) {
    pages.push('...')
  }

  if (totalPages > 1) {
    pages.push(totalPages)
  }

  return pages
}

export function Pagination({ totalItems, itemsPerPage, currentPage, onPageChange, className }: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(totalItems / itemsPerPage))

  if (totalPages <= 1) {
    return null
  }

  const pages = getPageNumbers(totalPages, currentPage)
  const startItem = (currentPage - 1) * itemsPerPage + 1
  const endItem = Math.min(currentPage * itemsPerPage, totalItems)

  const canPrev = currentPage > 1
  const canNext = currentPage < totalPages

  return (
    <div className={cn('flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between', className)}>
      <p className="text-xs tabular-nums text-muted-foreground">
        显示第 <span className="font-semibold text-foreground">{startItem}</span>
        <span className="mx-0.5">-</span>
        <span className="font-semibold text-foreground">{endItem}</span> 条，共 <span className="font-semibold text-foreground">{totalItems}</span> 条
      </p>

      <nav aria-label="分页导航" className="inline-flex items-center gap-0.5 rounded-lg border border-border/70 bg-card/60 p-1 backdrop-blur">
        <Button variant="ghost" size="icon-sm" onClick={() => onPageChange(currentPage - 1)} disabled={!canPrev} aria-label="上一页" className="text-muted-foreground hover:text-foreground">
          <ChevronLeft />
        </Button>

        {pages.map((page, index) =>
          page === '...' ? (
            <span key={`ellipsis-${index}`} className="flex h-7 items-center px-1 text-xs text-muted-foreground">
              ...
            </span>
          ) : page === currentPage ? (
            <Button key={page} size="icon-sm" onClick={() => onPageChange(page)} aria-label={`第 ${page} 页`} aria-current="page" className="bg-primary text-primary-foreground shadow-md shadow-primary/40 hover:bg-primary/90 dark:shadow-primary/30">
              {page}
            </Button>
          ) : (
            <Button key={page} variant="ghost" size="icon-sm" onClick={() => onPageChange(page)} aria-label={`第 ${page} 页`} className="text-muted-foreground hover:text-foreground">
              {page}
            </Button>
          ),
        )}

        <Button variant="ghost" size="icon-sm" onClick={() => onPageChange(currentPage + 1)} disabled={!canNext} aria-label="下一页" className="text-muted-foreground hover:text-foreground">
          <ChevronRight />
        </Button>
      </nav>
    </div>
  )
}

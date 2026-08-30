'use client'

import { KeyRound, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatRelativeTime } from '@/lib/format'
import type { User } from '@/types'

interface UserTableProps {
  users: User[]
  currentUsername: string | null
  onResetPassword: (user: User) => void
  onDelete: (user: User) => void
}

export function UserTable({ users, currentUsername, onResetPassword, onDelete }: UserTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>用户名</TableHead>
          <TableHead>角色</TableHead>
          <TableHead>创建时间</TableHead>
          <TableHead className="w-[100px]">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {users.map((user) => {
          const isSelf = user.username === currentUsername
          return (
            <TableRow key={user.id}>
              <TableCell className="font-medium">
                {user.username}
                {isSelf && (
                  <Badge variant="secondary" className="ml-2">
                    当前
                  </Badge>
                )}
              </TableCell>
              <TableCell className="text-muted-foreground">{user.role}</TableCell>
              <TableCell className="text-muted-foreground">{formatRelativeTime(user.created_at)}</TableCell>
              <TableCell>
                <div className="flex items-center gap-1">
                  <Button variant="ghost" size="icon" title="重置密码" onClick={() => onResetPassword(user)}>
                    <KeyRound />
                  </Button>
                  <Button variant="ghost" size="icon" title="删除" disabled={isSelf} onClick={() => onDelete(user)}>
                    <Trash2 />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}

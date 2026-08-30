'use client'

import { useActionState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Toaster, toast } from '@/components/ui/toast'
import { login } from '@/lib/auth'

interface LoginState {
  error: string
}

const initialState: LoginState = { error: '' }

export default function LoginPage() {
  const router = useRouter()

  async function loginAction(_prev: LoginState, formData: FormData): Promise<LoginState> {
    const username = formData.get('username') as string
    const password = formData.get('password') as string
    try {
      await login(username, password)
      router.push('/admin')
      return { error: '' }
    } catch (err) {
      const message = (err as Error).message
      toast.add({ title: '登录失败', description: message })
      return { error: message }
    }
  }

  const [state, formAction, pending] = useActionState(loginAction, initialState)

  return (
    <Toaster>
      <div className="mx-auto flex min-h-[60vh] max-w-sm flex-col justify-center gap-6">
        <div className="space-y-2 text-center">
          <h1 className="text-2xl font-bold">管理员登录</h1>
          <p className="text-sm text-muted-foreground">请输入用户名和密码以访问管理后台</p>
        </div>
        <form action={formAction} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="username">用户名</Label>
            <Input id="username" name="username" type="text" autoFocus required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="password">密码</Label>
            <Input id="password" name="password" type="password" required />
          </div>
          {state.error && <p className="text-sm text-destructive">{state.error}</p>}
          <Button type="submit" disabled={pending} className="w-full">
            {pending ? '登录中...' : '登录'}
          </Button>
        </form>
      </div>
    </Toaster>
  )
}

import { describe, it, expect } from 'vitest'
import { parseRepoInput } from './repo-url'

describe('parseRepoInput', () => {
  const tests: { name: string; input: string; expected: { owner: string; name: string } | null }[] = [
    { name: '完整 URL', input: 'https://github.com/ClashX-Pro/ClashX', expected: { owner: 'ClashX-Pro', name: 'ClashX' } },
    { name: 'release tag URL', input: 'https://github.com/ClashX-Pro/ClashX/releases/tag/1.140.0', expected: { owner: 'ClashX-Pro', name: 'ClashX' } },
    { name: 'blob 文件 URL', input: 'https://github.com/ClashX-Pro/ClashX/blob/master/README_zh-CN.md', expected: { owner: 'ClashX-Pro', name: 'ClashX' } },
    { name: 'owner/repo 简写', input: 'ClashX-Pro/ClashX', expected: { owner: 'ClashX-Pro', name: 'ClashX' } },
    { name: 'owner/repo 带多余路径', input: 'octocat/hello/extra/path', expected: { owner: 'octocat', name: 'hello' } },
    { name: '.git 后缀去除', input: 'https://github.com/octocat/hello.git', expected: { owner: 'octocat', name: 'hello' } },
    { name: '前后空格', input: '  octocat/hello  ', expected: { owner: 'octocat', name: 'hello' } },
    { name: '空字符串', input: '', expected: null },
    { name: '仅空格', input: '   ', expected: null },
    { name: '单段无斜杠', input: 'single', expected: null },
    { name: '仅域名', input: 'https://github.com/', expected: null },
    { name: '域名+单段', input: 'https://github.com/onlyone', expected: null },
  ]

  for (const tt of tests) {
    it(tt.name, () => {
      expect(parseRepoInput(tt.input)).toEqual(tt.expected)
    })
  }
})

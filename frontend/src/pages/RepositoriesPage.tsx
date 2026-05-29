import { useEffect, useState } from 'react'
import { Alert, Button, Paper, Stack, TextField, Typography } from '@mui/material'
import { api } from '../lib/api'

type Repo = { id: string; name: string; url: string; type: string; username: string; createdAt: string }

export function RepositoriesPage() {
  const [name, setName] = useState('')
  const [url, setURL] = useState('')
  const [result, setResult] = useState('')
  const [repos, setRepos] = useState<Repo[]>([])

  const load = async () => {
    try { setRepos(await api<Repo[]>('/api/v1/repositories/credentials')) } catch { setRepos([]) }
  }

  useEffect(() => { load() }, [])

  const create = async () => {
    try {
      await api('/api/v1/repositories/credentials', {
        method: 'POST',
        body: JSON.stringify({ name, url, type: 'https', username: 'git', secretRef: 'repo-secret' })
      })
      setResult('Credential saved')
      load()
    } catch (e) {
      setResult(e instanceof Error ? e.message : 'Request failed')
    }
  }

  return <Stack spacing={2}>
    <Typography variant='h5'>Repository Credentials</Typography>
    <Paper sx={{ p: 2 }}>
      <Stack spacing={2}>
        <TextField label='Name' value={name} onChange={e => setName(e.target.value)} />
        <TextField label='URL' value={url} onChange={e => setURL(e.target.value)} />
        <Button variant='contained' onClick={create}>Save</Button>
        {result && <Alert severity='info'>{result}</Alert>}
      </Stack>
    </Paper>
    <Paper sx={{ p: 2 }}>
      <Typography variant='h6'>Saved Credentials</Typography>
      {repos.map(r => <Typography key={r.id} variant='body2'>{r.name} | {r.url} | {r.type}</Typography>)}
    </Paper>
  </Stack>
}

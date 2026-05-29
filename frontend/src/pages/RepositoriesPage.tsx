import { useState } from 'react'
import { Button, Paper, Stack, TextField, Typography } from '@mui/material'

export function RepositoriesPage() {
  const [name, setName] = useState('')
  const [url, setURL] = useState('')
  const [result, setResult] = useState('')

  const create = async () => {
    const res = await fetch('/api/v1/repositories/credentials', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, url, type: 'https', username: 'git', secretRef: 'repo-secret' })
    })
    setResult(res.ok ? 'Credential saved' : 'Request failed')
  }

  return <Stack spacing={2}>
    <Typography variant='h5'>Repository Credentials</Typography>
    <Paper sx={{ p: 2 }}>
      <Stack spacing={2}>
        <TextField label='Name' value={name} onChange={e => setName(e.target.value)} />
        <TextField label='URL' value={url} onChange={e => setURL(e.target.value)} />
        <Button variant='contained' onClick={create}>Save</Button>
        <Typography variant='body2'>{result}</Typography>
      </Stack>
    </Paper>
  </Stack>
}

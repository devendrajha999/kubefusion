import { useEffect, useState } from 'react'
import { Alert, Button, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography } from '@mui/material'
import { api } from '../lib/api'

type App = { id: string; name: string; project: string; syncStatus: string; health: string }

export function ApplicationsPage() {
  const [apps, setApps] = useState<App[]>([])
  const [rollbackRevision, setRollbackRevision] = useState('main')
  const [error, setError] = useState('')

  const load = async () => {
    try {
      setError('')
      setApps(await api<App[]>('/api/v1/applications'))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Load failed')
      setApps([])
    }
  }
  useEffect(() => { load() }, [])

  const sync = async (id: string) => { await api(`/api/v1/applications/${id}/sync`, { method: 'POST' }); load() }
  const drift = async (id: string) => { const data = await api<{ status: string }>(`/api/v1/applications/${id}/drift`); alert(`Drift: ${data.status}`) }
  const rollback = async (id: string) => { await api(`/api/v1/applications/${id}/rollback`, { method: 'POST', body: JSON.stringify({ revision: rollbackRevision }) }); load() }

  return <Stack spacing={2}><Typography variant='h5'>Applications</Typography>{error && <Alert severity='error'>{error}</Alert>}<TextField label='Rollback Revision' value={rollbackRevision} onChange={e => setRollbackRevision(e.target.value)} /><Paper><Table><TableHead><TableRow><TableCell>Name</TableCell><TableCell>Project</TableCell><TableCell>Sync</TableCell><TableCell>Health</TableCell><TableCell>Actions</TableCell></TableRow></TableHead><TableBody>{apps.map(a => <TableRow key={a.id}><TableCell>{a.name}</TableCell><TableCell>{a.project}</TableCell><TableCell>{a.syncStatus}</TableCell><TableCell>{a.health}</TableCell><TableCell><Stack direction='row' spacing={1}><Button size='small' onClick={() => sync(a.id)}>Sync</Button><Button size='small' onClick={() => drift(a.id)}>Drift</Button><Button size='small' onClick={() => rollback(a.id)}>Rollback</Button></Stack></TableCell></TableRow>)}</TableBody></Table></Paper></Stack>
}

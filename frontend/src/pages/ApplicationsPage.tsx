import { useEffect, useState } from 'react'
import { Button, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography } from '@mui/material'

type App = { id: string; name: string; project: string; syncStatus: string; health: string }

export function ApplicationsPage() {
  const [apps, setApps] = useState<App[]>([])
  const [rollbackRevision, setRollbackRevision] = useState('main')

  const load = () => fetch('/api/v1/applications').then(r => r.json()).then(setApps).catch(() => setApps([]))
  useEffect(() => { load() }, [])

  const sync = async (id: string) => {
    await fetch(`/api/v1/applications/${id}/sync`, { method: 'POST' })
    load()
  }

  const drift = async (id: string) => {
    const res = await fetch(`/api/v1/applications/${id}/drift`)
    const data = await res.json()
    alert(`Drift: ${data.status}`)
  }

  const rollback = async (id: string) => {
    await fetch(`/api/v1/applications/${id}/rollback`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ revision: rollbackRevision })
    })
    load()
  }

  return <Stack spacing={2}><Typography variant='h5'>Applications</Typography><Button variant='contained'>Create Application</Button><TextField label='Rollback Revision' value={rollbackRevision} onChange={e => setRollbackRevision(e.target.value)} /><Paper><Table><TableHead><TableRow><TableCell>Name</TableCell><TableCell>Project</TableCell><TableCell>Sync</TableCell><TableCell>Health</TableCell><TableCell>Actions</TableCell></TableRow></TableHead><TableBody>{apps.map(a => <TableRow key={a.id}><TableCell>{a.name}</TableCell><TableCell>{a.project}</TableCell><TableCell>{a.syncStatus}</TableCell><TableCell>{a.health}</TableCell><TableCell><Stack direction='row' spacing={1}><Button size='small' onClick={() => sync(a.id)}>Sync</Button><Button size='small' onClick={() => drift(a.id)}>Drift</Button><Button size='small' onClick={() => rollback(a.id)}>Rollback</Button></Stack></TableCell></TableRow>)}</TableBody></Table></Paper></Stack>
}

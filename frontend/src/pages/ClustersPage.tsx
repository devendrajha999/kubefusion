import { useEffect, useState } from 'react'
import { Alert, Paper, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material'
import { api } from '../lib/api'

type Cluster = { name: string; status: string; server: string }

export function ClustersPage() {
  const [clusters, setClusters] = useState<Cluster[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    api<Cluster[]>('/api/v1/clusters').then(setClusters).catch((e: unknown) => {
      setError(e instanceof Error ? e.message : 'Failed to load clusters')
      setClusters([])
    })
  }, [])

  return <>{error && <Alert severity='error' sx={{ mb: 2 }}>{error}</Alert>}<Typography variant='h5' sx={{ mb: 2 }}>Clusters</Typography><Paper><Table><TableHead><TableRow><TableCell>Name</TableCell><TableCell>Status</TableCell><TableCell>Server</TableCell></TableRow></TableHead><TableBody>{clusters.map(c => <TableRow key={c.name}><TableCell>{c.name}</TableCell><TableCell>{c.status}</TableCell><TableCell>{c.server}</TableCell></TableRow>)}</TableBody></Table></Paper></>
}


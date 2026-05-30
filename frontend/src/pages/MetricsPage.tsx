import { useEffect, useState } from 'react'
import { Alert, Card, CardContent, Grid2, Typography } from '@mui/material'
import { api } from '../lib/api'

export function MetricsPage() {
  const [data, setData] = useState<Record<string, number>>({})
  const [error, setError] = useState('')

  useEffect(() => {
    api<Record<string, number>>('/api/v1/metrics/overview').then(setData).catch((e: unknown) => setError(e instanceof Error ? e.message : 'Failed to load metrics'))
  }, [])

  const cards = [
    { key: 'clusterCpu', title: 'Cluster CPU', fmt: (v: number) => v.toFixed(3) + ' cores/s' },
    { key: 'clusterMemory', title: 'Cluster Memory', fmt: (v: number) => (v / (1024 * 1024 * 1024)).toFixed(2) + ' GiB' },
    { key: 'podRestarts', title: 'Pod Restarts', fmt: (v: number) => Math.round(v).toString() },
    { key: 'nodeReady', title: 'Ready Nodes', fmt: (v: number) => Math.round(v).toString() },
  ]

  return <>
    {error && <Alert severity='warning' sx={{ mb: 2 }}>{error}</Alert>}
    <Grid2 container spacing={2}>
      {cards.map(c => <Grid2 size={{ xs: 12, md: 6 }} key={c.key}><Card><CardContent><Typography variant='h6'>{c.title}</Typography><Typography variant='h4'>{c.fmt(data[c.key] || 0)}</Typography></CardContent></Card></Grid2>)}
    </Grid2>
  </>
}


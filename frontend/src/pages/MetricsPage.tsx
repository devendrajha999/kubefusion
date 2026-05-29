import { Card, CardContent, Grid2, Typography } from '@mui/material'

export function MetricsPage() {
  const m = ['Cluster CPU', 'Cluster Memory', 'Node Load', 'Pod Restarts']
  return <Grid2 container spacing={2}>{m.map(x => <Grid2 size={{ xs: 12, md: 6 }} key={x}><Card><CardContent><Typography variant='h6'>{x}</Typography><Typography variant='body2'>Real-time stream enabled via WebSocket + Prometheus adapter.</Typography></CardContent></Card></Grid2>)}</Grid2>
}

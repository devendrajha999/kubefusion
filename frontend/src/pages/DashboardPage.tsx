import { Card, CardContent, Grid2, Typography } from '@mui/material'

export function DashboardPage() {
  const cards = [{ t: 'Applications', v: '12' }, { t: 'Clusters', v: '3' }, { t: 'Healthy', v: '92%' }, { t: 'Drift Alerts', v: '2' }]
  return <Grid2 container spacing={2}>{cards.map(c => <Grid2 size={{ xs: 12, md: 3 }} key={c.t}><Card><CardContent><Typography variant='h6'>{c.t}</Typography><Typography variant='h4'>{c.v}</Typography></CardContent></Card></Grid2>)}</Grid2>
}

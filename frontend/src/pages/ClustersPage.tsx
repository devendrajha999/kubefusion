import { Paper, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material'

const clusters = [{ name: 'in-cluster', status: 'Healthy', server: 'https://kubernetes.default.svc' }]

export function ClustersPage() {
  return <><Typography variant='h5' sx={{ mb: 2 }}>Clusters</Typography><Paper><Table><TableHead><TableRow><TableCell>Name</TableCell><TableCell>Status</TableCell><TableCell>Server</TableCell></TableRow></TableHead><TableBody>{clusters.map(c => <TableRow key={c.name}><TableCell>{c.name}</TableCell><TableCell>{c.status}</TableCell><TableCell>{c.server}</TableCell></TableRow>)}</TableBody></Table></Paper></>
}

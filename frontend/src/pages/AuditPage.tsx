import { useEffect, useState } from 'react'
import { Paper, Stack, Typography } from '@mui/material'

type Event = { id: string; actor: string; action: string; target: string; createdAt: string }

export function AuditPage() {
  const [events, setEvents] = useState<Event[]>([])
  useEffect(() => { fetch('/api/v1/audit/events').then(r => r.json()).then(setEvents).catch(() => setEvents([])) }, [])
  return <Stack spacing={2}><Typography variant='h5'>Audit Events</Typography>{events.map(e => <Paper key={e.id} sx={{ p: 2 }}><Typography variant='body2'>{e.createdAt} | {e.actor} | {e.action} | {e.target}</Typography></Paper>)}</Stack>
}

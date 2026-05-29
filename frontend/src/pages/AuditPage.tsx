import { useEffect, useState } from 'react'
import { Alert, Paper, Stack, Typography } from '@mui/material'
import { api } from '../lib/api'

type Event = { id: string; actor: string; action: string; target: string; createdAt: string }

export function AuditPage() {
  const [events, setEvents] = useState<Event[]>([])
  const [error, setError] = useState('')
  useEffect(() => {
    api<Event[]>('/api/v1/audit/events').then(setEvents).catch(e => {
      setError(e instanceof Error ? e.message : 'Failed to load audit events')
      setEvents([])
    })
  }, [])
  return <Stack spacing={2}><Typography variant='h5'>Audit Events</Typography>{error && <Alert severity='error'>{error}</Alert>}{events.map(e => <Paper key={e.id} sx={{ p: 2 }}><Typography variant='body2'>{e.createdAt} | {e.actor} | {e.action} | {e.target}</Typography></Paper>)}</Stack>
}

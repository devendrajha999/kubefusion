import { useEffect, useMemo, useRef, useState } from 'react'
import { Alert, Box, Button, FormControl, InputLabel, List, ListItemButton, ListItemText, MenuItem, Paper, Select, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography } from '@mui/material'
import { api, getToken } from '../lib/api'

type NamespaceItem = { name: string; status: string }
type Row = Record<string, string | number | boolean | null | undefined>

type KindDef = { id: string; label: string }

const KINDS: KindDef[] = [
  { id: 'pods', label: 'Pods' },
  { id: 'deployments', label: 'Deployments' },
  { id: 'statefulsets', label: 'StatefulSets' },
  { id: 'daemonsets', label: 'DaemonSets' },
  { id: 'replicasets', label: 'ReplicaSets' },
  { id: 'jobs', label: 'Jobs' },
  { id: 'cronjobs', label: 'CronJobs' },
  { id: 'services', label: 'Services' },
  { id: 'ingresses', label: 'Ingresses' },
  { id: 'configmaps', label: 'ConfigMaps' },
  { id: 'secrets', label: 'Secrets' },
  { id: 'persistentvolumeclaims', label: 'PersistentVolumeClaims' },
  { id: 'persistentvolumes', label: 'PersistentVolumes' },
  { id: 'storageclasses', label: 'StorageClasses' },
  { id: 'namespaces', label: 'Namespaces' },
  { id: 'events', label: 'Events' },
  { id: 'nodes', label: 'Nodes' },
]

const podKinds = new Set(['pods'])

export function PodsPage() {
  const [kind, setKind] = useState('pods')
  const [namespaces, setNamespaces] = useState<NamespaceItem[]>([])
  const [namespace, setNamespace] = useState('')
  const [query, setQuery] = useState('')
  const [rows, setRows] = useState<Row[]>([])
  const [error, setError] = useState('')

  const [targetPod, setTargetPod] = useState('')
  const [container, setContainer] = useState('')
  const [logs, setLogs] = useState<string[]>([])
  const [command, setCommand] = useState('ls -la')
  const [execOut, setExecOut] = useState('')
  const evt = useRef<EventSource | null>(null)

  const loadNamespaces = async () => {
    try { setNamespaces(await api<NamespaceItem[]>('/api/v1/clusters/in-cluster/namespaces')) } catch { setNamespaces([]) }
  }

  const loadRows = async () => {
    try {
      setError('')
      const nsQuery = namespace && !['nodes', 'persistentvolumes', 'storageclasses', 'namespaces'].includes(kind) ? `?namespace=${encodeURIComponent(namespace)}` : ''
      const data = await api<Row[]>(`/api/v1/clusters/in-cluster/resources/${kind}${nsQuery}`)
      setRows(data)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Load failed')
      setRows([])
    }
  }

  useEffect(() => { loadNamespaces() }, [])
  useEffect(() => { loadRows() }, [kind, namespace])

  const filtered = useMemo(() => rows.filter(r => JSON.stringify(r).toLowerCase().includes(query.toLowerCase())), [rows, query])
  const columns = useMemo(() => {
    const keys = new Set<string>()
    filtered.forEach(r => Object.keys(r).forEach(k => keys.add(k)))
    return Array.from(keys)
  }, [filtered])

  const fetchLogs = async () => {
    const ns = namespace || String(filtered.find(r => String(r.name) === targetPod)?.namespace || 'default')
    const data = await api<{ lines: string[] }>('/api/v1/clusters/in-cluster/pods/logs', { method: 'POST', body: JSON.stringify({ cluster: 'in-cluster', namespace: ns, pod: targetPod, container, tailLines: 200 }) })
    setLogs(data.lines || [])
  }

  const startStream = () => {
    if (evt.current) evt.current.close()
    const token = encodeURIComponent(getToken())
    const ns = encodeURIComponent(namespace || 'default')
    const es = new EventSource(`/api/v1/clusters/in-cluster/pods/logs/stream?namespace=${ns}&pod=${encodeURIComponent(targetPod)}&container=${encodeURIComponent(container)}&token=${token}`)
    es.onmessage = (e) => setLogs(prev => [...prev.slice(-500), e.data])
    es.onerror = () => es.close()
    evt.current = es
  }

  const runExec = async () => {
    const data = await api<{ stdout?: string; stderr?: string; error?: string }>('/api/v1/clusters/in-cluster/pods/exec', { method: 'POST', body: JSON.stringify({ namespace: namespace || 'default', pod: targetPod, container, command: ['/bin/sh', '-c', command] }) })
    setExecOut((data.stdout || '') + (data.stderr ? '\nERR:\n' + data.stderr : '') + (data.error ? '\nERROR: ' + data.error : ''))
  }

  return <Stack spacing={2}>
    <Typography variant='h5'>Cluster Navigator</Typography>
    {error && <Alert severity='error'>{error}</Alert>}
    <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '260px 1fr' }, gap: 2 }}>
      <Paper sx={{ p: 1, maxHeight: 600, overflow: 'auto' }}>
        <Typography variant='subtitle1' sx={{ px: 1, py: 1 }}>Navigator</Typography>
        <List dense>
          {KINDS.map(k => <ListItemButton key={k.id} selected={kind === k.id} onClick={() => setKind(k.id)}><ListItemText primary={k.label} /></ListItemButton>)}
        </List>
      </Paper>
      <Stack spacing={2}>
        <Stack direction='row' spacing={2}>
          <FormControl sx={{ minWidth: 240 }}>
            <InputLabel id='ns-label'>Namespace</InputLabel>
            <Select labelId='ns-label' label='Namespace' value={namespace} onChange={e => setNamespace(e.target.value)}>
              <MenuItem value=''>All namespaces</MenuItem>
              {namespaces.map(n => <MenuItem key={n.name} value={n.name}>{n.name}</MenuItem>)}
            </Select>
          </FormControl>
          <TextField label={`Search ${kind}`} value={query} onChange={e => setQuery(e.target.value)} sx={{ minWidth: 280 }} />
          <Button variant='outlined' onClick={loadRows}>Refresh</Button>
        </Stack>
        <Paper>
          <Table size='small'>
            <TableHead><TableRow>{columns.map(c => <TableCell key={c}>{c}</TableCell>)}</TableRow></TableHead>
            <TableBody>
              {filtered.map((r, i) => <TableRow key={i} onClick={() => { if (kind === 'pods' && typeof r.name === 'string') { setTargetPod(r.name); if (typeof r.namespace === 'string') setNamespace(r.namespace) } }}>{columns.map(c => <TableCell key={c}>{String(r[c] ?? '')}</TableCell>)}</TableRow>)}
            </TableBody>
          </Table>
        </Paper>
      </Stack>
    </Box>

    {podKinds.has(kind) && (
      <>
        <Paper sx={{ p: 2 }}><Stack spacing={2}><Typography variant='h6'>Pod Logs</Typography><TextField label='Pod Name' value={targetPod} onChange={e => setTargetPod(e.target.value)} /><TextField label='Container' value={container} onChange={e => setContainer(e.target.value)} /><Stack direction='row' spacing={1}><Button variant='contained' onClick={fetchLogs}>Fetch</Button><Button onClick={startStream}>Start Stream</Button></Stack><pre>{logs.join('\n')}</pre></Stack></Paper>
        <Paper sx={{ p: 2 }}><Stack spacing={2}><Typography variant='h6'>Pod Exec</Typography><TextField label='Shell Command' value={command} onChange={e => setCommand(e.target.value)} /><Button variant='contained' onClick={runExec}>Run Command</Button><pre>{execOut}</pre></Stack></Paper>
      </>
    )}
  </Stack>
}

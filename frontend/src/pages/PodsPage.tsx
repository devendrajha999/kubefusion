import { useEffect, useMemo, useRef, useState } from 'react'
import { Alert, Box, Button, Checkbox, Collapse, Dialog, DialogContent, DialogTitle, FormControl, InputLabel, List, ListItemButton, ListItemText, MenuItem, Paper, Select, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography } from '@mui/material'
import { api, getToken } from '../lib/api'

type NamespaceItem = { name: string; status: string }
type Row = Record<string, string | number | boolean | null | undefined>

type KindDef = { id: string; label: string }
type Group = { id: string; label: string; children?: KindDef[]; kind?: string }
const GROUPS: Group[] = [
  { id: 'overview', label: 'Overview', kind: 'pods' },
  { id: 'applications', label: 'Applications', kind: 'deployments' },
  { id: 'nodes', label: 'Nodes', kind: 'nodes' },
  {
    id: 'workloads', label: 'Workloads', children: [
      { id: 'pods', label: 'Pods' },
      { id: 'deployments', label: 'Deployments' },
      { id: 'daemonsets', label: 'Daemon Sets' },
      { id: 'statefulsets', label: 'Stateful Sets' },
      { id: 'replicasets', label: 'Replica Sets' },
      { id: 'jobs', label: 'Jobs' },
      { id: 'cronjobs', label: 'Cron Jobs' },
    ]
  },
  {
    id: 'config', label: 'Config', children: [
      { id: 'configmaps', label: 'Config Maps' },
      { id: 'secrets', label: 'Secrets' },
    ]
  },
  {
    id: 'network', label: 'Network', children: [
      { id: 'services', label: 'Services' },
      { id: 'ingresses', label: 'Ingresses' },
    ]
  },
  {
    id: 'storage', label: 'Storage', children: [
      { id: 'persistentvolumeclaims', label: 'Persistent Volume Claims' },
      { id: 'persistentvolumes', label: 'Persistent Volumes' },
      { id: 'storageclasses', label: 'Storage Classes' },
    ]
  },
  { id: 'namespaces', label: 'Namespaces', kind: 'namespaces' },
  { id: 'events', label: 'Events', kind: 'events' },
  { id: 'helm', label: 'Helm', kind: 'deployments' },
  { id: 'access-control', label: 'Access Control', kind: 'nodes' },
  { id: 'custom-resources', label: 'Custom Resources', kind: 'events' },
]

const podKinds = new Set(['pods'])

export function PodsPage() {
  const [kind, setKind] = useState('pods')
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({
    workloads: true, config: true, network: true, storage: true,
  })
  const [namespaces, setNamespaces] = useState<NamespaceItem[]>([])
  const [selectedNamespaces, setSelectedNamespaces] = useState<string[]>([])
  const [query, setQuery] = useState('')
  const [rows, setRows] = useState<Row[]>([])
  const [error, setError] = useState('')

  const [targetPod, setTargetPod] = useState('')
  const [container, setContainer] = useState('')
  const [logs, setLogs] = useState<string[]>([])
  const [command, setCommand] = useState('ls -la')
  const [execOut, setExecOut] = useState('')
  const [detailsOpen, setDetailsOpen] = useState(false)
  const [detailsText, setDetailsText] = useState('')
  const evt = useRef<EventSource | null>(null)

  const loadNamespaces = async () => {
    try { setNamespaces(await api<NamespaceItem[]>('/api/v1/clusters/in-cluster/namespaces')) } catch { setNamespaces([]) }
  }

  const loadRows = async () => {
    try {
      setError('')
      const data = await api<Row[]>(`/api/v1/clusters/in-cluster/resources/${kind}`)
      setRows(data)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Load failed')
      setRows([])
    }
  }

  useEffect(() => { loadNamespaces() }, [])
  useEffect(() => { loadRows() }, [kind])

  const filtered = useMemo(() => rows.filter(r => {
    const matchText = JSON.stringify(r).toLowerCase().includes(query.toLowerCase())
    const ns = typeof r.namespace === 'string' ? r.namespace : ''
    const matchNs = selectedNamespaces.length === 0 || selectedNamespaces.includes(ns)
    return matchText && matchNs
  }), [rows, query, selectedNamespaces])
  const columns = useMemo(() => {
    const keys = new Set<string>()
    filtered.forEach(r => Object.keys(r).forEach(k => keys.add(k)))
    const ordered = ['name', 'namespace', 'status', 'containers', 'cpu', 'memory', 'restarts', 'controlledBy', 'node', 'qos', 'age', 'replicas', 'ready', 'updated', 'type', 'host', 'reason', 'message']
    const all = Array.from(keys)
    return [...ordered.filter(k => keys.has(k)), ...all.filter(k => !ordered.includes(k))]
  }, [filtered])

  const fetchLogs = async () => {
    const ns = String(filtered.find(r => String(r.name) === targetPod)?.namespace || selectedNamespaces[0] || 'default')
    const data = await api<{ lines: string[] }>('/api/v1/clusters/in-cluster/pods/logs', { method: 'POST', body: JSON.stringify({ cluster: 'in-cluster', namespace: ns, pod: targetPod, container, tailLines: 200 }) })
    setLogs(data.lines || [])
  }

  const startStream = () => {
    if (evt.current) evt.current.close()
    const token = encodeURIComponent(getToken())
    const ns = encodeURIComponent(String(filtered.find(r => String(r.name) === targetPod)?.namespace || selectedNamespaces[0] || 'default'))
    const es = new EventSource(`/api/v1/clusters/in-cluster/pods/logs/stream?namespace=${ns}&pod=${encodeURIComponent(targetPod)}&container=${encodeURIComponent(container)}&token=${token}`)
    es.onmessage = (e) => setLogs(prev => [...prev.slice(-500), e.data])
    es.onerror = () => es.close()
    evt.current = es
  }

  const runExec = async () => {
    const data = await api<{ stdout?: string; stderr?: string; error?: string }>('/api/v1/clusters/in-cluster/pods/exec', { method: 'POST', body: JSON.stringify({ namespace: selectedNamespaces[0] || 'default', pod: targetPod, container, command: ['/bin/sh', '-c', command] }) })
    setExecOut((data.stdout || '') + (data.stderr ? '\nERR:\n' + data.stderr : '') + (data.error ? '\nERROR: ' + data.error : ''))
  }

  const openDetails = async (ns: string, pod: string) => {
    const data = await api<Record<string, unknown>>(`/api/v1/clusters/in-cluster/pods/${encodeURIComponent(ns)}/${encodeURIComponent(pod)}`)
    setDetailsText(JSON.stringify(data, null, 2))
    setDetailsOpen(true)
  }

  const deletePod = async (ns: string, pod: string) => {
    if (!confirm(`Delete pod ${ns}/${pod}?`)) return
    await api('/api/v1/clusters/in-cluster/pods/' + encodeURIComponent(ns) + '/' + encodeURIComponent(pod), { method: 'DELETE' })
    await loadRows()
  }

  return <Stack spacing={2}>
    <Typography variant='h5'>Cluster Navigator</Typography>
    {error && <Alert severity='error'>{error}</Alert>}
    <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '260px 1fr' }, gap: 2 }}>
      <Paper sx={{ p: 1, maxHeight: 600, overflow: 'auto' }}>
        <Typography variant='subtitle1' sx={{ px: 1, py: 1 }}>Navigator</Typography>
        <List dense>
          {GROUPS.map(g => g.children ? (
            <Box key={g.id}>
              <ListItemButton onClick={() => setOpenGroups(prev => ({ ...prev, [g.id]: !prev[g.id] }))}>
                <ListItemText primary={g.label} />
              </ListItemButton>
              <Collapse in={!!openGroups[g.id]}>
                <List dense sx={{ pl: 2 }}>
                  {g.children.map(c => <ListItemButton key={c.id} selected={kind === c.id} onClick={() => setKind(c.id)}><ListItemText primary={c.label} /></ListItemButton>)}
                </List>
              </Collapse>
            </Box>
          ) : (
            <ListItemButton key={g.id} selected={g.kind === kind} onClick={() => g.kind && setKind(g.kind)}>
              <ListItemText primary={g.label} />
            </ListItemButton>
          ))}
        </List>
      </Paper>
      <Stack spacing={2}>
        <Stack direction='row' spacing={2}>
          <FormControl sx={{ minWidth: 240 }}>
            <InputLabel id='ns-label'>Namespaces</InputLabel>
            <Select
              multiple
              labelId='ns-label'
              label='Namespaces'
              value={selectedNamespaces}
              renderValue={(v) => (v as string[]).length ? (v as string[]).join(', ') : 'All namespaces'}
              onChange={e => setSelectedNamespaces(typeof e.target.value === 'string' ? e.target.value.split(',') : e.target.value)}
            >
              {namespaces.map(n => <MenuItem key={n.name} value={n.name}><Checkbox checked={selectedNamespaces.includes(n.name)} /><ListItemText primary={n.name} /></MenuItem>)}
            </Select>
          </FormControl>
          <TextField label={`Search ${kind}`} value={query} onChange={e => setQuery(e.target.value)} sx={{ minWidth: 280 }} />
          <Button variant='outlined' onClick={loadRows}>Refresh</Button>
        </Stack>
        <Paper>
          <Table size='small'>
            <TableHead><TableRow>{columns.map(c => <TableCell key={c}>{c}</TableCell>)}{kind === 'pods' && <TableCell>Actions</TableCell>}</TableRow></TableHead>
            <TableBody>
              {filtered.map((r, i) => <TableRow key={i} onClick={() => { if (kind === 'pods' && typeof r.name === 'string') { setTargetPod(r.name) } }}>{columns.map(c => <TableCell key={c}>{String(r[c] ?? '')}</TableCell>)}{kind === 'pods' && <TableCell><Stack direction='row' spacing={1}><Button size='small' onClick={() => { setTargetPod(String(r.name || '')); fetchLogs() }}>Logs</Button><Button size='small' onClick={() => { setTargetPod(String(r.name || '')); runExec() }}>Exec</Button><Button size='small' onClick={() => openDetails(String(r.namespace || ''), String(r.name || ''))}>View</Button><Button size='small' color='error' onClick={() => deletePod(String(r.namespace || ''), String(r.name || ''))}>Delete</Button></Stack></TableCell>}</TableRow>)}
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
    <Dialog open={detailsOpen} onClose={() => setDetailsOpen(false)} fullWidth maxWidth='md'>
      <DialogTitle>Pod Details</DialogTitle>
      <DialogContent><pre>{detailsText}</pre></DialogContent>
    </Dialog>
  </Stack>
}

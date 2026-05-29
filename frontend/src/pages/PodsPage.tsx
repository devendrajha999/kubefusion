import { useEffect, useRef, useState } from 'react'
import { Button, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography } from '@mui/material'

type Pod = { namespace: string; name: string; status: string; restarts: number }

export function PodsPage() {
  const [pods, setPods] = useState<Pod[]>([])
  const [namespace, setNamespace] = useState('default')
  const [targetPod, setTargetPod] = useState('')
  const [container, setContainer] = useState('')
  const [logs, setLogs] = useState<string[]>([])
  const [command, setCommand] = useState('ls -la')
  const [execOut, setExecOut] = useState('')
  const evt = useRef<EventSource | null>(null)

  const loadPods = () => fetch('/api/v1/clusters/in-cluster/pods?namespace=' + namespace).then(r => r.json()).then(setPods).catch(() => setPods([]))
  useEffect(() => { loadPods() }, [namespace])

  const fetchLogs = async () => {
    const res = await fetch('/api/v1/clusters/in-cluster/pods/logs', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ cluster: 'in-cluster', namespace, pod: targetPod, container, tailLines: 200 }) })
    const data = await res.json()
    setLogs(data.lines || [])
  }

  const startStream = () => {
    if (evt.current) evt.current.close()
    const es = new EventSource(`/api/v1/clusters/in-cluster/pods/logs/stream?namespace=${encodeURIComponent(namespace)}&pod=${encodeURIComponent(targetPod)}&container=${encodeURIComponent(container)}`)
    es.onmessage = (e) => setLogs(prev => [...prev.slice(-500), e.data])
    es.onerror = () => es.close()
    evt.current = es
  }

  const stopStream = () => { evt.current?.close(); evt.current = null }

  const runExec = async () => {
    const res = await fetch('/api/v1/clusters/in-cluster/pods/exec', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ namespace, pod: targetPod, container, command: ['/bin/sh', '-c', command] }) })
    const data = await res.json()
    setExecOut((data.stdout || '') + (data.stderr ? '\nERR:\n' + data.stderr : '') + (data.error ? '\nERROR: ' + data.error : ''))
  }

  return <Stack spacing={2}><Typography variant='h5'>Pods</Typography>
    <Stack direction='row' spacing={2}><TextField label='Namespace' value={namespace} onChange={e => setNamespace(e.target.value)} /><Button variant='outlined' onClick={loadPods}>Reload</Button></Stack>
    <Paper><Table><TableHead><TableRow><TableCell>Namespace</TableCell><TableCell>Name</TableCell><TableCell>Status</TableCell><TableCell>Restarts</TableCell></TableRow></TableHead><TableBody>{pods.map(p => <TableRow key={p.name} onClick={() => { setTargetPod(p.name); setNamespace(p.namespace) }}><TableCell>{p.namespace}</TableCell><TableCell>{p.name}</TableCell><TableCell>{p.status}</TableCell><TableCell>{p.restarts}</TableCell></TableRow>)}</TableBody></Table></Paper>
    <Paper sx={{ p: 2 }}><Stack spacing={2}><Typography variant='h6'>Logs</Typography><TextField label='Pod Name' value={targetPod} onChange={e => setTargetPod(e.target.value)} /><TextField label='Container' value={container} onChange={e => setContainer(e.target.value)} /><Stack direction='row' spacing={1}><Button variant='contained' onClick={fetchLogs}>Fetch</Button><Button onClick={startStream}>Start Stream</Button><Button onClick={stopStream}>Stop Stream</Button></Stack><pre>{logs.join('\n')}</pre></Stack></Paper>
    <Paper sx={{ p: 2 }}><Stack spacing={2}><Typography variant='h6'>Exec</Typography><TextField label='Shell Command' value={command} onChange={e => setCommand(e.target.value)} /><Button variant='contained' onClick={runExec}>Run Command</Button><pre>{execOut}</pre></Stack></Paper>
  </Stack>
}

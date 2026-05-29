import { useState } from 'react'
import { Alert, Box, Button, Paper, Stack, TextField, Typography } from '@mui/material'
import { api, setToken } from '../lib/api'

type LoginResponse = { token: string; role: string }

export function LoginPage({ onLogin }: { onLogin: () => void }) {
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('admin')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async () => {
    setLoading(true)
    setError('')
    try {
      const data = await api<LoginResponse>('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      })
      setToken(data.token)
      onLogin()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box sx={{ minHeight: '100vh', display: 'grid', placeItems: 'center', p: 2 }}>
      <Paper sx={{ width: 420, p: 3 }}>
        <Stack spacing={2}>
          <Typography variant='h4'>KubeFusion Login</Typography>
          <TextField label='Username' value={username} onChange={e => setUsername(e.target.value)} />
          <TextField label='Password' type='password' value={password} onChange={e => setPassword(e.target.value)} />
          {error && <Alert severity='error'>{error}</Alert>}
          <Button variant='contained' disabled={loading} onClick={submit}>Sign In</Button>
        </Stack>
      </Paper>
    </Box>
  )
}

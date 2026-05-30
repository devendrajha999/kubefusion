import { useMemo, useState } from 'react'
import { AppBar, Box, Button, CssBaseline, ThemeProvider, Toolbar, Typography, createTheme } from '@mui/material'
import { Route, Routes } from 'react-router-dom'
import { DashboardPage } from '../pages/DashboardPage'
import { ApplicationsPage } from '../pages/ApplicationsPage'
import { ClustersPage } from '../pages/ClustersPage'
import { MetricsPage } from '../pages/MetricsPage'
import { RepositoriesPage } from '../pages/RepositoriesPage'
import { PodsPage } from '../pages/PodsPage'
import { AuditPage } from '../pages/AuditPage'
import { LoginPage } from '../pages/LoginPage'
import { clearToken, getToken } from '../lib/api'

export function App() {
  const [mode, setMode] = useState<'light' | 'dark'>('dark')
  const [token, setAuthToken] = useState(getToken())
  const theme = useMemo(() => createTheme({ palette: { mode } }), [mode])

  if (!token) return <ThemeProvider theme={theme}><CssBaseline /><LoginPage onLogin={() => setAuthToken(getToken())} /></ThemeProvider>

  return <ThemeProvider theme={theme}><CssBaseline />
    <Box sx={{ display: 'flex' }}>
      <AppBar position='fixed'><Toolbar>
        <Typography variant='h6' sx={{ flexGrow: 1 }}>KubeFusion</Typography>
        <Button color='inherit' onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}>Theme</Button>
        <Button color='inherit' onClick={() => { clearToken(); setAuthToken('') }}>Logout</Button>
      </Toolbar></AppBar>
      <Box component='main' sx={{ flexGrow: 1, p: 3, mt: 8 }}>
        <Routes>
          <Route path='/' element={<DashboardPage />} />
          <Route path='/applications' element={<ApplicationsPage />} />
          <Route path='/repositories' element={<RepositoriesPage />} />
          <Route path='/clusters' element={<ClustersPage />} />
          <Route path='/pods' element={<PodsPage />} />
          <Route path='/metrics' element={<MetricsPage />} />
          <Route path='/audit' element={<AuditPage />} />
        </Routes>
      </Box>
    </Box>
  </ThemeProvider>
}

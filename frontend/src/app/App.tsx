import { useMemo, useState } from 'react'
import { AppBar, Box, Button, CssBaseline, Drawer, IconButton, List, ListItemButton, ListItemText, ThemeProvider, Toolbar, Typography, createTheme } from '@mui/material'
import MenuIcon from '@mui/icons-material/Menu'
import { Link, Route, Routes } from 'react-router-dom'
import { DashboardPage } from '../pages/DashboardPage'
import { ApplicationsPage } from '../pages/ApplicationsPage'
import { ClustersPage } from '../pages/ClustersPage'
import { MetricsPage } from '../pages/MetricsPage'
import { RepositoriesPage } from '../pages/RepositoriesPage'
import { PodsPage } from '../pages/PodsPage'
import { AuditPage } from '../pages/AuditPage'
import { LoginPage } from '../pages/LoginPage'
import { clearToken, getToken } from '../lib/api'

const nav = ['Dashboard', 'Applications', 'Clusters', 'Workloads', 'Repositories', 'Metrics', 'Audit']

export function App() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const [mode, setMode] = useState<'light' | 'dark'>('dark')
  const [token, setAuthToken] = useState(getToken())
  const theme = useMemo(() => createTheme({ palette: { mode } }), [mode])

  if (!token) return <ThemeProvider theme={theme}><CssBaseline /><LoginPage onLogin={() => setAuthToken(getToken())} /></ThemeProvider>

  return <ThemeProvider theme={theme}><CssBaseline />
    <Box sx={{ display: 'flex' }}>
      <AppBar position='fixed'><Toolbar>
        <IconButton color='inherit' edge='start' onClick={() => setMobileOpen(!mobileOpen)} sx={{ mr: 2 }}><MenuIcon /></IconButton>
        <Typography variant='h6' sx={{ flexGrow: 1 }}>KubeFusion Navigator</Typography>
        <Button color='inherit' onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}>Theme</Button>
        <Button color='inherit' onClick={() => { clearToken(); setAuthToken('') }}>Logout</Button>
      </Toolbar></AppBar>
      <Drawer variant='permanent' sx={{ width: 260, display: { xs: 'none', md: 'block' }, '& .MuiDrawer-paper': { width: 260, boxSizing: 'border-box', mt: 8 } }}>
        <List>
          {nav.map(n => <ListItemButton key={n} component={Link} to={n === 'Dashboard' ? '/' : '/' + (n === 'Workloads' ? 'pods' : n.toLowerCase())}><ListItemText primary={n} /></ListItemButton>)}
        </List>
      </Drawer>
      <Drawer open={mobileOpen} onClose={() => setMobileOpen(false)} sx={{ display: { md: 'none' } }}>
        <List sx={{ width: 260 }}>
          {nav.map(n => <ListItemButton key={n} component={Link} to={n === 'Dashboard' ? '/' : '/' + (n === 'Workloads' ? 'pods' : n.toLowerCase())} onClick={() => setMobileOpen(false)}><ListItemText primary={n} /></ListItemButton>)}
        </List>
      </Drawer>
      <Box component='main' sx={{ flexGrow: 1, p: 3, mt: 8, ml: { md: '260px' } }}>
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

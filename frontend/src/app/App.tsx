import { useMemo, useState } from 'react'
import { AppBar, Box, CssBaseline, Drawer, IconButton, List, ListItemButton, ListItemText, ThemeProvider, Toolbar, Typography, createTheme } from '@mui/material'
import MenuIcon from '@mui/icons-material/Menu'
import { Link, Route, Routes } from 'react-router-dom'
import { DashboardPage } from '../pages/DashboardPage'
import { ApplicationsPage } from '../pages/ApplicationsPage'
import { ClustersPage } from '../pages/ClustersPage'
import { MetricsPage } from '../pages/MetricsPage'
import { RepositoriesPage } from '../pages/RepositoriesPage'
import { PodsPage } from '../pages/PodsPage'
import { AuditPage } from '../pages/AuditPage'

const nav = ['Dashboard', 'Applications', 'Repositories', 'Clusters', 'Pods', 'Metrics', 'Audit']

export function App() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const [mode, setMode] = useState<'light' | 'dark'>('dark')
  const theme = useMemo(() => createTheme({ palette: { mode } }), [mode])

  return <ThemeProvider theme={theme}><CssBaseline />
    <Box sx={{ display: 'flex' }}>
      <AppBar position='fixed'><Toolbar>
        <IconButton color='inherit' edge='start' onClick={() => setMobileOpen(!mobileOpen)} sx={{ mr: 2 }}><MenuIcon /></IconButton>
        <Typography variant='h6' sx={{ flexGrow: 1 }}>KubeFusion</Typography>
        <button onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}>Theme</button>
      </Toolbar></AppBar>
      <Drawer open={mobileOpen} onClose={() => setMobileOpen(false)}>
        <List sx={{ width: 260 }}>
          {nav.map(n => <ListItemButton key={n} component={Link} to={n === 'Dashboard' ? '/' : '/' + n.toLowerCase()}><ListItemText primary={n} /></ListItemButton>)}
        </List>
      </Drawer>
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

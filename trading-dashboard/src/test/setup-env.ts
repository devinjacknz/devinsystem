import { vi } from 'vitest'

const env = {
  VITE_API_URL: 'http://localhost:8080',
  MODE: 'test'
}

vi.stubGlobal('import.meta', { env })

const mockImportMeta = {
  env: {
    VITE_API_URL: 'http://localhost:8080',
    MODE: 'test'
  }
};

if (typeof window !== 'undefined') {
  window.import = { meta: mockImportMeta };
} else {
  global.import = { meta: mockImportMeta };
}

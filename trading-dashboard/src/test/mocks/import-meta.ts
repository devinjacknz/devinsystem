const mockImportMeta = {
  env: {
    VITE_API_URL: 'http://localhost:8080',
    MODE: 'test'
  }
};

jest.mock('import.meta', () => mockImportMeta, { virtual: true });

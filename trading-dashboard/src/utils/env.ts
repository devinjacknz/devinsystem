export const API_URL = process.env.NODE_ENV === 'test' 
  ? 'http://localhost:8080'
  : import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const getEnvVar = (key: string, defaultValue: string = ''): string => {
  if (process.env.NODE_ENV === 'test') {
    return process.env[key] || defaultValue
  }
  return import.meta.env[key] || defaultValue
}

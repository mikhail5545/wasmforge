export const getApiBaseUrl = (): string => {
  // During development, it points to the backend api path,
  // in production (while embedded in Go server) it will be the same origin,
  // so we return an empty string.
  if (process.env.NODE_ENV === 'development') {
    return 'http://localhost:8080'; 
  }

  return '';
};
import axios from 'axios';

// Dynamically determine the API base URL
// Default to the same hostname but port 8080, or use the environment variable if provided
const getApiBaseUrl = () => {
  if (import.meta.env.VITE_API_URL) {
    return import.meta.env.VITE_API_URL;
  }
  
  // If we are running in a browser, try to guess the backend location
  if (typeof window !== 'undefined') {
    const { protocol, hostname } = window.location;
    // Common development pattern: backend on 8080
    return `${protocol}//${hostname}:8080`;
  }
  
  return 'http://localhost:8080';
};

const API_BASE_URL = getApiBaseUrl();

export const login = async (username: string, password: string): Promise<string> => {
  const response = await axios.post(`${API_BASE_URL}/login`, {
    username,
    password,
  });
  return response.data.token;
};

export const getChannels = async (token: string): Promise<string[]> => {
  const response = await axios.get(`${API_BASE_URL}/api/channels`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  return response.data;
};

export const getChannelStatus = async (token: string, channelId: string): Promise<any> => {
  const response = await axios.get(`${API_BASE_URL}/api/channels/${channelId}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  return response.data;
};

export const setChannelStatus = async (token: string, channelId: string, active: boolean): Promise<any> => {
  const response = await axios.post(`${API_BASE_URL}/api/channels/${channelId}`, { active }, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  return response.data;
};

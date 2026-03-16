import axios from 'axios';

export const login = async (serverIp: string, username: string, password: string): Promise<string> => {
  const response = await axios.post(`${serverIp}/login`, {
    username,
    password,
  });
  return response.data.token;
};

export const getChannels = async (serverIp: string, token: string): Promise<string[]> => {
  const response = await axios.get(`${serverIp}/api/channels`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  return response.data;
};

export const getChannelStatus = async (serverIp: string, token: string, channelId: string): Promise<any> => {
  const response = await axios.get(`${serverIp}/api/channels/${channelId}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  return response.data;
};

export const setChannelStatus = async (serverIp: string, token: string, channelId: string, active: boolean): Promise<any> => {
  const response = await axios.post(`${serverIp}/api/channels/${channelId}`, { active }, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  return response.data;
};

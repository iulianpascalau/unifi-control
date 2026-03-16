import React, { useState, useEffect } from 'react';
import { StatusBar } from 'expo-status-bar';
import { View, StyleSheet, ActivityIndicator } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { LoginScreen } from './src/components/LoginScreen';
import { DashboardScreen } from './src/components/DashboardScreen';

export default function App() {
  const [token, setToken] = useState<string | null>(null);
  const [serverIp, setServerIp] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check for existing session token
    const loadSession = async () => {
      try {
        const savedToken = await AsyncStorage.getItem('@auth_token');
        const savedIp = await AsyncStorage.getItem('@server_ip');
        if (savedToken && savedIp) {
          setToken(savedToken);
          setServerIp(savedIp);
        }
      } catch (err) {
        console.error('Failed to load session data');
      } finally {
        setIsLoading(false);
      }
    };

    loadSession();
  }, []);

  const handleLoginSuccess = (newToken: string, ip: string) => {
    setToken(newToken);
    setServerIp(ip);
  };

  const handleLogout = async () => {
    await AsyncStorage.removeItem('@auth_token');
    setToken(null);
  };

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#00d2ff" />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <StatusBar style="light" />
      {!token || !serverIp ? (
        <LoginScreen onLoginSuccess={handleLoginSuccess} />
      ) : (
        <DashboardScreen 
          token={token} 
          serverIp={serverIp} 
          onLogout={handleLogout} 
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#0F2027',
  },
  container: {
    flex: 1,
    backgroundColor: '#0F2027',
  },
});

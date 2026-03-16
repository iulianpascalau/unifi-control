import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, ScrollView, Switch, Modal, ActivityIndicator, SafeAreaView, Platform } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { getChannels, getChannelStatus, setChannelStatus } from '../api';

interface DashboardScreenProps {
  token: string;
  serverIp: string;
  onLogout: () => void;
}

interface ChannelStatus {
  name: string;
  channel: string;
  status: boolean;
  error?: string;
}

export const DashboardScreen: React.FC<DashboardScreenProps> = ({ token, serverIp, onLogout }) => {
  const [channels, setChannels] = useState<ChannelStatus[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  
  const [modalVisible, setModalVisible] = useState(false);
  const [selectedError, setSelectedError] = useState('');
  const [selectedChannelId, setSelectedChannelId] = useState('');

  const loadData = async () => {
    try {
      const channelIds = await getChannels(serverIp, token);
      
      const statuses: ChannelStatus[] = [];
      const promises = channelIds.map((id: string) => 
        getChannelStatus(serverIp, token, id)
          .then(data => data)
          .catch(err => ({ 
            channel: id, 
            name: `Channel ${id}`, 
            status: false, 
            error: err.message || 'Unknown error fetching channel' 
          }))
      );
      
      const results = await Promise.all(promises);
      setChannels(results);
    } catch (err: any) {
      console.error(err);
      if (err.response?.status === 401) {
        onLogout(); // Token expired
      }
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [serverIp, token]);

  const toggleChannel = async (id: string, currentStatus: boolean) => {
    // Optimistic UI update
    setChannels(prev => prev.map(ch => 
      ch.channel === id ? { ...ch, status: !currentStatus, error: undefined } : ch
    ));

    try {
      await setChannelStatus(serverIp, token, id, !currentStatus);
    } catch (error: any) {
      // Revert on failure & capture error
      setChannels(prev => prev.map(ch => 
        ch.channel === id ? { ...ch, status: currentStatus, error: error.message || 'Failed to update channel status' } : ch
      ));
    }
  };

  const showErrorDetails = (channelId: string, errorMsg?: string) => {
    if (errorMsg) {
      setSelectedChannelId(channelId);
      setSelectedError(errorMsg);
      setModalVisible(true);
    }
  };

  const renderChannelCard = (ch: ChannelStatus) => {
    return (
      <View key={ch.channel} style={styles.card}>
        <View style={styles.cardLeft}>
          <View style={styles.iconContainer}>
            <Text style={styles.iconText}>📹</Text>
          </View>
          <View>
            <Text style={styles.channelName}>{ch.name}</Text>
            <Text style={styles.channelId}>Channel ID: {ch.channel}</Text>
          </View>
        </View>

        <View style={styles.cardRight}>
          <Switch
            trackColor={{ false: 'rgba(255,255,255,0.1)', true: '#00d2ff' }}
            thumbColor={ch.status ? '#ffffff' : '#a0bacc'}
            ios_backgroundColor="rgba(255,255,255,0.1)"
            onValueChange={() => toggleChannel(ch.channel, ch.status)}
            value={ch.status}
          />
        </View>

        {ch.error && (
          <TouchableOpacity 
            style={styles.errorContainer}
            onPress={() => showErrorDetails(ch.channel, ch.error)}
          >
            <Text style={styles.errorTextHeading}>⚠️ Issue Detected. Tap for details.</Text>
          </TouchableOpacity>
        )}
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      <LinearGradient colors={['#0F2027', '#203A43', '#2C5364']} style={styles.gradient}>
        
        <View style={styles.header}>
          <View>
            <Text style={styles.headerTitle}>Network NVR</Text>
            <Text style={styles.headerSubtitle}>{serverIp}</Text>
          </View>
          <TouchableOpacity style={styles.logoutButton} onPress={onLogout}>
            <Text style={styles.logoutText}>Logout</Text>
          </TouchableOpacity>
        </View>

        {isLoading ? (
          <View style={styles.centerContainer}>
            <ActivityIndicator size="large" color="#00d2ff" />
          </View>
        ) : (
          <ScrollView 
            contentContainerStyle={styles.scrollContent}
            showsVerticalScrollIndicator={false}
          >
            {channels.map(renderChannelCard)}
            {channels.length === 0 && (
              <Text style={styles.emptyText}>No channels found on this NVR.</Text>
            )}
          </ScrollView>
        )}

      </LinearGradient>

      {/* Error Details Modal */}
      <Modal
        animationType="slide"
        transparent={true}
        visible={modalVisible}
        onRequestClose={() => setModalVisible(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.modalView}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>Channel {selectedChannelId} Error</Text>
            </View>
            
            <View style={styles.modalBody}>
              <Text style={styles.modalErrorDesc}>{selectedError}</Text>
            </View>

            <TouchableOpacity
              style={styles.modalButton}
              onPress={() => setModalVisible(false)}
            >
              <Text style={styles.modalButtonText}>Dismiss</Text>
            </TouchableOpacity>
          </View>
        </View>
      </Modal>

    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0F2027',
  },
  gradient: {
    flex: 1,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 24,
    paddingTop: Platform.OS === 'android' ? 40 : 20,
    paddingBottom: 20,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255,255,255,0.05)',
  },
  headerTitle: {
    fontSize: 24,
    fontWeight: '800',
    color: '#ffffff',
    letterSpacing: 0.5,
  },
  headerSubtitle: {
    fontSize: 12,
    color: '#a0bacc',
    marginTop: 2,
  },
  logoutButton: {
    backgroundColor: 'rgba(255,107,107,0.15)',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: 'rgba(255,107,107,0.3)',
  },
  logoutText: {
    color: '#ff6b6b',
    fontWeight: '700',
    fontSize: 14,
  },
  scrollContent: {
    padding: 24,
    paddingBottom: 40,
  },
  card: {
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderRadius: 20,
    padding: 20,
    marginBottom: 16,
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 10,
    elevation: 4,
  },
  cardLeft: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  iconContainer: {
    width: 48,
    height: 48,
    borderRadius: 14,
    backgroundColor: 'rgba(0, 210, 255, 0.15)',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  iconText: {
    fontSize: 24,
  },
  channelName: {
    fontSize: 18,
    fontWeight: '700',
    color: '#ffffff',
  },
  channelId: {
    fontSize: 13,
    color: '#8aa6b5',
    marginTop: 4,
  },
  cardRight: {
    justifyContent: 'center',
  },
  errorContainer: {
    width: '100%',
    marginTop: 16,
    padding: 12,
    backgroundColor: 'rgba(255, 107, 107, 0.1)',
    borderRadius: 12,
    borderWidth: 1,
    borderColor: 'rgba(255, 107, 107, 0.3)',
  },
  errorTextHeading: {
    color: '#ff6b6b',
    fontWeight: '600',
    fontSize: 13,
    textAlign: 'center',
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyText: {
    color: '#a0bacc',
    textAlign: 'center',
    fontSize: 16,
    marginTop: 40,
  },
  modalOverlay: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
    padding: 20,
  },
  modalView: {
    width: '100%',
    backgroundColor: '#1E2F38',
    borderRadius: 24,
    padding: 0,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 10 },
    shadowOpacity: 0.5,
    shadowRadius: 20,
    elevation: 20,
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  modalHeader: {
    padding: 24,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255, 255, 255, 0.05)',
  },
  modalTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#ff6b6b',
    textAlign: 'center',
  },
  modalBody: {
    padding: 24,
  },
  modalErrorDesc: {
    color: '#E0E0E0',
    fontSize: 15,
    lineHeight: 24,
  },
  modalButton: {
    margin: 24,
    marginTop: 0,
    backgroundColor: '#00d2ff',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
  },
  modalButtonText: {
    color: '#001a24',
    fontSize: 16,
    fontWeight: '700',
    textTransform: 'uppercase',
  },
});

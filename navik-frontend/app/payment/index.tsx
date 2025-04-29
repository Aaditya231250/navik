import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  ActivityIndicator,
  ScrollView,
  TextInput,
  Alert,
  StatusBar,
  Platform,
  RefreshControl,
} from 'react-native';
import { MaterialIcons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { getAccessToken } from '@/utils/authStorage';

interface WalletData {
  amount: number;
}

export default function WalletScreen() {
  const router = useRouter();
  const [walletData, setWalletData] = useState<WalletData | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [addAmount, setAddAmount] = useState('');
  const [paymentMethod, setPaymentMethod] = useState<string | null>(null);

  const fetchWalletData = async () => {
    try {
      const token = await getAccessToken();
      
      const response = await fetch('http://172.31.115.2/wallet', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
        }
      });

      if (response.ok) {
        const data = await response.json();
        setWalletData(data);
      } else {
        console.error('Failed to fetch wallet data:', response.status);
        Alert.alert('Error', 'Failed to fetch wallet data');
      }
    } catch (error) {
      console.error('Error fetching wallet data:', error);
      Alert.alert('Error', 'Could not connect to server');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const onRefresh = () => {
    setRefreshing(true);
    fetchWalletData();
  };

  useEffect(() => {
    fetchWalletData();
  }, []);

  const handleAddMoney = async () => {
    if (!paymentMethod) {
      Alert.alert('Error', 'Please select a payment method');
      return;
    }

    const amount = parseInt(addAmount);
    if (isNaN(amount) || amount <= 0) {
      Alert.alert('Error', 'Please enter a valid amount');
      return;
    }

    // For demo purposes, just update the UI
    Alert.alert(
      'Add Money',
      `Add ₹${amount} using ${paymentMethod}?`,
      [
        {
          text: 'Cancel',
          style: 'cancel',
        },
        {
          text: 'Add',
          onPress: () => {
            // Simulate successful addition
            if (walletData) {
              setWalletData({
                ...walletData,
                amount: walletData.amount + amount
              });
            }
            setAddAmount('');
            setPaymentMethod(null);
            Alert.alert('Success', 'Amount added successfully');
          }
        }
      ]
    );
  };

  const PaymentMethodOption = ({ method, icon, name }: { method: string, icon: string, name: string }) => (
    <TouchableOpacity
      className={`flex-row items-center w-[48%] p-3 rounded-lg mb-2.5 ${paymentMethod === method ? 'bg-teal-500' : 'bg-gray-100'}`}
      onPress={() => setPaymentMethod(method)}
    >
      <MaterialIcons name={icon} size={24} color={paymentMethod === method ? '#fff' : '#333'} />
      <Text className={`ml-2 ${paymentMethod === method ? 'text-white' : 'text-gray-700'}`}>
        {name}
      </Text>
    </TouchableOpacity>
  );

  return (
    <View 
      style={{ 
        flex: 1, 
        backgroundColor: '#f5f5f5',
        marginTop: Platform.OS === 'android' ? StatusBar.currentHeight : 0 
      }}
    >
      <StatusBar barStyle="dark-content" backgroundColor="#ffffff" />
      
      {/* Header */}
      <View className="flex-row items-center justify-between px-4 py-3 bg-white border-b border-gray-200">
        <TouchableOpacity onPress={() => router.back()} className="p-1">
          <MaterialIcons name="arrow-back" size={24} color="#000" />
        </TouchableOpacity>
        <Text className="text-lg font-bold">Wallet</Text>
        <View className="w-6" />
      </View>

      <ScrollView
        className="flex-1"
        contentContainerClassName="p-4"
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
      >
        {/* Wallet Balance Card */}
        <View className="bg-white rounded-xl p-5 mb-4 items-center shadow">
          <Text className="text-sm text-gray-500 mb-2">Available Balance</Text>
          {loading ? (
            <ActivityIndicator size="large" color="#00a5a0" />
          ) : (
            <Text className="text-4xl font-bold text-teal-500">
              ₹{0}
            </Text>
          )}
        </View>

        {/* Add Money Section */}
        <View className="bg-white rounded-xl p-5 mb-4 shadow">
          <Text className="text-lg font-bold mb-4">Add Money</Text>
          
          <TextInput
            className="border border-gray-300 rounded-lg p-3 text-base mb-4"
            placeholder="Enter amount"
            keyboardType="number-pad"
            value={addAmount}
            onChangeText={setAddAmount}
          />
          
          <Text className="text-base font-medium mb-3">Select Payment Method</Text>
          
          <View className="flex-row flex-wrap justify-between mb-5">
            <PaymentMethodOption 
              method="credit_card" 
              icon="credit-card" 
              name="Credit Card" 
            />
            <PaymentMethodOption 
              method="debit_card" 
              icon="account-balance" 
              name="Debit Card" 
            />
            <PaymentMethodOption 
              method="upi" 
              icon="smartphone" 
              name="UPI" 
            />
            <PaymentMethodOption 
              method="net_banking" 
              icon="language" 
              name="Net Banking" 
            />
          </View>
          
          <TouchableOpacity 
            className="bg-teal-500 rounded-lg p-4 items-center"
            onPress={handleAddMoney}
          >
            <Text className="text-white font-bold text-base">Add Money</Text>
          </TouchableOpacity>
        </View>

        {/* Recent Transactions Section */}
        <View className="bg-white rounded-xl p-5 mb-4 shadow">
          <Text className="text-lg font-bold mb-4">Recent Transactions</Text>
          <View className="items-center justify-center py-8">
            <MaterialIcons name="history" size={48} color="#ddd" />
            <Text className="mt-3 text-gray-400 text-base">No recent transactions</Text>
          </View>
        </View>
      </ScrollView>
    </View>
  );
}
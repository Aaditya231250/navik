import React, { useState, useEffect }  from "react";
import {
  View,
  Text,
  Image,
  TouchableOpacity,
  StatusBar,
  Platform,
  Alert, 
} from "react-native";
import { useRouter } from "expo-router";
import { MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from '@react-navigation/native';
import { clearAuthData, getUserData, getUserType } from "@/utils/authStorage";

export default function ProfileScreen() {
  const router = useRouter();
  const navigation = useNavigation();
  const [profileData, setProfileData] = useState(null);
  const [userType, setUserType] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadProfileData() {
      try {
        const userData = await getUserData();
        const type = await getUserType();
        setProfileData(userData);
        setUserType(type);
      } catch (error) {
        console.error('Error loading profile data:', error);
      } finally {
        setLoading(false);
      }
    }
    
    loadProfileData();
  }, []);

  const handleLogout = async () => {
    try {
      Alert.alert(
        "Logout",
        "Are you sure you want to logout?",
        [
          {
            text: "Cancel",
            style: "cancel"
          },
          {
            text: "Logout",
            onPress: async () => {
              await clearAuthData();
              
              // Navigate to login screen
              router.replace("/auth/login");
            }
          }
        ]
      );
    } catch (error) {
      console.error("Logout error:", error);
      Alert.alert("Error", "Could not log out. Please try again.");
    }
  };

  return (
    <View
      style={{
        flex: 1,
        backgroundColor: "#fff",
        marginTop: Platform.OS === "android" ? StatusBar.currentHeight : 0,
      }}
    >
      {/* Header */}
      <View className="flex-row items-center justify-between px-4 py-3 border-b border-gray-200">
        <Text className="text-xl font-bold">Profile</Text>
        <TouchableOpacity onPress={() => router.back()}>
          <MaterialIcons name="arrow-back" size={24} color="#000" />
        </TouchableOpacity>
      </View>

      {/* Profile Section */}
      <View className="items-center mt-6">
        <Image
          source={require("@/assets/images/profile-avatar.jpeg")}
          className="w-24 h-24 rounded-full"
        />
        <Text className="text-lg font-semibold mt-4">
          {profileData ? `${profileData.first_name} ${profileData.last_name}` : 'User Name'}
        </Text>
        <Text className="text-sm text-gray-500">
          {profileData ? profileData.email : 'user@example.com'}
        </Text>
        {userType && (
          <Text className="text-xs text-blue-500 mt-1">
            {userType === 'driver' ? 'Driver Account' : 'Customer Account'}
          </Text>
        )}
      </View>

      {/* Profile Options */}
      <View className="mt-8 px-4 space-y-4">
        {/* Edit Profile */}
        <TouchableOpacity className="flex-row items-center justify-between p-4 bg-gray-100 rounded-lg">
          <Text className="text-base font-medium">Edit Profile</Text>
          <MaterialIcons name="edit" size={24} color="#000" />
        </TouchableOpacity>

        {/* Payment Methods */}
        <TouchableOpacity className="flex-row items-center justify-between p-4 bg-gray-100 rounded-lg">
          <Text className="text-base font-medium">Payment Methods</Text>
          <MaterialIcons name="payment" size={24} color="#000" />
        </TouchableOpacity>

        {/* Ride History */}
        <TouchableOpacity className="flex-row items-center justify-between p-4 bg-gray-100 rounded-lg" onPress={() => navigation.navigate('Activity')}>
          <Text className="text-base font-medium">Ride History</Text>
          <MaterialIcons name="history" size={24} color="#000" />
        </TouchableOpacity>

        {/* Logout */}
        <TouchableOpacity
          onPress={handleLogout}
          className="flex-row items-center justify-between p-4 bg-red-100 rounded-lg"
        >
          <Text className="text-base font-medium text-red-500">Logout</Text>
          <MaterialIcons name="logout" size={24} color="#d9534f" />
        </TouchableOpacity>
      </View>
    </View>
  );
}
// app/home/profile/index.tsx
import React from "react";
import { View, Text, Image, TouchableOpacity, SafeAreaView } from "react-native";
import { useRouter } from "expo-router";
import { MaterialIcons } from "@expo/vector-icons";

export default function ProfileScreen() {
  const router = useRouter();

  return (
    <SafeAreaView className="flex-1 bg-white">
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
          source={require("@/assets/images/profile-avatar.jpeg")} // Add your avatar image in assets/images
          className="w-24 h-24 rounded-full"
        />
        <Text className="text-lg font-semibold mt-4">John Doe</Text>
        <Text className="text-sm text-gray-500">john.doe@example.com</Text>
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
        <TouchableOpacity className="flex-row items-center justify-between p-4 bg-gray-100 rounded-lg">
          <Text className="text-base font-medium">Ride History</Text>
          <MaterialIcons name="history" size={24} color="#000" />
        </TouchableOpacity>

        {/* Logout */}
        <TouchableOpacity
          onPress={() => alert("Logged out")}
          className="flex-row items-center justify-between p-4 bg-red-100 rounded-lg"
        >
          <Text className="text-base font-medium text-red-500">Logout</Text>
          <MaterialIcons name="logout" size={24} color="#d9534f" />
        </TouchableOpacity>
      </View>
    </SafeAreaView>
  );
}
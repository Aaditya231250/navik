// app/home/activity/index.tsx
import { View, Text, FlatList, TouchableOpacity, Image, SafeAreaView } from "react-native";
import React from "react";
import { MaterialIcons } from "@expo/vector-icons";

export default function ActivityScreen() {
  // Dummy ride history data
  const rideHistory = [
    {
      id: "1",
      type: "Uber Go",
      date: "Apr 2, 2025",
      time: "8:30 PM",
      price: "₹170.71",
      origin: "Badarpur Metro Station",
      destination: "Select Citywalk Mall",
      status: "Completed",
      image: require("@/assets/images/uber-go.png"),
    },
    {
      id: "2",
      type: "Auto",
      date: "Mar 28, 2025",
      time: "5:15 PM",
      price: "₹120.50",
      origin: "IGI Airport Terminal 3",
      destination: "Connaught Place",
      status: "Completed",
      image: require("@/assets/images/auto.png"),
    },
    {
      id: "3",
      type: "Uber Premier",
      date: "Mar 25, 2025",
      time: "10:20 AM",
      price: "₹223.63",
      origin: "Home",
      destination: "Saket Metro Station",
      status: "Completed",
      image: require("@/assets/images/uber-premier.png"),
    },
    {
      id: "4",
      type: "Uber Go",
      date: "Mar 20, 2025",
      time: "7:45 PM",
      price: "₹185.30",
      origin: "Cyber Hub",
      destination: "Home",
      status: "Cancelled",
      image: require("@/assets/images/uber-go.png"),
    },
    {
      id: "5",
      type: "Auto",
      date: "Mar 18, 2025",
      time: "3:30 PM",
      price: "₹110.25",
      origin: "Lajpat Nagar Market",
      destination: "Nehru Place",
      status: "Completed",
      image: require("@/assets/images/auto.png"),
    },
  ];

  const renderRideCard = ({ item }) => (
    <TouchableOpacity 
      className="bg-white rounded-lg shadow-sm mb-4 mx-4 border border-gray-200 overflow-hidden"
      onPress={() => console.log(`View details for ride ${item.id}`)}
    >
      <View className="p-4">
        {/* Header with ride type, date and price */}
        <View className="flex-row justify-between items-center mb-3">
          <View className="flex-row items-center">
            <Image source={item.image} className="w-10 h-10 mr-3" />
            <View>
              <Text className="font-bold text-lg">{item.type}</Text>
              <Text className="text-gray-500">{item.date}, {item.time}</Text>
            </View>
          </View>
          <Text className="font-bold text-lg">{item.price}</Text>
        </View>

        {/* Ride details with locations */}
        <View className="flex-row mt-2">
          <View className="items-center mr-3">
            <View className="w-2 h-2 bg-gray-500 rounded-full" />
            <View className="w-1 h-16 bg-gray-300" />
            <View className="w-4 h-4 bg-black rounded-sm" />
          </View>
          
          <View className="flex-1">
            <View className="mb-4">
              <Text className="text-gray-500">From</Text>
              <Text className="font-medium">{item.origin}</Text>
            </View>
            <View>
              <Text className="text-gray-500">To</Text>
              <Text className="font-medium">{item.destination}</Text>
            </View>
          </View>
        </View>

        {/* Status badge */}
        <View className="mt-3 flex-row justify-end">
          <View className={`px-3 py-1 rounded-full ${item.status === 'Completed' ? 'bg-green-100' : 'bg-red-100'}`}>
            <Text className={`text-sm font-medium ${item.status === 'Completed' ? 'text-green-600' : 'text-red-600'}`}>
              {item.status}
            </Text>
          </View>
        </View>
      </View>
    </TouchableOpacity>
  );

  return (
    <SafeAreaView className="flex-1 bg-gray-100">
      <View className="py-4 px-4 bg-white border-b border-gray-200">
        <Text className="text-2xl font-bold">Activity</Text>
        <Text className="text-gray-500 mt-1">Your ride history</Text>
      </View>

      <FlatList
        data={rideHistory}
        renderItem={renderRideCard}
        keyExtractor={item => item.id}
        className="pt-4"
        showsVerticalScrollIndicator={false}
      />
    </SafeAreaView>
  );
}

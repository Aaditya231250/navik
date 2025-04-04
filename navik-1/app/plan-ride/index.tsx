// app/plan-ride/index.tsx
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  FlatList,
  SafeAreaView,
  StatusBar,
  Platform,
} from "react-native";
import React, { useState } from "react";
import { useRouter } from "expo-router";
import { MaterialIcons, Ionicons } from "@expo/vector-icons";

export default function PlanRideScreen() {
  const router = useRouter();
  const [origin, setOrigin] = useState("Badarpur-India Post Office");
  const [destination, setDestination] = useState("");

  // Sample saved and recent places data
  const places = [
    { id: "saved", title: "Saved places", icon: "star", isHeader: true },
    {
      id: "place1",
      title: "Select Citywalk Mall",
      address:
        "Saket District Center, District Center, Sector 6, Pushp Vihar, New Delhi",
      icon: "location-on",
    },
    {
      id: "place2",
      title: "Select Citywalk Mall",
      address:
        "Saket District Center, District Center, Sector 6, Pushp Vihar, New Delhi",
      icon: "location-on",
    },
    {
      id: "place3",
      title: "Select Citywalk Mall",
      address:
        "Saket District Center, District Center, Sector 6, Pushp Vihar, New Delhi",
      icon: "location-on",
    },
    {
      id: "place4",
      title: "Select Citywalk Mall",
      address:
        "Saket District Center, District Center, Sector 6, Pushp Vihar, New Delhi",
      icon: "location-on",
    },
  ];

  const renderItem = ({ item }) => {
    if (item.isHeader) {
      return (
        <TouchableOpacity
          className="flex-row items-center py-4 px-4 bg-white border-b border-gray-200"
          onPress={() => {}}
        >
          <View className="w-8 h-8 bg-gray-300 rounded-full items-center justify-center mr-3">
            <MaterialIcons name={item.icon} size={18} color="#333" />
          </View>
          <Text className="text-base font-medium">{item.title}</Text>
          <MaterialIcons
            name="chevron-right"
            size={24}
            color="#999"
            style={{ marginLeft: "auto" }}
          />
        </TouchableOpacity>
      );
    }

    return (
      <TouchableOpacity
        className="flex-row items-center py-4 px-4 bg-white border-b border-gray-200"
        onPress={() => {
          setDestination(item.title);
          router.push("/ride-details"); // ✅ Navigate after selection
        }}
      >
        <View className="w-8 h-8 bg-gray-300 rounded-full items-center justify-center mr-3">
          <MaterialIcons name="location-on" size={18} color="#333" />
        </View>
        <View>
          <Text className="text-base font-medium">{item.title}</Text>
          <Text className="text-sm text-gray-500" numberOfLines={1}>
            {item.address}
          </Text>
        </View>
      </TouchableOpacity>
    );
  };

  return (
    <SafeAreaView
      className="flex-1 bg-white"
      style={{
        paddingTop: Platform.OS === "android" ? StatusBar.currentHeight : 0,
      }}
    >
      {/* Header */}
      <View className="flex-row items-center px-4 py-3 border-b border-gray-200">
        <TouchableOpacity onPress={() => router.back()} className="mr-4">
          <MaterialIcons name="arrow-back" size={24} color="#000" />
        </TouchableOpacity>
        <Text className="text-xl font-semibold">Plan your ride</Text>
      </View>

      {/* Ride Options */}
      <View className="flex-row px-4 py-3 space-x-2">
        <TouchableOpacity className="flex-row items-center bg-gray-200 px-3 py-2 rounded-full">
          <Ionicons name="time-outline" size={16} color="#000" />
          <Text className="ml-1 font-medium">Pick up now</Text>
          <MaterialIcons name="keyboard-arrow-down" size={18} color="#000" />
        </TouchableOpacity>

        <TouchableOpacity className="flex-row items-center bg-gray-200 px-3 py-2 rounded-full">
          <MaterialIcons name="swap-horiz" size={16} color="#000" />
          <Text className="ml-1 font-medium">One way</Text>
          <MaterialIcons name="keyboard-arrow-down" size={18} color="#000" />
        </TouchableOpacity>
      </View>

      {/* Location Inputs */}
      <View className="px-4 py-2">
        <View className="flex-row items-center mb-2">
          <View className="mr-3 items-center">
            <View className="w-2 h-2 bg-gray-500 rounded-full" />
            <View className="w-1 h-12 bg-gray-300" />
            <View className="w-4 h-4 bg-black rounded-sm" />
          </View>

          <View className="flex-1">
            <TextInput
              value={origin}
              onChangeText={setOrigin}
              className="py-2 px-3 bg-gray-100 rounded-md mb-2"
              editable={false}
            />
            <TextInput
              value={destination}
              onChangeText={setDestination}
              placeholder="Where to?"
              className="py-2 px-3 bg-gray-100 rounded-md"
              autoFocus
            />
          </View>
        </View>
      </View>

      {/* Suggested Places */}
      <FlatList
        data={places}
        renderItem={renderItem}
        keyExtractor={(item) => item.id}
        className="flex-1"
      />
    </SafeAreaView>
  );
}

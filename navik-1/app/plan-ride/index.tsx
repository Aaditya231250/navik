// app/plan-ride/index.tsx
import { View, Text, TextInput, TouchableOpacity, FlatList, SafeAreaView, StatusBar, Platform, Modal } from "react-native";
import React, { useState, useEffect } from "react";
import { useRouter } from "expo-router";
import { MaterialIcons, Ionicons, FontAwesome } from "@expo/vector-icons";
import * as Location from "expo-location";

export default function PlanRideScreen() {
  const router = useRouter();
  const [origin, setOrigin] = useState("");
  const [destination, setDestination] = useState("");
  const [showPassengerModal, setShowPassengerModal] = useState(false);
  const [passengers, setPassengers] = useState(1);
  
  useEffect(() => {
    (async () => {
      let { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== "granted") {
        setOrigin("Location access denied");
        return;
      }

      try {
        let location = await Location.getCurrentPositionAsync({});
        let address = await Location.reverseGeocodeAsync({
          latitude: location.coords.latitude,
          longitude: location.coords.longitude,
        });
        
        if (address && address.length > 0) {
          const currentLocation = `${address[0].name || ''}, ${address[0].street || ''}, ${address[0].city || ''}`;
          setOrigin(currentLocation);
        } else {
          setOrigin("Current Location");
        }
      } catch (error) {
        setOrigin("Current Location");
      }
    })();
  }, []);
  
  const places = [
    { id: "saved", title: "Saved places", icon: "star", isHeader: true },
    { id: "place1", title: "Select Citywalk Mall", address: "Saket District Center, District Center, Sector 6, Pushp Vihar, New Delhi", icon: "location-on" },
    { id: "place2", title: "Select Citywalk Mall", address: "Saket District Center, District Center, Sector 6, Pushp Vihar, New Delhi", icon: "location-on" },
    { id: "place3", title: "Select Citywalk Mall", address: "Saket District Center, District Center, Sector 6, Pushp Vihar, New Delhi", icon: "location-on" },
    { id: "place4", title: "Select Citywalk Mall", address: "Saket District Center, District Center, Sector 6, Pushp Vihar, New Delhi", icon: "location-on" },
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
          <MaterialIcons name="chevron-right" size={24} color="#999" style={{ marginLeft: 'auto' }} />
        </TouchableOpacity>
      );
    }
    
    return (
      <TouchableOpacity 
        className="flex-row items-center py-4 px-4 bg-white border-b border-gray-200"
        onPress={() => {
          setDestination(item.title);
          router.push("/ride-details"); // Navigate after selection
        }}
      >
        <View className="w-8 h-8 bg-gray-300 rounded-full items-center justify-center mr-3">
          <MaterialIcons name="location-on" size={18} color="#333" />
        </View>
        <View className="flex-1">
          <Text className="text-base font-medium">{item.title}</Text>
          <Text className="text-sm text-gray-500" numberOfLines={1} ellipsizeMode="tail">{item.address}</Text>
        </View>
      </TouchableOpacity>
    );
  };

  return (
    <SafeAreaView className="flex-1 bg-white" style={{ paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 0 }}>
      <View className="flex-row items-center px-4 py-3 border-b border-gray-200">
        <TouchableOpacity onPress={() => router.back()} className="mr-4">
          <MaterialIcons name="arrow-back" size={24} color="#000" />
        </TouchableOpacity>
        <Text className="text-xl font-semibold">Plan your ride</Text>
      </View>
      
      <View className="flex-row px-4 py-3 space-x-2">
        <TouchableOpacity className="flex-row items-center bg-gray-200 px-3 py-2 rounded-full">
          <Ionicons name="time-outline" size={16} color="#000" />
          <Text className="ml-1 font-medium">Pick up now</Text>
          <MaterialIcons name="keyboard-arrow-down" size={18} color="#000" />
        </TouchableOpacity>
        
        <TouchableOpacity 
          className="flex-row items-center bg-gray-200 px-3 py-2 rounded-full"
          onPress={() => setShowPassengerModal(true)}
        >
          <FontAwesome name="user" size={16} color="#000" />
          <Text className="ml-1 font-medium">{passengers} {passengers === 1 ? 'Passenger' : 'Passengers'}</Text>
          <MaterialIcons name="keyboard-arrow-down" size={18} color="#000" />
        </TouchableOpacity>
      </View>
      
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
              editable={true}
              placeholder="Pick up location"
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
      
      <FlatList
        data={places}
        renderItem={renderItem}
        keyExtractor={item => item.id}
        className="flex-1"
      />

      <Modal
        visible={showPassengerModal}
        transparent={true}
        animationType="slide"
        onRequestClose={() => setShowPassengerModal(false)}
      >
        <TouchableOpacity 
          style={{ flex: 1, backgroundColor: 'rgba(0,0,0,0.5)' }}
          activeOpacity={1}
          onPress={() => setShowPassengerModal(false)}
        >
          <View className="bg-white rounded-t-xl absolute bottom-0 w-full p-4">
            <Text className="text-xl font-bold mb-4">Select passengers</Text>
            
            {[1, 2, 3, 4, '4+'].map((num) => (
              <TouchableOpacity
                key={num}
                className="py-3 border-b border-gray-200 flex-row items-center"
                onPress={() => {
                  setPassengers(typeof num === 'string' ? 4 : num);
                  setShowPassengerModal(false);
                }}
              >
                <Text className="text-lg">{num} {num === 1 ? 'Passenger' : 'Passengers'}</Text>
                {passengers === num && (
                  <MaterialIcons name="check" size={24} color="#000" style={{ marginLeft: 'auto' }} />
                )}
              </TouchableOpacity>
            ))}
          </View>
        </TouchableOpacity>
      </Modal>
    </SafeAreaView>
  );
}

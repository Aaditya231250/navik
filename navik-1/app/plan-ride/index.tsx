import { 
  View, 
  Text, 
  TextInput, 
  TouchableOpacity, 
  FlatList, 
  SafeAreaView, 
  StatusBar, 
  Platform, 
  Modal, 
  ActivityIndicator,
  Alert
} from "react-native";
import React, { useState, useEffect, useCallback } from "react";
import { useRouter } from "expo-router";
import { MaterialIcons, Ionicons, FontAwesome } from "@expo/vector-icons";
import * as Location from "expo-location";
import { useMapWebSocket, SearchPlace } from "@/utils/websocket";
import { debounce } from "lodash";

export default function PlanRideScreen() {
  const router = useRouter();
  const [activeField, setActiveField] = useState<"origin" | "destination">("origin");
  const [origin, setOrigin] = useState<{ text: string; coords: { latitude: number; longitude: number } | null }>({ 
    text: "Current Location", 
    coords: null 
  });
  const [destination, setDestination] = useState<{ text: string; coords: { latitude: number; longitude: number } | null }>({ 
    text: "", 
    coords: null 
  });
  const [showPassengerModal, setShowPassengerModal] = useState(false);
  const [passengers, setPassengers] = useState(1);
  const [isSearching, setIsSearching] = useState(false);
  
  const wsUrl = "ws://172.31.115.2/search";
  const { 
    isConnected, 
    searchPlaces, 
    searchResults, 
    error, 
    reconnect 
  } = useMapWebSocket(wsUrl);

  useEffect(() => {
    (async () => {
      const { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== "granted") return;

      try {
        const location = await Location.getCurrentPositionAsync();
        const address = await Location.reverseGeocodeAsync(location.coords);
        setOrigin({
          text: address[0]?.name || "Current Location",
          coords: location.coords
        });
      } catch (error) {
        Alert.alert("Error", "Failed to get current location");
      }
    })();
  }, []);

  const handleSearch = useCallback(
    debounce((text: string) => {
      if (text.length > 2) {
        setIsSearching(true);
        searchPlaces(text);
      }
    }, 500),
    [searchPlaces]
  );

  const handleLocationSelect = (place: SearchPlace) => {
    const update = {
      text: place.name,
      coords: { latitude: place.latitude, longitude: place.longitude }
    };

    if (activeField === "origin") {
      setOrigin(update);
    } else {
      setDestination(update);
    }
    setIsSearching(false);
  };

  const handleConfirm = () => {
    if (!origin.coords || !destination.coords) {
      Alert.alert("Please select both pickup and destination locations.");
      return;
    }
    
    router.push({
      pathname: "/ride-details",
      params: {
        originLat: origin.coords.latitude.toString(),
        originLng: origin.coords.longitude.toString(),
        destLat: destination.coords.latitude.toString(),
        destLng: destination.coords.longitude.toString(),
        originTitle: origin.text,
        destTitle: destination.text
      }
    });
  };

  const renderItem = ({ item }: { item: SearchPlace }) => (
    <TouchableOpacity
      className="flex-row items-center py-4 px-4 bg-white border-b border-gray-200"
      onPress={() => handleLocationSelect(item)}
    >
      <View className="w-8 h-8 bg-gray-300 rounded-full items-center justify-center mr-3">
        <MaterialIcons name="location-on" size={18} color="#333" />
      </View>
      <View className="flex-1">
        <Text className="text-base font-medium">{item.name}</Text>
        <Text className="text-sm text-gray-500">{item.address}</Text>
      </View>
    </TouchableOpacity>
  );

  return (
    <SafeAreaView className="flex-1 bg-white" style={{ paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 0 }}>
      <View className="flex-row items-center px-4 py-3 border-b border-gray-200">
        <TouchableOpacity onPress={() => router.back()} className="mr-4">
          <MaterialIcons name="arrow-back" size={24} color="#000" />
        </TouchableOpacity>
        <Text className="text-xl font-semibold">Plan your ride</Text>
      </View>

      <View className="px-4 py-2">
        <View className="flex-row items-center mb-2">
          <View className="mr-3 items-center">
            <View className="w-2 h-2 bg-gray-500 rounded-full" />
            <View className="w-1 h-12 bg-gray-300" />
            <View className="w-4 h-4 bg-black rounded-sm" />
          </View>
          
          <View className="flex-1">
            <TouchableOpacity onPress={() => setActiveField("origin")}>
              <TextInput
                value={origin.text}
                placeholder="Pick up location"
                className="py-2 px-3 bg-gray-100 rounded-md mb-2"
                editable={false}
              />
            </TouchableOpacity>
            
            <TouchableOpacity onPress={() => setActiveField("destination")}>
              <TextInput
                value={destination.text}
                placeholder="Where to?"
                className="py-2 px-3 bg-gray-100 rounded-md"
                editable={false}
              />
            </TouchableOpacity>
          </View>
        </View>
      </View>

      {activeField && (
        <View className="absolute top-32 left-0 right-0 bottom-0 bg-white z-50">
          <View className="flex-row items-center px-4 py-3 border-b border-gray-200">
            <MaterialIcons name="search" size={24} color="#666" />
            <TextInput
              autoFocus
              style={{ flex: 1, marginLeft: 8, fontSize: 16 }}
              placeholder={`Search for ${activeField}...`}
              onChangeText={handleSearch}
            />
            <TouchableOpacity onPress={() => setActiveField(null)}>
              <MaterialIcons name="close" size={24} color="#666" />
            </TouchableOpacity>
          </View>

          {!isConnected && (
            <View className="bg-yellow-50 p-2 border-b border-yellow-200">
              <Text className="text-yellow-700 text-center">
                Connecting to search service...
                <TouchableOpacity onPress={reconnect}>
                  <Text className="text-blue-600 font-bold"> Retry</Text>
                </TouchableOpacity>
              </Text>
            </View>
          )}

          {error && (
            <View className="bg-red-50 p-2 border-b border-red-200">
              <Text className="text-red-700 text-center">{error}</Text>
            </View>
          )}

          <FlatList
            data={searchResults}
            renderItem={renderItem}
            keyExtractor={(item) => item.id}
            ListHeaderComponent={
              <TouchableOpacity
                className="flex-row items-center py-4 px-4 bg-white border-b border-gray-200"
                onPress={async () => {
                  try {
                    const location = await Location.getCurrentPositionAsync();
                    const address = await Location.reverseGeocodeAsync(location.coords);
                    setOrigin({
                      text: address[0]?.name || "Current Location",
                      coords: location.coords
                    });
                    setActiveField(null);
                  } catch (error) {
                    Alert.alert("Error", "Could not get current location");
                  }
                }}
              >
                <View className="w-8 h-8 bg-blue-100 rounded-full items-center justify-center mr-3">
                  <Ionicons name="locate" size={18} color="#007AFF" />
                </View>
                <View className="flex-1">
                  <Text className="text-base font-medium">Current Location</Text>
                  <Text className="text-sm text-gray-500">Use your current position</Text>
                </View>
              </TouchableOpacity>
            }
          />
        </View>
      )}

      <View className="px-4 py-2">
        <TouchableOpacity
          className="bg-black py-3 rounded-lg items-center"
          onPress={handleConfirm}
        >
          <Text className="text-white font-semibold text-lg">Confirm Ride</Text>
        </TouchableOpacity>
      </View>

      <Modal
        visible={showPassengerModal}
        transparent={true}
        animationType="slide"
        onRequestClose={() => setShowPassengerModal(false)}
      >
        <TouchableOpacity
          style={{ flex: 1, backgroundColor: "rgba(0,0,0,0.5)" }}
          activeOpacity={1}
          onPress={() => setShowPassengerModal(false)}
        >
          <View className="bg-white rounded-t-xl absolute bottom-0 w-full p-4">
            <Text className="text-xl font-bold mb-4">Select passengers</Text>
            {[1, 2, 3, 4, "4+"].map((num) => (
              <TouchableOpacity
                key={num}
                className="py-3 border-b border-gray-200 flex-row items-center"
                onPress={() => {
                  setPassengers(typeof num === "string" ? 4 : num);
                  setShowPassengerModal(false);
                }}
              >
                <Text className="text-lg">
                  {num} {num === 1 ? "Passenger" : "Passengers"}
                </Text>
                {passengers === num && (
                  <MaterialIcons name="check" size={24} color="#000" style={{ marginLeft: "auto" }} />
                )}
              </TouchableOpacity>
            ))}
          </View>
        </TouchableOpacity>
      </Modal>
    </SafeAreaView>
  );
}
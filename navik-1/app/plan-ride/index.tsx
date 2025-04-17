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
import React, { useState, useEffect, useCallback, useRef } from "react";
import { useRouter } from "expo-router";
import { MaterialIcons, Ionicons } from "@expo/vector-icons";
import * as Location from "expo-location";
import { useMapWebSocket, SearchPlace } from "@/utils/websocket";
import { debounce } from "lodash";

export default function PlanRideScreen() {
  const router = useRouter();
  const inputsContainerRef = useRef(null);
  const [inputsHeight, setInputsHeight] = useState(140);
  const [activeField, setActiveField] = useState<"origin" | "destination" | null>("destination");
  const [searchText, setSearchText] = useState("");
  const [isSearching, setIsSearching] = useState(false);
  
  // Updated state to include placeId
  const [origin, setOrigin] = useState<{ 
    text: string;
    placeId: string | null;
    coords: { latitude: number; longitude: number } | null 
  }>({ 
    text: "Current Location", 
    placeId: null,
    coords: null 
  });
  
  const [destination, setDestination] = useState<{ 
    text: string;
    placeId: string | null;
    coords: { latitude: number; longitude: number } | null 
  }>({ 
    text: "", 
    placeId: null,
    coords: null 
  });
  
  const [passengers, setPassengers] = useState(1);
  const [showPassengerModal, setShowPassengerModal] = useState(false);
  
  const wsUrl = "ws://172.31.115.2/search";
  const { 
    isConnected, 
    searchPlaces, 
    searchResults, 
    error, 
    reconnect 
  } = useMapWebSocket(wsUrl);

  // Get current location on mount
  useEffect(() => {
    (async () => {
      const { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== "granted") return;

      try {
        const location = await Location.getCurrentPositionAsync();
        const address = await Location.reverseGeocodeAsync(location.coords);
        setOrigin({
          text: address[0]?.name || "Current Location",
          placeId: null, // Current location doesn't have a placeId
          coords: location.coords
        });
      } catch (error) {
        Alert.alert("Error", "Failed to get current location");
      }
    })();
  }, []);

  // Measure inputs container for proper modal positioning
  useEffect(() => {
    if (inputsContainerRef.current) {
      setTimeout(() => {
        inputsContainerRef.current.measure((x, y, width, height, pageX, pageY) => {
          setInputsHeight(pageY + height + 20);
        });
      }, 300);
    }
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

  // Updated to store placeId instead of coordinates
  const handleLocationSelect = (place: SearchPlace) => {
    if (!place || !place.id) {
      console.warn("Invalid place data", place);
      return;
    }

    const update = {
      text: place.name,
      placeId: place.id,
      coords: null // We don't need coordinates anymore
    };

    if (activeField === "origin") {
      setOrigin(update);
    } else {
      setDestination(update);
    }
    
    setIsSearching(false);
    setSearchText("");
    setActiveField(null);
  };

  // Updated to use placeId for navigation
  const handleConfirm = () => {
    // Special case for origin if it's current location (has coords but no placeId)
    if (!origin.text || !destination.text || (!destination.placeId && !destination.coords)) {
      Alert.alert("Please select both pickup and destination locations");
      return;
    }
    
    // Navigate with placeIds (or coords for current location)
    router.push({
      pathname: "/ride-details",
      params: {
        // For origin, we might use coordinates if it's current location
        originPlaceId: origin.placeId || "",
        destPlaceId: destination.placeId || "",
        // Include coordinates as fallback if available
        originLat: origin.coords?.latitude?.toString() || "",
        originLng: origin.coords?.longitude?.toString() || "",
        destLat: destination.coords?.latitude?.toString() || "",
        destLng: destination.coords?.longitude?.toString() || "",
        // Always include display text
        originTitle: origin.text,
        destTitle: destination.text,
        // Include passenger count
        passengers: passengers.toString()
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

      <View 
        ref={inputsContainerRef} 
        className="px-4 py-2"
        onLayout={() => {
          if (inputsContainerRef.current) {
            inputsContainerRef.current.measure((x, y, width, height, pageX, pageY) => {
              setInputsHeight(pageY + height + 20);
            });
          }
        }}
      >
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
        <View 
          className="absolute left-0 right-0 bottom-0 bg-white z-50"
          style={{ 
            top: inputsHeight, 
            elevation: 5,
            shadowColor: "#000",
            shadowOffset: { width: 0, height: 2 },
            shadowOpacity: 0.25,
            shadowRadius: 3.84,
          }}
        >
          <View className="flex-row items-center px-4 py-3 border-b border-gray-200">
            <MaterialIcons name="search" size={24} color="#666" />
            <TextInput
              autoFocus
              style={{ flex: 1, marginLeft: 8, fontSize: 16 }}
              placeholder={`Search for ${activeField}...`}
              onChangeText={(text) => {
                setSearchText(text);
                handleSearch(text);
              }}
              value={searchText}
            />
            <TouchableOpacity onPress={() => {
              setActiveField(null);
              setSearchText("");
            }}>
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

          {isSearching && searchResults.length === 0 && (
            <View className="items-center justify-center py-20">
              <ActivityIndicator size="large" color="#0000ff" />
              <Text className="mt-4 text-gray-500">Searching locations...</Text>
            </View>
          )}

          <FlatList
            data={searchResults}
            renderItem={renderItem}
            keyExtractor={(item) => item.id.toString()}
            ListHeaderComponent={
              <TouchableOpacity
                className="flex-row items-center py-4 px-4 bg-white border-b border-gray-200"
                onPress={async () => {
                  try {
                    const location = await Location.getCurrentPositionAsync();
                    const address = await Location.reverseGeocodeAsync(location.coords);
                    const update = {
                      text: address[0]?.name || "Current Location",
                      placeId: null, // Current location doesn't have a placeId
                      coords: location.coords // Keep coords for current location
                    };
                    
                    if (activeField === "origin") {
                      setOrigin(update);
                    } else if (activeField === "destination") {
                      setDestination(update);
                    }
                    
                    setActiveField(null);
                    setSearchText("");
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
            ListEmptyComponent={
              !isSearching && searchText.length > 0 && (
                <View className="items-center justify-center py-20">
                  <Text className="text-gray-500">No results found</Text>
                </View>
              )
            }
          />
        </View>
      )}

      <View className="px-4 py-2 mt-2">
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
                key={num.toString()}
                className="py-3 border-b border-gray-200 flex-row items-center"
                onPress={() => {
                  setPassengers(typeof num === "string" ? 4 : Number(num));
                  setShowPassengerModal(false);
                }}
              >
                <Text className="text-lg">
                  {num} {num === 1 ? "Passenger" : "Passengers"}
                </Text>
                {passengers === (typeof num === "string" ? 4 : Number(num)) && (
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

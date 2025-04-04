// app/home/index.tsx
import {
  View,
  Text,
  Image,
  TouchableOpacity,
  ScrollView,
  RefreshControl,
} from "react-native";
import SearchBar from "@/components/SearchBar";
import MapView, { Marker } from "react-native-maps";
import * as Location from "expo-location";
import React, { useEffect, useState } from "react";

export default function HomeScreen() {
  const [location, setLocation] = useState<{
    latitude: number;
    longitude: number;
  } | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState<boolean>(false);

  // Function to fetch current location
  const fetchLocation = async () => {
    try {
      let { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== "granted") {
        setErrorMsg("Permission to access location was denied");
        return;
      }

      let currentLocation = await Location.getCurrentPositionAsync({});
      console.log(currentLocation);
      setLocation(currentLocation.coords);
    } catch (error) {
      console.error(error);
    }
  };

  // Fetch location on mount
  useEffect(() => {
    fetchLocation();
  }, []);

  // Handle pull-to-refresh
  const onRefresh = async () => {
    setRefreshing(true);
    await fetchLocation();
    setRefreshing(false);
  };

  return (
    <ScrollView
      className="flex-1 bg-white"
      refreshControl={
        <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
      }
    >
      {/* Search Bar */}
      <SearchBar />

      {/* Suggestions Section */}
      <View className="mt-6 px-4">
        <Text className="text-lg font-bold text-gray-800 mb-4">
          Suggestions
        </Text>
        <View className="flex-row justify-between">
          {/* Ride Option */}
          <TouchableOpacity className="items-center bg-gray-100 p-4 rounded-lg w-[48%] shadow-sm">
            <Image
              source={require("@/assets/images/homeScreen/ride.png")}
              className="h-16 mb-2"
            />
            <Text className="text-gray-800 font-semibold">Ride</Text>
          </TouchableOpacity>

          {/* Package Option */}
          <TouchableOpacity className="items-center bg-gray-100 p-4 rounded-lg w-[48%] shadow-sm">
            <Image
              source={require("@/assets/images/homeScreen/package.png")}
              className="w-16 h-16 mb-2"
            />
            <Text className="text-gray-800 font-semibold">Package</Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Around You Section */}
      <View className="mt-6 px-4">
        <Text className="text-lg font-bold text-gray-800 mb-4">Around You</Text>
        {/* Map Container */}
        <View
          style={{
            marginHorizontal: 0,
            borderRadius: 15,
            overflow: "hidden",
          }}
        >
          <MapView
            style={{ height: 300, width: "100%" }}
            region={
              location
                ? {
                    latitude: location.latitude,
                    longitude: location.longitude,
                    latitudeDelta: 0.01,
                    longitudeDelta: 0.01,
                  }
                : undefined // Render nothing if location is null
            }
          >
            {/* Location Marker */}
            {location && (
              <Marker
                coordinate={{
                  latitude: location.latitude,
                  longitude: location.longitude,
                }}
                title="Your Location"
                description="You are here"
              >
                <Image
                  source={require("@/assets/images/homeScreen/pin.png")}
                  style={{ width: 30, height: 30 }}
                  resizeMode="contain"
                />
              </Marker>
            )}
          </MapView>
        </View>
        {errorMsg && (
          <Text style={{ color: "red", marginTop: 10 }}>{errorMsg}</Text>
        )}
      </View>
    </ScrollView>
  );
}

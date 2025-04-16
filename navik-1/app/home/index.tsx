import {
  View,
  Text,
  Image,
  SafeAreaView,
  StatusBar,
  Platform,
  TouchableOpacity,
} from "react-native";
import SearchBar from "@/components/SearchBar";
import MapView, { Marker, Region } from "react-native-maps";
import * as Location from "expo-location";
import React, { useEffect, useState, useRef } from "react";
import { Ionicons } from "@expo/vector-icons";

export default function HomeScreen() {
  const [location, setLocation] = useState<{
    latitude: number;
    longitude: number;
  } | null>(null);

  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [isMapOffCenter, setIsMapOffCenter] = useState(false);
  const [currentRegion, setCurrentRegion] = useState<Region | null>(null);
  const mapRef = useRef<MapView>(null);

  const fetchLocation = async () => {
    try {
      let { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== "granted") {
        setErrorMsg("Permission to access location was denied");
        return;
      }

      let currentLocation = await Location.getCurrentPositionAsync({});
      setLocation(currentLocation.coords);
      
      // Set initial region
      if (currentLocation.coords) {
        const initialRegion = {
          latitude: currentLocation.coords.latitude,
          longitude: currentLocation.coords.longitude,
          latitudeDelta: 0.01,
          longitudeDelta: 0.01,
        };
        setCurrentRegion(initialRegion);
      }
    } catch (error) {
      console.error(error);
      setErrorMsg("Failed to get location");
    }
  };

  useEffect(() => {
    fetchLocation();
  }, []);

  // Function to recenter map to user's location
  const recenterMap = () => {
    if (location && mapRef.current) {
      const region = {
        latitude: location.latitude,
        longitude: location.longitude,
        latitudeDelta: 0.01,
        longitudeDelta: 0.01,
      };
      mapRef.current.animateToRegion(region, 500);
      setIsMapOffCenter(false);
    }
  };

  const checkIfMapOffCenter = (region: Region) => {
    if (!location) return;
    
    const distanceThreshold = 0.005;
    const zoomThreshold = 0.08; 
    
    const latDiff = Math.abs(region.latitude - location.latitude);
    const lngDiff = Math.abs(region.longitude - location.longitude);
    const isOffCenter = latDiff > distanceThreshold || lngDiff > distanceThreshold;
    const isZoomedOut = region.latitudeDelta > zoomThreshold || region.longitudeDelta > zoomThreshold;
    
    setIsMapOffCenter(isOffCenter || isZoomedOut);
    setCurrentRegion(region);
  };

  return (
    <View className="flex-1 relative bg-white">
      {/* Map as full background */}
      <MapView
        ref={mapRef}
        style={{ flex: 1 }}
        initialRegion={
          location
            ? {
                latitude: location.latitude,
                longitude: location.longitude,
                latitudeDelta: 0.01,
                longitudeDelta: 0.01,
              }
            : undefined
        }
        region={currentRegion || undefined}
        // showsUserLocation={true}
        onRegionChangeComplete={checkIfMapOffCenter}
      >
        {location && (
          <Marker
            coordinate={{
              latitude: location.latitude,
              longitude: location.longitude,
            }}
            title="Your Location"
            description="You're here"
          >
            <Image
              source={require("@/assets/images/homeScreen/pin.png")}
              style={{ width: 30, height: 30 }}
              resizeMode="contain"
            />
          </Marker>
        )}
      </MapView>

      {/* Floating SearchBar */}
      <SafeAreaView
        style={{
          position: "absolute",
          top: Platform.OS === "android" ? StatusBar.currentHeight! + 10 : 10,
          left: 0,
          right: 0,
          paddingHorizontal: 16,
        }}
      >
        <SearchBar />
      </SafeAreaView>

      {/* Recenter button - appears only when needed */}
      {isMapOffCenter && location && (
        <TouchableOpacity
          className="absolute bottom-8 right-4 bg-white rounded-full w-12 h-12 flex items-center justify-center shadow-lg"
          style={{
            elevation: 5, // Android shadow
            shadowColor: "#000", // iOS shadow
            shadowOffset: { width: 0, height: 2 },
            shadowOpacity: 0.15,
            shadowRadius: 4,
          }}
          onPress={recenterMap}
        >
          <Ionicons name="locate" size={24} color="#007AFF" />
        </TouchableOpacity>
      )}

      {/* Error Message */}
      {errorMsg && (
        <View className="absolute bottom-4 left-4 right-4 bg-white p-2 rounded shadow">
          <Text className="text-red-600">{errorMsg}</Text>
        </View>
      )}
    </View>
  );
}

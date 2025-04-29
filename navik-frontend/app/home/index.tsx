import {
  View,
  Text,
  Image,
  SafeAreaView,
  StatusBar,
  Platform,
  TouchableOpacity,
} from "react-native";
import MapView, { Marker, Region } from "react-native-maps";
import * as Location from "expo-location";
import React, { useEffect, useState, useRef } from "react";
import { Ionicons } from "@expo/vector-icons";
import SearchBar from "@/components/SearchBar";

export default function HomeScreen() {
  const [location, setLocation] = useState<{
    latitude: number;
    longitude: number;
  } | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [isMapOffCenter, setIsMapOffCenter] = useState(false);
  const [currentRegion, setCurrentRegion] = useState<Region | null>(null);
  const mapRef = useRef<MapView>(null);

  const fetchLocation = async (retries = 3) => {
    try {
      const { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== "granted") {
        setErrorMsg("Permission to access location was denied");
        return;
      }

      let location;
      for (let attempt = 0; attempt < retries; attempt++) {
        try {
          location = await Location.getCurrentPositionAsync({
            accuracy: Location.Accuracy.High,
            timeInterval: 5000,
          });
          break;
        } catch (error) {
          if (attempt === retries - 1) throw error;
          await new Promise(resolve => setTimeout(resolve, 1000));
        }
      }

      const coords = location!.coords;
      setLocation(coords);
      setCurrentRegion({
        latitude: coords.latitude,
        longitude: coords.longitude,
        latitudeDelta: 0.02,
        longitudeDelta: 0.02,
      });

    } catch (error) {
      console.error(error);
      setErrorMsg("Failed to get location. Ensure location services are enabled.");
      setLocation(null);
    }
  };

  const recenterMap = () => {
    if (location && mapRef.current) {
      const region = {
        latitude: location.latitude,
        longitude: location.longitude,
        latitudeDelta: 0.02,
        longitudeDelta: 0.02,
      };
      mapRef.current.animateToRegion(region, 500);
      setIsMapOffCenter(false);
    }
  };

  useEffect(() => {
    fetchLocation();
  }, []);

  return (
    <View className="flex-1 relative bg-white">
      <SafeAreaView
        style={{
          position: "absolute",
          top: Platform.OS === "android" ? StatusBar.currentHeight! + 10 : 10,
          left: 0,
          right: 0,
          paddingHorizontal: 16,
          zIndex: 2,
        }}
      >
        <SearchBar />
      </SafeAreaView>

      <MapView
        ref={mapRef}
        style={{ flex: 1 }}
        initialRegion={currentRegion || undefined}
        region={currentRegion || undefined}
        onRegionChangeComplete={(region) => {
          if (!location) return;
          const latDiff = Math.abs(region.latitude - location.latitude);
          const lngDiff = Math.abs(region.longitude - location.longitude);
          setIsMapOffCenter(latDiff > 0.005 || lngDiff > 0.005);
        }}
      >
        {location && (
          <Marker
            coordinate={location}
            title="Your Location"
            description="Current position"
          >
            <Image
              source={require("@/assets/images/homeScreen/pin.png")}
              style={{ width: 30, height: 30 }}
              resizeMode="contain"
            />
          </Marker>
        )}
      </MapView>

      {isMapOffCenter && location && (
        <TouchableOpacity
          className="absolute bottom-8 right-4 bg-white rounded-full w-12 h-12 flex items-center justify-center shadow-lg"
          style={{
            elevation: 5,
            shadowColor: "#000",
            shadowOffset: { width: 0, height: 2 },
            shadowOpacity: 0.15,
            shadowRadius: 4,
          }}
          onPress={recenterMap}
        >
          <Ionicons name="locate" size={24} color="#007AFF" />
        </TouchableOpacity>
      )}

      {!location && (
        <TouchableOpacity
          className="absolute bottom-20 right-4 bg-white p-3 rounded-full shadow-lg"
          onPress={fetchLocation}
        >
          <Ionicons name="refresh" size={24} color="#007AFF" />
        </TouchableOpacity>
      )}

      {errorMsg && (
        <View className="absolute bottom-4 left-4 right-4 bg-white p-2 rounded shadow">
          <Text className="text-red-600">{errorMsg}</Text>
        </View>
      )}
    </View>
  );
}

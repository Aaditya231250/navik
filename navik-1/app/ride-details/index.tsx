// app/ride-details/index.tsx
import React, { useState, useEffect } from "react";
import {
  View,
  Text,
  TouchableOpacity,
  Image,
  FlatList,
  ActivityIndicator,
} from "react-native";
import { useRouter, useLocalSearchParams } from "expo-router";
import { MaterialIcons } from "@expo/vector-icons";
import MapView, { Marker, Polyline, PROVIDER_GOOGLE } from "react-native-maps";
import polyline from "@mapbox/polyline";
import * as Location from 'expo-location';

// Pricing configuration
const PRICING = {
  BASE_FARE: 40,
  PER_KM: 14,
  PER_MIN: 2,
  SERVICE_FEE: 10,
  SURGE_MULTIPLIER: 1.2,
};

export default function RideDetailsScreen() {
  const router = useRouter();
  const params = useLocalSearchParams();
  const [selectedRide, setSelectedRide] = useState(null);
  const [loading, setLoading] = useState(true);
  const [routeCoordinates, setRouteCoordinates] = useState([]);
  const [mapRegion, setMapRegion] = useState(null);
  const [eta, setEta] = useState('');
  const [distance, setDistance] = useState('');
  const [rideOptions, setRideOptions] = useState([]);
  const [userLocation, setUserLocation] = useState(null);
  const [destinationCoords, setDestinationCoords] = useState(null);

  // Get destination from navigation params
  const destination = {
    latitude: parseFloat(params.destLat),
    longitude: parseFloat(params.destLng),
    title: params.destTitle,
    address: params.destAddress
  };

  // Get user's current location
  useEffect(() => {
    (async () => {
      let { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== 'granted') return;

      let location = await Location.getCurrentPositionAsync({});
      setUserLocation({
        latitude: location.coords.latitude,
        longitude: location.coords.longitude
      });
    })();
  }, []);

  // Calculate route when locations are available
  useEffect(() => {
    if (userLocation && destination.latitude) {
      setupMap();
    }
  }, [userLocation, destination]);

  const setupMap = async () => {
    try {
      // Set map region to show both points
      const newRegion = {
        latitude: (userLocation.latitude + destination.latitude) / 2,
        longitude: (userLocation.longitude + destination.longitude) / 2,
        latitudeDelta: Math.abs(userLocation.latitude - destination.latitude) * 2,
        longitudeDelta: Math.abs(userLocation.longitude - destination.longitude) * 2,
      };
      setMapRegion(newRegion);
      
      // Get route details
      await getRouteDirections(userLocation, destination);
      
      setLoading(false);
    } catch (error) {
      console.error("Error setting up map:", error);
      setLoading(false);
    }
  };

  const getRouteDirections = async (startLoc, destinationLoc) => {
    try {
      const apiKey = "AIzaSyDDpFzLAd-65b3ouKzDDXKqt1VEQk3ZfOw";
      const response = await fetch(
        `https://maps.googleapis.com/maps/api/directions/json?origin=${startLoc.latitude},${startLoc.longitude}&destination=${destinationLoc.latitude},${destinationLoc.longitude}&key=${apiKey}`
      );
      
      const json = await response.json();
      
      if (json.routes?.[0]?.legs?.[0]) {
        const leg = json.routes[0].legs[0];
        setEta(leg.duration.text);
        setDistance(leg.distance.text);
        
        // Calculate dynamic pricing
        const km = leg.distance.value / 1000;
        const min = leg.duration.value / 60;
        calculatePricing(km, min);

        // Decode polyline
        const points = polyline.decode(json.routes[0].overview_polyline.points);
        setRouteCoordinates(points.map(point => ({
          latitude: point[0],
          longitude: point[1]
        })));
      }
    } catch (error) {
      console.error("Error getting directions:", error);
    }
  };

  const calculatePricing = (km, min) => {
    const basePrice = PRICING.BASE_FARE + 
                     (km * PRICING.PER_KM) + 
                     (min * PRICING.PER_MIN) + 
                     PRICING.SERVICE_FEE;
                     
    setRideOptions([
      {
        id: "uber-go",
        name: "Uber Go",
        time: eta,
        image: require("@/assets/images/uber-go.png"),
        price: `₹${Math.round(basePrice * PRICING.SURGE_MULTIPLIER)}`,
        multiplier: PRICING.SURGE_MULTIPLIER,
        type: 'Economy'
      },
      {
        id: "auto",
        name: "Auto",
        time: eta,
        image: require("@/assets/images/auto.png"),
        price: `₹${Math.round(basePrice * 0.9)}`,
        multiplier: 0.9,
        type: 'Budget'
      },
      {
        id: "uber-premier",
        name: "Uber Premier",
        time: eta,
        image: require("@/assets/images/uber-premier.png"),
        price: `₹${Math.round(basePrice * 1.5)}`,
        multiplier: 1.5,
        type: 'Premium'
      },
    ]);
  };

  const renderRideOption = ({ item }) => (
    <TouchableOpacity
      onPress={() => setSelectedRide(item.id)}
      className={`flex-row items-center justify-between p-4 rounded-lg border ${
        selectedRide === item.id ? "border-black" : "border-gray-200"
      }`}
    >
      <View className="flex-row items-center">
        <Image source={item.image} className="w-12 h-12 mr-4" />
        <View>
          <Text className="text-lg font-semibold">{item.name}</Text>
          <Text className="text-sm text-gray-500">{item.time}</Text>
          <Text className="text-xs text-blue-500 font-bold mt-1">
            {item.type} • {distance}
          </Text>
        </View>
      </View>
      <View className="items-end">
        <Text className="text-lg font-bold">{item.price}</Text>
      </View>
    </TouchableOpacity>
  );

  if (loading) {
    return (
      <View className="flex-1 justify-center items-center">
        <ActivityIndicator size="large" color="#000" />
        <Text className="mt-2">Calculating best route...</Text>
      </View>
    );
  }

  return (
    <View className="flex-1">
      <MapView
        provider={PROVIDER_GOOGLE}
        style={{ flex: 1 }}
        region={mapRegion}
      >
        {userLocation && (
          <Marker coordinate={userLocation} title="Your Location">
            <View className="bg-blue-500 p-2 rounded-full">
              <MaterialIcons name="my-location" size={16} color="#fff" />
            </View>
          </Marker>
        )}

        {destination.latitude && (
          <Marker coordinate={destination} title={destination.title}>
            <View className="bg-red-500 p-2 rounded-full">
              <MaterialIcons name="location-on" size={16} color="#fff" />
            </View>
          </Marker>
        )}

        {routeCoordinates.length > 0 && (
          <Polyline
            coordinates={routeCoordinates}
            strokeWidth={4}
            strokeColor="#007bff"
          />
        )}
      </MapView>

      {/* Ride Options Overlay */}
      <View className="absolute bottom-0 left-0 right-0 bg-white rounded-t-2xl p-4 shadow-lg">
        <Text className="text-xl font-bold mb-4">Select your ride</Text>

        <FlatList
          data={rideOptions}
          renderItem={renderRideOption}
          keyExtractor={(item) => item.id}
          ItemSeparatorComponent={() => <View style={{ height: 10 }} />}
        />

        <TouchableOpacity
          onPress={() => router.push({
            pathname: "/searching",
            params: {
              originTitle: "Current Location",
              originAddress: "Your current location",
              destTitle: destination.title,
              destAddress: destination.address,
              fare: rideOptions.find(r => r.id === selectedRide)?.price || '₹0',
              eta,
              distance
            }
          })}
          className="mt-4 px-6 py-3 bg-black rounded-lg"
        >
          <Text className="text-white text-center text-lg font-semibold">
            {selectedRide ? `Confirm ${rideOptions.find(r => r.id === selectedRide).name}` : "Select a Ride"}
          </Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

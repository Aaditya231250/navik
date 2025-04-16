import React, { useState, useEffect } from "react";
import {
  View,
  Text,
  TouchableOpacity,
  Image,
  FlatList,
  ActivityIndicator,
} from "react-native";
import { useRouter } from "expo-router";
import { MaterialIcons } from "@expo/vector-icons";
import MapView, { Marker, Polyline, PROVIDER_GOOGLE } from "react-native-maps";
import polyline from "@mapbox/polyline";

export default function RideDetailsScreen() {
  const router = useRouter();
  const [selectedRide, setSelectedRide] = useState(null);
  const [loading, setLoading] = useState(true);
  const [routeCoordinates, setRouteCoordinates] = useState([]);
  const [mapRegion, setMapRegion] = useState(null);
  const [nearbyDrivers, setNearbyDrivers] = useState([]);

  // Define origin and destination coordinates
  const origin = { latitude: 26.446653, longitude: 73.103040 };
  const destination = { latitude: 26.406325, longitude: 73.059779 };

  // Ride options data
  const rideOptions = [
    {
      id: "uber-go",
      name: "Uber Go",
      price: "₹170.71",
      time: "8:46pm - 4 min away",
      image: require("@/assets/images/uber-go.png"),
      isFaster: true,
    },
    {
      id: "auto",
      name: "Auto",
      price: "₹170.71",
      time: "8:46pm - 4 min away",
      image: require("@/assets/images/auto.png"),
      oldPrice: "₹189.71",
    },
    {
      id: "uber-premier",
      name: "Uber Premier",
      price: "₹223.63",
      time: "8:46pm - 5 min away",
      image: require("@/assets/images/uber-premier.png"),
    },
  ];

  // Initialize map and fetch route on component mount
  useEffect(() => {
    setupMap();
  }, []);

  // Setup map with origin, destination and route
  const setupMap = async () => {
    try {
      // Set initial map region to show both points
      setMapRegion({
        latitude: (origin.latitude + destination.latitude) / 2,
        longitude: (origin.longitude + destination.longitude) / 2,
        latitudeDelta: 0.05,
        longitudeDelta: 0.05,
      });
      
      // Get route between points
      await getRouteDirections(origin, destination);
      
      // Generate nearby drivers
      generateNearbyDrivers();
      
      setLoading(false);
    } catch (error) {
      console.error("Error setting up map:", error);
      setLoading(false);
    }
  };

  // Generate nearby drivers
  const generateNearbyDrivers = () => {
    // Create drivers along the route
    const drivers = [
      {
        id: "driver-1",
        coordinate: {
          latitude: 26.440653,
          longitude: 73.095040,
        },
        type: "uber-go",
      },
      {
        id: "driver-2",
        coordinate: {
          latitude: 26.430653,
          longitude: 73.085040,
        },
        type: "auto",
      },
      {
        id: "driver-3",
        coordinate: {
          latitude: 26.420653,
          longitude: 73.075040,
        },
        type: "uber-premier",
      },
    ];
    setNearbyDrivers(drivers);
  };

  const getRouteDirections = async (startLoc, destinationLoc) => {
    try {
      const apiKey = "AIzaSyDDpFzLAd-65b3ouKzDDXKqt1VEQk3ZfOw"; // Replace with your actual API key
      const response = await fetch(
        `https://maps.googleapis.com/maps/api/directions/json?origin=${startLoc.latitude},${startLoc.longitude}&destination=${destinationLoc.latitude},${destinationLoc.longitude}&key=${apiKey}`
      );
      
      const json = await response.json();
      
      if (json.routes && json.routes.length > 0) {
        const points = polyline.decode(json.routes[0].overview_polyline.points);
        const coords = points.map(point => ({
          latitude: point[0],
          longitude: point[1]
        }));
        
        setRouteCoordinates(coords);
      } else {
        console.error("No routes found");
        // Fallback to empty route
        setRouteCoordinates([]);
      }
    } catch (error) {
      console.error("Error getting directions:", error);
      setRouteCoordinates([]);
    }
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
          {item.isFaster && (
            <Text className="text-xs text-blue-500 font-bold mt-1">Faster</Text>
          )}
        </View>
      </View>
      <View className="items-end">
        <Text className="text-lg font-bold">{item.price}</Text>
        {item.oldPrice && (
          <Text className="text-sm text-gray-400 line-through">
            {item.oldPrice}
          </Text>
        )}
      </View>
    </TouchableOpacity>
  );

  if (loading) {
    return (
      <View className="flex-1 justify-center items-center">
        <ActivityIndicator size="large" color="#000" />
        <Text className="mt-2">Loading route...</Text>
      </View>
    );
  }

  return (
    <View className="flex-1">
      {/* Full Screen Map */}
      <MapView
        provider={PROVIDER_GOOGLE}
        style={{ flex: 1 }}
        region={mapRegion}
      >
        {/* Origin Marker */}
        <Marker coordinate={origin} title="Pickup Location">
          <View className="bg-blue-500 p-2 rounded-full">
            <MaterialIcons name="my-location" size={16} color="#fff" />
          </View>
        </Marker>
        
        {/* Destination Marker */}
        <Marker coordinate={destination} title="Destination">
          <View className="bg-red-500 p-2 rounded-full">
            <MaterialIcons name="location-on" size={16} color="#fff" />
          </View>
        </Marker>
        
        {/* Route Polyline */}
        {routeCoordinates.length > 0 && (
          <Polyline
            coordinates={routeCoordinates}
            strokeWidth={4}
            strokeColor="#007bff"
          />
        )}
        
        {/* Nearby Drivers */}
        {nearbyDrivers.map((driver) => (
        <Marker key={driver.id} coordinate={driver.coordinate}>
          <View className="bg-white p-1 rounded-full">
            {driver.type && (
              <Image 
                source={
                  driver.type === "uber-go" 
                    ? require("@/assets/images/uber-go.png") 
                    : driver.type === "auto" 
                    ? require("@/assets/images/auto.png")
                    : require("@/assets/images/uber-premier.png")
                }
                style={{ width: 20, height: 20 }}
                defaultSource={require("@/assets/images/uber-go.png")}
                />
                )}
              </View>
            </Marker>
          ))}
      </MapView>

      {/* Back Button */}
      <TouchableOpacity
        onPress={() => router.back()}
        className="absolute top-6 left-4 z-10 bg-white p-2 rounded-full shadow"
      >
        <MaterialIcons name="arrow-back" size={24} color="#000" />
      </TouchableOpacity>

      {/* Ride Options Overlay */}
      <View className="absolute bottom-0 left-0 right-0 bg-white rounded-t-2xl p-4 shadow-lg">
        <Text className="text-xl font-bold mb-4">Choose a trip</Text>

        <FlatList
          data={rideOptions}
          renderItem={renderRideOption}
          keyExtractor={(item) => item.id}
          ItemSeparatorComponent={() => <View style={{ height: 10 }} />}
        />

        {/* Payment Option */}
        <TouchableOpacity
          onPress={() => {}}
          className="flex-row items-center justify-between mt-6 px-4 py-3 bg-gray-100 rounded-lg"
        >
          <MaterialIcons name="payment" size={24} color="#000" />
          <Text className="text-gray-700 flex-grow ml-2">
            user.name@okhdfcbank
          </Text>
          <MaterialIcons name="keyboard-arrow-right" size={24} color="#000" />
        </TouchableOpacity>

        <TouchableOpacity
          onPress={() => {
            // Navigate to the searching screen with ride details
            router.push({
              pathname: "/searching",
              params: {
                originTitle: "562/11-A",
                originAddress: "Kaikondrahalli, Bengaluru, Karnataka",
                destTitle: "Third Wave Coffee",
                destAddress: "17th Cross Rd, PWD Quarters, 1st Sector, HSR Layout, Bengaluru, Karnataka",
                fare: "₹193.20"
              }
            });
          }}
          disabled={!selectedRide}
          className={`mt-4 px-6 py-3 rounded-lg ${
            selectedRide ? "bg-black" : "bg-gray-300"
          }`}
        >
          <Text className="text-white text-center text-lg font-semibold">
            {selectedRide
              ? `Choose ${
                  rideOptions.find((r) => r.id === selectedRide)?.name
                }`
              : "Select a Ride"}
          </Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

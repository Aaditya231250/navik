import React, { useState, useEffect, useRef } from "react";
import {
  View,
  Text,
  TouchableOpacity,
  Image,
  FlatList,
  ActivityIndicator,
  Alert,
  Dimensions
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
  const mapRef = useRef<MapView>(null);
  const [selectedRide, setSelectedRide] = useState(null);
  const [loading, setLoading] = useState(true);
  const [routeCoordinates, setRouteCoordinates] = useState([]);
  const [mapRegion, setMapRegion] = useState(null);
  const [eta, setEta] = useState('');
  const [distance, setDistance] = useState('');
  const [rideOptions, setRideOptions] = useState([]);
  const [userLocation, setUserLocation] = useState(null);
  const [destinationCoords, setDestinationCoords] = useState(null);
  
  // Array of nearby drivers near IIT Jodhpur
  const [nearbyDrivers, setNearbyDrivers] = useState([
    {
      id: 1,
      coordinate: {
        latitude: 26.2771,  // Near IIT Jodhpur campus
        longitude: 73.0108
      },
      name: "Driver 1"
    },
    {
      id: 2,
      coordinate: {
        latitude: 26.2763,
        longitude: 73.0155
      },
      name: "Driver 2"
    },
    {
      id: 3,
      coordinate: {
        latitude: 26.2792,
        longitude: 73.0075
      },
      name: "Driver 3"
    },
    {
      id: 4,
      coordinate: {
        latitude: 26.2740,
        longitude: 73.0128
      },
      name: "Driver 4"
    },
    {
      id: 5,
      coordinate: {
        latitude: 26.2782,
        longitude: 73.0197
      },
      name: "Driver 5"
    }
  ]);
  
  // Flag to prevent repeated map fitting
  const [mapFitted, setMapFitted] = useState(false);
  
  // Get screen dimensions
  const { width, height } = Dimensions.get('window');

  // Extract place IDs and fallback coordinates from params
  const originPlaceId = params.originPlaceId as string;
  const destPlaceId = params.destPlaceId as string;
  const destination = {
    placeId: destPlaceId,
    latitude: parseFloat(params.destLat as string) || 0,
    longitude: parseFloat(params.destLng as string) || 0,
    title: params.destTitle as string,
    address: params.destAddress as string
  };
  
  const origin = {
    placeId: originPlaceId,
    latitude: parseFloat(params.originLat as string) || 0,
    longitude: parseFloat(params.originLng as string) || 0,
    title: params.originTitle as string || "Current Location"
  };

  // Get user's current location if no origin coordinates provided
  useEffect(() => {
    (async () => {
      if (origin.placeId || (origin.latitude && origin.longitude)) {
        if (origin.latitude && origin.longitude) {
          setUserLocation({
            latitude: origin.latitude,
            longitude: origin.longitude
          });
        }
        return;
      }
      
      let { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== 'granted') {
        Alert.alert("Location Access", "Permission to access location was denied");
        return;
      }

      try {
        let location = await Location.getCurrentPositionAsync({});
        setUserLocation({
          latitude: location.coords.latitude,
          longitude: location.coords.longitude
        });
      } catch (error) {
        console.error("Error getting current location:", error);
        Alert.alert("Location Error", "Could not determine your current location");
      }
    })();
  }, []);

  // Calculate route when locations are available
  useEffect(() => {
    if (
      originPlaceId || 
      destPlaceId || 
      (userLocation && destination.latitude && destination.longitude)
    ) {
      setupMap();
    }
  }, [userLocation, destination, originPlaceId, destPlaceId]);

  const setupMap = async () => {
    try {
      // Set initial region
      if (
        (userLocation || (origin.latitude && origin.longitude)) && 
        (destination.latitude && destination.longitude)
      ) {
        const originLat = userLocation?.latitude || origin.latitude;
        const originLng = userLocation?.longitude || origin.longitude;
        
        // Set a looser initial region - this won't restrict panning
        const newRegion = {
          latitude: (originLat + destination.latitude) / 2,
          longitude: (originLng + destination.longitude) / 2,
          latitudeDelta: Math.abs(originLat - destination.latitude) * 1.5,
          longitudeDelta: Math.abs(originLng - destination.longitude) * 1.5,
        };
        setMapRegion(newRegion);
      }
      
      // Get route
      if (originPlaceId || destPlaceId) {
        await getRouteDirectionsWithPlaceId();
      } else if (userLocation && destination.latitude && destination.longitude) {
        await getRouteDirections(
          userLocation, 
          { latitude: destination.latitude, longitude: destination.longitude }
        );
      } else {
        throw new Error("Insufficient location data to calculate route");
      }
      
    } catch (error) {
      console.error("Error setting up map:", error);
      Alert.alert("Route Error", "Could not calculate the route between locations");
      setLoading(false);
    }
  };

  // Get directions using place IDs
  const getRouteDirectionsWithPlaceId = async () => {
    try {
      const apiKey = "AIzaSyDDpFzLAd-65b3ouKzDDXKqt1VEQk3ZfOw";
      let originParam = '';
      let destParam = '';
      
      // Prepare origin parameter
      if (originPlaceId) {
        originParam = `place_id:${originPlaceId}`;
      } else if (userLocation) {
        originParam = `${userLocation.latitude},${userLocation.longitude}`;
      } else if (origin.latitude && origin.longitude) {
        originParam = `${origin.latitude},${origin.longitude}`;
      }
      
      // Prepare destination parameter
      if (destPlaceId) {
        destParam = `place_id:${destPlaceId}`;
      } else if (destination.latitude && destination.longitude) {
        destParam = `${destination.latitude},${destination.longitude}`;
      }
      
      if (!originParam || !destParam) {
        throw new Error("Missing origin or destination for route calculation");
      }
      
      const response = await fetch(
        `https://maps.googleapis.com/maps/api/directions/json?origin=${originParam}&destination=${destParam}&key=${apiKey}`
      );
      
      const json = await response.json();
      processDirectionsResponse(json);
    } catch (error) {
      console.error("Error getting directions with place ID:", error);
      throw error;
    }
  };

  // Get directions using coordinates
  const getRouteDirections = async (startLoc, destinationLoc) => {
    try {
      const apiKey = "AIzaSyDDpFzLAd-65b3ouKzDDXKqt1VEQk3ZfOw";
      const response = await fetch(
        `https://maps.googleapis.com/maps/api/directions/json?origin=${startLoc.latitude},${startLoc.longitude}&destination=${destinationLoc.latitude},${destinationLoc.longitude}&key=${apiKey}`
      );
      
      const json = await response.json();
      processDirectionsResponse(json);
    } catch (error) {
      console.error("Error getting directions:", error);
      throw error;
    }
  };
  
  // Process directions response
  const processDirectionsResponse = (json) => {
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
      const coordinates = points.map(point => ({
        latitude: point[0],
        longitude: point[1]
      }));
      
      setRouteCoordinates(coordinates);
      
      // Get start and end coordinates from route for markers
      if (!userLocation && leg.start_location) {
        setUserLocation({
          latitude: leg.start_location.lat,
          longitude: leg.start_location.lng
        });
      }
      
      if ((!destination.latitude || !destination.longitude) && leg.end_location) {
        setDestinationCoords({
          latitude: leg.end_location.lat,
          longitude: leg.end_location.lng
        });
      }
      
      // Fit map only once - this is key to allow free panning afterward
      setTimeout(() => {
        if (mapRef.current && coordinates.length > 0 && !mapFitted) {
          // Calculate if route is primarily north-south
          const latitudes = coordinates.map(c => c.latitude);
          const longitudes = coordinates.map(c => c.longitude);
          const latDiff = Math.max(...latitudes) - Math.min(...latitudes);
          const lngDiff = Math.max(...longitudes) - Math.min(...longitudes);
          const isNorthSouth = latDiff > lngDiff;
          
          const originPoint = userLocation || { 
            latitude: leg.start_location.lat, 
            longitude: leg.start_location.lng 
          };
          
          const destPoint = destinationCoords || { 
            latitude: leg.end_location.lat, 
            longitude: leg.end_location.lng 
          };
          
          const allCoordinates = [originPoint, ...coordinates, destPoint];
          
          // Determine paddings based on route orientation and screen size
          const faresHeight = height * 0.45; // 45% of screen for fares panel
          const topPadding = isNorthSouth ? height * 0.25 : height * 0.15;
          
          mapRef.current.fitToCoordinates(allCoordinates, {
            edgePadding: { 
              top: topPadding, 
              right: width * 0.1, 
              bottom: faresHeight, 
              left: width * 0.1 
            },
            animated: true
          });
          
          // Mark as fitted so we don't override user panning
          setMapFitted(true);
        }
      }, 500);
      
      setLoading(false);
    } else {
      Alert.alert("Route Error", "No route found between these locations");
      setLoading(false);
      throw new Error("No route found");
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
        ref={mapRef}
        provider={PROVIDER_GOOGLE}
        style={{ flex: 1 }}
        initialRegion={mapRegion}
        showsUserLocation={false}
        rotateEnabled={true}
        scrollEnabled={true}
        zoomEnabled={true}
        pitchEnabled={true}
      >
        {userLocation && (
          <Marker coordinate={userLocation} title={origin.title || "Your Location"}>
            <View className="bg-blue-500 p-2 rounded-full">
              <MaterialIcons name="my-location" size={16} color="#fff" />
            </View>
          </Marker>
        )}

        {(destination.latitude && destination.longitude) ? (
          <Marker 
            coordinate={{
              latitude: destination.latitude,
              longitude: destination.longitude
            }} 
            title={destination.title}
          >
            <View className="bg-red-500 p-2 rounded-full">
              <MaterialIcons name="location-on" size={16} color="#fff" />
            </View>
          </Marker>
        ) : destinationCoords && (
          <Marker coordinate={destinationCoords} title={destination.title}>
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
        
        {/* Render nearby drivers */}
        {nearbyDrivers.map(driver => (
          <Marker
            key={driver.id}
            coordinate={driver.coordinate}
            title={driver.name}
          >
            <View className="bg-red-500 p-2 rounded-full">
              <MaterialIcons name="local-taxi" size={16} color="#fff" />
            </View>
          </Marker>
        ))}
      </MapView>

      {/* Recenter button for manual navigation */}
      <TouchableOpacity
        className="absolute top-4 right-4 bg-white p-3 rounded-full shadow-lg"
        onPress={() => {
          if (mapRef.current && routeCoordinates.length > 0 && userLocation && (destinationCoords || destination)) {
            const destPoint = destination.latitude && destination.longitude
              ? { latitude: destination.latitude, longitude: destination.longitude }
              : destinationCoords;
              
            if (destPoint) {
              // Create route coordinates array
              const allCoordinates = [userLocation, ...routeCoordinates, destPoint];
              
              // Dynamic padding based on screen size
              const faresHeight = height * 0.45; // 45% of screen for fares panel
              
              // Refit the map
              mapRef.current.fitToCoordinates(allCoordinates, {
                edgePadding: { 
                  top: height * 0.25, 
                  right: width * 0.1, 
                  bottom: faresHeight, 
                  left: width * 0.1 
                },
                animated: true
              });
            }
          }
        }}
      >
        <MaterialIcons name="my-location" size={24} color="#007AFF" />
      </TouchableOpacity>

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
          disabled={!selectedRide}
          onPress={() => router.push({
            pathname: "/searching",
            params: {
              originTitle: origin.title || "Current Location",
              originAddress: "Your current location",
              destTitle: destination.title,
              destAddress: destination.address,
              fare: rideOptions.find(r => r.id === selectedRide)?.price || '₹0',
              eta,
              distance
            }
          })}
          className={`mt-4 px-6 py-3 ${selectedRide ? "bg-black" : "bg-gray-300"} rounded-lg`}
        >
          <Text className="text-white text-center text-lg font-semibold">
            {selectedRide ? `Confirm ${rideOptions.find(r => r.id === selectedRide)?.name}` : "Select a Ride"}
          </Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

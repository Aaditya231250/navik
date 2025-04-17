import React, { useEffect, useState, useRef } from "react";
import { View, Text, StyleSheet, TouchableOpacity } from "react-native";
import { useRouter, useLocalSearchParams } from "expo-router";
import MapView, { Marker, Polyline, PROVIDER_GOOGLE } from "react-native-maps";
import { MaterialIcons, FontAwesome } from "@expo/vector-icons";

export default function InProgressScreen() {
  const router = useRouter();
  const params = useLocalSearchParams();
  const mapRef = useRef(null);
  
  // State for locations and journey progress
  const [journeyProgress, setJourneyProgress] = useState(0);
  const [currentUserLocation, setCurrentUserLocation] = useState(null);
  const [routeCoordinates, setRouteCoordinates] = useState([]);
  
  // Extract location information from params
  const origin = {
    title: params.originTitle || "562/11-A",
    address: params.originAddress || "Kaikondrahalli, Bengaluru, Karnataka",
    latitude: parseFloat(params.originLat as string) || 26.2720,
    longitude: parseFloat(params.originLng as string) || 73.0120,
  };
  
  const destination = {
    title: params.destTitle || "Third Wave Coffee",
    address: params.destAddress || "17th Cross Rd, PWD Quarters, 1st Sector, HSR Layout, Bengaluru, Karnataka",
    latitude: parseFloat(params.destLat as string) || 26.2650,
    longitude: parseFloat(params.destLng as string) || 73.0200,
  };
  
  const fare = params.fare || "₹193.20";
  const eta = params.eta || "15 mins";
  const distance = params.distance || "5.2 km";
  
  // Extract driver information from params
  const selectedDriver = {
    name: params.driverName || "Driver 1",
    carModel: params.driverCarModel || "Toyota Camry",
    coordinate: { latitude: origin.latitude, longitude: origin.longitude },
    heading: 0
  };

  useEffect(() => {
    // Set initial user and driver location to origin
    setCurrentUserLocation({
      latitude: origin.latitude,
      longitude: origin.longitude
    });
    
    // Generate journey route
    const journeyRoute = generateRouteCoordinates(
      { latitude: origin.latitude, longitude: origin.longitude },
      { latitude: destination.latitude, longitude: destination.longitude },
      20
    );
    setRouteCoordinates(journeyRoute);
    
    // Start journey animation
    startJourneyAnimation(journeyRoute);
    
    // Fit map to show route
    if (mapRef.current) {
      mapRef.current.fitToCoordinates(
        [
          { latitude: origin.latitude, longitude: origin.longitude },
          { latitude: destination.latitude, longitude: destination.longitude }
        ],
        {
          edgePadding: { top: 100, right: 50, bottom: 300, left: 50 },
          animated: true
        }
      );
    }
  }, []);

  // Generate route coordinates between two points
  const generateRouteCoordinates = (start, end, numPoints) => {
    const coordinates = [];
    for (let i = 0; i <= numPoints; i++) {
      const fraction = i / numPoints;
      coordinates.push({
        latitude: start.latitude + (end.latitude - start.latitude) * fraction,
        longitude: start.longitude + (end.longitude - start.longitude) * fraction
      });
    }
    return coordinates;
  };

  // Simulate journey animation
  const startJourneyAnimation = (journeyRoute) => {
    let journeyIndex = 0;
    const totalPoints = journeyRoute.length;
    
    const journeyInterval = setInterval(() => {
      if (journeyIndex < totalPoints - 1) {
        journeyIndex++;
        
        // Update driver and passenger location (they move together)
        selectedDriver.coordinate = journeyRoute[journeyIndex];
        setCurrentUserLocation(journeyRoute[journeyIndex]);
        
        // Update journey progress
        setJourneyProgress(journeyIndex / (totalPoints - 1));
        
        // Calculate heading
        if (journeyIndex > 0) {
          const prevPoint = journeyRoute[journeyIndex - 1];
          const currentPoint = journeyRoute[journeyIndex];
          const heading = Math.atan2(
            currentPoint.longitude - prevPoint.longitude,
            currentPoint.latitude - prevPoint.latitude
          ) * (180 / Math.PI);
          
          selectedDriver.heading = heading;
        }
        
        // When journey is complete
        if (journeyIndex === totalPoints - 1) {
          clearInterval(journeyInterval);
          
          // Navigate to completed screen
          setTimeout(() => {
            router.push({
              pathname: "/ride-journey/completed",
              params: {
                ...params,
                driverName: selectedDriver.name
              }
            });
          }, 1000);
        }
      }
    }, 1000);
    
    return () => clearInterval(journeyInterval);
  };

  return (
    <View style={styles.container}>
      {/* Map View */}
      <MapView
        ref={mapRef}
        provider={PROVIDER_GOOGLE}
        style={styles.map}
        initialRegion={{
          latitude: origin.latitude,
          longitude: origin.longitude,
          latitudeDelta: 0.02,
          longitudeDelta: 0.02,
        }}
      >
        {/* Current User Location during journey */}
        {currentUserLocation && (
          <Marker
            coordinate={currentUserLocation}
            title="Your location"
          >
            <View style={styles.userLocationMarker}>
              <MaterialIcons name="person-pin-circle" size={24} color="#007bff" />
            </View>
          </Marker>
        )}

        {/* Destination Marker */}
        <Marker
          coordinate={{
            latitude: destination.latitude,
            longitude: destination.longitude,
          }}
          title={destination.title}
        >
          <View style={styles.destinationMarker}>
            <MaterialIcons name="flag" size={24} color="#d32f2f" />
          </View>
        </Marker>

        {/* Selected driver marker */}
        <Marker
          coordinate={selectedDriver.coordinate}
          title={selectedDriver.name}
          rotation={selectedDriver.heading}
        >
          <View style={styles.driverMarker}>
            <MaterialIcons 
              name="directions-car"
              size={18} 
              color="#fff" 
            />
          </View>
        </Marker>

        {/* Route polyline */}
        {routeCoordinates.length > 0 && (
          <Polyline
            coordinates={routeCoordinates}
            strokeWidth={4}
            strokeColor="#4CAF50"
          />
        )}
      </MapView>

      {/* Origin Label */}
      <View style={styles.originLabel}>
        <Text style={styles.originText}>To {destination.title}</Text>
      </View>

      {/* Journey Info Card */}
      <View style={styles.bottomJourneyCard}>
        <View style={styles.journeyProgressContainer}>
          <View style={styles.journeyProgressBar}>
            <View style={[styles.journeyProgressFill, { width: `${journeyProgress * 100}%` }]} />
          </View>
          <Text style={styles.journeyStatusText}>
            {journeyProgress < 0.1 ? 'Starting trip' : 
             journeyProgress < 0.9 ? 'On the way to destination' : 
             'Arriving soon'}
          </Text>
        </View>

        <View style={styles.journeyDetails}>
          <View style={styles.journeyDetailItem}>
            <MaterialIcons name="access-time" size={20} color="#666" />
            <Text style={styles.journeyDetailText}>ETA: {eta}</Text>
          </View>
          <View style={styles.journeyDetailItem}>
            <MaterialIcons name="straighten" size={20} color="#666" />
            <Text style={styles.journeyDetailText}>{distance}</Text>
          </View>
          <View style={styles.journeyDetailItem}>
            <MaterialIcons name="attach-money" size={20} color="#666" />
            <Text style={styles.journeyDetailText}>{fare}</Text>
          </View>
        </View>

        <View style={styles.driverInfoMini}>
          <FontAwesome name="user-circle" size={40} color="#007bff" />
          <View style={styles.driverDetailsMini}>
            <Text style={styles.driverNameMini}>{selectedDriver.name}</Text>
            <Text style={styles.carInfoMini}>{selectedDriver.carModel}</Text>
          </View>
          <View style={styles.actionButtonsMini}>
            <TouchableOpacity style={styles.actionButtonMini}>
              <MaterialIcons name="phone" size={20} color="#007bff" />
            </TouchableOpacity>
          </View>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  // ... Include all necessary styles from the original component
  container: {
    flex: 1,
    backgroundColor: "#fff",
  },
  map: {
    flex: 1,
  },
  originLabel: {
    position: "absolute",
    top: 60,
    right: 20,
    backgroundColor: "white",
    padding: 8,
    borderRadius: 4,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.2,
    shadowRadius: 2,
    elevation: 2,
  },
  originText: {
    fontWeight: "bold",
  },
  userLocationMarker: {
    alignItems: "center",
    justifyContent: "center",
  },
  destinationMarker: {
    alignItems: "center",
    justifyContent: "center",
  },
  driverMarker: {
    backgroundColor: "#4CAF50",
    padding: 6,
    borderRadius: 12,
    alignItems: "center",
    justifyContent: "center",
  },
  bottomJourneyCard: {
    position: "absolute",
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: "white",
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    padding: 15,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: -2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 5,
  },
  journeyProgressContainer: {
    marginBottom: 15,
  },
  journeyProgressBar: {
    height: 6,
    backgroundColor: "#e0e0e0",
    borderRadius: 3,
    marginBottom: 8,
  },
  journeyProgressFill: {
    height: 6,
    backgroundColor: "#4CAF50",
    borderRadius: 3,
  },
  journeyStatusText: {
    fontSize: 14,
    fontWeight: "500",
    color: "#4CAF50",
  },
  journeyDetails: {
    flexDirection: "row",
    justifyContent: "space-between",
    marginBottom: 15,
    paddingBottom: 15,
    borderBottomWidth: 1,
    borderBottomColor: "#e0e0e0",
  },
  journeyDetailItem: {
    flexDirection: "row",
    alignItems: "center",
  },
  journeyDetailText: {
    marginLeft: 5,
    fontSize: 14,
    fontWeight: "500",
  },
  driverInfoMini: {
    flexDirection: "row",
    alignItems: "center",
  },
  driverDetailsMini: {
    flex: 1,
    marginLeft: 10,
  },
  driverNameMini: {
    fontSize: 16,
    fontWeight: "bold",
  },
  carInfoMini: {
    fontSize: 14,
    color: "#666",
  },
  actionButtonsMini: {
    flexDirection: "row",
  },
  actionButtonMini: {
    padding: 8,
  },
});

import React from "react";
import { View, Text, StyleSheet, TouchableOpacity } from "react-native";
import { useRouter, useLocalSearchParams } from "expo-router";
import MapView, { Marker, PROVIDER_GOOGLE } from "react-native-maps";
import { MaterialIcons, FontAwesome } from "@expo/vector-icons";

export default function CompletedScreen() {
  const router = useRouter();
  const params = useLocalSearchParams();
  
  // Extract location information from params
  const destination = {
    title: params.destTitle || "Third Wave Coffee",
    address: params.destAddress || "17th Cross Rd, PWD Quarters, 1st Sector, HSR Layout, Bengaluru, Karnataka",
    latitude: parseFloat(params.destLat as string) || 26.2650,
    longitude: parseFloat(params.destLng as string) || 73.0200,
  };
  
  const fare = params.fare || "₹193.20";
  const eta = params.eta || "15 mins";
  const distance = params.distance || "5.2 km";
  const driverName = params.driverName || "Driver";

  return (
    <View style={styles.container}>
      {/* Map View showing destination */}
      <MapView
        provider={PROVIDER_GOOGLE}
        style={styles.map}
        initialRegion={{
          latitude: destination.latitude,
          longitude: destination.longitude,
          latitudeDelta: 0.01,
          longitudeDelta: 0.01,
        }}
      >
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
      </MapView>

      {/* Origin Label */}
      <View style={styles.originLabel}>
        <Text style={styles.originText}>Trip Completed</Text>
      </View>

      {/* Completion Card */}
      <View style={styles.completeCard}>
        <View style={styles.completedIconContainer}>
          <MaterialIcons name="check-circle" size={60} color="#4CAF50" />
        </View>
        <Text style={styles.completedTitleText}>Trip Completed</Text>
        <Text style={styles.completedSubText}>Thank you for riding with us!</Text>
        
        <View style={styles.completedDetails}>
          <View style={styles.completedItem}>
            <MaterialIcons name="location-on" size={24} color="#666" />
            <Text style={styles.completedItemText}>{distance}</Text>
          </View>
          <View style={styles.completedItem}>
            <MaterialIcons name="access-time" size={24} color="#666" />
            <Text style={styles.completedItemText}>{eta}</Text>
          </View>
          <View style={styles.completedItem}>
            <MaterialIcons name="attach-money" size={24} color="#666" />
            <Text style={styles.completedItemText}>{fare}</Text>
          </View>
        </View>
        
        <View style={styles.rateDriverSection}>
          <Text style={styles.rateDriverText}>Rate your driver {driverName}</Text>
          <View style={styles.starsContainer}>
            {[1, 2, 3, 4, 5].map(star => (
              <TouchableOpacity key={star}>
                <FontAwesome name="star" size={30} color="#FFD700" />
              </TouchableOpacity>
            ))}
          </View>
        </View>
        
        <TouchableOpacity 
          style={styles.homeButton}
          onPress={() => router.replace("/")}
        >
          <Text style={styles.homeButtonText}>Back to Home</Text>
        </TouchableOpacity>
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
  destinationMarker: {
    alignItems: "center",
    justifyContent: "center",
  },
  completeCard: {
    position: "absolute",
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: "white",
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    padding: 20,
    alignItems: "center",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: -2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 5,
  },
  completedIconContainer: {
    marginBottom: 15,
  },
  completedTitleText: {
    fontSize: 22,
    fontWeight: "bold",
    marginBottom: 5,
  },
  completedSubText: {
    fontSize: 16,
    color: "#666",
    marginBottom: 20,
  },
  completedDetails: {
    flexDirection: "row",
    justifyContent: "space-around",
    width: "100%",
    marginBottom: 25,
  },
  completedItem: {
    alignItems: "center",
  },
  completedItemText: {
    marginTop: 5,
    fontSize: 16,
    fontWeight: "500",
  },
  rateDriverSection: {
    width: "100%",
    marginBottom: 20,
  },
  rateDriverText: {
    fontSize: 16,
    fontWeight: "bold",
    textAlign: "center",
    marginBottom: 10,
  },
  starsContainer: {
    flexDirection: "row",
    justifyContent: "center",
    marginBottom: 20,
  },
  homeButton: {
    backgroundColor: "#007bff",
    paddingVertical: 12,
    paddingHorizontal: 30,
    borderRadius: 25,
  },
  homeButtonText: {
    color: "white",
    fontSize: 16,
    fontWeight: "bold",
  },
});

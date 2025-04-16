// app/driver-assigned/index.tsx
import React, { useState } from "react";
import {
  View,
  Text,
  Image,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  Platform,
  StatusBar,
  Linking,
} from "react-native";
import { useRouter, useLocalSearchParams } from "expo-router";
import MapView, { Marker, Polyline, PROVIDER_GOOGLE } from "react-native-maps";
import { MaterialIcons, FontAwesome, Ionicons } from "@expo/vector-icons";

export default function DriverAssignedScreen() {
  const router = useRouter();
  const params = useLocalSearchParams();
  const [showTripDetails, setShowTripDetails] = useState(false);

  // Dummy driver data (will be replaced with backend data)
  const driver = {
    name: "Rishi Raj Bhagat",
    rating: 4.8,
    phone: "+91 98765 43210",
    car: {
      model: "Bajaj Auto",
      color: "Yellow Balak",
      plate: "KA 01 AB 1234",
    },
    eta: "4 min",
    photo: require("@/assets/images/auto.png"),
  };

  // Trip details from params or dummy data
  const tripDetails = {
    originTitle: params.originTitle || "562/11-A",
    originAddress: params.originAddress || "Kaikondrahalli, Bengaluru, Karnataka",
    destTitle: params.destTitle || "Third Wave Coffee",
    destAddress: params.destAddress || "17th Cross Rd, PWD Quarters, 1st Sector, HSR Layout, Bengaluru, Karnataka",
    fare: params.fare || "₹193.20",
  };

  // Dummy coordinates (will be replaced with actual data)
  const origin = { latitude: 12.9259, longitude: 77.6229 };
  const destination = { latitude: 12.9139, longitude: 77.6380 };
  const driverLocation = { latitude: 12.9229, longitude: 77.6209 };

  // Dummy route coordinates (will be replaced with actual route)
  const routeCoordinates = [
    origin,
    { latitude: 12.9239, longitude: 77.6219 },
    { latitude: 12.9219, longitude: 77.6239 },
    { latitude: 12.9199, longitude: 77.6259 },
    { latitude: 12.9179, longitude: 77.6279 },
    { latitude: 12.9159, longitude: 77.6299 },
    { latitude: 12.9139, longitude: 77.6380 },
  ];

  const handleCall = () => {
    Linking.openURL(`tel:${driver.phone}`);
  };

  const handleCancel = () => {
    // Show confirmation dialog before canceling
    alert("Are you sure you want to cancel this ride?");
    // In a real app, you would make an API call to cancel the ride
    router.replace("/home");
  };

  return (
    <SafeAreaView style={styles.container}>
      {/* Map View */}
      <MapView
        provider={PROVIDER_GOOGLE}
        style={styles.map}
        initialRegion={{
          latitude: origin.latitude,
          longitude: origin.longitude,
          latitudeDelta: 0.02,
          longitudeDelta: 0.02,
        }}
      >
        {/* Origin Marker */}
        <Marker coordinate={origin} title={tripDetails.originTitle}>
          <View style={styles.originMarker}>
            <MaterialIcons name="my-location" size={16} color="#fff" />
          </View>
        </Marker>

        {/* Destination Marker */}
        <Marker coordinate={destination} title={tripDetails.destTitle}>
          <View style={styles.destMarker}>
            <MaterialIcons name="location-on" size={16} color="#fff" />
          </View>
        </Marker>

        {/* Driver Marker */}
        <Marker coordinate={driverLocation} title={`${driver.name}'s car`}>
          <View style={styles.driverMarker}>
            <MaterialIcons name="directions-car" size={16} color="#000" />
          </View>
        </Marker>

        {/* Route Polyline */}
        <Polyline
          coordinates={routeCoordinates}
          strokeWidth={4}
          strokeColor="#007bff"
        />
      </MapView>

      {/* Driver Card */}
      <View style={styles.driverCard}>
        <View style={styles.driverHeader}>
          <Text style={styles.driverEta}>{driver.eta} away</Text>
          <TouchableOpacity 
            style={styles.tripDetailsToggle}
            onPress={() => setShowTripDetails(!showTripDetails)}
          >
            <Text style={styles.tripDetailsText}>
              {showTripDetails ? "Hide trip details" : "Show trip details"}
            </Text>
            <MaterialIcons 
              name={showTripDetails ? "keyboard-arrow-up" : "keyboard-arrow-down"} 
              size={24} 
              color="#000" 
            />
          </TouchableOpacity>
        </View>

        {/* Trip Details (Collapsible) */}
        {showTripDetails && (
          <View style={styles.tripDetails}>
            <View style={styles.locationItem}>
              <MaterialIcons name="location-on" size={20} color="#000" />
              <View style={styles.locationText}>
                <Text style={styles.locationTitle}>{tripDetails.originTitle}</Text>
                <Text style={styles.locationAddress}>{tripDetails.originAddress}</Text>
              </View>
            </View>
            
            <View style={styles.locationItem}>
              <MaterialIcons name="flag" size={20} color="#000" />
              <View style={styles.locationText}>
                <Text style={styles.locationTitle}>{tripDetails.destTitle}</Text>
                <Text style={styles.locationAddress}>{tripDetails.destAddress}</Text>
              </View>
            </View>
            
            <View style={styles.fareContainer}>
              <MaterialIcons name="attach-money" size={20} color="#000" />
              <Text style={styles.fareText}>{tripDetails.fare}</Text>
              <Text style={styles.fareSubtext}>Cash</Text>
            </View>
          </View>
        )}

        <View style={styles.divider} />

        {/* Driver Info */}
        <View style={styles.driverInfo}>
          <Image source={driver.photo} style={styles.driverPhoto} />
          
          <View style={styles.driverDetails}>
            <Text style={styles.driverName}>{driver.name}</Text>
            <View style={styles.ratingContainer}>
              <FontAwesome name="star" size={14} color="#FFD700" />
              <Text style={styles.ratingText}>{driver.rating}</Text>
            </View>
          </View>
          
          <TouchableOpacity style={styles.callButton} onPress={handleCall}>
            <Ionicons name="call" size={24} color="white" />
          </TouchableOpacity>
        </View>

        {/* Car Details */}
        <View style={styles.carDetails}>
          <Text style={styles.carModel}>{driver.car.model}</Text>
          <Text style={styles.carPlate}>{driver.car.color} • {driver.car.plate}</Text>
        </View>

        {/* Cancel Button */}
        <TouchableOpacity style={styles.cancelButton} onPress={handleCancel}>
          <Text style={styles.cancelText}>Cancel Ride</Text>
        </TouchableOpacity>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff",
    paddingTop: Platform.OS === "android" ? StatusBar.currentHeight : 0,
  },
  map: {
    flex: 1,
  },
  originMarker: {
    backgroundColor: "#4285F4",
    padding: 8,
    borderRadius: 50,
  },
  destMarker: {
    backgroundColor: "#EA4335",
    padding: 8,
    borderRadius: 50,
  },
  driverMarker: {
    backgroundColor: "#FFFFFF",
    padding: 8,
    borderRadius: 50,
    borderWidth: 1,
    borderColor: "#000",
  },
  driverCard: {
    position: "absolute",
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: "white",
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    padding: 20,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: -2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 5,
  },
  driverHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 10,
  },
  driverEta: {
    fontSize: 18,
    fontWeight: "bold",
  },
  tripDetailsToggle: {
    flexDirection: "row",
    alignItems: "center",
  },
  tripDetailsText: {
    color: "#000",
    marginRight: 5,
  },
  tripDetails: {
    marginVertical: 10,
    backgroundColor: "#f9f9f9",
    padding: 10,
    borderRadius: 10,
  },
  locationItem: {
    flexDirection: "row",
    alignItems: "flex-start",
    marginBottom: 10,
  },
  locationText: {
    marginLeft: 10,
    flex: 1,
  },
  locationTitle: {
    fontWeight: "bold",
  },
  locationAddress: {
    color: "#666",
    fontSize: 12,
  },
  fareContainer: {
    flexDirection: "row",
    alignItems: "center",
    borderTopWidth: 1,
    borderTopColor: "#eee",
    paddingTop: 10,
  },
  fareText: {
    fontWeight: "bold",
    marginLeft: 10,
  },
  fareSubtext: {
    color: "#666",
    marginLeft: 5,
  },
  divider: {
    height: 1,
    backgroundColor: "#eee",
    marginVertical: 10,
  },
  driverInfo: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 10,
  },
  driverPhoto: {
    width: 50,
    height: 50,
    borderRadius: 25,
    marginRight: 15,
  },
  driverDetails: {
    flex: 1,
  },
  driverName: {
    fontSize: 18,
    fontWeight: "bold",
  },
  ratingContainer: {
    flexDirection: "row",
    alignItems: "center",
    marginTop: 5,
  },
  ratingText: {
    marginLeft: 5,
  },
  callButton: {
    backgroundColor: "#34A853",
    width: 40,
    height: 40,
    borderRadius: 20,
    justifyContent: "center",
    alignItems: "center",
  },
  carDetails: {
    marginBottom: 15,
  },
  carModel: {
    fontSize: 16,
    fontWeight: "500",
  },
  carPlate: {
    color: "#666",
  },
  cancelButton: {
    borderWidth: 1,
    borderColor: "#EA4335",
    borderRadius: 8,
    padding: 12,
    alignItems: "center",
  },
  cancelText: {
    color: "#EA4335",
    fontWeight: "bold",
  },
});

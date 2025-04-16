import React, { useEffect, useState } from "react";
import { View, Text, Image, StyleSheet, Animated, Easing } from "react-native";
import { useRouter, useLocalSearchParams } from "expo-router";
import MapView, { Marker, PROVIDER_GOOGLE } from "react-native-maps";
import { MaterialIcons } from "@expo/vector-icons";

export default function SearchingScreen() {
  const router = useRouter();
  const params = useLocalSearchParams();
  const [pulseAnim] = useState(new Animated.Value(0.8));
  
  // Extract location information from params or use defaults
  const origin = {
    title: params.originTitle || "562/11-A",
    address: params.originAddress || "Kaikondrahalli, Bengaluru, Karnataka",
    latitude: parseFloat(params.originLat) || 12.9259,
    longitude: parseFloat(params.originLng) || 77.6229,
  };
  
  const destination = {
    title: params.destTitle || "Third Wave Coffee",
    address: params.destAddress || "17th Cross Rd, PWD Quarters, 1st Sector, HSR Layout, Bengaluru, Karnataka",
    latitude: parseFloat(params.destLat) || 12.9139,
    longitude: parseFloat(params.destLng) || 77.6380,
  };
  
  const fare = params.fare || "₹193.20";

  // Animation for the pulsing effect
  useEffect(() => {
    Animated.loop(
      Animated.sequence([
        Animated.timing(pulseAnim, {
          toValue: 1.2,
          duration: 800,
          easing: Easing.ease,
          useNativeDriver: true,
        }),
        Animated.timing(pulseAnim, {
          toValue: 0.8,
          duration: 800,
          easing: Easing.ease,
          useNativeDriver: true,
        }),
      ])
    ).start();

    // In the useEffect of searching/index.tsx
    const timer = setTimeout(() => {
      router.push({
        pathname: "/driver-assigned",
        params: {
          // Pass through the same params
          originTitle: params.originTitle,
          originAddress: params.originAddress,
          destTitle: params.destTitle,
          destAddress: params.destAddress,
          fare: params.fare
        }
      });
    }, 5000);
 
    return () => clearTimeout(timer);
  }, []);

  return (
    <View style={styles.container}>
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
        <Marker
          coordinate={{
            latitude: origin.latitude,
            longitude: origin.longitude,
          }}
          title={origin.title}
        />
      </MapView>

      {/* Origin Label */}
      <View style={styles.originLabel}>
        <Text style={styles.originText}>From {origin.title}</Text>
      </View>

      {/* Bottom Card */}
      <View style={styles.bottomCard}>
        <Text style={styles.searchingText}>Looking for nearby drivers</Text>
        
        <View style={styles.divider} />
        
        {/* Animated Car */}
        <View style={styles.carContainer}>
          <View style={styles.pulseCircle}>
            <Animated.View
              style={[
                styles.pulseCircleInner,
                {
                  transform: [{ scale: pulseAnim }],
                },
              ]}
            />
            <Image
              source={require("@/assets/images/car.png")}
              style={styles.carImage}
              resizeMode="contain"
            />
          </View>
        </View>

        {/* Location Details */}
        <View style={styles.locationDetails}>
          <View style={styles.locationItem}>
            <MaterialIcons name="location-on" size={24} color="black" />
            <View style={styles.locationText}>
              <Text style={styles.locationTitle}>{origin.title}</Text>
              <Text style={styles.locationAddress}>{origin.address}</Text>
            </View>
          </View>
          
          <View style={styles.locationItem}>
            <MaterialIcons name="flag" size={24} color="black" />
            <View style={styles.locationText}>
              <Text style={styles.locationTitle}>{destination.title}</Text>
              <Text style={styles.locationAddress}>{destination.address}</Text>
            </View>
          </View>
          
          <View style={styles.fareContainer}>
            <MaterialIcons name="attach-money" size={24} color="black" />
            <Text style={styles.fareText}>{fare}</Text>
            <Text style={styles.fareSubtext}>Cash Cash</Text>
          </View>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
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
  bottomCard: {
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
  searchingText: {
    fontSize: 18,
    fontWeight: "bold",
    textAlign: "center",
    marginBottom: 10,
  },
  divider: {
    height: 1,
    backgroundColor: "#e0e0e0",
    marginVertical: 10,
  },
  carContainer: {
    alignItems: "center",
    justifyContent: "center",
    marginVertical: 20,
  },
  pulseCircle: {
    width: 100,
    height: 100,
    borderRadius: 50,
    alignItems: "center",
    justifyContent: "center",
    position: "relative",
  },
  pulseCircleInner: {
    width: 100,
    height: 100,
    borderRadius: 50,
    backgroundColor: "rgba(173, 216, 230, 0.5)",
    position: "absolute",
  },
  carImage: {
    width: 60,
    height: 60,
    zIndex: 1,
  },
  locationDetails: {
    marginTop: 10,
  },
  locationItem: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 15,
    paddingHorizontal: 10,
  },
  locationText: {
    marginLeft: 10,
    flex: 1,
  },
  locationTitle: {
    fontWeight: "bold",
    fontSize: 16,
  },
  locationAddress: {
    color: "#666",
    fontSize: 14,
  },
  fareContainer: {
    flexDirection: "row",
    alignItems: "center",
    borderTopWidth: 1,
    borderTopColor: "#e0e0e0",
    paddingTop: 15,
    paddingHorizontal: 10,
  },
  fareText: {
    fontWeight: "bold",
    fontSize: 18,
    marginLeft: 10,
  },
  fareSubtext: {
    color: "#666",
    fontSize: 14,
    marginLeft: 10,
  },
});

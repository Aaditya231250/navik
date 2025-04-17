import React, { useEffect, useState, useRef } from "react";
import { View, Text, Image, StyleSheet, Animated, Easing, TouchableOpacity, Alert } from "react-native";
import { useRouter, useLocalSearchParams } from "expo-router";
import MapView, { Marker, PROVIDER_GOOGLE } from "react-native-maps";
import { MaterialIcons } from "@expo/vector-icons";
import * as Location from 'expo-location';

export default function SearchingScreen() {
  const router = useRouter();
  const params = useLocalSearchParams();
  const [pulseAnim] = useState(new Animated.Value(0.8));
  
  // Extract location information from params or use defaults
  const origin = {
    title: params.originTitle || "562/11-A",
    address: params.originAddress || "Kaikondrahalli, Bengaluru, Karnataka",
    latitude: parseFloat(params.originLat as string) || 26.2720,  // Default to IIT Jodhpur
    longitude: parseFloat(params.originLng as string) || 73.0120,
  };
  
  const destination = {
    title: params.destTitle || "Third Wave Coffee",
    address: params.destAddress || "17th Cross Rd, PWD Quarters, 1st Sector, HSR Layout, Bengaluru, Karnataka",
    latitude: parseFloat(params.destLat as string) || 26.2650,  // Nearby location in Jodhpur
    longitude: parseFloat(params.destLng as string) || 73.0200,
  };
  
  const fare = params.fare || "₹193.20";
  const eta = params.eta || "15 mins";
  const distance = params.distance || "5.2 km";
  
  // Generate random drivers around origin
  const [nearbyDrivers, setNearbyDrivers] = useState([]);
  
  useEffect(() => {
    // Setup location
    (async () => {
      let { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== 'granted') {
        Alert.alert("Location Access", "Permission to access location was denied");
        return;
      }
      
      // Generate random drivers nearby
      generateRandomDrivers();
    })();
    
    // Animation for the pulsing effect
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

    // Simulate finding a driver after 3 seconds
    const driverTimer = setTimeout(() => {
      if (nearbyDrivers.length > 0) {
        const randomDriver = nearbyDrivers[Math.floor(Math.random() * nearbyDrivers.length)];
        
        // Navigate to driver-assigned screen passing all needed params
        router.push({
          pathname: "/ride-journey/driver-assigned",
          params: {
            ...params,
            driverId: randomDriver.id,
            driverName: randomDriver.name,
            driverRating: randomDriver.rating,
            driverCarModel: randomDriver.carModel,
            driverPlateNumber: randomDriver.plateNumber,
            driverLatitude: randomDriver.coordinate.latitude,
            driverLongitude: randomDriver.coordinate.longitude,
            driverHeading: randomDriver.heading,
          }
        });
      }
    }, 3000);
    
    return () => clearTimeout(driverTimer);
  }, [nearbyDrivers]);

  // Generate random drivers around origin
  const generateRandomDrivers = () => {
    const drivers = [];
    for (let i = 1; i <= 5; i++) {
      // Generate random offset (within ~1km)
      const latOffset = (Math.random() - 0.5) * 0.01;
      const lngOffset = (Math.random() - 0.5) * 0.01;
      
      drivers.push({
        id: i,
        name: `Driver ${i}`,
        rating: (4 + Math.random()).toFixed(1),
        carModel: ["Toyota Camry", "Honda Accord", "Maruti Swift", "Hyundai i20", "Tata Nexon"][i-1],
        plateNumber: `JH-${10 + i}X-${1000 + Math.floor(Math.random() * 9000)}`,
        coordinate: {
          latitude: origin.latitude + latOffset,
          longitude: origin.longitude + lngOffset
        },
        heading: Math.floor(Math.random() * 360)
      });
    }
    setNearbyDrivers(drivers);
  };

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
        {/* Origin Marker */}
        <Marker
          coordinate={{
            latitude: origin.latitude,
            longitude: origin.longitude,
          }}
          title={origin.title}
        >
          <View style={styles.originMarker}>
            <MaterialIcons name="person-pin-circle" size={28} color="#007bff" />
          </View>
        </Marker>

        {/* Nearby drivers markers */}
        {nearbyDrivers.map(driver => (
          <Marker
            key={`driver-${driver.id}`}
            coordinate={driver.coordinate}
            title={driver.name}
            rotation={driver.heading}
          >
            <View style={styles.driverMarker}>
              <MaterialIcons name="local-taxi" size={18} color="#fff" />
            </View>
          </Marker>
        ))}
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
                { transform: [{ scale: pulseAnim }] },
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
            <Text style={styles.fareSubtext}>Cash</Text>
          </View>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  // ... Include all the necessary styles from the original component
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
  driverMarker: {
    backgroundColor: "#FF5722",
    padding: 6,
    borderRadius: 12,
    alignItems: "center",
    justifyContent: "center",
  },
  originMarker: {
    alignItems: "center",
    justifyContent: "center",
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

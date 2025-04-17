import React, { useEffect, useState, useRef } from "react";
import { View, Text, StyleSheet, Animated, Easing, TouchableOpacity } from "react-native";
import { useRouter, useLocalSearchParams } from "expo-router";
import MapView, { Marker, Polyline, PROVIDER_GOOGLE } from "react-native-maps";
import { MaterialIcons, FontAwesome } from "@expo/vector-icons";

// Ride status constants
const RIDE_STATUS = {
  DRIVER_ASSIGNED: 'driver_assigned',
  DRIVER_ARRIVING: 'driver_arriving',
  ARRIVED: 'arrived',
};

export default function DriverAssignedScreen() {
  const router = useRouter();
  const params = useLocalSearchParams();
  const mapRef = useRef(null);
  const [slideAnim] = useState(new Animated.Value(0));
  
  // State for ride status and locations
  const [rideStatus, setRideStatus] = useState(RIDE_STATUS.DRIVER_ASSIGNED);
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
  
  // Extract driver information from params
  const selectedDriver = {
    id: params.driverId || 1,
    name: params.driverName || "Driver 1",
    rating: params.driverRating || "4.8",
    carModel: params.driverCarModel || "Toyota Camry",
    plateNumber: params.driverPlateNumber || "JH-11X-1234",
    coordinate: {
      latitude: parseFloat(params.driverLatitude as string) || origin.latitude + 0.005,
      longitude: parseFloat(params.driverLongitude as string) || origin.longitude + 0.005
    },
    heading: parseFloat(params.driverHeading as string) || 0
  };

  useEffect(() => {
    // Set user location
    setCurrentUserLocation({
      latitude: origin.latitude,
      longitude: origin.longitude
    });
    
    // Generate route coordinates for driver to reach user
    const routePoints = generateRouteCoordinates(
      selectedDriver.coordinate,
      { latitude: origin.latitude, longitude: origin.longitude },
      10
    );
    setRouteCoordinates(routePoints);
    
    // Start driver movement animation
    startDriverMovement(routePoints);
    
    // Slide in animation for driver card
    Animated.timing(slideAnim, {
      toValue: 1,
      duration: 500,
      useNativeDriver: true,
      easing: Easing.out(Easing.ease)
    }).start();
    
    // Fit map to show both user and driver
    if (mapRef.current) {
      mapRef.current.fitToCoordinates(
        [currentUserLocation, selectedDriver.coordinate],
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

  // Simulate driver movement
  const startDriverMovement = (routePoints) => {
    let currentPointIndex = 0;
    
    const movementInterval = setInterval(() => {
      if (currentPointIndex < routePoints.length - 1) {
        currentPointIndex++;
        
        // Update driver location
        selectedDriver.coordinate = routePoints[currentPointIndex];
        
        // Calculate heading (direction)
        if (currentPointIndex > 0) {
          const prevPoint = routePoints[currentPointIndex - 1];
          const currentPoint = routePoints[currentPointIndex];
          const heading = Math.atan2(
            currentPoint.longitude - prevPoint.longitude,
            currentPoint.latitude - prevPoint.latitude
          ) * (180 / Math.PI);
          
          selectedDriver.heading = heading;
        }
        
        // When driver is halfway to pickup
        if (currentPointIndex === Math.floor(routePoints.length / 2) && rideStatus === RIDE_STATUS.DRIVER_ASSIGNED) {
          setRideStatus(RIDE_STATUS.DRIVER_ARRIVING);
        }
        
        // When driver arrives at pickup
        if (currentPointIndex === routePoints.length - 1) {
          clearInterval(movementInterval);
          setRideStatus(RIDE_STATUS.ARRIVED);
          
          // Simulate pickup after 3 seconds
          setTimeout(() => {
            // Navigate to in-progress screen
            router.push({
              pathname: "/ride-journey/in-progress",
              params: {
                ...params,
                driverName: selectedDriver.name,
                driverRating: selectedDriver.rating,
                driverCarModel: selectedDriver.carModel,
                driverPlateNumber: selectedDriver.plateNumber,
              }
            });
          }, 3000);
        }
      }
    }, 1000);
    
    return () => clearInterval(movementInterval);
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

        {/* Selected driver marker */}
        <Marker
          coordinate={selectedDriver.coordinate}
          title={selectedDriver.name}
          rotation={selectedDriver.heading}
        >
          <View style={styles.driverMarker}>
            <MaterialIcons name="local-taxi" size={18} color="#fff" />
          </View>
        </Marker>

        {/* Route polyline */}
        {routeCoordinates.length > 0 && (
          <Polyline
            coordinates={routeCoordinates}
            strokeWidth={4}
            strokeColor="#FF5722"
          />
        )}
      </MapView>

      {/* Origin Label */}
      <View style={styles.originLabel}>
        <Text style={styles.originText}>
          {`Driver is ${rideStatus === RIDE_STATUS.ARRIVED ? 'here' : 'on the way'}`}
        </Text>
      </View>

      {/* Driver Info Card */}
      <Animated.View 
        style={[
          styles.bottomCard, 
          { transform: [{ translateY: slideAnim.interpolate({
            inputRange: [0, 1],
            outputRange: [200, 0]
          })}] }
        ]}
      >
        <Text style={styles.driverAssignedText}>Your driver is on the way</Text>
        <View style={styles.driverInfoContainer}>
          <View style={styles.driverImageContainer}>
            <FontAwesome name="user-circle" size={60} color="#007bff" />
          </View>
          <View style={styles.driverDetails}>
            <Text style={styles.driverName}>{selectedDriver.name}</Text>
            <View style={styles.ratingContainer}>
              <FontAwesome name="star" size={16} color="#FFD700" />
              <Text style={styles.ratingText}>{selectedDriver.rating}</Text>
            </View>
            <Text style={styles.carInfo}>{selectedDriver.carModel} • {selectedDriver.plateNumber}</Text>
          </View>
          <View style={styles.actionButtons}>
            <TouchableOpacity style={styles.actionButton}>
              <MaterialIcons name="phone" size={24} color="#007bff" />
            </TouchableOpacity>
            <TouchableOpacity style={styles.actionButton}>
              <MaterialIcons name="message" size={24} color="#007bff" />
            </TouchableOpacity>
          </View>
        </View>

        <View style={styles.divider} />
        
        <View style={styles.etaContainer}>
          <Text style={styles.etaText}>
            {rideStatus === RIDE_STATUS.DRIVER_ASSIGNED ? 'Driver will arrive in about 3 minutes' : 
             rideStatus === RIDE_STATUS.DRIVER_ARRIVING ? 'Driver is almost there' : 
             'Driver has arrived'}
          </Text>
        </View>

        <View style={styles.divider} />
        
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
      </Animated.View>
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
  originMarker: {
    alignItems: "center",
    justifyContent: "center",
  },
  driverMarker: {
    backgroundColor: "#FF5722",
    padding: 6,
    borderRadius: 12,
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
  driverAssignedText: {
    fontSize: 18,
    fontWeight: "bold",
    textAlign: "center",
    marginBottom: 15,
  },
  driverInfoContainer: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 10,
    marginVertical: 10,
  },
  driverImageContainer: {
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
    marginVertical: 4,
  },
  ratingText: {
    marginLeft: 5,
    fontSize: 16,
  },
  carInfo: {
    color: "#666",
    fontSize: 14,
  },
  actionButtons: {
    flexDirection: "row",
  },
  actionButton: {
    padding: 10,
    marginLeft: 5,
  },
  divider: {
    height: 1,
    backgroundColor: "#e0e0e0",
    marginVertical: 10,
  },
  etaContainer: {
    alignItems: "center",
    marginVertical: 10,
  },
  etaText: {
    fontSize: 16,
    fontWeight: "600",
    color: "#007bff",
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

import React, { useEffect, useState, useRef } from "react";
import { View, Text, Image, StyleSheet, Animated, Easing, TouchableOpacity, Alert } from "react-native";
import { useRouter, useLocalSearchParams } from "expo-router";
import MapView, { Marker, Polyline, PROVIDER_GOOGLE } from "react-native-maps";
import { MaterialIcons, FontAwesome } from "@expo/vector-icons";
import * as Location from 'expo-location';

// Ride status constants
const RIDE_STATUS = {
  SEARCHING: 'searching',
  DRIVER_ASSIGNED: 'driver_assigned',
  DRIVER_ARRIVING: 'driver_arriving',
  ARRIVED: 'arrived',
  IN_JOURNEY: 'in_journey',
  COMPLETED: 'completed'
};

export default function SearchingScreen() {
  const router = useRouter();
  const params = useLocalSearchParams();
  const mapRef = useRef(null);
  const [pulseAnim] = useState(new Animated.Value(0.8));
  const [slideAnim] = useState(new Animated.Value(0));
  
  // State for ride status and locations
  const [rideStatus, setRideStatus] = useState(RIDE_STATUS.SEARCHING);
  const [nearbyDrivers, setNearbyDrivers] = useState([]);
  const [selectedDriver, setSelectedDriver] = useState(null);
  const [journeyProgress, setJourneyProgress] = useState(0);
  const [currentUserLocation, setCurrentUserLocation] = useState(null);
  const [routeCoordinates, setRouteCoordinates] = useState([]);
  
  // Extract location information from params or use defaults
  const origin = {
    title: params.originTitle || "562/11-A",
    address: params.originAddress || "Kaikondrahalli, Bengaluru, Karnataka",
    latitude: parseFloat(params.originLat) || 26.2720,  // Default to IIT Jodhpur
    longitude: parseFloat(params.originLng) || 73.0120,
  };
  
  const destination = {
    title: params.destTitle || "Third Wave Coffee",
    address: params.destAddress || "17th Cross Rd, PWD Quarters, 1st Sector, HSR Layout, Bengaluru, Karnataka",
    latitude: parseFloat(params.destLat) || 26.2650,  // Nearby location in Jodhpur
    longitude: parseFloat(params.destLng) || 73.0200,
  };
  
  const fare = params.fare || "₹193.20";
  const eta = params.eta || "15 mins";
  const distance = params.distance || "5.2 km";

  // Get user's current location
  useEffect(() => {
    (async () => {
      let { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== 'granted') {
        Alert.alert("Location Access", "Permission to access location was denied");
        return;
      }

      // Use the origin from params instead of actual location for demo
      setCurrentUserLocation({
        latitude: origin.latitude,
        longitude: origin.longitude
      });
      
      // Generate random drivers nearby
      generateRandomDrivers();
    })();
  }, []);

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

    // Simulate finding a driver after 3 seconds
    const driverTimer = setTimeout(() => {
      if (nearbyDrivers.length > 0) {
        const randomDriver = nearbyDrivers[Math.floor(Math.random() * nearbyDrivers.length)];
        setSelectedDriver(randomDriver);
        setRideStatus(RIDE_STATUS.DRIVER_ASSIGNED);
        
        // Generate route coordinates
        const routePoints = generateRouteCoordinates(
          randomDriver.coordinate,
          { latitude: origin.latitude, longitude: origin.longitude },
          10
        );
        setRouteCoordinates(routePoints);
        
        // Start driver movement animation
        startDriverMovement(randomDriver, routePoints);
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
  const startDriverMovement = (driver, routePoints) => {
    let currentPointIndex = 0;
    
    const movementInterval = setInterval(() => {
      if (currentPointIndex < routePoints.length - 1) {
        currentPointIndex++;
        
        // Update driver location
        setSelectedDriver(prevDriver => ({
          ...prevDriver,
          coordinate: routePoints[currentPointIndex]
        }));
        
        // Calculate heading (direction)
        if (currentPointIndex > 0) {
          const prevPoint = routePoints[currentPointIndex - 1];
          const currentPoint = routePoints[currentPointIndex];
          const heading = Math.atan2(
            currentPoint.longitude - prevPoint.longitude,
            currentPoint.latitude - prevPoint.latitude
          ) * (180 / Math.PI);
          
          setSelectedDriver(prevDriver => ({
            ...prevDriver,
            heading: heading
          }));
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
            setRideStatus(RIDE_STATUS.IN_JOURNEY);
            
            // Generate journey route
            const journeyRoute = generateRouteCoordinates(
              { latitude: origin.latitude, longitude: origin.longitude },
              { latitude: destination.latitude, longitude: destination.longitude },
              20
            );
            setRouteCoordinates(journeyRoute);
            
            // Start journey animation
            startJourneyAnimation(journeyRoute);
          }, 3000);
        }
      }
    }, 1000);
    
    return () => clearInterval(movementInterval);
  };

  // Simulate journey animation
  const startJourneyAnimation = (journeyRoute) => {
    let journeyIndex = 0;
    const totalPoints = journeyRoute.length;
    
    const journeyInterval = setInterval(() => {
      if (journeyIndex < totalPoints - 1) {
        journeyIndex++;
        
        // Update driver and passenger location (they move together)
        setSelectedDriver(prevDriver => ({
          ...prevDriver,
          coordinate: journeyRoute[journeyIndex]
        }));
        
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
          
          setSelectedDriver(prevDriver => ({
            ...prevDriver,
            heading: heading
          }));
        }
        
        // When journey is complete
        if (journeyIndex === totalPoints - 1) {
          clearInterval(journeyInterval);
          setRideStatus(RIDE_STATUS.COMPLETED);
          
          // Show completion message and redirect after 5 seconds
          setTimeout(() => {
            router.replace("/home");
          }, 5000);
        }
      }
    }, 1000);
    
    return () => clearInterval(journeyInterval);
  };

  // Slide in animation for driver card
  useEffect(() => {
    if (rideStatus === RIDE_STATUS.DRIVER_ASSIGNED) {
      Animated.timing(slideAnim, {
        toValue: 1,
        duration: 500,
        useNativeDriver: true,
        easing: Easing.out(Easing.ease)
      }).start();
    }
  }, [rideStatus]);

  // Fit map to show both user and driver
  useEffect(() => {
    if (mapRef.current && selectedDriver && currentUserLocation) {
      const coordinates = [
        currentUserLocation,
        selectedDriver.coordinate
      ];
      
      // If journey has started, include destination
      if (rideStatus === RIDE_STATUS.IN_JOURNEY) {
        coordinates.push({
          latitude: destination.latitude,
          longitude: destination.longitude
        });
      }
      
      mapRef.current.fitToCoordinates(coordinates, {
        edgePadding: { top: 100, right: 50, bottom: 300, left: 50 },
        animated: true
      });
    }
  }, [selectedDriver, currentUserLocation, rideStatus]);

  // Render searching UI
  const renderSearchingUI = () => (
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
  );

  // Render driver assigned UI
  const renderDriverAssignedUI = () => (
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
          <Text style={styles.driverName}>{selectedDriver?.name}</Text>
          <View style={styles.ratingContainer}>
            <FontAwesome name="star" size={16} color="#FFD700" />
            <Text style={styles.ratingText}>{selectedDriver?.rating}</Text>
          </View>
          <Text style={styles.carInfo}>{selectedDriver?.carModel} â€¢ {selectedDriver?.plateNumber}</Text>
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
  );

  // Render in-journey UI
  const renderInJourneyUI = () => (
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
          <Text style={styles.driverNameMini}>{selectedDriver?.name}</Text>
          <Text style={styles.carInfoMini}>{selectedDriver?.carModel}</Text>
        </View>
        <View style={styles.actionButtonsMini}>
          <TouchableOpacity style={styles.actionButtonMini}>
            <MaterialIcons name="phone" size={20} color="#007bff" />
          </TouchableOpacity>
        </View>
      </View>
    </View>
  );

  // Render completed UI
  const renderCompletedUI = () => (
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
        <Text style={styles.rateDriverText}>Rate your driver</Text>
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
        onPress={() => router.replace("/home")}
      >
        <Text style={styles.homeButtonText}>Back to Home</Text>
      </TouchableOpacity>
    </View>
  );

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
        {(rideStatus === RIDE_STATUS.SEARCHING || 
          rideStatus === RIDE_STATUS.DRIVER_ASSIGNED || 
          rideStatus === RIDE_STATUS.DRIVER_ARRIVING || 
          rideStatus === RIDE_STATUS.ARRIVED) && (
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
        )}

        {/* Current User Location during journey */}
        {rideStatus === RIDE_STATUS.IN_JOURNEY && currentUserLocation && (
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
        {(rideStatus === RIDE_STATUS.IN_JOURNEY || rideStatus === RIDE_STATUS.COMPLETED) && (
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
        )}

        {/* Nearby drivers markers */}
        {rideStatus === RIDE_STATUS.SEARCHING && nearbyDrivers.map(driver => (
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

        {/* Selected driver marker */}
        {selectedDriver && (
          <Marker
            coordinate={selectedDriver.coordinate}
            title={selectedDriver.name}
            rotation={selectedDriver.heading}
          >
            <View style={[
              styles.driverMarker, 
              { backgroundColor: rideStatus === RIDE_STATUS.IN_JOURNEY ? '#4CAF50' : '#FF5722' }
            ]}>
              <MaterialIcons 
                name={rideStatus === RIDE_STATUS.IN_JOURNEY ? "directions-car" : "local-taxi"} 
                size={18} 
                color="#fff" 
              />
            </View>
          </Marker>
        )}

        {/* Route polyline */}
        {routeCoordinates.length > 0 && (
          <Polyline
            coordinates={routeCoordinates}
            strokeWidth={4}
            strokeColor={rideStatus === RIDE_STATUS.IN_JOURNEY ? "#4CAF50" : "#FF5722"}
          />
        )}
      </MapView>

      {/* Origin Label */}
      <View style={styles.originLabel}>
        <Text style={styles.originText}>
          {rideStatus === RIDE_STATUS.SEARCHING ? `From ${origin.title}` :
           rideStatus === RIDE_STATUS.IN_JOURNEY ? `To ${destination.title}` :
           rideStatus === RIDE_STATUS.COMPLETED ? "Trip Completed" :
           `Driver is ${rideStatus === RIDE_STATUS.ARRIVED ? 'here' : 'on the way'}`}
        </Text>
      </View>

      {/* Bottom Cards based on ride status */}
      {rideStatus === RIDE_STATUS.SEARCHING && renderSearchingUI()}
      
      {(rideStatus === RIDE_STATUS.DRIVER_ASSIGNED || 
        rideStatus === RIDE_STATUS.DRIVER_ARRIVING || 
        rideStatus === RIDE_STATUS.ARRIVED) && renderDriverAssignedUI()}
      
      {rideStatus === RIDE_STATUS.IN_JOURNEY && renderInJourneyUI()}
      
      {rideStatus === RIDE_STATUS.COMPLETED && renderCompletedUI()}
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
  // Base bottom card
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
  // Driver assigned styles
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
  etaContainer: {
    alignItems: "center",
    marginVertical: 10,
  },
  etaText: {
    fontSize: 16,
    fontWeight: "600",
    color: "#007bff",
  },
  // Journey UI styles
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
  // Completed UI styles
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
  // Map markers
  originMarker: {
    alignItems: "center",
    justifyContent: "center",
  },
  destinationMarker: {
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
  userLocationMarker: {
    alignItems: "center",
    justifyContent: "center",
  }
});
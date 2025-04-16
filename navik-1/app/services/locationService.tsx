import * as Location from 'expo-location';

interface Coordinates {
  latitude: number;
  longitude: number;
}

interface LocationUpdateData {
  driver_id: string;
  city: string;
  latitude: number;
  longitude: number;
  vehicle_type: string;
  status: string;
}

interface LocationUpdateController {
  stop: () => void;
  updateNow: () => Promise<void>;
}

const DRIVER_CONFIG = {
  driver_id: "CS-12345678",
  city: "mumbai",
  vehicle_type: "STANDARD",
  status: "ACTIVE"
};

const LOCATION_API_ENDPOINT = 'http://172.31.115.2/api/location';

// Tracking state
let isUpdating = false;
let lastUpdateTimestamp = 0;
const THROTTLE_TIME = 3000;


export const sendLocationUpdate = async (coords: Coordinates): Promise<any> => {
  if (isUpdating) {
    console.log('Update already in progress, skipping');
    return null;
  }
  
  const now = Date.now();
  if (now - lastUpdateTimestamp < THROTTLE_TIME) {
    console.log(`Update throttled (last update was ${now - lastUpdateTimestamp}ms ago)`);
    return null;
  }
  
  try {
    isUpdating = true;
    lastUpdateTimestamp = now;
    
    const locationData: LocationUpdateData = {
      ...DRIVER_CONFIG,
      latitude: coords.latitude,
      longitude: coords.longitude,
    };
    
    console.log('Sending location update:', locationData);
    
    const response = await fetch(LOCATION_API_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(locationData),
    });
    
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to update location: ${errorText}`);
    }
    
    console.log('Location updated successfully');
    return await response.json();
  } catch (error) {
    console.error('Error sending location update:', error);
    throw error;
  } finally {
    isUpdating = false;
  }
};

let activeIntervalId: NodeJS.Timeout | null = null;

export const startLocationUpdates = (
  locationCallback: (coords: Coordinates) => void, 
  updateCallback: (time: Date) => void, 
  interval: number = 5000
): LocationUpdateController => {
  if (activeIntervalId) {
    console.log('Stopping existing location updates before starting new ones');
    clearInterval(activeIntervalId);
    activeIntervalId = null;
  }
  
  const stopUpdates = (): void => {
    if (activeIntervalId) {
      clearInterval(activeIntervalId);
      activeIntervalId = null;
      console.log('Location updates stopped');
    }
  };
  
  const getCurrentPositionAndUpdate = async (): Promise<void> => {
    // Skip if already processing an update
    if (isUpdating) {
      console.log('Skipping location update - previous update still in progress');
      return;
    }
    
    try {
      // Get current position from Expo Location
      const currentLocation = await Location.getCurrentPositionAsync({});
      const coords = currentLocation.coords;
      
      // Update local state via callback
      locationCallback(coords);
      
      // Send update to backend (with built-in throttling)
      const result = await sendLocationUpdate(coords);
      
      // Only notify if update was actually sent (not throttled)
      if (result !== null) {
        updateCallback(new Date());
      }
    } catch (error) {
      console.error('Error in location update cycle:', error);
    }
  };
  
  // Start the interval with a safe minimum
  const safeInterval = Math.max(interval, THROTTLE_TIME);
  console.log(`Starting location updates with interval: ${safeInterval}ms`);
  activeIntervalId = setInterval(getCurrentPositionAndUpdate, safeInterval);
  
  // Return control object
  return {
    stop: stopUpdates,
    updateNow: getCurrentPositionAndUpdate
  };
};
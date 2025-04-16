import json
import random
import requests
import time
import threading
import sys
import logging
from concurrent.futures import ThreadPoolExecutor, as_completed

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler("driver_simulation.log"),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger('driver_simulation')

# Constants
VEHICLE_TYPES = ["STANDARD", "PREMIUM", "COMPACT"]
DEFAULT_CITY = "delhi"
DEFAULT_STATUS = "ACTIVE"
API_ENDPOINT = "http://172.31.115.2/api/location"

def load_drivers(filename, n=10):
    try:
        with open(filename, 'r') as f:
            users = json.load(f)
        drivers = [u for u in users if u['user_type'] == 'driver']
        selected = random.sample(drivers, min(n, len(drivers)))
        return selected
    except FileNotFoundError:
        logger.error(f"Error: {filename} not found. Make sure to generate it first.")
        return []

def load_map_coords(filename):
    try:
        with open(filename, 'r') as f:
            data = json.load(f)
        # Extract all node coordinates
        coords = [(el['lat'], el['lon']) for el in data.get('elements', []) 
                 if el.get('type') == 'node' and 'lat' in el and 'lon' in el]
        return coords
    except FileNotFoundError:
        logger.error(f"Error: {filename} not found. Make sure to download it first.")
        return []

def haversine_distance(lat1, lon1, lat2, lon2):
    import math
    lat1, lon1, lat2, lon2 = map(math.radians, [lat1, lon1, lat2, lon2])
    
    dlat = lat2 - lat1
    dlon = lon2 - lon1
    a = math.sin(dlat/2)**2 + math.cos(lat1) * math.cos(lat2) * math.sin(dlon/2)**2
    c = 2 * math.asin(math.sqrt(a))
    r = 6371
    return c * r

def find_next_coordinate(current_pos, coordinates, last_direction=None, max_distance=0.05):
    lat, lon = current_pos
    
    # Filter coordinates within reasonable distance
    nearby_coords = []
    for coord in coordinates:
        distance = haversine_distance(lat, lon, coord[0], coord[1])
        if distance < max_distance and (lat, lon) != coord:  # Avoid same position
            nearby_coords.append((coord, distance))
    
    if not nearby_coords:
        # No nearby coordinates, expand search radius
        for coord in coordinates:
            distance = haversine_distance(lat, lon, coord[0], coord[1])
            if distance < max_distance * 2 and (lat, lon) != coord:
                nearby_coords.append((coord, distance))
                
        # If still no nearby coords, return a random one
        if not nearby_coords:
            return random.choice(coordinates)
    
    # Sort by distance
    nearby_coords.sort(key=lambda x: x[1])
    
    # If we have a previous direction, try to maintain it with some probability
    if last_direction and random.random() < 0.7:
        last_lat_dir = 1 if last_direction[0] > 0 else -1
        last_lon_dir = 1 if last_direction[1] > 0 else -1
        
        # Find coordinates that continue in similar direction
        continuing_coords = []
        for coord, dist in nearby_coords:
            lat_dir = 1 if coord[0] - lat > 0 else -1
            lon_dir = 1 if coord[1] - lon > 0 else -1
            
            # If direction is similar, add to list
            if lat_dir == last_lat_dir and lon_dir == last_lon_dir:
                continuing_coords.append((coord, dist))
        
        # If we have coords that continue in similar direction, choose one
        if continuing_coords:
            # Sometimes choose closest, sometimes random for variety
            if random.random() < 0.8:
                return continuing_coords[0][0]  # Closest
            else:
                return random.choice(continuing_coords[:min(5, len(continuing_coords))])[0]
    
    # No direction constraint or couldn't find continuing coords
    # 80% choose closest, 20% choose random from top 5 for variety
    if random.random() < 0.8:
        return nearby_coords[0][0]  # Closest
    else:
        return random.choice(nearby_coords[:min(5, len(nearby_coords))])[0]

# Driver class to keep track of state
class Driver:
    def __init__(self, email, city=DEFAULT_CITY, coordinates=None):
        self.driver_id = email  # Using email as driver_id as specified
        self.email = email
        self.city = city
        self.vehicle_type = random.choice(VEHICLE_TYPES)
        self.status = DEFAULT_STATUS
        self.coordinates = coordinates
        
        # Initialize with random position from coordinates
        if coordinates and len(coordinates) > 0:
            self.position = random.choice(coordinates)
        else:
            # Default position (Jodhpur center)
            self.position = (26.28, 73.03)
            
        self.last_direction = None
        self.active = True
        self.update_count = 0

    def update_position(self, coordinates):
        if not self.active:
            return False
            
        old_pos = self.position
        self.position = find_next_coordinate(self.position, coordinates, self.last_direction)
        
        # Calculate direction for next update
        lat_diff = self.position[0] - old_pos[0]
        lon_diff = self.position[1] - old_pos[1]
        self.last_direction = (lat_diff, lon_diff)
        
        self.update_count += 1
        return True
        
    def to_location_json(self):
        return {
            "driver_id": self.email,
            "city": self.city,
            "latitude": self.position[0],
            "longitude": self.position[1],
            "vehicle_type": self.vehicle_type,
            "status": self.status
        }

def update_driver_location(driver, coordinates, api_endpoint=API_ENDPOINT):
    if driver.update_position(coordinates):
        location_data = driver.to_location_json()
        try:
            response = requests.post(
                api_endpoint,
                json=location_data,
                headers={"Content-Type": "application/json"}
            )
            logger.debug(response.text)
            logger.info(f"Updated {driver.email}: Status {response.status_code}")
            return True
        except Exception as e:
            logger.error(f"Error updating {driver.email}: {str(e)}")
    return False

# Main function to run driver simulation
def run_driver_simulation():
    # Load data
    drivers_data = load_drivers('random_users.json', 10)
    if not drivers_data:
        logger.error("No drivers found. Exiting.")
        return
        
    coordinates = load_map_coords('jodhpur_map_data.json')
    if not coordinates:
        logger.error("No coordinates found. Exiting.")
        return
    
    logger.info(f"Loaded {len(drivers_data)} drivers and {len(coordinates)} coordinates")
    
    # Create driver objects
    drivers = [Driver(driver['email'], DEFAULT_CITY, coordinates) for driver in drivers_data]
    
    # Initial position update
    logger.info("Setting initial positions...")
    for driver in drivers:
        update_driver_location(driver, coordinates)
    
    # Ask user if they want to continue with updates
    do_updates = input("Do you want to update driver locations at regular intervals? (y/n): ").strip().lower()
    if do_updates != 'y':
        logger.info("Simulation ended. Initial positions set.")
        return
    
    # Get update interval
    update_interval = 5  # Default
    try:
        user_interval = input(f"Update interval in seconds (default: {update_interval}): ")
        if user_interval.strip():
            update_interval = int(user_interval)
    except ValueError:
        logger.warning(f"Invalid input. Using default interval of {update_interval} seconds.")
    
    # Get number of parallel updates
    max_parallel = 100  # Default
    try:
        user_parallel = input(f"Maximum number of parallel updates (default: {max_parallel}): ")
        if user_parallel.strip():
            max_parallel = int(user_parallel)
    except ValueError:
        logger.warning(f"Invalid input. Using default of {max_parallel} parallel updates.")
    
    logger.info(f"Starting location updates every {update_interval} seconds with {max_parallel} parallel updates")
    try:
        while True:
            start_time = time.time()
            random_drivers = random.sample(drivers, min(max_parallel, len(drivers)))
            with ThreadPoolExecutor(max_workers=max_parallel) as executor:
                futures = [executor.submit(update_driver_location, driver, coordinates) 
                          for driver in random_drivers]
                
                for future in as_completed(futures):
                    try:
                        future.result()
                    except Exception as e:
                        logger.error(f"Error in thread: {str(e)}")
            
            # Calculate sleep time to maintain update interval
            elapsed = time.time() - start_time
            sleep_time = max(0, update_interval - elapsed)
            if sleep_time > 0:
                logger.info(f"Waiting {sleep_time:.2f} seconds until next update...")
                time.sleep(sleep_time)
    except KeyboardInterrupt:
        logger.info("\nSimulation stopped by user")
    finally:
        logger.info("Simulation ended")

if __name__ == "__main__":
    run_driver_simulation()
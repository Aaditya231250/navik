import requests
import json

def fetch_jodhpur_map_data():
    jodhpur_bbox = {
        "min_lat": 26.2,
        "min_lon": 73.0,
        "max_lat": 26.4,
        "max_lon": 73.2
    }
    
    highway_exclude = [
        "footway", "street_lamp", "steps", "pedestrian", "track", "path"
    ]
    
    exclusion = ''.join([f'[highway!="{exclude}"]' for exclude in highway_exclude])
    
    query = f"""
    [out:json];
    (
      way[highway]{exclusion}[footway!="*"]
      ({jodhpur_bbox['min_lat']},{jodhpur_bbox['min_lon']},{jodhpur_bbox['max_lat']},{jodhpur_bbox['max_lon']});
      node(w);
    );
    out skel;
    """
    
    response = requests.post(
        "https://overpass-api.de/api/interpreter",
        data=query
    )
    
    print(f"Status Code: {response.status_code}")
    
    if response.status_code == 200:
        # Parse and save the response to a file
        with open("jodhpur_map_data.json", "w") as f:
            f.write(response.text)
        print("Map data saved to jodhpur_map_data.json")
    else:
        print(f"Error: {response.text}")

if __name__ == "__main__":
    fetch_jodhpur_map_data()

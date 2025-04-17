import asyncio
import json
from websockets import connect
import aiohttp
from datetime import datetime


async def test_websocket_flow():
    user_id = "customer-rajesh876@example.com"
    ws_url = f"ws://localhost/ws?user_id={user_id}"
    api_url = "http://localhost/api/matching"
    
    test_location = {
        "user_id": user_id,
    "city": "delhi",
    "latitude": 26.32504,
    "longitude": 73.12539,
    "request_type": "RIDE_REQUEST"
    }

    async with connect(ws_url) as websocket:
        async with aiohttp.ClientSession() as session:
            async with session.post(api_url, json=test_location) as resp:
                if resp.status != 200:
                    print(resp.text)
                    print(f"❌ HTTP POST failed: {resp.status}")
                    return
                print(f"✅ HTTP POST successful: {resp.text}")

        # Wait for WebSocket response
        try:
            response = await asyncio.wait_for(websocket.recv(), timeout=10)
            data = json.loads(response)
            
            if data.get("status") == "SUCCESS" and "drivers" in data:
                print(f"✅ Received {len(data['drivers'])} drivers!")
                print(json.dumps(data, indent=2))
            else:
                print(f"❌ Unexpected response: {response}")
                
        except asyncio.TimeoutError:
            print("❌ No response received within 10 seconds")

if __name__ == "__main__":
    asyncio.run(test_websocket_flow())

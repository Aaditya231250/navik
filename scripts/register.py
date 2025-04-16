import asyncio
import aiohttp
import json
import time
from tqdm import tqdm

API_URL = "http://172.31.115.2/api/auth/register"
USER_FILE = "random_users.json"
CONCURRENT_REQUESTS = 50  # Number of concurrent connections (tune as needed)
TOTAL_REQUESTS_PER_SECOND = 100  # Target rate (requests/sec)
CUSTOM_USER_AGENT = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
    "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"
)

async def register_user(session, user):
    try:
        async with session.post(API_URL, json=user) as response:
            status = response.status
            if status in (200, 201):
                return True
            else:
                return False
    except Exception:
        return False

async def main():
    # Load users
    with open(USER_FILE, "r") as f:
        users = json.load(f)
    total_users = len(users)

    # Prepare session and connector
    timeout = aiohttp.ClientTimeout(total=5)
    connector = aiohttp.TCPConnector(limit=CONCURRENT_REQUESTS, force_close=True)
    headers = {"User-Agent": CUSTOM_USER_AGENT, "Content-Type": "application/json"}

    # Progress bar
    pbar = tqdm(total=total_users, desc="Registering", unit="user")
    success, fail = 0, 0

    async with aiohttp.ClientSession(timeout=timeout, connector=connector, headers=headers) as session:
        tasks = []
        start = time.perf_counter()
        for i, user in enumerate(users):
            # Rate limiting: sleep if needed to maintain desired rate
            if i > 0 and i % TOTAL_REQUESTS_PER_SECOND == 0:
                elapsed = time.perf_counter() - start
                if elapsed < i / TOTAL_REQUESTS_PER_SECOND:
                    await asyncio.sleep((i / TOTAL_REQUESTS_PER_SECOND) - elapsed)
            tasks.append(register_user(session, user))

            # Dispatch in batches for memory efficiency
            if len(tasks) >= CONCURRENT_REQUESTS or i == total_users - 1:
                results = await asyncio.gather(*tasks)
                for result in results:
                    if result:
                        success += 1
                    else:
                        fail += 1
                    pbar.update(1)
                tasks = []

        pbar.close()
        total_time = time.perf_counter() - start
        print(f"\nTotal: {total_users} | Success: {success} | Fail: {fail}")
        print(f"Elapsed: {total_time:.2f}s | Rate: {total_users/total_time:.2f} req/sec")

if __name__ == "__main__":
    asyncio.run(main())

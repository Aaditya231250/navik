"""
Authentication Service API Test Script
Tests all endpoints for both customer and driver users including:
- Registration
- Login
- Token refresh
- Profile retrieval
- Profile updates
- Password reset
- Error cases
"""

import requests
import json
import time
import random
import string
import sys
from datetime import datetime, timedelta

# Configuration
BASE_URL = "http://localhost/api"
VERBOSE = True  # Set to True for detailed output

# Test data
CUSTOMER_EMAIL = f"customer_{int(time.time())}@example.com"
CUSTOMER_PASSWORD = "Password123!"
DRIVER_EMAIL = f"driver_{int(time.time())}@example.com"
DRIVER_PASSWORD = "Password123!"

# Keep track of tokens and user IDs
tokens = {
    "customer": {"access": None, "refresh": None, "user_id": None},
    "driver": {"access": None, "refresh": None, "user_id": None}
}

# Helper functions
def log(message, level="INFO"):
    """Log messages with timestamp"""
    if VERBOSE or level != "INFO":
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        print(f"[{timestamp}] [{level}] {message}")

def random_string(length=8):
    """Generate a random string of fixed length"""
    return ''.join(random.choices(string.ascii_letters + string.digits, k=length))

def make_request(method, endpoint, data=None, auth=None, expected_status=None):
    """Make HTTP request with error handling and logging"""
    url = f"{BASE_URL}{endpoint}"
    headers = {"Content-Type": "application/json"}
    
    if auth:
        headers["Authorization"] = f"Bearer {auth}"
    
    log(f"Making {method} request to {url}")
    if data:
        log(f"Request data: {json.dumps(data, indent=2)}")
    
    try:
        if method == "GET":
            response = requests.get(url, headers=headers)
        elif method == "POST":
            response = requests.post(url, headers=headers, json=data)
        elif method == "PUT":
            response = requests.put(url, headers=headers, json=data)
        else:
            log(f"Unsupported method: {method}", "ERROR")
            return None
        
        if expected_status and response.status_code != expected_status:
            log(f"Expected status {expected_status}, got {response.status_code}", "ERROR")
            log(f"Response: {response.text}", "ERROR")
            return None
        
        log(f"Response status: {response.status_code}")
        
        if response.headers.get("Content-Type", "").startswith("application/json"):
            resp_data = response.json()
            log(f"Response data: {json.dumps(resp_data, indent=2)}")
            return resp_data
        
        return response.text
    
    except Exception as e:
        log(f"Request error: {str(e)}", "ERROR")
        return None

def run_test(name, test_fn):
    """Run a test function with proper logging"""
    log(f"\n{'=' * 80}")
    log(f"Starting test: {name}")
    log(f"{'=' * 80}")
    
    try:
        result = test_fn()
        if result:
            log(f"Test '{name}' PASSED", "SUCCESS")
        else:
            log(f"Test '{name}' FAILED", "ERROR")
        return result
    except Exception as e:
        log(f"Test '{name}' FAILED with exception: {str(e)}", "ERROR")
        return False

# Test functions
def test_customer_registration():
    """Test customer registration"""
    data = {
        "email": CUSTOMER_EMAIL,
        "password": CUSTOMER_PASSWORD,
        "user_type": "customer",
        "phone": "123-456-7890",
        "first_name": "Test",
        "last_name": "Customer",
        "address": "123 Customer St"
    }
    
    response = make_request("POST", "/auth/register", data, expected_status=201)
    if not response:
        return False
    
    # Save tokens
    tokens["customer"]["access"] = response.get("access_token")
    tokens["customer"]["refresh"] = response.get("refresh_token")
    tokens["customer"]["user_id"] = response.get("user_id")
    
    return tokens["customer"]["access"] is not None

def test_driver_registration():
    """Test driver registration"""
    expiry_date = (datetime.now() + timedelta(days=365)).strftime("%Y-%m-%dT%H:%M:%SZ")
    
    data = {
        "email": DRIVER_EMAIL,
        "password": DRIVER_PASSWORD,
        "user_type": "driver",
        "phone": "123-456-7891",
        "first_name": "Test",
        "last_name": "Driver",
        "license_number": "DL12345678",
        "license_expiry": expiry_date,
        "vehicle_info": "Toyota Camry, 2020, White"
    }
    
    response = make_request("POST", "/auth/register", data, expected_status=201)
    if not response:
        return False
    
    # Save tokens
    tokens["driver"]["access"] = response.get("access_token")
    tokens["driver"]["refresh"] = response.get("refresh_token")
    tokens["driver"]["user_id"] = response.get("user_id")
    
    return tokens["driver"]["access"] is not None

def test_duplicate_registration():
    """Test registering with an email that already exists"""
    data = {
        "email": CUSTOMER_EMAIL,  # Use existing email
        "password": CUSTOMER_PASSWORD,
        "user_type": "customer",
        "phone": "123-456-7890",
        "first_name": "Duplicate",
        "last_name": "User"
    }
    
    response = make_request("POST", "/auth/register", data, expected_status=409)
    print(f"Duplicate registration response: {response}")
    return response is None

def test_invalid_registration():
    """Test registration with invalid data"""
    data = {
        "email": "not_an_email",
        "password": "short",
        "user_type": "unknown_type",
        "phone": "123-456-7890",
        "first_name": "Invalid",
        "last_name": "User"
    }
    response = make_request("POST", "/auth/register", data, expected_status=400)
    d = {
  "error": "Email email, Password min, UserType oneof"
    }
    if response == d:
        print(f"Invalid registration response: {response}")
        return response
    return None

def test_customer_login():
    """Test customer login"""
    data = {
        "email": CUSTOMER_EMAIL,
        "password": CUSTOMER_PASSWORD
    }
    
    response = make_request("POST", "/auth/login", data, expected_status=200)
    if not response:
        return False
    
    # Save tokens
    tokens["customer"]["access"] = response.get("access_token")
    tokens["customer"]["refresh"] = response.get("refresh_token")
    
    return tokens["customer"]["access"] is not None

def test_driver_login():
    """Test driver login"""
    data = {
        "email": DRIVER_EMAIL,
        "password": DRIVER_PASSWORD
    }
    
    response = make_request("POST", "/auth/login", data, expected_status=200)
    if not response:
        return False
    
    # Save tokens
    tokens["driver"]["access"] = response.get("access_token")
    tokens["driver"]["refresh"] = response.get("refresh_token")
    
    return tokens["driver"]["access"] is not None

def test_invalid_login():
    """Test login with invalid credentials"""
    data = {
        "email": CUSTOMER_EMAIL,
        "password": "wrong_password"
    }
    
    response = make_request("POST", "/auth/login", data, expected_status=401)
    # We expect this to fail with a 401 unauthorized
    print(f"Invalid login response: {response}")
    d = {'error': 'Invalid email or password'}
    if d == response:
        print(f"Invalid login response: {response}")
        return response
    return  None

def test_refresh_token():
    """Test refresh token endpoint"""
    data = {
        "refresh_token": tokens["customer"]["refresh"]
    }
    
    response = make_request("POST", "/auth/refresh", data, expected_status=200)
    if not response:
        return False
    
    # Save new tokens
    tokens["customer"]["access"] = response.get("access_token")
    tokens["customer"]["refresh"] = response.get("refresh_token")
    
    return tokens["customer"]["access"] is not None

def test_get_customer_profile():
    """Test getting customer profile"""
    response = make_request(
        "GET", 
        "/customer/profile", 
        auth=tokens["customer"]["access"],
        expected_status=200
    )
    
    if not response:
        return False
        
    return (
        response.get("email") == CUSTOMER_EMAIL and
        response.get("first_name") == "Test" and
        response.get("last_name") == "Customer"
    )

def test_get_driver_profile():
    """Test getting driver profile"""
    response = make_request(
        "GET", 
        "/driver/profile", 
        auth=tokens["driver"]["access"],
        expected_status=200
    )
    print(f"Driver profile response: {response}")
    if not response:
        return False
        
    return (
        response.get("email") == DRIVER_EMAIL and
        response.get("first_name") == "Test" and
        response.get("last_name") == "Driver"
        # response.get("license_number") == "DL12345678"
    )

def test_update_customer_profile():
    """Test updating customer profile"""
    data = {
        "first_name": f"Updated{random_string(4)}",
        "last_name": f"Customer{random_string(4)}",
        "phone": f"555-{random_string(3)}-{random_string(4)}",
        "address": f"{random_string(5)} Main St, Anytown"
    }
    
    response = make_request(
        "PUT", 
        "/customer/profile", 
        data,
        auth=tokens["customer"]["access"],
        expected_status=200
    )
    
    if not response:
        return False
        
    return (
        response.get("first_name") == data["first_name"] and
        response.get("last_name") == data["last_name"] and
        response.get("phone") == data["phone"] and
        response.get("address") == data["address"]
    )

def test_update_driver_profile():
    """Test updating driver profile"""
    data = {
        "first_name": f"Updated{random_string(4)}",
        "last_name": f"Driver{random_string(4)}",
        "phone": f"555-{random_string(3)}-{random_string(4)}",
        "vehicle_info": f"{random_string(10)} Model {random_string(2)}, 2023, Black"
    }
    
    response = make_request(
        "PUT", 
        "/driver/profile", 
        data,
        auth=tokens["driver"]["access"],
        expected_status=200
    )
    
    if not response:
        return False
        
    return (
        response.get("first_name") == data["first_name"] and
        response.get("last_name") == data["last_name"] and
        response.get("phone") == data["phone"] and
        response.get("vehicle_info") == data["vehicle_info"]
    )

def test_unauthorized_access():
    """Test accessing customer endpoint with driver token and vice versa"""
    # Driver trying to access customer endpoint
    response = make_request(
        "GET", 
        "/customer/profile", 
        auth=tokens["driver"]["access"],
        expected_status=403
    )
    d= {
  "error": "Insufficient permissions"
}
    if response != d:
        return None
        
    # Customer trying to access driver endpoint
    response = make_request(
        "GET", 
        "/driver/profile", 
        auth=tokens["customer"]["access"],
        expected_status=403
    )
    if response != d:
        return None
    return response

def test_password_reset_request():
    """Test password reset request"""
    data = {
        "email": CUSTOMER_EMAIL
    }
    
    response = make_request("POST", "/auth/password-reset/request", data, expected_status=200)
    if not response:
        return False
        
    return "message" in response

def main():
    """Main test function"""
    tests = [
        # Registration tests
        ("Customer Registration", test_customer_registration),
        ("Driver Registration", test_driver_registration),
        ("Duplicate Registration", test_duplicate_registration),
        ("Invalid Registration", test_invalid_registration),
        
        # Login tests
        ("Customer Login", test_customer_login),
        ("Driver Login", test_driver_login),
        ("Invalid Login", test_invalid_login),
        
        # Token tests
        ("Refresh Token", test_refresh_token),
        
        # Profile tests
        ("Get Customer Profile", test_get_customer_profile),
        ("Get Driver Profile", test_get_driver_profile),
        ("Update Customer Profile", test_update_customer_profile),
        ("Update Driver Profile", test_update_driver_profile),
        
        # Authorization tests
        ("Unauthorized Access", test_unauthorized_access),
        
        # Password reset
        ("Password Reset Request", test_password_reset_request)
    ]
    
    results = {}
    all_passed = True
    
    log("Starting API tests...")
    
    for name, test_fn in tests:
        results[name] = run_test(name, test_fn)
        all_passed = all_passed and results[name]
    
    log("\n\n" + "=" * 80)
    log("TEST SUMMARY")
    log("=" * 80)
    
    for name, result in results.items():
        status = "PASSED" if result else "FAILED"
        log(f"{name.ljust(30)} : {status}")
    
    log("=" * 80)
    log(f"OVERALL RESULT: {'PASSED' if all_passed else 'FAILED'}")
    log("=" * 80)
    
    # Exit with appropriate status code
    sys.exit(0 if all_passed else 1)

if __name__ == "__main__":
    main()

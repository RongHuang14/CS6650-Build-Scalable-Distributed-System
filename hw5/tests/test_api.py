"""
test script to demonstrate all API response codes
For CS 6650 Assignment - Product API
"""
import requests
import json

BASE_URL = "http://localhost:8080"

print("="*50)
print("Product API Test - All Response Codes")
print("="*50)

# GET ENDPOINT TESTS
print("\n### GET /products/{productId} ###")

# Test 1: GET 200 - Success
print("\n1. GET 200 - Existing Product:")
r = requests.get(f"{BASE_URL}/products/1")
print(f"   Request: GET /products/1")
print(f"   Status: {r.status_code}")
if r.status_code == 200:
    print(f"   ✓ Response: {r.json()}")

# Test 2: GET 404 - Not Found
print("\n2. GET 404 - Non-existing Product:")
r = requests.get(f"{BASE_URL}/products/999")
print(f"   Request: GET /products/999")
print(f"   Status: {r.status_code}")
if r.status_code == 404:
    print(f"   ✓ Error: {r.json()}")

# Test 3: GET 500 - Server Error (using special trigger)
print("\n3. GET 500 - Server Error:")
r = requests.get(f"{BASE_URL}/products/500")
print(f"   Request: GET /products/500 (test trigger)")
print(f"   Status: {r.status_code}")
if r.status_code == 500:
    print(f"   ✓ Error: {r.json()}")

# POST ENDPOINT TESTS
print("\n### POST /products/{productId}/details ###")

# Test 4: POST 204 - Success
print("\n4. POST 204 - Valid Product:")
product = {
    "product_id": 3,
    "sku": "TEST-003",
    "manufacturer": "Test Corp",
    "category_id": 1,
    "weight": 500,
    "some_other_id": 103
}
r = requests.post(f"{BASE_URL}/products/3/details", json=product)
print(f"   Request: POST /products/3/details")
print(f"   Body: {json.dumps(product, indent=6)}")
print(f"   Status: {r.status_code}")
if r.status_code == 204:
    print("   ✓ Product added successfully (no content)")

# Test 5: POST 400 - Bad Request (ID mismatch)
print("\n5. POST 400 - ID Mismatch:")
product["product_id"] = 999  # Different from path
r = requests.post(f"{BASE_URL}/products/3/details", json=product)
print(f"   Request: POST /products/3/details (path ID: 3)")
print(f"   Body product_id: 999 (mismatch)")
print(f"   Status: {r.status_code}")
if r.status_code == 400:
    print(f"   ✓ Error: {r.json()}")

# Test 6: POST 404 - Not Found
print("\n6. POST 404 - Product Not Found:")
print("   Note: This depends on business logic")
product_404 = {
    "product_id": 404,
    "sku": "TEST-404",
    "manufacturer": "Test Corp",
    "category_id": 1,
    "weight": 500,
    "some_other_id": 404
}
r = requests.post(f"{BASE_URL}/products/404/details", json=product_404)
print(f"   Request: POST /products/404/details")
print(f"   Status: {r.status_code}")
if r.status_code == 404:
    print("   ✓ Product must exist before adding details")
elif r.status_code == 204:
    print("   Note: System allows creating via details endpoint")

# Test 7: POST 500 - Server Error
print("\n7. POST 500 - Server Error:")
error_product = {
    "product_id": 500,
    "sku": "ERROR-TEST",
    "manufacturer": "Test Corp",
    "category_id": 1,
    "weight": 100,
    "some_other_id": 500
}
r = requests.post(f"{BASE_URL}/products/500/details", json=error_product)
print(f"   Request: POST /products/500/details (test trigger)")
print(f"   Status: {r.status_code}")
if r.status_code == 500:
    print(f"   ✓ Error: {r.json()}")

print("\n" + "="*50)
print("✅ All response codes tested!")
print("="*50)
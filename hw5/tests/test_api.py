import requests
import json

BASE_URL = "http://localhost:8080"

print("Testing Product API...")

# Test GET - 200
print("\n1. GET /products/1 (expecting 200):")
r = requests.get(f"{BASE_URL}/products/1")
print(f"   Status: {r.status_code}")
if r.status_code == 200:
    print(f"   ✓ Response: {r.json()}")

# Test GET - 404
print("\n2. GET /products/999 (expecting 404):")
r = requests.get(f"{BASE_URL}/products/999")
print(f"   Status: {r.status_code}")
if r.status_code == 404:
    print(f"   ✓ Error: {r.json()}")

# Test POST - 204
print("\n3. POST /products/3/details (expecting 204):")
product = {
    "product_id": 3,
    "sku": "TEST-003",
    "manufacturer": "Test Corp",
    "category_id": 1,
    "weight": 500,
    "some_other_id": 103
}
r = requests.post(f"{BASE_URL}/products/3/details", json=product)
print(f"   Status: {r.status_code}")
if r.status_code == 204:
    print("   ✓ Product added successfully")

# Test POST - 400 (ID mismatch)
print("\n4. POST with ID mismatch (expecting 400):")
product["product_id"] = 999
r = requests.post(f"{BASE_URL}/products/3/details", json=product)
print(f"   Status: {r.status_code}")
if r.status_code == 400:
    print(f"   ✓ Error: {r.json()}")

print("\n✅ All tests completed!")
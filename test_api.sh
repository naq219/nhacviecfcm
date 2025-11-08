#!/bin/bash

# Test API với Git Bash

echo "Testing API with Git Bash..."

# Lấy token trước
echo "Getting auth token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8090/api/collections/users/auth-with-password \
  -H "Content-Type: application/json" \
  -d '{"identity":"test@example.com","password":"test123456"}')

TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "Token: $TOKEN"

# Tạo reminder với next_recurring cụ thể
echo "Creating reminder with specific next_recurring..."
RESPONSE=$(curl -s -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Test Reminder Git Bash",
    "description": "Test next_recurring fix with Git Bash",
    "type": "recurring",
    "calendar_type": "solar",
    "next_recurring": "2025-11-08T08:05:00.000Z",
    "recurrence_pattern": {
      "interval": 180,
      "unit": "minute"
    },
    "crp_interval_sec": 20,
    "max_crp": 3
  }')

echo "Response: $RESPONSE"

# Kiểm tra kết quả
if echo "$RESPONSE" | grep -q '"id"'; then
    echo "✅ SUCCESS: Reminder created successfully!"
    echo "Full response:"
    echo "$RESPONSE" | python -m json.tool
else
    echo "❌ ERROR: Failed to create reminder"
    echo "Error details:"
    echo "$RESPONSE" | python -m json.tool
fi
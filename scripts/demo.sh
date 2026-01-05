#!/bin/bash

# Demo script to showcase the Voucher & Payment Service

set -e

BASE_URL="http://localhost:8080"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

print_header() {
    echo -e "\n${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}$1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_step() {
    echo -e "${GREEN}➜${NC} $1"
}

print_result() {
    echo -e "${CYAN}$1${NC}\n"
}

# Check if service is running
if ! curl -s "$BASE_URL/v1/vouchers/search" > /dev/null 2>&1; then
    echo -e "${RED}❌ Service is not running. Please start it with: make run${NC}"
    exit 1
fi

print_header "🎮 Voucher & Payment Service Demo"

# 1. Search Vouchers
print_header "1️⃣  Search for Amazon Vouchers"
print_step "Searching for 'amazon' vouchers..."
curl -s "$BASE_URL/v1/vouchers/search?query=amazon&page=1&page_size=5" | jq '.'
sleep 2

# 2. Get Voucher Details
print_header "2️⃣  Get Voucher Details"
print_step "Getting details for voucher-001..."
curl -s "$BASE_URL/v1/vouchers/voucher-001" | jq '.'
sleep 2

# 3. Check Wallet Balance
print_header "3️⃣  Check Wallet Balance"
print_step "Checking balance for user-001..."
BALANCE=$(curl -s "$BASE_URL/v1/users/user-001/balance" | jq -r '.balance')
echo -e "Current balance: ${GREEN}₹${BALANCE}${NC}"
sleep 2

# 4. Buy Voucher
print_header "4️⃣  Buy Voucher"
print_step "Purchasing 1 Amazon voucher using wallet..."
PURCHASE_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/vouchers/buy" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-001",
    "voucher_id": "voucher-001",
    "quantity": 1,
    "payment_method": "wallet"
  }')

echo "$PURCHASE_RESPONSE" | jq '.'

TRANSACTION_ID=$(echo "$PURCHASE_RESPONSE" | jq -r '.transaction_id')
VOUCHER_CODE=$(echo "$PURCHASE_RESPONSE" | jq -r '.purchased_vouchers[0].voucher_code')
PIN=$(echo "$PURCHASE_RESPONSE" | jq -r '.purchased_vouchers[0].pin')

echo -e "\n${GREEN}✅ Purchase successful!${NC}"
echo -e "Transaction ID: ${YELLOW}$TRANSACTION_ID${NC}"
echo -e "Voucher Code: ${YELLOW}$VOUCHER_CODE${NC}"
echo -e "PIN: ${YELLOW}$PIN${NC}"
sleep 2

# 5. Check Updated Balance
print_header "5️⃣  Check Updated Balance"
print_step "Checking balance after purchase..."
NEW_BALANCE=$(curl -s "$BASE_URL/v1/users/user-001/balance" | jq -r '.balance')
echo -e "New balance: ${GREEN}₹${NEW_BALANCE}${NC}"
sleep 2

# 6. List User Vouchers
print_header "6️⃣  List User Vouchers"
print_step "Listing all vouchers for user-001..."
curl -s "$BASE_URL/v1/users/user-001/vouchers?status=active&page=1&page_size=10" | jq '.'
sleep 2

# 7. View Transactions
print_header "7️⃣  View Transaction History"
print_step "Listing recent transactions..."
curl -s "$BASE_URL/v1/users/user-001/transactions?type=all&page=1&page_size=5" | jq '.'
sleep 2

# 8. Add Money to Wallet
print_header "8️⃣  Add Money to Wallet"
print_step "Adding ₹1000 to wallet via UPI..."
curl -s -X POST "$BASE_URL/v1/users/user-001/wallet/add" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1000.0,
    "payment_method": "upi",
    "transaction_reference": "UPI-DEMO-123456"
  }' | jq '.'

FINAL_BALANCE=$(curl -s "$BASE_URL/v1/users/user-001/balance" | jq -r '.balance')
echo -e "\nFinal balance: ${GREEN}₹${FINAL_BALANCE}${NC}"
sleep 2

# 9. Search by Category
print_header "9️⃣  Search by Category"
print_step "Searching for 'Food & Beverage' vouchers..."
curl -s "$BASE_URL/v1/vouchers/search?category=Food%20%26%20Beverage&page=1&page_size=5" | jq '.'
sleep 2

# Summary
print_header "📊 Demo Summary"
echo -e "${GREEN}✅ Completed operations:${NC}"
echo "  1. ✓ Searched for vouchers"
echo "  2. ✓ Retrieved voucher details"
echo "  3. ✓ Checked wallet balance"
echo "  4. ✓ Purchased voucher"
echo "  5. ✓ Listed user vouchers"
echo "  6. ✓ Viewed transactions"
echo "  7. ✓ Added money to wallet"
echo "  8. ✓ Searched by category"
echo ""
echo -e "${YELLOW}🔍 Explore more:${NC}"
echo "  - Jaeger Traces:  http://localhost:16686"
echo "  - RabbitMQ:       http://localhost:15672 (guest/guest)"
echo "  - API Docs:       See API_GUIDE.md"
echo ""
print_header "🎉 Demo Complete!"


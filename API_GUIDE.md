# API Usage Guide

This guide provides detailed examples for using the Voucher & Payment Service API.

## Authentication

Currently, the service uses a mock authentication interceptor. In production, add a JWT token:

```bash
curl -H "Authorization: Bearer <token>" ...
```

## Voucher Operations

### 1. Search Vouchers

Find vouchers by query, category, brand, or price range.

**Request:**
```bash
curl -X GET "http://localhost:8080/v1/vouchers/search?query=amazon&category=E-Commerce&min_price=100&max_price=500&page=1&page_size=10"
```

**Response:**
```json
{
  "vouchers": [
    {
      "id": "voucher-001",
      "name": "Amazon Gift Card ₹500",
      "description": "Amazon shopping voucher worth ₹500",
      "brand": "Amazon",
      "category": "E-Commerce",
      "face_value": 500.0,
      "discount_percentage": 5.0,
      "selling_price": 475.0,
      "stock_quantity": 100,
      "is_active": true,
      "valid_from": "2024-01-01T00:00:00Z",
      "valid_until": "2025-01-01T00:00:00Z",
      "terms_and_conditions": "Valid for 1 year. Non-refundable."
    }
  ],
  "total_count": 1,
  "page": 1,
  "page_size": 10
}
```

### 2. Get Voucher Details

Get detailed information about a specific voucher.

**Request:**
```bash
curl -X GET "http://localhost:8080/v1/vouchers/voucher-001"
```

**Response:**
```json
{
  "id": "voucher-001",
  "name": "Amazon Gift Card ₹500",
  "description": "Amazon shopping voucher worth ₹500",
  "brand": "Amazon",
  "category": "E-Commerce",
  "face_value": 500.0,
  "discount_percentage": 5.0,
  "selling_price": 475.0,
  "stock_quantity": 100,
  "is_active": true
}
```

### 3. Buy Voucher

Purchase one or more vouchers using wallet or UPI.

**Request (Wallet):**
```bash
curl -X POST "http://localhost:8080/v1/vouchers/buy" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-001",
    "voucher_id": "voucher-001",
    "quantity": 2,
    "payment_method": "wallet"
  }'
```

**Request (UPI):**
```bash
curl -X POST "http://localhost:8080/v1/vouchers/buy" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-001",
    "voucher_id": "voucher-001",
    "quantity": 1,
    "payment_method": "upi",
    "upi_id": "user@paytm"
  }'
```

**Response:**
```json
{
  "transaction_id": "txn-1234567890",
  "status": "completed",
  "amount_paid": 950.0,
  "purchased_vouchers": [
    {
      "id": "pv-001",
      "voucher_id": "voucher-001",
      "voucher_code": "AMZN-1234-5678-9012",
      "pin": "1234",
      "status": "active",
      "valid_until": "2025-01-01T00:00:00Z"
    },
    {
      "id": "pv-002",
      "voucher_id": "voucher-001",
      "voucher_code": "AMZN-3456-7890-1234",
      "pin": "5678",
      "status": "active",
      "valid_until": "2025-01-01T00:00:00Z"
    }
  ],
  "purchased_at": "2024-01-15T10:30:00Z"
}
```

### 4. List User Vouchers

Get all vouchers owned by a user.

**Request:**
```bash
curl -X GET "http://localhost:8080/v1/users/user-001/vouchers?status=active&page=1&page_size=10"
```

**Response:**
```json
{
  "vouchers": [
    {
      "id": "pv-001",
      "voucher_id": "voucher-001",
      "voucher_code": "AMZN-1234-5678-9012",
      "pin": "1234",
      "status": "active",
      "valid_until": "2025-01-01T00:00:00Z"
    }
  ],
  "total_count": 1,
  "page": 1,
  "page_size": 10
}
```

## Wallet Operations

### 1. Get Balance

Check wallet balance for a user.

**Request:**
```bash
curl -X GET "http://localhost:8080/v1/users/user-001/balance"
```

**Response:**
```json
{
  "user_id": "user-001",
  "balance": 5000.0,
  "currency": "INR"
}
```

### 2. Add Money

Add money to wallet via UPI/card.

**Request:**
```bash
curl -X POST "http://localhost:8080/v1/users/user-001/wallet/add" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1000.0,
    "payment_method": "upi",
    "transaction_reference": "UPI-REF-123456"
  }'
```

**Response:**
```json
{
  "transaction_id": "txn-9876543210",
  "status": "completed",
  "new_balance": 6000.0,
  "completed_at": "2024-01-15T11:00:00Z"
}
```

### 3. List Transactions

Get transaction history.

**Request:**
```bash
curl -X GET "http://localhost:8080/v1/users/user-001/transactions?type=all&page=1&page_size=20"
```

**Response:**
```json
{
  "transactions": [
    {
      "id": "txn-1234567890",
      "user_id": "user-001",
      "type": "debit",
      "amount": 950.0,
      "balance_before": 6000.0,
      "balance_after": 5050.0,
      "description": "Voucher purchase: Amazon Gift Card ₹500 x2",
      "reference_id": "voucher-001",
      "status": "completed",
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "txn-9876543210",
      "user_id": "user-001",
      "type": "credit",
      "amount": 1000.0,
      "balance_before": 5000.0,
      "balance_after": 6000.0,
      "description": "Money added via upi",
      "reference_id": "UPI-REF-123456",
      "status": "completed",
      "created_at": "2024-01-15T11:00:00Z"
    }
  ],
  "total_count": 2,
  "page": 1,
  "page_size": 20
}
```

## gRPC Usage

### Using grpcurl

Install grpcurl:
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**List services:**
```bash
grpcurl -plaintext localhost:50051 list
```

**List methods:**
```bash
grpcurl -plaintext localhost:50051 list voucher.v1.VoucherService
```

**Call a method:**
```bash
grpcurl -plaintext -d '{
  "user_id": "user-001",
  "voucher_id": "voucher-001",
  "quantity": 1,
  "payment_method": "wallet"
}' localhost:50051 voucher.v1.VoucherService/BuyVoucher
```

### Using Go Client

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    
    voucherv1 "github.com/voucher-payment-service/api/generated"
)

func main() {
    // Connect to server
    conn, err := grpc.Dial("localhost:50051", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    // Create client
    client := voucherv1.NewVoucherServiceClient(conn)
    
    // Search vouchers
    resp, err := client.SearchVouchers(context.Background(), 
        &voucherv1.SearchVouchersRequest{
            Query:    "amazon",
            Category: "E-Commerce",
            Page:     1,
            PageSize: 10,
        })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d vouchers", resp.TotalCount)
}
```

## Error Handling

### Common Error Codes

| gRPC Code | HTTP Status | Description |
|-----------|-------------|-------------|
| OK | 200 | Success |
| INVALID_ARGUMENT | 400 | Invalid request parameters |
| UNAUTHENTICATED | 401 | Missing or invalid auth |
| NOT_FOUND | 404 | Resource not found |
| FAILED_PRECONDITION | 412 | Business rule violation |
| INTERNAL | 500 | Internal server error |

### Error Response Format

```json
{
  "error": {
    "code": "INSUFFICIENT_BALANCE",
    "message": "Insufficient wallet balance",
    "details": []
  }
}
```

### Common Error Scenarios

**Insufficient Balance:**
```bash
# Returns 412 Precondition Failed
curl -X POST "http://localhost:8080/v1/vouchers/buy" \
  -d '{"user_id": "user-001", "voucher_id": "voucher-001", "quantity": 100, "payment_method": "wallet"}'
```

**Out of Stock:**
```bash
# Returns 412 Precondition Failed
curl -X POST "http://localhost:8080/v1/vouchers/buy" \
  -d '{"user_id": "user-001", "voucher_id": "voucher-001", "quantity": 999, "payment_method": "wallet"}'
```

**Not Found:**
```bash
# Returns 404 Not Found
curl -X GET "http://localhost:8080/v1/vouchers/invalid-id"
```

## Rate Limiting

The service implements rate limiting to prevent abuse:

- **Default**: 100 requests per second per client
- **Burst**: 200 requests
- **Response**: 429 Too Many Requests

## Best Practices

1. **Use pagination** for large result sets
2. **Cache responses** on client side
3. **Handle errors gracefully** with retries
4. **Use timeouts** for all requests
5. **Monitor request IDs** for debugging
6. **Validate input** before sending requests

## Testing with Postman

Import the API collection (coming soon) or use the examples above to create your own collection.

## Additional Resources

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
- [RESTful API Best Practices](https://restfulapi.net/)


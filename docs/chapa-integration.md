# Chapa Gateway Integration

## Overview

This integration adds Chapa as a payment gateway in UniBee merchant configuration and checkout.

It supports:

- Gateway setup in admin Payment Gateways tab
- Checkout link creation via Chapa initialize API
- Payment verification via Chapa verify API
- Redirect/callback handling for hosted checkout
- Webhook event handling with optional HMAC signature verification
- Transaction cancellation API call

## Where It Appears

After backend deployment, Chapa appears in merchant setup list endpoint:

- `GET /merchant/gateway/setup_list`

That endpoint is used by admin portal configuration page:

- `http://localhost:5175/configuration?tab=paymentGateways`

## Required Merchant Configuration

In gateway setup modal for Chapa:

- `Gateway Key`: Chapa secret key (`CHASECK_...` or `CHASECK_TEST_...`)
- `Gateway Secret`: optional webhook secret (if empty, key is used as fallback)

## Chapa API Endpoints Used

- Initialize payment: `POST /v1/transaction/initialize`
- Verify payment: `GET /v1/transaction/verify/{tx_ref}`
- Cancel payment: `PUT /v1/transaction/cancel/{tx_ref}`

Base URL:

- `https://api.chapa.co`

## Checkout Flow

1. UniBee creates payment and sets `tx_ref = paymentId`.
2. UniBee calls Chapa initialize API.
3. UniBee returns Chapa `checkout_url` to frontend as payment link.
4. User completes payment on Chapa hosted page.
5. Chapa redirects back to UniBee redirect entrance.
6. UniBee verifies payment status via Chapa verify API.
7. UniBee marks payment success/failure and redirects to merchant return URL.

## Webhook Notes

- Endpoint format in UniBee: `/payment/gateway_webhook_entry/{gatewayId}/notifications`
- Signature headers supported: `x-chapa-signature`, `chapa-signature`
- Signature algorithm: `HMAC SHA256` over raw body using webhook secret
- If no signature header is sent, webhook payload is still processed

## Limitations

- Refund operations are not implemented in this adapter yet
- Stored payment methods and auto-charge operations are not implemented
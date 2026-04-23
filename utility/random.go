package utility

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/gogf/gf/v2/os/gtime"
	"math/big"
	"runtime"
	"time"
)

func GetLineSeparator() string {
	switch runtime.GOOS {
	case "windows":
		return "\r\n"
	default:
		return "\n"
	}
}

func CurrentTimeMillis() (s int64) {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GenerateRandomAlphanumeric(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(fmt.Sprintf("crypto/rand failed: %v", err))
		}
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

func JodaTimePrefix() (prefix string) {
	return time.Now().Format("20060102")
}

func CreateHistoryEventId() string {
	return fmt.Sprintf("hisev%s%s", JodaTimePrefix(), GenerateRandomAlphanumeric(15))
}

func CreateEventId() string {
	return fmt.Sprintf("ev%s%s", JodaTimePrefix(), GenerateRandomAlphanumeric(15))
}

func CreateSessionId(userId string) string {
	return fmt.Sprintf("us%s%s%s", userId, JodaTimePrefix(), GenerateRandomAlphanumeric(15))
}

func CreateRequestId() string {
	return fmt.Sprintf("req%s%s", JodaTimePrefix(), GenerateRandomAlphanumeric(15))
}

func CreateCreditRechargeId() string {
	return fmt.Sprintf("crrecharge%d%s", gtime.Now().Timestamp(), GenerateRandomAlphanumeric(8))
}

func CreateCreditPaymentId() string {
	return fmt.Sprintf("crpayment%d%s", gtime.Now().Timestamp(), GenerateRandomAlphanumeric(8))
}

func CreateCreditRefundId() string {
	return fmt.Sprintf("crrefund%d%s", gtime.Now().Timestamp(), GenerateRandomAlphanumeric(8))
}

func CreateSubscriptionId() string {
	return fmt.Sprintf("sub%s%s", JodaTimePrefix(), GenerateRandomAlphanumeric(15))
}

func CreateInvoiceId() string {
	return fmt.Sprintf("8%d%s", gtime.Now().Timestamp(), GenerateRandomAlphanumeric(8))
}

func CreateInvoiceSt() string {
	return fmt.Sprintf("iv%s%s", JodaTimePrefix(), GenerateRandomAlphanumeric(30))
}

func CreatePendingUpdateId() string {
	return fmt.Sprintf("subup%s%s", JodaTimePrefix(), GenerateRandomAlphanumeric(15))
}

func CreatePaymentId() string {
	return fmt.Sprintf("pay%s%s", JodaTimePrefix(), GenerateRandomAlphanumeric(15))
}

func CreateRefundId() string {
	return fmt.Sprintf("ref%s%s", JodaTimePrefix(), GenerateRandomAlphanumeric(15))
}

const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func GenerateRandomCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(fmt.Sprintf("crypto/rand failed: %v", err))
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func GenerateRandomNumber(length int) string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return fmt.Sprintf("%06v", n.Int64())
}

func GenerateRandomOpenApiKey(length int) (string, error) {
	// We need enough raw bytes so that after base64 encoding the result is at
	// least `length` characters long.  base64 expands by ~4/3, so reading
	// ceil(length * 3 / 4) bytes is always sufficient.
	rawLen := (length*3 + 3) / 4
	key := make([]byte, rawLen)

	// Read cryptographically secure random bytes.
	if _, err := rand.Read(key); err != nil {
		return "", err
	}

	encodedKey := base64.URLEncoding.EncodeToString(key)
	if len(encodedKey) < length {
		return "", fmt.Errorf("GenerateRandomOpenApiKey: encoded length %d is less than requested %d", len(encodedKey), length)
	}
	return encodedKey[:length], nil
}

func Base64EncodeToString(source string) string {
	if source == "" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(source))
}

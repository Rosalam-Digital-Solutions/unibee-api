package merchant_config_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"unibee/internal/logic/merchant_config"
	"unibee/internal/logic/merchant_config/update"
	entity "unibee/internal/model/entity/default"
	"unibee/test"
)

func TestMerchantConfig(t *testing.T) {
	ctx := context.Background()
	var one *entity.MerchantConfig
	var err error
	t.Run("Test for Merchant Config Set|Get", func(t *testing.T) {
		one = merchant_config.GetMerchantConfig(ctx, test.TestMerchant.Id, "test_config_key")
		require.Equal(t, true, one == nil || len(one.ConfigValue) == 0)
		err = update.SetMerchantConfig(ctx, test.TestMerchant.Id, "test_config_key", "test")
		require.Nil(t, err)
		one = merchant_config.GetMerchantConfig(ctx, test.TestMerchant.Id, "test_config_key")
		require.Equal(t, true, one != nil && len(one.ConfigValue) > 0)
	})
	t.Run("Test for Clean Merchant Config", func(t *testing.T) {
		err = update.SetMerchantConfig(ctx, test.TestMerchant.Id, "test_config_key", "")
		require.Nil(t, err)
		one = merchant_config.GetMerchantConfig(ctx, test.TestMerchant.Id, "test_config_key")
		require.Equal(t, true, one == nil || len(one.ConfigValue) == 0)
	})
}

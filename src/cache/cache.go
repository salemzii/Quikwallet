package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	entity "github.com/salemzii/Quikwallet/src/entities"
	"github.com/shopspring/decimal"
)

var mycache *cache.Cache

func init() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	mycache = cache.New(&cache.Options{
		Redis: rdb,
		// cache 1,000 keys for 1 minute
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
}

func SetWalletBalanceInCache(wallet_id int, wallet_balance decimal.Decimal) error {
	ctx := context.TODO()
	key := strconv.Itoa(wallet_id)
	wallet := &entity.Wallet{
		Balance: wallet_balance,
	}
	if err := mycache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: wallet,
		TTL:   time.Hour,
	}); err != nil {
		return err
	}
	return nil
}

func GetWalletBalanceInCache(wallet_id int) (w *entity.Wallet, e error) {
	ctx := context.TODO()
	key := strconv.Itoa(wallet_id)
	var wallet entity.Wallet
	if err := mycache.Get(ctx, key, &wallet); err == nil {
		return &wallet, nil
	} else {
		return &wallet, err
	}
}

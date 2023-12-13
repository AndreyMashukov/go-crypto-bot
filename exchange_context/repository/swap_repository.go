package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	model "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
	"time"
)

type SwapRepository struct {
	DB         *sql.DB
	RDB        *redis.Client
	Ctx        *context.Context
	CurrentBot *model.Bot
}

func (s *SwapRepository) GetSwapChain(hash string) (model.SwapChainEntity, error) {
	var swapChain model.SwapChainEntity
	swapChain.SwapOne = &model.SwapTransitionEntity{}
	swapChain.SwapTwo = &model.SwapTransitionEntity{}
	swapChain.SwapThree = &model.SwapTransitionEntity{}

	err := s.DB.QueryRow(`
		SELECT
		    sc.id as Id,
		    sc.title as Title,
		    sc.type as Type,
		    sc.hash as Hash,
		    sc.percent as Percent,
		    sc.timestamp as Timestamp,
		    one.id as OneId,
		    one.type as OneType,
		    one.symbol as OneSymbol,
		    one.base_asset as OneBaseAsset,
		    one.quote_asset as OneQuoteAsset,
		    one.operation as OneOperation,
		    one.quantity as OneQuantity,
		    one.price as OnePrice,
		    one.level as OneLevel,
		    two.id as TwoId,
		    two.type as TwoType,
		    two.symbol as TwoSymbol,
		    two.base_asset as TwoBaseAsset,
		    two.quote_asset as TwoQuoteAsset,
		    two.operation as TwoOperation,
		    two.quantity as TwoQuantity,
		    two.price as TwoPrice,
		    two.level as TwoLevel,
		    three.id as ThreeId,
		    three.type as ThreeType,
		    three.symbol as ThreeSymbol,
		    three.base_asset as ThreeBaseAsset,
		    three.quote_asset as ThreeQuoteAsset,
		    three.operation as ThreeOperation,
		    three.quantity as ThreeQuantity,
		    three.price as ThreePrice,
		    three.level as ThreeLevel
		FROM swap_chain sc
		INNER JOIN swap_transition one ON one.id = sc.swap_one
		INNER JOIN swap_transition two ON two.id = sc.swap_two
		INNER JOIN swap_transition three ON three.id = sc.swap_three
		WHERE sc.hash = ?
	`,
		hash,
	).Scan(
		&swapChain.Id,
		&swapChain.Title,
		&swapChain.Type,
		&swapChain.Hash,
		&swapChain.Percent,
		&swapChain.Timestamp,
		&swapChain.SwapOne.Id,
		&swapChain.SwapOne.Type,
		&swapChain.SwapOne.Symbol,
		&swapChain.SwapOne.BaseAsset,
		&swapChain.SwapOne.QuoteAsset,
		&swapChain.SwapOne.Operation,
		&swapChain.SwapOne.Quantity,
		&swapChain.SwapOne.Price,
		&swapChain.SwapOne.Level,
		&swapChain.SwapTwo.Id,
		&swapChain.SwapTwo.Type,
		&swapChain.SwapTwo.Symbol,
		&swapChain.SwapTwo.BaseAsset,
		&swapChain.SwapTwo.QuoteAsset,
		&swapChain.SwapTwo.Operation,
		&swapChain.SwapTwo.Quantity,
		&swapChain.SwapTwo.Price,
		&swapChain.SwapTwo.Level,
		&swapChain.SwapThree.Id,
		&swapChain.SwapThree.Type,
		&swapChain.SwapThree.Symbol,
		&swapChain.SwapThree.BaseAsset,
		&swapChain.SwapThree.QuoteAsset,
		&swapChain.SwapThree.Operation,
		&swapChain.SwapThree.Quantity,
		&swapChain.SwapThree.Price,
		&swapChain.SwapThree.Level,
	)
	if err != nil {
		return swapChain, err
	}

	return swapChain, nil
}

func (s *SwapRepository) CreateSwapTransition(transition model.SwapTransitionEntity) (*int64, error) {
	res, err := s.DB.Exec(`
		INSERT INTO swap_transition SET
		    type = ?,
		    symbol = ?,
		    base_asset = ?,
		    quote_asset = ?,
		    operation = ?,
		    quantity = ?,
		    price = ?,
		    level = ?
	`,
		transition.Type,
		transition.Symbol,
		transition.BaseAsset,
		transition.QuoteAsset,
		transition.Operation,
		transition.Quantity,
		transition.Price,
		transition.Level,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	lastId, err := res.LastInsertId()

	return &lastId, err
}

func (s *SwapRepository) CreateSwapChain(swapChain model.SwapChainEntity) (*int64, error) {
	_, _ = s.DB.Exec("START TRANSACTION")
	swapIdOne, _ := s.CreateSwapTransition(*swapChain.SwapOne)
	swapIdTwo, _ := s.CreateSwapTransition(*swapChain.SwapTwo)
	swapIdThree, _ := s.CreateSwapTransition(*swapChain.SwapThree)

	res, err := s.DB.Exec(`
		INSERT INTO swap_chain SET
		    title = ?,
		    type = ?,
		    hash = ?,
		    percent = ?,
		    timestamp = ?,
		    swap_one = ?,
		    swap_two = ?,
		    swap_three = ?
	`,
		swapChain.Title,
		swapChain.Type,
		swapChain.Hash,
		swapChain.Percent,
		swapChain.Timestamp,
		swapIdOne,
		swapIdTwo,
		swapIdThree,
	)

	if err != nil {
		_, _ = s.DB.Exec("ROLLBACK")
		log.Println(err)
		return nil, err
	}

	_, _ = s.DB.Exec("COMMIT")

	lastId, err := res.LastInsertId()

	return &lastId, err
}

func (s *SwapRepository) UpdateSwapChain(swapChain model.SwapChainEntity) error {
	_, _ = s.DB.Exec("START TRANSACTION")
	_ = s.UpdateSwapTransition(*swapChain.SwapOne)
	_ = s.UpdateSwapTransition(*swapChain.SwapTwo)
	_ = s.UpdateSwapTransition(*swapChain.SwapThree)

	_, err := s.DB.Exec(`
		UPDATE swap_chain SET
		    title = ?,
		    type = ?,
		    hash = ?,
		    percent = ?,
		    timestamp = ?
		WHERE id = ?
	`,
		swapChain.Title,
		swapChain.Type,
		swapChain.Hash,
		swapChain.Percent,
		swapChain.Timestamp,
		swapChain.Id,
	)

	if err != nil {
		_, _ = s.DB.Exec("ROLLBACK")
		log.Println(err)
		return err
	}

	_, _ = s.DB.Exec("COMMIT")

	return nil
}

func (s *SwapRepository) UpdateSwapTransition(transition model.SwapTransitionEntity) error {
	_, err := s.DB.Exec(`
		UPDATE swap_transition st SET
		    st.type = ?,
		    st.symbol = ?,
		    st.base_asset = ?,
		    st.quote_asset = ?,
		    st.operation = ?,
		    st.quantity = ?,
		    st.price = ?,
		    st.level = ?
		WHERE st.id = ?
	`,
		transition.Type,
		transition.Symbol,
		transition.BaseAsset,
		transition.QuoteAsset,
		transition.Operation,
		transition.Quantity,
		transition.Price,
		transition.Level,
		transition.Id,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *SwapRepository) SaveSwapChainCache(asset string, entity model.SwapChainEntity) {
	encoded, _ := json.Marshal(entity)

	s.RDB.Set(*s.Ctx, s.getSwapCacheKey(asset), string(encoded), time.Minute*10)
}

func (s *SwapRepository) GetSwapChainCache(asset string) *model.SwapChainEntity {
	cached := s.RDB.Get(*s.Ctx, s.getSwapCacheKey(asset)).Val()

	if len(cached) > 0 {
		var entity model.SwapChainEntity
		json.Unmarshal([]byte(cached), &entity)

		return &entity
	}

	return nil
}

func (s *SwapRepository) getSwapCacheKey(asset string) string {
	return fmt.Sprintf("swap-chain-%s", asset)
}

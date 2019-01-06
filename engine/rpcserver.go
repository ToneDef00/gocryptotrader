package engine

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/thrasher-/gocryptotrader/currency/pair"

	"github.com/thrasher-/gocryptotrader/engine/events"

	"github.com/thrasher-/gocryptotrader/currency"
	"github.com/thrasher-/gocryptotrader/exchanges/assets"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/gctrpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// RPCServer struct
type RPCServer struct {
}

func StartRPCRESTProxy() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gctrpc.RegisterGoCryptoTraderHandlerFromEndpoint(ctx, mux, ":4444", opts)
	if err != nil {
		log.Fatalf("failed to register gRPC proxy server: %s", err)
	}

	http.ListenAndServe(":4445", mux)
}

func StartRPCServer() {
	log.Printf("Starting gRPC server on port 4444")
	lis, err := net.Listen("tcp", ":4444")
	if err != nil {
		log.Fatalf("failed to bind to port")
	}

	s := RPCServer{}
	server := grpc.NewServer()
	gctrpc.RegisterGoCryptoTraderServer(server, &s)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to start gRPC server: %s", err)
	}
}

func (s *RPCServer) GetExchanges(ctx context.Context, r *gctrpc.GetExchangesRequest) (*gctrpc.GetExchangesResponse, error) {
	exchanges := common.JoinStrings(GetExchanges(r.Enabled), ",")
	return &gctrpc.GetExchangesResponse{Exchanges: exchanges}, nil
}

func (s *RPCServer) DisableExchange(ctx context.Context, r *gctrpc.GenericExchangeNameRequest) (*gctrpc.GenericExchangeNameResponse, error) {
	err := UnloadExchange(r.Exchange)
	return &gctrpc.GenericExchangeNameResponse{}, err
}

func (s *RPCServer) EnableExchange(ctx context.Context, r *gctrpc.GenericExchangeNameRequest) (*gctrpc.GenericExchangeNameResponse, error) {
	err := LoadExchange(r.Exchange, false, nil)
	return &gctrpc.GenericExchangeNameResponse{}, err
}

func (s *RPCServer) GetTicker(ctx context.Context, r *gctrpc.GetTickerRequest) (*gctrpc.TickerResponse, error) {
	t, err := GetSpecificTicker(
		pair.CurrencyPair{
			Delimiter:      r.Pair.Delimiter,
			FirstCurrency:  pair.CurrencyItem(r.Pair.FirstCurrency),
			SecondCurrency: pair.CurrencyItem(r.Pair.SecondCurrency),
		},
		r.Exchange,
		assets.AssetType(r.AssetType),
	)
	if err != nil {
		return nil, err
	}

	resp := &gctrpc.TickerResponse{
		Pair:         r.Pair,
		LastUpdated:  t.LastUpdated.Unix(),
		CurrencyPair: t.CurrencyPair,
		Last:         t.Last,
		High:         t.High,
		Low:          t.Low,
		Bid:          t.Bid,
		Ask:          t.Ask,
		Volume:       t.Volume,
		PriceAth:     t.PriceATH,
	}

	return resp, nil
}

func (s *RPCServer) GetTickers(ctx context.Context, r *gctrpc.GetTickersRequest) (*gctrpc.GetTickersResponse, error) {
	activeTickers := GetAllActiveTickers()
	var tickers []*gctrpc.Tickers

	for x := range activeTickers {
		var ticker gctrpc.Tickers
		ticker.Exchange = activeTickers[x].ExchangeName
		for y := range activeTickers[x].ExchangeValues {
			t := activeTickers[x].ExchangeValues[y]
			ticker.Tickers = append(ticker.Tickers, &gctrpc.TickerResponse{
				Pair: &gctrpc.CurrencyPair{
					Delimiter:      t.Pair.Delimiter,
					FirstCurrency:  t.Pair.FirstCurrency.String(),
					SecondCurrency: t.Pair.SecondCurrency.String(),
				},
				LastUpdated:  t.LastUpdated.Unix(),
				CurrencyPair: t.CurrencyPair,
				Last:         t.Last,
				High:         t.High,
				Low:          t.Low,
				Bid:          t.Bid,
				Ask:          t.Ask,
				Volume:       t.Volume,
				PriceAth:     t.PriceATH,
			})
		}
		tickers = append(tickers, &ticker)
	}

	return &gctrpc.GetTickersResponse{Tickers: tickers}, nil
}

func (s *RPCServer) GetOrderbook(ctx context.Context, r *gctrpc.GetOrderbookRequest) (*gctrpc.GetOrderbookResponse, error) {
	ob, err := GetSpecificOrderbook(
		pair.CurrencyPair{
			Delimiter:      r.Pair.Delimiter,
			FirstCurrency:  pair.CurrencyItem(r.Pair.FirstCurrency),
			SecondCurrency: pair.CurrencyItem(r.Pair.SecondCurrency),
		},
		r.Exchange,
		assets.AssetType(r.AssetType),
	)
	if err != nil {
		return nil, err
	}

	var bids []*gctrpc.OrderbookItem
	for x := range ob.Bids {
		bids = append(bids, &gctrpc.OrderbookItem{
			Amount: ob.Bids[x].Amount,
			Price:  ob.Bids[x].Price,
		})
	}

	var asks []*gctrpc.OrderbookItem
	for x := range ob.Asks {
		asks = append(asks, &gctrpc.OrderbookItem{
			Amount: ob.Asks[x].Amount,
			Price:  ob.Asks[x].Price,
		})
	}

	resp := &gctrpc.GetOrderbookResponse{
		Pair:        r.Pair,
		Bids:        bids,
		Asks:        asks,
		LastUpdated: ob.LastUpdated.Unix(),
		AssetType:   r.AssetType,
	}

	return resp, nil
}

func (s *RPCServer) GetOrderbooks(ctx context.Context, r *gctrpc.GetOrderbooksRequest) (*gctrpc.GetOrderbooksResponse, error) {
	return &gctrpc.GetOrderbooksResponse{}, common.ErrNotYetImplemented
}

func (s *RPCServer) GetConfig(ctx context.Context, r *gctrpc.GetConfigRequest) (*gctrpc.GetConfigResponse, error) {
	return &gctrpc.GetConfigResponse{}, common.ErrNotYetImplemented
}

func (s *RPCServer) GetPortfolio(ctx context.Context, r *gctrpc.GetPortfolioRequest) (*gctrpc.GetPortfolioResponse, error) {
	var addrs []*gctrpc.PortfolioAddress
	botAddrs := Bot.Portfolio.Addresses

	for x := range botAddrs {
		addrs = append(addrs, &gctrpc.PortfolioAddress{
			Address:     botAddrs[x].Address,
			CoinType:    botAddrs[x].CoinType,
			Description: botAddrs[x].Description,
			Balance:     botAddrs[x].Balance,
		})
	}

	resp := &gctrpc.GetPortfolioResponse{
		Portfolio: addrs,
	}

	return resp, nil
}

func (s *RPCServer) AddPortfolioAddress(ctx context.Context, r *gctrpc.AddPortfolioAddressRequest) (*gctrpc.AddPortfolioAddressResponse, error) {
	Bot.Portfolio.AddAddress(r.Address, r.CoinType, r.Description, r.Balance)
	return &gctrpc.AddPortfolioAddressResponse{}, nil
}

func (s *RPCServer) RemovePortfolioAddress(ctx context.Context, r *gctrpc.RemovePortfolioAddressRequest) (*gctrpc.RemovePortfolioAddressResponse, error) {
	Bot.Portfolio.RemoveAddress(r.Address, r.CoinType, r.Description)
	return &gctrpc.RemovePortfolioAddressResponse{}, nil
}

func (s *RPCServer) GetForexRates(ctx context.Context, r *gctrpc.GetForexRatesRequest) (*gctrpc.GetForexRatesResponse, error) {
	rates := currency.GetExchangeRates()
	if rates == nil {
		return nil, fmt.Errorf("forex rates is empty")
	}
	return &gctrpc.GetForexRatesResponse{ForexRates: rates}, nil
}

func (s *RPCServer) GetOrders(ctx context.Context, r *gctrpc.GetOrdersRequest) (*gctrpc.GetOrdersResponse, error) {
	return &gctrpc.GetOrdersResponse{}, common.ErrNotYetImplemented
}

func (s *RPCServer) GetOrder(ctx context.Context, r *gctrpc.GetOrderRequest) (*gctrpc.OrderDetails, error) {
	return &gctrpc.OrderDetails{}, common.ErrNotYetImplemented
}

func (s *RPCServer) SubmitOrder(ctx context.Context, r *gctrpc.SubmitOrderRequest) (*gctrpc.SubmitOrderResponse, error) {
	return &gctrpc.SubmitOrderResponse{}, common.ErrNotYetImplemented
}

func (s *RPCServer) CancelOrder(ctx context.Context, r *gctrpc.CancelOrderRequest) (*gctrpc.CancelOrderResponse, error) {
	return &gctrpc.CancelOrderResponse{}, common.ErrNotYetImplemented
}

func (s *RPCServer) CancelAllOrders(ctx context.Context, r *gctrpc.CancelAllOrdersRequest) (*gctrpc.CancelAllOrdersResponse, error) {
	return &gctrpc.CancelAllOrdersResponse{}, common.ErrNotYetImplemented
}

func (s *RPCServer) GetEvents(ctx context.Context, r *gctrpc.GetEventsRequest) (*gctrpc.GetEventsResponse, error) {
	return &gctrpc.GetEventsResponse{}, common.ErrNotYetImplemented
}

func (s *RPCServer) AddEvent(ctx context.Context, r *gctrpc.AddEventRequest) (*gctrpc.AddEventResponse, error) {
	evtCondition := events.ConditionParams{
		CheckBids:        r.ConditionParams.CheckBids,
		CheckBidsAndAsks: r.ConditionParams.CheckBidsAndAsks,
		Condition:        r.ConditionParams.Condition,
		OrderbookAmount:  r.ConditionParams.OrderbookAmount,
		Price:            r.ConditionParams.Price,
	}

	p := pair.CurrencyPair{
		Delimiter:      r.Pair.Delimiter,
		FirstCurrency:  pair.CurrencyItem(r.Pair.FirstCurrency),
		SecondCurrency: pair.CurrencyItem(r.Pair.SecondCurrency),
	}

	id, err := events.Add(r.Exchange, r.Item, evtCondition, p, assets.AssetType(r.AssetType), r.Action)
	if err != nil {
		return nil, err
	}

	return &gctrpc.AddEventResponse{Id: id}, nil
}

func (s *RPCServer) RemoveEvent(ctx context.Context, r *gctrpc.RemoveEventRequest) (*gctrpc.RemoveEventResponse, error) {
	events.Remove(r.Id)
	return &gctrpc.RemoveEventResponse{}, nil
}

func (s *RPCServer) GetCryptocurrencyDepositAddresses(ctx context.Context, r *gctrpc.GetCryptocurrencyDepositAddressesRequest) (*gctrpc.GetCryptocurrencyDepositAddressesResponse, error) {
	return &gctrpc.GetCryptocurrencyDepositAddressesResponse{}, common.ErrNotYetImplemented
}

func (s *RPCServer) GetCryptocurrencyDepositAddress(ctx context.Context, r *gctrpc.GetCryptocurrencyDepositAddressRequest) (*gctrpc.GetCryptocurrencyDepositAddressResponse, error) {
	return &gctrpc.GetCryptocurrencyDepositAddressResponse{}, common.ErrNotYetImplemented
}

func (s *RPCServer) WithdrawCryptocurrencyFunds(ctx context.Context, r *gctrpc.WithdrawCurrencyRequest) (*gctrpc.WithdrawResponse, error) {
	return &gctrpc.WithdrawResponse{}, common.ErrNotYetImplemented
}

func (s *RPCServer) WithdrawFiatFunds(ctx context.Context, r *gctrpc.WithdrawCurrencyRequest) (*gctrpc.WithdrawResponse, error) {
	return &gctrpc.WithdrawResponse{}, common.ErrNotYetImplemented
}

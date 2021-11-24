package main

import (
	"log"
	"time"

	"gitlab.com/tsuchinaga/gridon"
)

func main() {
	service, err := gridon.NewService()
	if err != nil {
		log.Fatalln(err)
	}

	orderPassword := "Gmryy715"

	strategies := make([]*gridon.Strategy, 0)
	strategies = append(strategies, &gridon.Strategy{
		Code:            "1475-buy",
		SymbolCode:      "1475",
		Exchange:        gridon.ExchangeToushou,
		Product:         gridon.ProductMargin,
		MarginTradeType: gridon.MarginTradeTypeDay,
		EntrySide:       gridon.SideBuy,
		Cash:            100_000,
		RebalanceStrategy: gridon.RebalanceStrategy{
			Runnable: true,
			Timings:  []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)},
		},
		GridStrategy: gridon.GridStrategy{
			Runnable:      true,
			Width:         2,
			Quantity:      1,
			NumberOfGrids: 3,
			TimeRanges: []gridon.TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
			},
		},
		CancelStrategy: gridon.CancelStrategy{
			Runnable: true,
			Timings: []time.Time{
				time.Date(0, 1, 1, 12, 20, 0, 0, time.Local),
				time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
		},
		ExitStrategy: gridon.ExitStrategy{
			Runnable:      true,
			ExecutionType: gridon.ExecutionTypeMarketAfternoonClose,
			Timings:       []time.Time{time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
		},
		Account: gridon.Account{
			Password:    orderPassword,
			AccountType: gridon.AccountTypeSpecific,
		},
	})

	strategies = append(strategies, &gridon.Strategy{
		Code:            "1476-buy",
		SymbolCode:      "1476",
		Exchange:        gridon.ExchangeToushou,
		Product:         gridon.ProductMargin,
		MarginTradeType: gridon.MarginTradeTypeDay,
		EntrySide:       gridon.SideBuy,
		Cash:            100_000,
		RebalanceStrategy: gridon.RebalanceStrategy{
			Runnable: true,
			Timings:  []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)},
		},
		GridStrategy: gridon.GridStrategy{
			Runnable:      true,
			Width:         2,
			Quantity:      1,
			NumberOfGrids: 3,
			TimeRanges: []gridon.TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
			},
		},
		CancelStrategy: gridon.CancelStrategy{
			Runnable: true,
			Timings: []time.Time{
				time.Date(0, 1, 1, 12, 20, 0, 0, time.Local),
				time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
		},
		ExitStrategy: gridon.ExitStrategy{
			Runnable:      true,
			ExecutionType: gridon.ExecutionTypeMarketAfternoonClose,
			Timings:       []time.Time{time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
		},
		Account: gridon.Account{
			Password:    orderPassword,
			AccountType: gridon.AccountTypeSpecific,
		},
	})

	strategies = append(strategies, &gridon.Strategy{
		Code:            "1459-buy",
		SymbolCode:      "1459",
		Exchange:        gridon.ExchangeToushou,
		Product:         gridon.ProductMargin,
		MarginTradeType: gridon.MarginTradeTypeDay,
		EntrySide:       gridon.SideBuy,
		Cash:            80_000,
		RebalanceStrategy: gridon.RebalanceStrategy{
			Runnable: true,
			Timings:  []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)},
		},
		GridStrategy: gridon.GridStrategy{
			Runnable:      true,
			Width:         3,
			Quantity:      1,
			NumberOfGrids: 3,
			TimeRanges: []gridon.TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
			},
		},
		CancelStrategy: gridon.CancelStrategy{
			Runnable: true,
			Timings: []time.Time{
				time.Date(0, 1, 1, 12, 20, 0, 0, time.Local),
				time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
		},
		ExitStrategy: gridon.ExitStrategy{
			Runnable:      true,
			ExecutionType: gridon.ExecutionTypeMarketAfternoonClose,
			Timings:       []time.Time{time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
		},
		Account: gridon.Account{
			Password:    orderPassword,
			AccountType: gridon.AccountTypeSpecific,
		},
	})

	strategies = append(strategies, &gridon.Strategy{
		Code:            "1459-sell",
		SymbolCode:      "1459",
		Exchange:        gridon.ExchangeToushou,
		Product:         gridon.ProductMargin,
		MarginTradeType: gridon.MarginTradeTypeDay,
		EntrySide:       gridon.SideSell,
		Cash:            80_000,
		RebalanceStrategy: gridon.RebalanceStrategy{
			Runnable: true,
			Timings:  []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)},
		},
		GridStrategy: gridon.GridStrategy{
			Runnable:      true,
			Width:         3,
			Quantity:      1,
			NumberOfGrids: 3,
			TimeRanges: []gridon.TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
			},
		},
		CancelStrategy: gridon.CancelStrategy{
			Runnable: true,
			Timings: []time.Time{
				time.Date(0, 1, 1, 12, 20, 0, 0, time.Local),
				time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
		},
		ExitStrategy: gridon.ExitStrategy{
			Runnable:      true,
			ExecutionType: gridon.ExecutionTypeMarketAfternoonClose,
			Timings:       []time.Time{time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
		},
		Account: gridon.Account{
			Password:    orderPassword,
			AccountType: gridon.AccountTypeSpecific,
		},
	})

	strategies = append(strategies, &gridon.Strategy{
		Code:            "1458-buy",
		SymbolCode:      "1458",
		Exchange:        gridon.ExchangeToushou,
		Product:         gridon.ProductMargin,
		MarginTradeType: gridon.MarginTradeTypeDay,
		EntrySide:       gridon.SideBuy,
		Cash:            500_000,
		RebalanceStrategy: gridon.RebalanceStrategy{
			Runnable: true,
			Timings:  []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)},
		},
		GridStrategy: gridon.GridStrategy{
			Runnable:      true,
			Width:         3,
			Quantity:      1,
			NumberOfGrids: 3,
			TimeRanges: []gridon.TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
			},
		},
		CancelStrategy: gridon.CancelStrategy{
			Runnable: true,
			Timings: []time.Time{
				time.Date(0, 1, 1, 12, 20, 0, 0, time.Local),
				time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
		},
		ExitStrategy: gridon.ExitStrategy{
			Runnable:      true,
			ExecutionType: gridon.ExecutionTypeMarketAfternoonClose,
			Timings:       []time.Time{time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
		},
		Account: gridon.Account{
			Password:    orderPassword,
			AccountType: gridon.AccountTypeSpecific,
		},
	})

	strategies = append(strategies, &gridon.Strategy{
		Code:            "1458-sell",
		SymbolCode:      "1458",
		Exchange:        gridon.ExchangeToushou,
		Product:         gridon.ProductMargin,
		MarginTradeType: gridon.MarginTradeTypeDay,
		EntrySide:       gridon.SideSell,
		Cash:            500_000,
		RebalanceStrategy: gridon.RebalanceStrategy{
			Runnable: true,
			Timings:  []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)},
		},
		GridStrategy: gridon.GridStrategy{
			Runnable:      true,
			Width:         3,
			Quantity:      1,
			NumberOfGrids: 3,
			TimeRanges: []gridon.TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
			},
		},
		CancelStrategy: gridon.CancelStrategy{
			Runnable: true,
			Timings: []time.Time{
				time.Date(0, 1, 1, 12, 20, 0, 0, time.Local),
				time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
		},
		ExitStrategy: gridon.ExitStrategy{
			Runnable:      true,
			ExecutionType: gridon.ExecutionTypeMarketAfternoonClose,
			Timings:       []time.Time{time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
		},
		Account: gridon.Account{
			Password:    orderPassword,
			AccountType: gridon.AccountTypeSpecific,
		},
	})

	for _, strategy := range strategies {
		if err := service.SaveStrategy(strategy); err != nil {
			log.Fatalln(err)
		}
		time.Sleep(500 * time.Millisecond) // DBには非同期で反映するため、反映されるまで少し待機
	}
}

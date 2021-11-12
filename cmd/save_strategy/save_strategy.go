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

	strategies := make([]*gridon.Strategy, 0)
	strategies = append(strategies, &gridon.Strategy{
		Code:            "1475-buy",
		SymbolCode:      "1475",
		Exchange:        gridon.ExchangeToushou,
		Product:         gridon.ProductMargin,
		MarginTradeType: gridon.MarginTradeTypeDay,
		EntrySide:       gridon.SideBuy,
		Cash:            301_065,
		RebalanceStrategy: gridon.RebalanceStrategy{
			Runnable: true,
			Timings:  []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)},
		},
		GridStrategy: gridon.GridStrategy{
			Runnable:      true,
			Width:         1,
			Quantity:      2,
			NumberOfGrids: 5,
			TimeRanges: []gridon.TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)},
			},
		},
		CancelStrategy: gridon.CancelStrategy{
			Runnable: true,
			Timings: []time.Time{
				time.Date(0, 1, 1, 11, 30, 0, 0, time.Local),
				time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)},
		},
		ExitStrategy: gridon.ExitStrategy{
			Runnable: true,
			Timings:  []time.Time{time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
		},
		Account: gridon.Account{
			Password:    "Password1234",
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
		Cash:            299_440,
		RebalanceStrategy: gridon.RebalanceStrategy{
			Runnable: true,
			Timings:  []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)},
		},
		GridStrategy: gridon.GridStrategy{
			Runnable:      true,
			Width:         1,
			Quantity:      2,
			NumberOfGrids: 5,
			TimeRanges: []gridon.TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)},
			},
		},
		CancelStrategy: gridon.CancelStrategy{
			Runnable: true,
			Timings: []time.Time{
				time.Date(0, 1, 1, 11, 30, 0, 0, time.Local),
				time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)},
		},
		ExitStrategy: gridon.ExitStrategy{
			Runnable: true,
			Timings:  []time.Time{time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
		},
		Account: gridon.Account{
			Password:    "Password1234",
			AccountType: gridon.AccountTypeSpecific,
		},
	})

	for _, strategy := range strategies {
		if err := service.SaveStrategy(strategy); err != nil {
			log.Fatalln(err)
		}
	}

	time.Sleep(3 * time.Second) // DBには非同期で反映するため、反映されるまで少し待機
}

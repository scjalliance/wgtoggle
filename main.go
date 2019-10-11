// +build windows

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gentlemanautomaton/winservice"
	"github.com/jackpal/gateway"
)

func main() {
	var (
		app  = App()
		conf = AppConfig(app)
	)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		app.Usage(os.Args[1:])
		os.Exit(1)
	}

	tunnel := "WireGuardTunnel$" + conf.Tunnel
	svcFound, err := winservice.Exists(tunnel)
	if err != nil {
		fmt.Printf("Service check failed: %v.\n", err)
		os.Exit(2)
	}

	if !svcFound {
		// The WireGuard tunnel service doesn't exist. There's nothing to
		// manage so we just exit.
		return
	}

	fmt.Printf("Tunnel found: %s\n", tunnel)

	gw, err := gateway.DiscoverGateway()
	if err != nil {
		fmt.Printf("Gateway query failed: %s\n", tunnel)
	}

	fmt.Printf("Gateway discovered: %s\n", gw)

	gwMatched := false
	for i := range conf.Gateway {
		if gw.Equal(conf.Gateway[i]) {
			gwMatched = true
			break
		}
	}
	if gwMatched {
		fmt.Printf("Gateway matched: %s\n", gw)
	} else {
		fmt.Printf("Gateway not matched.\n")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if !gwMatched {
		fmt.Printf("Starting %s...", tunnel)
		if err := winservice.Start(ctx, tunnel); err != nil {
			fmt.Printf(" failed: %v\n", err)
			os.Exit(3)
		}
		fmt.Printf(" succeeded.")
	} else {
		fmt.Printf("Stopping %s...", tunnel)
		if err := winservice.Stop(ctx, tunnel); err != nil {
			fmt.Printf(" failed: %v\n", err)
			os.Exit(3)
		}
		fmt.Printf(" succeeded.")
	}
}

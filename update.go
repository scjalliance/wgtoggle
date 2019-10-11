// +build windows

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gentlemanautomaton/globber"

	"github.com/gentlemanautomaton/winservice"
	"github.com/jackpal/gateway"
	"gopkg.in/alecthomas/kingpin.v2"
)

// UpdateCommand returns an update command and configuration for app.
func UpdateCommand(app *kingpin.Application) (*kingpin.CmdClause, *UpdateConfig) {
	cmd := app.Command("update", "Updates the requested tunnel state as necessary.")
	conf := &UpdateConfig{}
	conf.Bind(cmd)
	return cmd, conf
}

// UpdateConfig holds configuration for the update command.
type UpdateConfig struct {
	Tunnel   string
	Gateways []string
}

// Bind binds the update configuration to the command.
func (conf *UpdateConfig) Bind(cmd *kingpin.CmdClause) {
	cmd.Flag("tunnel", "WireGuard tunnel name to toggle").Short('t').Required().StringVar(&conf.Tunnel)
	cmd.Flag("gateway", "Gateway IP addresses match").Short('g').Required().StringsVar(&conf.Gateways)
}

func update(ctx context.Context, conf UpdateConfig) {
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

	gateways := globber.NewSet(conf.Gateways...)
	matched := gateways.Match(gw.String())
	if matched {
		fmt.Printf("Gateway matched: %s\n", gw)
	} else {
		fmt.Printf("Gateway not matched.\n")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if !matched {
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

package client

import (
	"context"
	"sync"

	"github.com/dharmab/skyeye/pkg/sim"
)

type Client interface {
	sim.Sim
	Run(context.Context, *sync.WaitGroup) error
}

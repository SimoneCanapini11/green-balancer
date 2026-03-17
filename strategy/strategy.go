package strategy

import (
	"green-balancer/balancer"
)

// Interfaccia per tutti gli algoritmi
type Strategy interface {
	NextNode(pool *balancer.ServerPool) *balancer.Node
}
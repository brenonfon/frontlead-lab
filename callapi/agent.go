package callapi

import (
	"fmt"
	"log"
)

var (
	activeAgents = make(map[AgentPhone]*ActiveAgent)
)

type AgentPhone string

type SalesAgent struct {
	Customer    string
	Extension   string
	ExternalNr  string
	IsAvailable bool
}

func (s *SalesAgent) String() string {
	return fmt.Sprintf("%s@%s-%s", s.Extension, s.Customer, s.ExternalNr)
}

type ActiveAgent struct {
	customer string
	isReady  bool
}

func (s *SalesAgent) SetAvailability(isAvailable bool) {
	s.IsAvailable = isAvailable
}

// AddActiveAgent adds or updates an active agent for a customer
func AddActiveAgent(customer string, agent *SalesAgent) {
	activeAgents[AgentPhone(agent.ExternalNr)] = &ActiveAgent{
		customer: customer,
		isReady:  false,
	}
	log.Printf("[CallAPI] Added/Updated active agent for customer %s: agent=%s", customer, agent)
}

// SetActiveAgentReady updates the ready status of an active agent
func SetActiveAgentReady(agentExternalNr string, isReady bool) bool {
	if agent, exists := activeAgents[AgentPhone(agentExternalNr)]; exists {
		agent.isReady = isReady
		return true
	}
	return false
}

// IsAgentReady checks if an agent is ready for a given customer
func IsAgentReady(customer string) bool {
	for _, agent := range activeAgents {
		if agent.customer == customer {
			return agent.isReady
		}
	}
	return false
}

// GetAgentPhone retrieves the phone number of the active agent for a customer
func GetAgentPhone(customer string) (string, bool) {
	for agentPhone, agent := range activeAgents {
		if agent.customer == customer {
			return string(agentPhone), true
		}
	}
	return "", false
}

// GetCustomerByAgentPhone retrieves the customer phone number by agent phone
func GetCustomerByAgentPhone(agentPhone string) (string, bool) {
	if agent, exists := activeAgents[AgentPhone(agentPhone)]; exists {
		return agent.customer, true
	}
	return "", false
}

// GetActiveAgents returns a copy of all active agents for debugging
func GetActiveAgents() map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	for agentPhone, agent := range activeAgents {
		result[string(agentPhone)] = map[string]interface{}{
			"customer": agent.customer,
			"isReady":  agent.isReady,
		}
	}
	return result
}

// ClearActiveAgents removes all entries from the activeAgents map
func ClearActiveAgents() {
	activeAgents = make(map[AgentPhone]*ActiveAgent)
	log.Printf("[CallAPI] Cleared all active agents")
}

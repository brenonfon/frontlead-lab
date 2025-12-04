package callapi

import (
	"log"
)

var (
	activeAgents = make(map[Customer]*ActiveSalesAgent)
)

type Customer string

type SalesAgent struct {
	Phone       string
	IsAvailable bool
}

type ActiveSalesAgent struct {
	phone   string
	isReady bool
}

func (s *SalesAgent) SetAvailability(isAvailable bool) {
	s.IsAvailable = isAvailable
}

// AddActiveAgent adds or updates an active agent for a customer
func AddActiveAgent(customer string, agent *SalesAgent) {
	activeAgents[Customer(customer)] = &ActiveSalesAgent{
		phone:   agent.Phone,
		isReady: false,
	}
	log.Printf("[CallAPI] Added/Updated active agent for customer %s: agent=%s", customer, agent.Phone)
}

// DeleteActiveAgent removes an active agent for a customer
func DeleteActiveAgent(customer string) {
	if _, exists := activeAgents[Customer(customer)]; exists {
		delete(activeAgents, Customer(customer))
		log.Printf("[CallAPI] Deleted active agent for customer %s", customer)
	}
}

// SetActiveAgentReady updates the ready status of an active agent
func SetActiveAgentReady(customer string, isReady bool) bool {
	if agent, exists := activeAgents[Customer(customer)]; exists {
		agent.isReady = isReady
		log.Printf("[CallAPI] Updated ready status for customer %s: isReady=%t", customer, isReady)
		return true
	}
	log.Printf("[CallAPI] Customer %s not found in active agents", customer)
	return false
}

// IsAgentReady checks if an agent is ready for a given customer
func IsAgentReady(customer string) bool {
	if agent, exists := activeAgents[Customer(customer)]; exists {
		return agent.isReady
	}
	return false
}

// GetAgentPhone retrieves the phone number of the active agent for a customer
func GetAgentPhone(customer string) (string, bool) {
	if agent, exists := activeAgents[Customer(customer)]; exists {
		return agent.phone, true
	}
	return "", false
}

// GetActiveAgent retrieves the active agent for a customer
func GetActiveAgent(customer string) (*ActiveSalesAgent, bool) {
	agent, exists := activeAgents[Customer(customer)]
	return agent, exists
}

// GetAllActiveAgents returns all active agents
func GetAllActiveAgents() map[Customer]*ActiveSalesAgent {
	return activeAgents
}

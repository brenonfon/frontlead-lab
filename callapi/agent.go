package callapi

import (
	"fmt"
	"log"
)

var (
	activeAgents = make(map[Customer]*ActiveSalesAgent)
)

type Customer string

type SalesAgent struct {
	Customer    string
	Extension   string
	ExternalNr  string
	IsAvailable bool
}

func (s *SalesAgent) String() string {
	return fmt.Sprintf("%s@%s-%s", s.Extension, s.Customer, s.ExternalNr)
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
		phone:   agent.ExternalNr,
		isReady: false,
	}
	log.Printf("[CallAPI] Added/Updated active agent for customer %s: agent=%s", customer, agent)
}

// SetActiveAgentReady updates the ready status of an active agent
func SetActiveAgentReady(agentExternalNr string, isReady bool) bool {
	for _, agent := range activeAgents {
		if agent.phone == agentExternalNr {
			agent.isReady = isReady
			return true
		}
	}
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

// GetCustomerByAgentPhone retrieves the customer phone number by agent phone
func GetCustomerByAgentPhone(agentPhone string) (string, bool) {
	for customer, agent := range activeAgents {
		if agent.phone == agentPhone {
			return string(customer), true
		}
	}
	return "", false
}

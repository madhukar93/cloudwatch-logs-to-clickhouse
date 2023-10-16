package main

import (
	"encoding/json"

	uuid "github.com/satori/go.uuid"
)

type MockScenario int

const (
	Success MockScenario = iota
	InternalServerError
	DelayedResponse
	ConnectionReset
	NotFound
	ValidationFailure
)

func (m MockScenario) String() string {
	return [...]string{"success", "internalServerError", "delayedResponse", "connectionReset", "notFound", "validationFailure"}[m]
}

type event struct {
	TenantId             string
	ThirdPartyName       string
	EntityId             string
	EntityType           string
	OperationCategory    string
	OperationSubCategory string
	MockScenario         string
}

// return json of events
func payloads() ([][]byte, error) {
	e := []event{
		{
			TenantId:             "tenant1",
			ThirdPartyName:       "thirdPartyA",
			EntityId:             uuid.NewV4().String(),
			EntityType:           "product",
			OperationCategory:    "created",
			OperationSubCategory: "subCat1",
			MockScenario:         Success.String(),
		},
		{
			TenantId:             "tenant1",
			ThirdPartyName:       "thirdPartyB",
			EntityId:             uuid.NewV4().String(),
			EntityType:           "product",
			OperationCategory:    "priceUpdated",
			OperationSubCategory: "subCat2",
			MockScenario:         Success.String(),
		},
		{
			TenantId:             "tenant2",
			ThirdPartyName:       "thirdPartyA",
			EntityId:             uuid.NewV4().String(),
			EntityType:           "order",
			OperationCategory:    "checkout",
			OperationSubCategory: "subCat1",
			MockScenario:         InternalServerError.String(),
		},
		{
			TenantId:             "tenant2",
			ThirdPartyName:       "thirdPartyB",
			EntityId:             uuid.NewV4().String(),
			EntityType:           "order",
			OperationCategory:    "paid",
			OperationSubCategory: "subCat3",
			MockScenario:         DelayedResponse.String(),
		},
		{
			TenantId:             "tenant1",
			ThirdPartyName:       "thirdPartyA",
			EntityId:             uuid.NewV4().String(),
			EntityType:           "order",
			OperationCategory:    "enroute",
			OperationSubCategory: "subCat2",
			MockScenario:         ConnectionReset.String(),
		},
		{
			TenantId:             "tenant2",
			ThirdPartyName:       "thirdPartyB",
			EntityId:             uuid.NewV4().String(),
			EntityType:           "product",
			OperationCategory:    "created",
			OperationSubCategory: "subCat3",
			MockScenario:         NotFound.String(),
		},
		{
			TenantId:             "tenant1",
			ThirdPartyName:       "thirdPartyA",
			EntityId:             uuid.NewV4().String(),
			EntityType:           "order",
			OperationCategory:    "delivered",
			OperationSubCategory: "subCat1",
			MockScenario:         ValidationFailure.String(),
		},
		{
			TenantId:             "tenant2",
			ThirdPartyName:       "thirdPartyB",
			EntityId:             uuid.NewV4().String(),
			EntityType:           "product",
			OperationCategory:    "priceUpdated",
			OperationSubCategory: "subCat2",
			MockScenario:         DelayedResponse.String(),
		},
	}
	p := make([][]byte, len(e))
	for event := range e {
		payload, err := json.Marshal(e[event])
		if err != nil {
			return nil, err
		}
		p = append(p, payload)
	}
	return p, nil
}

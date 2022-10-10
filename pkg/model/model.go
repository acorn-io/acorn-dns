package model

import (
	"fmt"
)

const (
	RecordTypeA     = "A"
	RecordTypeAAAA  = "AAAA"
	RecordTypeCname = "CNAME"
	RecordTypeTxt   = "TXT"
)

func IsValidRecordType(rt string) error {
	switch rt {
	case RecordTypeA, RecordTypeAAAA, RecordTypeCname, RecordTypeTxt:
		return nil
	}

	return fmt.Errorf("invalid record type")
}

type DomainResponse struct {
	Name  string `json:"name,omitempty"`
	Token string `json:"token,omitempty"`
}

type RenewRequest struct {
	Records []RecordRequest `json:"records,omitempty"`
	Version string          `json:"version,omitempty"`
}

type RenewResponse struct {
	Name string `json:"name,omitempty"`
	// TODO - These records contain the "short" name, not the FQDN. Refactor the api to reflect that, but it needs
	// to be coordinated with acorn.
	OutOfSyncRecords []FQDNTypePair `json:"outOfSyncRecords,omitempty"`
}

type RecordRequest struct {
	Name   string   `json:"name,omitempty"`
	Type   string   `json:"type,omitempty"`
	Values []string `json:"values,omitempty"`
}

type RecordResponse struct {
	RecordRequest
	FQDN string `json:"fqdn,omitempty"`
}

type ErrorResponse struct {
	Status  int         `json:"status,omitempty"`
	Message string      `json:"msg,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type FQDNTypePair struct {
	FQDN string `json:"fqdn,omitempty"`
	Type string `json:"type,omitempty"`
}

package backend

import (
	"github.com/acorn-io/acorn-dns/pkg/db"
	"github.com/acorn-io/acorn-dns/pkg/model"
)

type Backend interface {
	GetDomain(domainName string) (db.Domain, error)
	CreateDomain() (model.DomainResponse, error)
	// TODO delete? not used
	DeleteDomain(domainName string) error
	Renew(domain string, domainID uint, records []model.RecordRequest) ([]model.FQDNTypePair, error)
	GetRootDomain() string
	CreateRecord(domain string, domainID uint, input model.RecordRequest) (model.RecordResponse, error)
	DeleteRecord(recordPrefix string, domain string, domainID uint) error
	StartPurgerDaemon(done <-chan struct{})
}

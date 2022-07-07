package backend

import (
	"strings"
	"time"

	"github.com/acorn-io/acorn-dns/pkg/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	intervalSeconds     int64 = 24 * 60 * 60      // Check for domains and records to purge once a day
	domainMaxAgeSeconds int64 = 30 * 24 * 60 * 60 // A domain is purged after 30 days
	recordMaxAgeSeconds int64 = 48 * 60 * 60      // A record is purged after 2 days
)

func (b *backend) StartPurgerDaemon(stopCh <-chan struct{}) {
	logrus.Info("starting purge daemon")
	wait.JitterUntil(b.purge, time.Duration(intervalSeconds)*time.Second, .002, true, stopCh)
}

func (b *backend) purge() {
	logrus.Infof("Beginning purge ☠️")

	domainDeleted, recordsDeleted, err := b.db.PurgeOldDomainsAndRecords(domainMaxAgeSeconds, recordMaxAgeSeconds)
	if err != nil {
		logrus.Errorf("problem purging old domains: %v", err)
	}
	logrus.Infof("Domains purged from DB: %v", domainDeleted)
	logrus.Infof("Records purged from DB: %v", recordsDeleted)

	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(b.ZoneID),
	}

	recordsToDelete := make(map[model.FQDNTypePair]*route53.ResourceRecordSet)
	err = b.Svc.ListResourceRecordSetsPages(input,
		func(page *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
			currentPageRecords := make(map[model.FQDNTypePair]*route53.ResourceRecordSet)
			pairsToQuery := make(map[model.FQDNTypePair]bool)
			for _, recordSet := range page.ResourceRecordSets {
				if aws.StringValue(recordSet.Type) != model.RecordTypeA &&
					aws.StringValue(recordSet.Type) != model.RecordTypeCname &&
					aws.StringValue(recordSet.Type) != model.RecordTypeTxt {
					continue
				}
				// key is name (fqdn) + type
				pair := model.FQDNTypePair{
					FQDN: strings.TrimSuffix(aws.StringValue(recordSet.Name), "."),
					Type: aws.StringValue(recordSet.Type),
				}
				currentPageRecords[pair] = recordSet
				pairsToQuery[pair] = true

			}

			// Young records should not be deleted. Remove them from the map for this page. Once they are removed, records that
			// are old or not in our DB at all will be left. These are the purge-worthy records. Add them to the recordsToDelete map
			youngRecordsByPair, err := b.db.GetYoungRecords(recordMaxAgeSeconds, pairsToQuery)
			if err != nil {
				logrus.Errorf("Could not load records from database. Error: %v", err)
				return false
			}
			for pair := range youngRecordsByPair {
				delete(currentPageRecords, pair)
			}
			maps.Copy(recordsToDelete, currentPageRecords)
			return true
		})
	if err != nil {
		logrus.Errorf("Error communicating with Route53: %v", err)
		return
	}

	if len(recordsToDelete) == 0 {
		logrus.Infof("Records purged from Route53: 0")
		return
	}

	var myChanges []*route53.Change
	for _, recordSet := range recordsToDelete {
		change := &route53.Change{
			Action:            aws.String("DELETE"),
			ResourceRecordSet: recordSet,
		}
		myChanges = append(myChanges, change)
	}

	changeInput := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(b.ZoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: myChanges,
		},
	}

	if _, err := b.Svc.ChangeResourceRecordSets(changeInput); err != nil {
		logrus.Errorf("Unable to delete recordSets from Route53. Error: %v", err)
	}

	logrus.Infof("Records purged from Route53: %v", len(recordsToDelete))
}

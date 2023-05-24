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

func (b *backend) StartPurgerDaemon(stopCh <-chan struct{}) {
	logrus.Infof("starting purge daemon. Purge interval: %v, max domain age: %v, record max age: %v",
		b.purgeIntervalSeconds, b.domainMaxAgeSeconds, b.recordMaxAgeSeconds)
	wait.JitterUntil(b.purge, time.Duration(b.purgeIntervalSeconds)*time.Second, .002, true, stopCh)
}

func (b *backend) purge() {
	logrus.Infof("Beginning purge ☠️")

	domainDeleted, recordsDeleted, err := b.db.PurgeOldDomainsAndRecords(b.domainMaxAgeSeconds, b.recordMaxAgeSeconds)
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
			youngRecordsByPair, err := b.db.GetYoungRecords(b.recordMaxAgeSeconds, pairsToQuery)
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

	// Elide records for "local" FQDNs, these won't exist in the database but shouldn't be deleted.
	localSuffix := ".local." + b.baseDomain
	for pair := range recordsToDelete {
		if strings.HasSuffix(pair.FQDN, localSuffix) {
			delete(recordsToDelete, pair)
		}
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

package lxAuditRepos

import (
	"github.com/globalsign/mgo"
	"github.com/litixsoft/lx-golib/audit"
	"github.com/litixsoft/lx-golib/db"
	"log"
	"time"
)

// auditMongo, mongo repository
type auditMongo struct {
	serviceName string
	serviceHost string
	db          *lxDb.MongoDb
}

// NewAuditMongo, return instance of auditMongo repository
func NewAuditMongo(db *lxDb.MongoDb, serviceName, serviceHost string) lxAudit.IAudit {
	return &auditMongo{db: db, serviceName: serviceName, serviceHost: serviceHost}
}

// SetupAudit, set the indexes for mongoDb
func (repo *auditMongo) SetupAudit() error {
	// Copy mongo session (thread safe) and close after function
	conn := repo.db.Conn.Copy()
	defer conn.Close()

	// Setup indexes
	return repo.db.Setup([]mgo.Index{
		{Key: []string{"timestamp"}},
	})
}

// Log, save log entry to mongoDb
func (repo *auditMongo) Log(user, message, data interface{}) chan bool {
	// channel for done
	done := make(chan bool, 1)

	go func() {
		// Copy mongo session (thread safe) and close after function
		conn := repo.db.Conn.Copy()
		defer conn.Close()

		// Log entry
		entry := &lxAudit.AuditModel{
			TimeStamp:   time.Now(),
			ServiceName: repo.serviceName,
			ServiceHost: repo.serviceHost,
			User:        user,
			Message:     message,
			Data:        data,
		}

		// Insert entry
		if err := conn.DB(repo.db.Name).C(repo.db.Collection).Insert(entry); err != nil {
			log.Printf("mongoDb can't insert audit entry, error: %v\n", err)
		}

		// inform when worker is done
		done <- true

	}()

	return done
}

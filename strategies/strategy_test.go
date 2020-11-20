package strategies

import (
	"github.com/pradykaushik/task-ranker/logger"
	"log"
	"testing"
)

func TestMain(m *testing.M) {
	err := logger.Configure()
	if err != nil {
		log.Println("could not configure logger for strategy testing")
	}
	m.Run()
	err = logger.Done()
	if err != nil {
		log.Println("could not close logger after strategy testing")
	}
}

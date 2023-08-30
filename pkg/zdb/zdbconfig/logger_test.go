package zdbconfig

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type LoggerSuite struct {
	suite.Suite
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}

func (s *LoggerSuite) TestGetDBLogger() {
	config := LogConfig{
		LogLevel: "info",
	}
	logger := getDBLogger(config)
	s.NotNil(logger)
}

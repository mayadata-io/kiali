package serviceentries

import (
	"testing"

	"github.com/mayadata-io/kiali/config"
	"github.com/mayadata-io/kiali/models"
	"github.com/mayadata-io/kiali/tests/data"
	"github.com/stretchr/testify/assert"
)

func TestValidPortDefinition(t *testing.T) {
	conf := config.NewConfig()
	config.Set(conf)

	assert := assert.New(t)

	se := data.AddPortDefinitionToServiceEntry(
		data.CreateEmptyPortDefinition(80, "http", "HTTP"),
		data.CreateEmptyMeshExternalServiceEntry("valid-se", "test", []string{"localhost"}),
	)

	pc := PortChecker{ServiceEntry: se}
	validations, valid := pc.Check()
	assert.True(valid)
	assert.Empty(validations)
}

func TestInvalidPortDefinition(t *testing.T) {
	conf := config.NewConfig()
	config.Set(conf)

	assert := assert.New(t)

	se := data.AddPortDefinitionToServiceEntry(
		data.CreateEmptyPortDefinition(80, "example-http", "HTTP"),
		data.CreateEmptyMeshExternalServiceEntry("notvalid-se", "test", []string{"localhost"}),
	)

	pc := PortChecker{ServiceEntry: se}
	validations, valid := pc.Check()
	assert.False(valid)
	assert.NotEmpty(validations)
	assert.Equal(models.ErrorSeverity, validations[0].Severity)
	assert.Equal(models.CheckMessage("port.name.mismatch"), validations[0].Message)
	assert.Equal("spec/ports[0]/name", validations[0].Path)
}

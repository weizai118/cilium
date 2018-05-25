package models

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type ModelsSuite struct{}

var _ = Suite(&ModelsSuite{})

var (
	drc1 = &DirectRoutingConfiguration{InstallRoutes: true}
	drc2 = &DirectRoutingConfiguration{}

	rc1 = RoutingConfiguration{Encapsulation: RoutingConfigurationEncapsulationVxlan, DirectRouting: drc1}
	rc2 = RoutingConfiguration{Encapsulation: RoutingConfigurationEncapsulationGeneve, DirectRouting: drc1}
	rc3 = RoutingConfiguration{DirectRouting: drc1}
)

func (s *ModelsSuite) TestDirectRoutingConfigurationEqual(c *C) {
	c.Assert(drc1.Equal(drc1), Equals, true)
	c.Assert(drc1.Equal(drc2), Equals, false)

	c.Assert(drc2.Equal(drc1), Equals, false)
	c.Assert(drc2.Equal(drc2), Equals, true)
}

func (s *ModelsSuite) TestRoutingConfigurationEqual(c *C) {
	c.Assert(rc1.Equal(&rc1), Equals, true)
	c.Assert(rc1.Equal(&rc2), Equals, false)

	c.Assert(rc2.Equal(&rc1), Equals, false)
	c.Assert(rc2.Equal(&rc2), Equals, true)

	c.Assert(rc3.Equal(&rc1), Equals, false)
	c.Assert(rc3.Equal(&rc2), Equals, false)
}

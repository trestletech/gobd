package main

import (
	"math"

	"gopkg.in/check.v1"
)

type PidsSuite struct{}

var _ = check.Suite(&PidsSuite{})

func (ps *PidsSuite) TestUnsupportedEngineLoad(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 03")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	_, err = obd.GetEngineLoad()
	c.Assert(err, check.NotNil)
}

func (ps *PidsSuite) TestEngineLoad(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 10")
	ms.addResponse("0104", "41 04 0A")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	l, err := obd.GetEngineLoad()
	c.Assert(err, check.IsNil)
	// Floats may not be identical. That's fine.
	c.Check(math.Abs(l-10/2.55) < .00001, check.Equals, true)

	ms.addResponse("0104", "41 04 FF")
	l, err = obd.GetEngineLoad()
	c.Assert(err, check.IsNil)
	// Floats may not be identical. That's fine.
	c.Check(math.Abs(l-255/2.55) < .00001, check.Equals, true)

}

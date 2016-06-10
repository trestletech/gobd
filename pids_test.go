package gobd

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

func (ps *PidsSuite) TestCoolantTemp(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 08")
	ms.addResponse("0105", "41 05 29")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	val, err := obd.GetCoolantTemp()
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 1)

	ms.addResponse("0105", "41 05 FF")
	val, err = obd.GetCoolantTemp()
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 215)

	ms.addResponse("0105", "41 05 00")
	val, err = obd.GetCoolantTemp()
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, -40)
}

func (ps *PidsSuite) TestRPM(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 00 10")
	ms.addResponse("010c", "41 0c 55 55")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	val, err := obd.GetRPM()
	c.Assert(err, check.IsNil)
	c.Check(math.Abs(val-5461.25) < 0.00001, check.Equals, true)

	ms.addResponse("010c", "41 0c 00 00")
	val, err = obd.GetRPM()
	c.Assert(err, check.IsNil)
	c.Check(math.Abs(val) < 0.00001, check.Equals, true)

	ms.addResponse("010c", "41 0c FF FF")
	val, err = obd.GetRPM()
	c.Assert(err, check.IsNil)
	c.Check(math.Abs(val-16383.75) < 0.00001, check.Equals, true)
}

func (ps *PidsSuite) TestSpeed(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 00 08")
	ms.addResponse("010d", "41 0d 55")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	val, err := obd.GetSpeed()
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 85)

	ms.addResponse("010d", "41 0d 00")
	val, err = obd.GetSpeed()
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 0)

	ms.addResponse("010d", "41 0d FF")
	val, err = obd.GetSpeed()
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 255)
}

func (ps *PidsSuite) TestThrottle(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 00 00 80")
	ms.addResponse("0111", "41 11 55")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	val, err := obd.GetThrottlePosition()
	c.Assert(err, check.IsNil)
	c.Check(math.Abs(val-33.333333) < 0.00001, check.Equals, true)

	ms.addResponse("0111", "41 11 00")
	val, err = obd.GetThrottlePosition()
	c.Assert(err, check.IsNil)
	c.Check(math.Abs(val) < 0.00001, check.Equals, true)

	ms.addResponse("0111", "41 11 FF")
	val, err = obd.GetThrottlePosition()
	c.Assert(err, check.IsNil)
	c.Check(math.Abs(val-100) < 0.00001, check.Equals, true)
}

func (ps *PidsSuite) TestFuelLevel(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 00 00 00 01")
	ms.addResponse("0120", "41 20 00 02 00 00")
	ms.addResponse("012f", "41 2f 55")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	val, err := obd.GetFuelLevel()
	c.Assert(err, check.IsNil)
	c.Check(math.Abs(val-33.333333) < 0.00001, check.Equals, true)

	ms.addResponse("012f", "41 2f 00")
	val, err = obd.GetFuelLevel()
	c.Assert(err, check.IsNil)
	c.Check(math.Abs(val) < 0.00001, check.Equals, true)

	ms.addResponse("012f", "41 2F FF")
	val, err = obd.GetFuelLevel()
	c.Assert(err, check.IsNil)
	c.Check(math.Abs(val-100) < 0.00001, check.Equals, true)
}

func (ps *PidsSuite) TestBarometricPressure(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 00 00 00 01")
	ms.addResponse("0120", "41 20 00 00 20 00")
	ms.addResponse("0133", "41 33 55")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	val, err := obd.GetBarometricPressure()
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 85)

	ms.addResponse("0133", "41 33 00")
	val, err = obd.GetBarometricPressure()
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 0)

	ms.addResponse("0133", "41 33 ff")
	val, err = obd.GetBarometricPressure()
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 255)
}

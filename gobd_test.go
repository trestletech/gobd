package gobd

import (
	"testing"

	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type GobdSuite struct{}

var _ = check.Suite(&GobdSuite{})

type mockSerial struct {
	messages     [][]byte
	responses    map[string]string
	nextResponse string
}

func newMockSerial() *mockSerial {
	ms := &mockSerial{
		messages:  make([][]byte, 0),
		responses: make(map[string]string, 0),
	}

	ms.addResponse("ATZ", " ELM327/ELM-USB v1.0 (c) SECONS Ltd.\r\n > ")
	ms.addResponse("ATE0", " ATE0\r\n > OK \r\n > ")

	return ms
}

func (ms *mockSerial) Read(b []byte) (int, error) {
	if ms.nextResponse != "" {
		res := ms.nextResponse
		ms.nextResponse = ""

		// Copy the canned response into the result slice
		for i := 0; i < len(res); i++ {
			b[i] = res[i]
		}

		return len(res), nil
	}

	// Return a newline so that the reader can advance.
	b[0] = byte('\r')
	b[1] = byte('\n')
	// Nothing in the queue
	return 2, nil
}

func (ms *mockSerial) Write(b []byte) (int, error) {
	nr, ok := ms.responses[string(b)]
	if ok {
		ms.nextResponse = nr
	}
	ms.messages = append(ms.messages, b)
	return len(b), nil
}

func (ms *mockSerial) addResponse(in string, response string) {
	ms.responses[in+"\r\n"] = response
}

func (s *GobdSuite) TestInclude(c *check.C) {
	c.Check(includes([]uint{1, 2, 3}, 1), check.Equals, true)
	c.Check(includes([]uint{1, 2, 3}, 2), check.Equals, true)
	c.Check(includes([]uint{1, 2, 3}, 3), check.Equals, true)
	c.Check(includes([]uint{1, 2, 3}, 4), check.Equals, false)
}

func (s *GobdSuite) TestNewOBD(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 03")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)
	c.Check(ms.messages, check.HasLen, 3)
	// Check that the version got set
	c.Check(obd.id, check.Equals, "ELM327/ELM-USB v1.0 (c) SECONS Ltd.")

	// Check that the echo got disabled
	c.Check(string(ms.messages[1]), check.Equals, "ATE0\r\n")

	// Check that we polled for PIDs once.
	c.Check(string(ms.messages[2]), check.Equals, "0100\r\n")

	// Check that the PIDs got set correctly
	c.Check(obd.pids, check.DeepEquals, []uint{0, 7, 8})
}

//TODO
/*
func (s *GobdSuite) TestReadMulti(c *check.C) {
	ms := newMockSerial()
	ms.AddResponseNoNewline("0101", "41 ")
	ms.AddResponseNoNewline("0101", "01")
	ms.AddResponseNoNewline("0101", " 00")
	ms.AddResponseNoNewline("0101", " \r")
	ms.AddResponseNoNewline("0101", "\n ")
}
*/

func (s *GobdSuite) TestParseOutput(c *check.C) {
	c.Check(parseOutput([]byte("\n > \r hi > there! \r\n >")), check.DeepEquals, []byte("hi > there!"))
	c.Check(parseOutput([]byte("hi > there!")), check.DeepEquals, []byte("hi > there!"))
}

func (s *GobdSuite) TestUnsupportedPID(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 03")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	_, err = obd.current(5)
	c.Check(err, check.NotNil)
}

func (s *GobdSuite) TestCurrent(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 03")
	ms.addResponse("0107", "41 07 55")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	byt, err := obd.current(7)
	c.Assert(err, check.IsNil)
	c.Check(byt, check.DeepEquals, []byte{'5', '5'})
}

func (s *GobdSuite) TestCurrentInt(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 07")
	ms.addResponse("0106", "41 06 00 00")
	ms.addResponse("0107", "41 07 55 82 a0")
	ms.addResponse("0108", "41 08 F")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)

	val, err := obd.currentInt(6)
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 0)

	val, err = obd.currentInt(7)
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 5604000)

	val, err = obd.currentInt(8)
	c.Assert(err, check.IsNil)
	c.Check(val, check.Equals, 15)
}

func (s *GobdSuite) TestListPidsFull(c *check.C) {
	ms := newMockSerial()
	ms.addResponse("0100", "41 00 01 00 00 01")
	ms.addResponse("0120", "41 20 01 00 00 01")
	ms.addResponse("0140", "41 40 01 00 00 01")
	ms.addResponse("0160", "41 60 01 00 00 01")
	ms.addResponse("0180", "41 80 01 00 00 01")
	ms.addResponse("01a0", "41 a0 01 00 00 01")
	ms.addResponse("01c0", "41 c0 01 00 00 01")

	obd, err := NewDebugOBD(ms, c.Logf)
	c.Assert(err, check.IsNil)
	c.Check(obd.pids, check.DeepEquals, []uint{0, 8, 32, 40, 64, 72, 96, 104, 128, 136, 160, 168, 192, 200, 224})
}

func (s *GobdSuite) ParseMode1(c *check.C) {
	_, err := parseMode1Response([]byte(""), "0101")
	c.Check(err, check.NotNil)

	_, err = parseMode1Response([]byte("UNAVAILABLE"), "0101")
	c.Check(err, check.NotNil)

	_, err = parseMode1Response([]byte("00 00 00"), "0100")
	c.Check(err, check.NotNil)

	_, err = parseMode1Response([]byte("41 00 00 11 22"), "01ff")
	c.Check(err, check.IsNil)

	val, err := parseMode1Response([]byte("41 00 00 11 22"), "0100")
	c.Check(err, check.IsNil)
	c.Check(val, check.DeepEquals, []byte{'0', '0', '1', '1', '2', '2'})
}

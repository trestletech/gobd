package gobd

// [1 3 4 5 6 7 11 12 13 14 15 16 17 19 21 28 31 32 33 44 46 47 48 49 50 51 52 60 64 65 66 67 68 69 70 71 73 74 76
//var monitorStatusPID int = 1
//var fuelStatusPID int = 3
var engineLoadPID int = 4
var coolantTempPID int = 5

/*var shortTermFuelTrim1PID int = 6
var longTermFuelTrim1PID int = 7
var shortTermFuelTrim2PID int = 8
var longTermFuelTrim2PID int = 9
var intakeManifoldAbsPressurePID int = 11
*/
var engineRPMPID int = 12
var speedPID int = 13

//var timingAdvancePID int = 14
//var intakeAirTempPID int = 15

var throttlePositionPID int = 17

var fuelLevelPID int = 47
var barometricPressurePID int = 51

/*
type monitorStatus struct{}

func (o *OBD) GetMonitorStatus() (monitorStatus, error) {
	//TODO
	return monitorStatus{}, nil
}

type fuelStatus struct{}

func (o *OBD) GetFuelStatus() (fuelStatus, error) {
	//TODO
	return fuelStatus{}, nil
}
*/

func (o *OBD) GetEngineLoad() (float64, error) {
	if !includes(o.pids, engineLoadPID) {
		return float64(0), pidNotSupportedError(engineLoadPID)
	}
	val, err := o.currentInt(engineLoadPID)
	return float64(val) / 2.55, err
}

func (o *OBD) GetCoolantTemp() (int, error) {
	if !includes(o.pids, coolantTempPID) {
		return 0, pidNotSupportedError(coolantTempPID)
	}
	val, err := o.currentInt(coolantTempPID)
	return val - 40, err
}

func (o *OBD) GetRPM() (float64, error) {
	if !includes(o.pids, engineRPMPID) {
		return float64(0), pidNotSupportedError(engineRPMPID)
	}
	val, err := o.currentInt(engineRPMPID)
	return float64(val) / 4, err
}

func (o *OBD) GetSpeed() (int, error) {
	if !includes(o.pids, speedPID) {
		return 0, pidNotSupportedError(speedPID)
	}
	val, err := o.currentInt(speedPID)
	return val, err
}

func (o *OBD) GetThrottlePosition() (float64, error) {
	if !includes(o.pids, throttlePositionPID) {
		return float64(0), pidNotSupportedError(throttlePositionPID)
	}
	val, err := o.currentInt(throttlePositionPID)
	return float64(val) / 2.55, err
}

func (o *OBD) GetFuelLevel() (float64, error) {
	if !includes(o.pids, fuelLevelPID) {
		return float64(0), pidNotSupportedError(fuelLevelPID)
	}
	val, err := o.currentInt(fuelLevelPID)
	return float64(val) / 2.55, err
}

func (o *OBD) GetBarometricPressure() (int, error) {
	if !includes(o.pids, barometricPressurePID) {
		return 0, pidNotSupportedError(barometricPressurePID)
	}
	val, err := o.currentInt(barometricPressurePID)
	return val, err
}

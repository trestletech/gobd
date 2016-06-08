package main

var engineLoadPID int = 4

func (o *OBD) GetEngineLoad() (float64, error) {
	if !includes(o.pids, engineLoadPID) {
		return float64(0), pidNotSupportedError(engineLoadPID)
	}
	val, err := o.currentInt(engineLoadPID)
	return float64(val) / 2.55, err
}

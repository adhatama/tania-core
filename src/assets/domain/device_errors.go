package domain

// DeviceError is a custom error from Go built-in error
type DeviceError struct {
	Code int
}

const (
	DeviceErrorDeviceNotFoundCode = iota
	DeviceErrorDeviceIDAlreadyExistsCode
	DeviceErrorNameEmptyCode
)

func (e DeviceError) Error() string {
	switch e.Code {
	case DeviceErrorDeviceNotFoundCode:
		return "Device not found."
	case DeviceErrorDeviceIDAlreadyExistsCode:
		return "Device ID is already exists"
	case DeviceErrorNameEmptyCode:
		return "Device name cannot be empty"
	default:
		return "Unrecognized device error code"
	}
}

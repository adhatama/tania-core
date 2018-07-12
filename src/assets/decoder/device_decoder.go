package decoder

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Tanibox/tania-core/src/assets/domain"
	"github.com/mitchellh/mapstructure"
)

type DeviceEventWrapper EventWrapper

func (w *DeviceEventWrapper) UnmarshalJSON(b []byte) error {
	wrapper := EventWrapper{}

	err := json.Unmarshal(b, &wrapper)
	if err != nil {
		return err
	}

	mapped, ok := wrapper.EventData.(map[string]interface{})
	if !ok {
		return errors.New("Error type assertion")
	}

	f := mapstructure.ComposeDecodeHookFunc(
		UIDHook(),
		TimeHook(time.RFC3339),
	)

	switch wrapper.EventName {
	case "DeviceCreated":
		e := domain.DeviceCreated{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	case "DeviceIDChanged":
		e := domain.DeviceIDChanged{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	case "DeviceNameChanged":
		e := domain.DeviceNameChanged{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	case "DeviceDescriptionChanged":
		e := domain.DeviceDescriptionChanged{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	case "DeviceStatusChanged":
		e := domain.DeviceStatusChanged{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	case "DeviceRemoved":
		e := domain.DeviceRemoved{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	}

	return nil
}

package decoder

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Tanibox/tania-core/src/user/domain"
	"github.com/mitchellh/mapstructure"
)

type OrganizationEventWrapper EventWrapper

func (w *OrganizationEventWrapper) UnmarshalJSON(b []byte) error {
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
	case "OrganizationCreated":
		e := domain.OrganizationCreated{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	}

	return nil
}
